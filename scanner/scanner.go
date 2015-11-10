package scanner

import (
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
)

func Scan(path string) (ifs []resmatch.IntegrityFields) {
	zipper := app_files.ApplicationZipper{}

	if zipper.IsZipFile(path) {
		fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
			cliutil.Check(err)

			cliutil.Check(zipper.Unzip(path, tmpDir))

			ifs = doScan(tmpDir)
		})
	} else {
		ifs = doScan(path)
	}
	return
}

func doScan(dirPath string) []resmatch.IntegrityFields {
	af := app_files.ApplicationFiles{}
	affs, err := af.AppFilesInDir(dirPath)
	cliutil.Check(err)
	ifs := make([]resmatch.IntegrityFields, 0, len(affs))

	for _, aff := range affs {
		ifs = append(ifs, resmatch.IntegrityFields{
			Sha1: aff.Sha1,
			Size: aff.Size,
		})
	}

	return ifs
}
