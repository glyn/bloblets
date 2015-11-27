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
	"strings"
	"time"
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
		log.Println("AppHandler unzipping uploaded portion of application")
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

		var blobUploadTime time.Duration = 0
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
			fi, err := dest.Stat()
			if err != nil {
				servutil.Fail(w, "Failed to stat application file: %s", err)
				return
			}
			dest.Close()
			fn := dest.Name()
			if !fi.IsDir() && fi.Size() > 65535 {
				uploadStart := time.Now()
				sha := sha(fn)
				//			log.Printf("AppHandler adding file %s to the blob store: %s\n", f.Name, sha)
				blobstore.Add(sha, fn)
				blobUploadTime += time.Now().Sub(uploadStart)
			}
		}
		log.Printf("AppHandler unzipped uploaded portion of application and spent %d seconds uploading blobs to the blobstore\n", blobUploadTime/time.Second)
	}

	// Process any uploaded bloblets.
	log.Println("AppHandler unzipping uploaded bloblets")
	var blobletUploadTime time.Duration = 0
	for i := 0; true; i++ {
		hdrs, ok := mpForm.File[fmt.Sprintf("bloblet-%d", i)]
		if !ok {
			break
		}
		blobletHeader := hdrs[0]
		cd := blobletHeader.Header.Get("Content-Disposition")
		var (
			j               int
			path, bfn, hash string
		)

		k := strings.Replace(cd, `form-data; name="bloblet-`, "", 1)
		k = strings.Replace(k, `"; path="`, " ", 1)
		k = strings.Replace(k, `"; filename="`, " ", 1)
		k = strings.Replace(k, `"; hash="`, " ", 1)
		k = strings.Replace(k, `"`, "", 1)
		_, err := fmt.Sscanf(k, `%d %s %s %s`, &j, &path, &bfn, &hash)
		if err != nil {
			servutil.Fail(w, "Malformed Content-Disposition %s: %s", cd, err)
			return
		}
		if j != i || bfn != bloblet.BlobletFileName {
			servutil.Fail(w, "Invalid Content-Disposition %d %s: %s", i, cd, err)
			return
		}

		// Read bloblet zip file.
		blobletZipFile, err := blobletHeader.Open()
		if err != nil {
			servutil.Fail(w, "failed to open bloblet zip: %s", err)
			return
		}
		tmpFile, err := ioutil.TempFile("", "bloblet-zip")
		if err != nil {
			servutil.Fail(w, "failed to create bloblet temporary zip file: %s", err)
			return
		}
		_, err = io.Copy(tmpFile, blobletZipFile)
		if err != nil {
			servutil.Fail(w, "failed to copy bloblet zip file: %s", err)
			return
		}
		tmpFilename := tmpFile.Name()
		err = tmpFile.Close()
		if err != nil {
			servutil.Fail(w, "failed to close bloblet temporary zip file: %s", err)
			return
		}
		err = blobletZipFile.Close()
		if err != nil {
			servutil.Fail(w, "failed to close bloblet zip file: %s", err)
			return
		}

		// Store bloblet zip file in blob store.
		uploadStart := time.Now()
		blobstore.Add(hash, tmpFilename)
		blobletUploadTime += time.Now().Sub(uploadStart)

		// Unzip bloblet into the application.
		dest := filepath.Join(appDir, path)
		destDir, _ := filepath.Split(dest)
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			servutil.Fail(w, "creating destination directory for bloblet failed: %s", err)
			return
		}
		err = unzip(tmpFilename, destDir, w)
		if err != nil {
			return
		}

		err = os.RemoveAll(tmpFilename)
		if err != nil {
			servutil.Fail(w, "failed to delete bloblet temporary zip file: %s", err)
			return
		}

	}
	log.Printf("AppHandler unzipped uploaded bloblets and spent %d seconds uploading them to the blobstore\n", blobletUploadTime/time.Second)

	// Add blobs to the application
	log.Println("AppHandler demarshalling resources")
	res := mpForm.Value["resources"]
	presentFiles := []resources.AppFileResource{}
	err = json.Unmarshal([]byte(res[0]), &presentFiles)
	if err != nil {
		servutil.Fail(w, "demarshalling resources failed: %s", err)
		return
	}
	log.Println("AppHandler downloading & unzipping resources from the blobstore")
	var blobletDownloadTime time.Duration = 0
	for _, pf := range presentFiles {
		dest := filepath.Join(appDir, pf.Path)
		destDir, destFile := filepath.Split(dest)
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			servutil.Fail(w, "creating destination directory for downloaded file failed: %s", err)
			return
		}
		if destFile == bloblet.BlobletFileName {
			fileutils.TempDir("bloblet-download-dir", func(downloadDir string, err error) {
				if err != nil {
					servutil.Fail(w, "bloblet download directory error: %s", err)
					return
				}
				z := filepath.Join(downloadDir, "bloblet.zip")
				downloadStart := time.Now()
				blobstore.Get(pf.Sha1, z)
				blobletDownloadTime += time.Now().Sub(downloadStart)

				err = unzip(z, destDir, w)
				if err != nil {
					return
				}

			})
		} else {
			downloadStart := time.Now()
			blobstore.Get(pf.Sha1, dest)
			blobletDownloadTime += time.Now().Sub(downloadStart)
		}
	}
	log.Printf("AppHandler downloaded & unzipped resources from the blobstore and spent %d seconds downloading them from the blobstore\n", blobletDownloadTime/time.Second)

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
