package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"log"

	"strings"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/appbits"
	"github.com/glyn/bloblets/bloblet"
	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
	"github.com/glyn/bloblets/scanner"
)

func main() {
	if len(os.Args) != 3 || len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Println("Usage: client host:port application-path")
		os.Exit(1)
	}

	log.Println("Push started")
	client := &http.Client{}
	server := os.Args[1]

	scanner.Scan(os.Args[2], func(appDir string, condenser bloblet.Condensate, appFiles []models.AppFileFields) {
		log.Println("Resource matching started")
		request := resmatch.ResMatchRequest(server, integrityFields(appFiles))

		log.Println("Resource matching sending request")
		response := cliutil.Converse(client, request)
		log.Println("Resource matching received response")

		presentIffs := resmatch.ProcessResponse(response)
		log.Println("Resource matching response demarshalled")

		presentFiles := intersectAppFilesIntegrityFields(appFiles, presentIffs)
		log.Println("Resource matching complete")

		appFilesToUpload := make([]models.AppFileFields, len(appFiles))
		copy(appFilesToUpload, appFiles)
		for _, file := range presentFiles {
			appFilesToUpload = deleteAppFile(appFilesToUpload, file.Path)
		}

		// Populate file modes
		var err error
		presentFiles, err = PopulateFileMode(appDir, presentFiles)
		cliutil.Check(err)

		hasFileToUpload := len(appFilesToUpload) > 0

		fileutils.TempDir("upload-dir", func(uploadDir string, err error) {
			cliutil.Check(err)

			var blobletsToUpload []bloblet.Bloblet
			blobletsToUpload, appFilesToUpload = condenser.Bloblets(appFilesToUpload)

			cliutil.Check(app_files.ApplicationFiles{}.CopyFiles(appFilesToUpload, appDir, uploadDir))
			// FIXME: upload these as HTTP parts
			_ = blobletsToUpload

			fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
				if hasFileToUpload {
					log.Println("Zipping application files")
					cliutil.Check(zipAppFiles(zipFile, appDir, uploadDir))
					log.Println("Zipped application files")
				}

				log.Println("Uploading application")
				appbits.UploadApp(server, "test-guid", zipFile, presentFiles)
				log.Println("Uploaded application")
				log.Println("Deleting uploads file")
			})
			log.Println("Deleting upload directory")
		})
	})
	log.Println("Push completed")
}

func deleteAppFile(appFiles []models.AppFileFields, path string) []models.AppFileFields {
	for i, file := range appFiles {
		if file.Path == path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func PopulateFileMode(appDir string, presentFiles []models.AppFileFields) ([]models.AppFileFields, error) {
	for i, _ := range presentFiles {
		path := presentFiles[i].Path
		if strings.HasSuffix(path, bloblet.BlobletFileNamePathTerminator) {
			continue // filemode makes no sense for a bloblet
		}
		fileInfo, err := os.Lstat(filepath.Join(appDir, path))
		if err != nil {
			return presentFiles, err
		}
		presentFiles[i].Mode = fmt.Sprintf("%#o", fileInfo.Mode())
	}

	return presentFiles, nil
}

func zipAppFiles(zipFile *os.File, appDir string, uploadDir string) (zipErr error) {
	zipErr = zipWithBetterErrors(uploadDir, zipFile)
	if zipErr != nil {
		return
	}

	_, zipErr = app_files.ApplicationZipper{}.GetZipSize(zipFile)

	return
}

func zipWithBetterErrors(uploadDir string, zipFile *os.File) error {
	zipper := app_files.ApplicationZipper{}
	zipError := zipper.Zip(uploadDir, zipFile)
	switch err := zipError.(type) {
	case nil:
		return nil
	case *errors.EmptyDirError:
		zipFile = nil
		return zipError
	default:
		return fmt.Errorf("%s: %s", "Error zipping application", err.Error())
	}
}

func integrityFields(affs []models.AppFileFields) []resmatch.IntegrityFields {
	ifs := []resmatch.IntegrityFields{} // FIXME: make([]resmatch.IntegrityFields, 0, len(affs))

	for _, aff := range affs {
		ifs = append(ifs, resmatch.IntegrityFields{
			Sha1: aff.Sha1,
			Size: aff.Size,
		})
	}

	return ifs
}

func intersectAppFilesIntegrityFields(appFiles []models.AppFileFields, integrityFields []resmatch.IntegrityFields) (out []models.AppFileFields) {
	inputFiles := appFilesBySha(appFiles)
	for _, responseFields := range integrityFields {
		item, found := inputFiles[responseFields.Sha1]
		if found {
			out = append(out, item)
		}
	}
	return out
}

func appFilesBySha(in []models.AppFileFields) (out map[string]models.AppFileFields) {
	out = map[string]models.AppFileFields{} // FIXME
	for _, inputFileResource := range in {
		out[inputFileResource.Sha1] = inputFileResource
	}
	return out
}
