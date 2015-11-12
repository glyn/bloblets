package scanner

import (
	"log"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/cliutil"
)

func Scan(path string, processAppFiles func(appDir string, affs []models.AppFileFields)) {
	zipper := app_files.ApplicationZipper{}

	if zipper.IsZipFile(path) {
		fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
			cliutil.Check(err)

			log.Println("Unzipping application")
			cliutil.Check(zipper.Unzip(path, tmpDir))
			log.Println("Unzipped application")

			doScan(tmpDir, processAppFiles)

			log.Println("Deleting unzipped application")
		})
	} else {
		doScan(path, processAppFiles)
	}
}

func doScan(appDir string, processAppFiles func(appDir string, affs []models.AppFileFields)) {
	af := app_files.ApplicationFiles{}
	log.Println("Scanning application files and computing SHA1s")
	affs, err := af.AppFilesInDir(appDir)
	cliutil.Check(err)
	log.Println("Scanned application files and computed SHA1s")

	processAppFiles(appDir, affs)
}
