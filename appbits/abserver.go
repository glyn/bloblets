package appbits

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"io"
	"io/ioutil"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"github.com/glyn/bloblets/bloblet"
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

	// Unzip the uploaded portion of the application.
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
		fi, fn, err := unzipFile(f, appDir, w)
		if err != nil {
			return
		}
		if !fi.IsDir() && fi.Size() > 65535 {
			sha := sha(fn)
			log.Printf("AppHandler adding file %s to the blob store: %s\n", f.Name, sha)
			blobstore.Add(sha, fn)
		}
	}
	log.Println("AppHandler unzipped uploaded portion of application")

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
		destDir, destFile := filepath.Split(dest)
		if destFile == bloblet.BlobletFileName {
			fileutils.TempDir("bloblet-download-dir", func(downloadDir string, err error) {
				if err != nil {
					servutil.Fail(w, "bloblet download directory error: %s", err)
					return
				}
				z := filepath.Join(downloadDir, "bloblet.zip")
				blobstore.Get(pf.Sha1, z)

				err = unzip(z, destDir, w)
				if err != nil {
					return
				}

			})
		} else {
			blobstore.Get(pf.Sha1, dest)
		}
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

func unzip(zipFile, destDir string, w http.ResponseWriter) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		servutil.Fail(w, "failed to unzip: %s", err)
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		_, _, err := unzipFile(f, destDir, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func unzipFile(zipFile *zip.File, unzipPath string, w http.ResponseWriter) (os.FileInfo, string, error) {
	destPath := filepath.Join(unzipPath, zipFile.Name)
	destDir, _ := filepath.Split(destPath)
	err := os.MkdirAll(destDir[:len(destDir)-1], 0755)
	if err != nil {
		servutil.Fail(w, "Failed to create destination directory: %s", err)
		return nil, "", err
	}

	if zipFile.FileInfo().IsDir() {
		return zipFile.FileInfo(), "", nil
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		servutil.Fail(w, "Failed to create destination file: %s", err)
		return nil, "", err
	}

	of, err := zipFile.Open()
	if err != nil {
		servutil.Fail(w, "Failed to open application file: %s", err)
		return nil, "", err
	}

	io.Copy(dest, of)
	of.Close()
	fi, err := dest.Stat()
	if err != nil {
		servutil.Fail(w, "Failed to stat application file: %s", err)
		return nil, "", err
	}
	dest.Close()
	fn := dest.Name()
	return fi, fn, nil
}
