package appbits

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"io"
	"io/ioutil"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"github.com/glyn/bloblets/blobstore"
	"github.com/glyn/bloblets/servutil"

	"os"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AppHandler entered")
	r.ParseMultipartForm(100 * 1024 * 1024)
	mpForm := r.MultipartForm

	res := mpForm.Value["resources"]
	presentFiles := []resources.AppFileResource{}
	err := json.Unmarshal([]byte(res[0]), &presentFiles)
	if err != nil {
		servutil.Fail(w, "demarshalling resources failed: %s", err)
		return
	}
	log.Println("AppHandler demarshalled resources")

	application := mpForm.File["application"][0]
	zipFile, err := application.Open()
	if err != nil {
		servutil.Fail(w, "failed to open application zip: %s", err)
		return
	}
	cl := application.Header.Get("Content-Length")
	if cl == "" {
		servutil.Fail(w, "Content-Length header not supplied for application zip: %s", err)
		return
	}
	zipLen, err := strconv.ParseInt(cl, 10, 64)
	if err != nil {
		servutil.Fail(w, "Invalid Content-Length %s: %s", cl, err)
		return
	}

	zr, err := zip.NewReader(zipFile, zipLen)

	for _, f := range zr.File {
		t, err := ioutil.TempFile("", "tempFile")
		if err != nil {
			servutil.Fail(w, "Failed to create temporary file: %s", err)
			return
		}

		of, err := f.Open()
		if err != nil {
			servutil.Fail(w, "Failed to open application file: %s", err)
			return
		}

		io.Copy(t, of)
		of.Close()
		fi, err := t.Stat()
		if err != nil {
			servutil.Fail(w, "Failed to stat application file: %s", err)
			return
		}
		t.Close()
		fn := t.Name()
		if !fi.IsDir() && fi.Size() > 65535 {
			sha := sha(fn)
			log.Printf("AppHandler adding file %s to the blob store: %s\n", f.Name, sha)
			blobstore.Add(sha, fn)
		}
		os.RemoveAll(fn)
	}
	log.Println("AppHandler exiting")
}

func sha(path string) string {
	hash := sha1.New()
	err := fileutils.CopyPathToWriter(path, hash)
	if err != nil {
		panic("Failed to compute sha1")
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
