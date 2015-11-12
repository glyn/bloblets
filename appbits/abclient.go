package appbits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"runtime"
	"strings"
	"time"

	basenet "net"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/cliutil"
)

const (
	DefaultAppUploadBitsTimeout = 15 * time.Minute
)

func UploadApp(appGuid string, zipFile *os.File, presentFiles []models.AppFileFields) {
	afrs := make([]resources.AppFileResource, 0, len(presentFiles))
	for _, pf := range presentFiles {
		afrs = append(afrs, resources.AppFileResource{
			Sha1: pf.Sha1,
			Size: pf.Size,
			Path: pf.Path,
			Mode: pf.Mode,
		})
	}
	uploadBits(appGuid, zipFile, afrs)
}

func uploadBits(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) {
	apiUrl := fmt.Sprintf("http://localhost:8080/v2/apps/%s/bits", appGuid)
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		cliutil.Check(err)

		// json.Marshal represents a nil value as "null" instead of an empty slice "[]"
		if presentFiles == nil {
			presentFiles = []resources.AppFileResource{}
		}

		presentFilesJSON, err := json.Marshal(presentFiles)
		cliutil.Check(err)

		boundary, err := writeUploadBody(zipFile, requestFile, presentFilesJSON)
		cliutil.Check(err)

		var request *Request
		request, err = newRequestForFile("PUT", apiUrl, requestFile)
		cliutil.Check(err)

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)

		response := &resources.Resource{}
		_, err = performPollingRequestForJSONResponse("localhost:8080", request, response, DefaultAppUploadBitsTimeout)
		cliutil.Check(err)
	})
}

func writeUploadBody(zipFile *os.File, body *os.File, presentResourcesJson []byte) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBuffer(presentResourcesJson))
	if err != nil {
		return
	}

	if zipFile != nil {
		zipStats, zipErr := zipFile.Stat()
		if zipErr != nil {
			return
		}

		if zipStats.Size() == 0 {
			return
		}

		part, zipErr = createZipPartWriter(zipStats, writer)
		if zipErr != nil {
			return
		}

		_, zipErr = io.Copy(part, zipFile)
		if zipErr != nil {
			return
		}
	}

	return
}

func createZipPartWriter(zipStats os.FileInfo, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipStats.Size()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}

type Request struct {
	HttpReq      *http.Request
	SeekableBody io.ReadSeeker
}

func newRequestForFile(method, fullUrl string, body *os.File) (req *Request, apiErr error) {
	ui := terminal.NewUI(os.Stdin, &cliutil.NonPrinter{})
	progressReader := net.NewProgressReader(body, ui, 5*time.Second)
	progressReader.Seek(0, 0)
	fileStats, err := body.Stat()

	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error getting file info", err.Error())
		return
	}

	request, err := http.NewRequest(method, fullUrl, progressReader)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error building request", err.Error())
		return
	}

	fileSize := fileStats.Size()
	progressReader.SetTotalSize(fileSize)
	request.ContentLength = fileSize

	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error building request", err.Error())
		return
	}

	return newRequest(request, progressReader)
}

func newRequest(request *http.Request, body io.ReadSeeker) (*Request, error) {
	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "bloblets client v0.1 / "+runtime.GOOS)
	return &Request{HttpReq: request, SeekableBody: body}, nil
}

type AsyncResource struct {
	Metadata struct {
		URL string
	}
}

func performPollingRequestForJSONResponse(endpoint string, request *Request, response interface{}, timeout time.Duration) (headers http.Header, apiErr error) {
	log.Println("Polling entered")
	defer log.Println("Polling exiting")
	query := request.HttpReq.URL.Query()
	query.Add("async", "true")
	request.HttpReq.URL.RawQuery = query.Encode()

	bytes, headers, rawResponse, apiErr := performRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}
	defer rawResponse.Body.Close()

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Invalid JSON response from server", err.Error())
		return
	}

	asyncResource := &AsyncResource{}
	err = json.Unmarshal(bytes, &asyncResource)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Invalid async response from server", err.Error())
		return
	}

	jobUrl := asyncResource.Metadata.URL
	if jobUrl == "" {
		return
	}

	log.Printf("Polling waiting for job %s\n", jobUrl)

	if !strings.Contains(jobUrl, "/jobs/") {
		return
	}

	apiErr = waitForJob(endpoint+jobUrl, request.HttpReq.Header.Get("Authorization"), timeout)
	return
}

func performRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, rawResponse *http.Response, apiErr error) {
	rawResponse, apiErr = doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}
	defer rawResponse.Body.Close()

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error reading response", err.Error())
	}

	headers = rawResponse.Header
	return
}

func doRequestHandlingAuth(request *Request) (rawResponse *http.Response, err error) {
	httpReq := request.HttpReq

	if request.SeekableBody != nil {
		httpReq.Body = ioutil.NopCloser(request.SeekableBody)
	}

	// perform request
	rawResponse, err = gwDoRequestAndHandlerError(request)

	return
}

func gwDoRequestAndHandlerError(request *Request) (rawResponse *http.Response, err error) {
	rawResponse, err = gwDoRequest(request.HttpReq)
	cliutil.Check(err)
	if rawResponse.StatusCode > 299 {
		jsonBytes, _ := ioutil.ReadAll(rawResponse.Body)
		rawResponse.Body.Close()
		rawResponse.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBytes))
		cliutil.Check(fmt.Errorf("gwDoRequestAndHandlerError: %d %s", rawResponse.StatusCode, string(jsonBytes)))
	}

	return
}

func gwDoRequest(request *http.Request) (response *http.Response, err error) {
	transport := makeHttpTransport()

	httpClient := net.NewHttpClient(transport)

	for i := 0; i < 3; i++ {
		response, err = httpClient.Do(request) // prints some spaces
		fmt.Println()                          // align subsequent log entry
		if response == nil && err != nil {
			continue
		} else {
			break
		}
	}

	return
}

func makeHttpTransport() *http.Transport {
	return &http.Transport{
		Dial:  (&basenet.Dialer{Timeout: 5 * time.Second}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}
}

const (
	JOB_FINISHED             = "finished"
	JOB_FAILED               = "failed"
	DEFAULT_POLLING_THROTTLE = 5 * time.Second
)

type JobResource struct {
	Entity struct {
		Status       string
		ErrorDetails struct {
			Description string
		} `json:"error_details"`
	}
}

func waitForJob(jobUrl, accessToken string, timeout time.Duration) (err error) {
	startTime := time.Now()
	for true {
		if time.Now().Sub(startTime) > timeout && timeout != 0 {
			return errors.NewAsyncTimeoutError(jobUrl)
		}
		var request *Request
		request, err = NewRequest("GET", jobUrl, accessToken, nil)
		response := &JobResource{}
		_, err = gwPerformRequestForJSONResponse(request, response)
		if err != nil {
			return
		}

		switch response.Entity.Status {
		case JOB_FINISHED:
			return
		case JOB_FAILED:
			err = errors.New(response.Entity.ErrorDetails.Description)
			return
		}

		accessToken = request.HttpReq.Header.Get("Authorization")

		time.Sleep(DEFAULT_POLLING_THROTTLE)
	}
	return
}

func NewRequest(method, path, accessToken string, body io.ReadSeeker) (req *Request, apiErr error) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error building request", err.Error())
		return
	}
	return gwNewRequest(request, accessToken, body)
}

func gwNewRequest(request *http.Request, accessToken string, body io.ReadSeeker) (*Request, error) {
	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "bloblets client v0.1 / "+runtime.GOOS)
	return &Request{HttpReq: request, SeekableBody: body}, nil
}

func gwPerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiErr error) {
	bytes, headers, rawResponse, apiErr := gwPerformRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Invalid JSON response from server", err.Error())
	}
	return
}

func gwPerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, rawResponse *http.Response, apiErr error) {
	rawResponse, apiErr = doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}
	defer rawResponse.Body.Close()

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", "Error reading response", err.Error())
	}

	headers = rawResponse.Header
	return
}
