package bloblet

import (
	"archive/zip"
	"io"
	"os"

	"path/filepath"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"github.com/glyn/bloblets/bloblet/filehash"
	"github.com/glyn/bloblets/cliutil"
)

type Bloblet interface {
	Path() string
	Size() int64
	Hash() string
	Write(io.Writer)
}

type bloblet struct {
	path  string
	hash  filehash.Hash
	size  int64
	files map[string]*models.AppFileFields
}

func (b *bloblet) Compress(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE+os.O_RDWR, 0644)
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

func (b *bloblet) Path() string {
	return b.path
}

func (b *bloblet) Size() int64 {
	return b.size
}

func (b *bloblet) Hash() string {
	return b.hash.String()
}

func (b *bloblet) Write(w io.Writer) {
	fileutils.TempDir("bloblet-zipdir", func(zipDir string, err error) {
		cliutil.Check(err)
		zipPath := filepath.Join(zipDir, BlobletFileName)
		cliutil.Check(b.Compress(zipPath))
		cliutil.Check(fileutils.CopyPathToWriter(zipPath, w))
	})
}
