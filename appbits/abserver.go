package appbits

import (
	"archive/zip"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"io"
	"io/ioutil"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/glyn/bloblets/blobstore"
	"github.com/glyn/bloblets/servutil"

	"os"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AppHandler entered")
	appDir, err := ioutil.TempDir("", "recomposed-application")
	if err != nil {
		servutil.Fail(w, "creating recomposed application directory failed: %s", err)
		return
	}
	defer os.RemoveAll(appDir)

	r.ParseMultipartForm(100 * 1024 * 1024)
	mpForm := r.MultipartForm

	// Unzip any uploaded portion of the application.
	appHdrs, ok := mpForm.File["application"]
	if ok {
		application := appHdrs[0]
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
			destPath := filepath.Join(appDir, f.Name)
			destDir, _ := filepath.Split(destPath)
			err := os.MkdirAll(destDir[:len(destDir)-1], 0755)
			if err != nil {
				servutil.Fail(w, "Failed to create destination directory: %s", err)
				return
			}

			if f.FileInfo().IsDir() {
				continue
			}

			dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				servutil.Fail(w, "Failed to create destination file: %s", err)
				return
			}

			of, err := f.Open()
			if err != nil {
				servutil.Fail(w, "Failed to open application file: %s", err)
				return
			}

			io.Copy(dest, of)
			of.Close()
			_, err = dest.Stat()
			if err != nil {
				servutil.Fail(w, "Failed to stat application file: %s", err)
				return
			}
			dest.Close()
		}
		log.Println("AppHandler unzipped uploaded portion of application")
	}

	// Add blobs to the application
	res := mpForm.Value["resources"]
	presentFiles := []resources.AppFileResource{}
	err = json.Unmarshal([]byte(res[0]), &presentFiles)
	if err != nil {
		servutil.Fail(w, "demarshalling resources failed: %s", err)
		return
	}
	for _, pf := range presentFiles {
		dest := filepath.Join(appDir, pf.Path)
		destDir, _ := filepath.Split(dest)
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			servutil.Fail(w, "creating destination directory for downloaded file failed: %s", err)
			return
		}
		blobstore.Get(pf.Sha1, dest)
	}

	log.Println("AppHandler exiting")
}
