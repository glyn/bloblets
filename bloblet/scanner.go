package bloblet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/glyn/bloblets/bloblet/filehash"
)

type Condenser interface {
	Condense(minBlobletSize, minCompressedBlobletSize int64) []models.AppFileFields
}

type directory struct {
	path     string
	hash     filehash.Hash
	size     int64
	files    map[string]*models.AppFileFields
	children map[string]*directory
	bloblet  *bloblet // non-nil iff this directory has an associated bloblet
}

func Scan(path string) (*directory, error) {
	return doScan(path, ".")
}

func doScan(appDir, path string) (*directory, error) {
	fullPath := filepath.Join(appDir, path)
	fi, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s for path %s failed: %s", appDir, path, err)
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("Not a directory: %s", path)
	}

	dir := directory{
		path:     path,
		hash:     filehash.Zero(),
		size:     0,
		files:    make(map[string]*models.AppFileFields, 50),
		children: make(map[string]*directory, 10),
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s failed on Open: %s", fullPath, err)
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s failed on Readdir: %s", fullPath, err)
	}

	for _, fi := range fis {
		fp := filepath.Join(fullPath, fi.Name())
		fpRelative := filepath.Join(path, fi.Name())
		if !fi.IsDir() {
			h := filehash.New(fp)
			dir.files[fp] = &models.AppFileFields{
				Sha1: h.String(),
				Size: fi.Size(),
				Path: fpRelative,
				Mode: fmt.Sprintf("%#o", fi.Mode()),
			}
			dir.size += fi.Size()
			dir.hash.Combine(h)
		} else {
			s, err := doScan(appDir, fpRelative)
			if err != nil {
				return nil, fmt.Errorf("Scan of %s failed to scan subdirectory %s: %s", fullPath, fp, err)
			}
			dir.children[fp] = s
		}
	}

	return &dir, nil
}

func subdirectories(dirPath string) ([]string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdirnames(-1)
}
