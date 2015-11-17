package bloblet

import (
	"fmt"
	"path/filepath"

	"strings"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/glyn/bloblets/bloblet/filehash"
)

type Bloblet interface {
}

type Condensate interface {
	Bloblets([]models.AppFileFields) ([]Bloblet, []models.AppFileFields)
}

const blobletFileName = "com-github-glyn-bloblet.zip"

var blobletFileNamePathTerminator = filepath.Join("", blobletFileName)

var noFiles = make(map[string]*models.AppFileFields)
var zeroHash = filehash.Zero()

// Condenses the files in the given directory and returns details of (a) files which
// are already larger than minCompressedBlobletSize, and (b) bloblets.
func (dir *directory) Condense(minBlobletSize, minCompressedBlobletSize int64) []models.AppFileFields {
	affs := make([]models.AppFileFields, 0, 100)
	for _, child := range dir.children {
		caffs := child.Condense(minBlobletSize, minCompressedBlobletSize)
		affs = appendAll(affs, caffs)
		// If the child has not been condensed, copy its files to its parent (this directory).
		if child.bloblet == nil {
			addAll(dir.files, child.files)
			dir.size += child.size
			dir.hash.Combine(child.hash)

			// Show child has been condensed into parent - only for use in testing.
			child.files = noFiles
			child.size = 0
			child.hash = zeroHash
		}
	}

	// Extract any individual files which should not be condensed into bloblets
	largeFiles := []string{}
	for f, aff := range dir.files {
		if aff.Size > minCompressedBlobletSize {
			largeFiles = append(largeFiles, f)
			affs = append(affs, *aff)

			// Remove the file's contribution to the directory's hash and size.
			dir.hash.Remove(filehash.StringToHash(aff.Sha1))
			dir.size -= aff.Size
		}
	}
	for _, f := range largeFiles {
		delete(dir.files, f)
	}

	// If the remaining files in this directory have sufficient total size, then condense them into a bloblet.
	// Note: the bloblet is zipped later if it needs uploading.
	if dir.size >= minBlobletSize {
		dir.bloblet = &bloblet{
			path:  dir.path,
			hash:  dir.hash,
			size:  dir.size,
			files: dir.files,
		}
		affs = append(affs, models.AppFileFields{
			Path: filepath.Join(dir.path, blobletFileName),
			Sha1: dir.hash.String(),
			Size: dir.size,
			Mode: "0755", // precise value unimportant
		})
	}
	return affs
}

func addAll(m1, m2 map[string]*models.AppFileFields) {
	for k, v := range m2 {
		m1[k] = v
	}
}

func appendAll(s1, s2 []models.AppFileFields) []models.AppFileFields {
	for _, aff := range s2 {
		s1 = append(s1, aff)
	}
	return s1
}

// Returns the bloblets references in affs and affs minus the bloblets.
func (dir *directory) Bloblets(affs []models.AppFileFields) ([]Bloblet, []models.AppFileFields) {
	bloblets := make([]Bloblet, 0, 100)
	nonBloblets := make([]models.AppFileFields, 0, len(affs))
	for _, aff := range affs {
		b := dir.findBloblet(aff.Path)
		if b == nil {
			nonBloblets = append(nonBloblets, aff)
		} else {
			bloblets = append(bloblets, b)
		}
	}
	return bloblets, nonBloblets
}

func (dir *directory) findBloblet(path string) Bloblet {
	if !strings.HasSuffix(path, blobletFileNamePathTerminator) {
		return nil
	}
	var p string
	if path == blobletFileNamePathTerminator {
		p = "."
	} else {
		p = path[:len(path)-len(blobletFileNamePathTerminator)-1]
	}
	b, ok := dir.findBlobletWithPath(p)
	if !ok {
		panic(fmt.Sprintf("bad path %s", p))
	}
	return b
}

func (dir *directory) findBlobletWithPath(path string) (Bloblet, bool) {
	if dir.path == path {
		return dir.bloblet, true
	}
	for _, child := range dir.children {
		b, ok := child.findBlobletWithPath(path)
		if ok {
			return b, ok
		}
	}
	return nil, false
}
