package bloblet

import (
	"archive/zip"
	"os"

	"path/filepath"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"github.com/glyn/bloblets/bloblet/filehash"
)

type app struct {
	bloblets map[string]*bloblet
	extra    map[string]*directory
}

type bloblet struct {
	path  string
	hash  filehash.Hash
	size  int64
	files map[string]*models.AppFileFields
}

func (b *bloblet) Compress(dir string) error {
	f, err := os.OpenFile(filepath.Join(dir, "bloblet.zip"), os.O_CREATE+os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	for filePath, _ := range b.files {
		w, err := zw.Create(filePath)
		if err != nil {
			return err
		}
		err = fileutils.CopyPathToWriter(filePath, w)
		if err != nil {
			return err
		}
	}
	return nil
}

//       Scan               Condense
// path ------> *directory ----------> *app
