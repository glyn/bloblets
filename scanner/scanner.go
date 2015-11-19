package scanner

import (
	"log"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/bloblet"
	"github.com/glyn/bloblets/cliutil"
)

func Scan(path string, processAppFiles func(appDir string, condensate bloblet.Condensate, affs []models.AppFileFields)) {
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

func doScan(appDir string, processAppFiles func(appDir string, condensate bloblet.Condensate, affs []models.AppFileFields)) {
	log.Println("Scanning application files and computing SHA1s")
	affs, condensate, err := AppFilesInDir(appDir)
	cliutil.Check(err)
	log.Println("Scanned application files and computed SHA1s")

	processAppFiles(appDir, condensate, affs)
}

// Returns bloblets and other files for resource matching.
func AppFilesInDir(dir string) (appFiles []models.AppFileFields, condensate bloblet.Condensate, err error) {
	d, err := bloblet.Scan(dir)
	if err != nil {
		return []models.AppFileFields{}, nil, err
	}

	log.Println("Condensing")

	affs := d.Condense(int64(4*65536), int64(65536))
	return affs, d, nil
}
