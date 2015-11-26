package main

import (
	"fmt"
	"os"

	"log"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/appbits"
	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/scanner"
)

func main() {
	if len(os.Args) != 3 || len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Println("Usage: client host:port application-path")
		os.Exit(1)
	}

	log.Println("Push started")
	server := os.Args[1]

	scanner.Scan(os.Args[2], func(appDir string, appFiles []models.AppFileFields) {
		presentFiles := []models.AppFileFields{}

		appFilesToUpload := make([]models.AppFileFields, len(appFiles))
		copy(appFilesToUpload, appFiles)

		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			log.Println("Zipping application files")
			cliutil.Check(zipAppFiles(zipFile, appDir, appDir))
			log.Println("Zipped application files")

			log.Println("Uploading application")
			appbits.UploadApp(server, "test-guid", zipFile, presentFiles)
			log.Println("Uploaded application")
			log.Println("Deleting uploads file")
		})
	})
	log.Println("Push completed")
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
