package scanner

import (
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

			cliutil.Check(zipper.Unzip(path, tmpDir))

			doScan(tmpDir, processAppFiles)
		})
	} else {
		doScan(path, processAppFiles)
	}
}

func doScan(appDir string, processAppFiles func(appDir string, affs []models.AppFileFields)) {
	af := app_files.ApplicationFiles{}
	affs, err := af.AppFilesInDir(appDir)
	cliutil.Check(err)
	processAppFiles(appDir, affs)
}
