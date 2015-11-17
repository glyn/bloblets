package bloblet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glyn/bloblets/bloblet/filehash"
)

type directory struct {
	path     string
	hash     filehash.Hash
	size     int64
	files    map[string]filehash.Hash
	children map[string]*directory
	bloblet  *bloblet // non-nil iff this directory has an associated bloblet
}

func Scan(path string) (*directory, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s failed: %s", path, err)
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("Not a directory: %s", path)
	}

	dir := directory{
		path:     path,
		hash:     filehash.Zero(),
		size:     0,
		files:    make(map[string]filehash.Hash, 50),
		children: make(map[string]*directory, 10),
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s failed on Open: %s", path, err)
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("Scan of %s failed on Readdir: %s", path, err)
	}

	for _, fi := range fis {
		fp := filepath.Join(path, fi.Name())
		if !fi.IsDir() {
			h := filehash.New(fp)
			dir.files[fp] = h
			dir.size += fi.Size()
			dir.hash.Combine(h)
		} else {
			s, err := Scan(fp)
			if err != nil {
				return nil, fmt.Errorf("Scan of %s failed to scan subdirectory %s: %s", path, fp, err)
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
