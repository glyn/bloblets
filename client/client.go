package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"log"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/appbits"
	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
	"github.com/glyn/bloblets/scanner"
)

func main() {
	client := &http.Client{}

	scanner.Scan(os.Args[1], func(appDir string, appFiles []models.AppFileFields) {
		request := resmatch.ResMatchRequest(integrityFields(appFiles))

		response := cliutil.Converse(client, request)

		presentIffs := resmatch.ProcessResponse(response)

		presentFiles := intersectAppFilesIntegrityFields(appFiles, presentIffs)

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

			cliutil.Check(app_files.ApplicationFiles{}.CopyFiles(appFilesToUpload, appDir, uploadDir))

			fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
				if hasFileToUpload {
					log.Println("about to zipAppFiles")
					cliutil.Check(zipAppFiles(zipFile, appDir, uploadDir))
					log.Println("zipAppFiles completed")
				}

				log.Println("about to UploadApp")
				appbits.UploadApp("test-guid", zipFile, presentFiles)
				log.Println("UploadApp completed")
			})
		})
	})
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
		fileInfo, err := os.Lstat(filepath.Join(appDir, presentFiles[i].Path))
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
