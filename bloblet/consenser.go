package bloblet

import (
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/glyn/bloblets/bloblet/filehash"
)

var noChildren = make(map[string]*models.AppFileFields)
var zeroHash = filehash.Zero()

// Condenses the files in the given directory and returns details of (a) files which
// are already larger than minCompressedBlobletSize, and (b) bloblets.
func (dir *directory) Condense(minBlobletSize, minCompressedBlobletSize int64) []models.AppFileFields {
	affs := make([]models.AppFileFields, 0, 100)
	for _, child := range dir.children {
		caffs := child.Condense(minBlobletSize, minCompressedBlobletSize)
		affs = appendAll(affs, caffs)
		if child.bloblet == nil {
			addAll(dir.files, child.files)
			dir.size += child.size
			dir.hash.Combine(child.hash)

			// Show child has been condensed into parent
			child.files = noChildren
			child.size = 0
			child.hash = zeroHash
		}
	}

	// Extract any individual files which need not go into bloblets
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

	if dir.size >= minBlobletSize {
		dir.bloblet = &bloblet{
			path:  dir.path,
			hash:  dir.hash,
			size:  dir.size,
			files: dir.files,
		}
		affs = append(affs, models.AppFileFields{
			Path: dir.path,
			Sha1: dir.hash.String(),
			Size: dir.size,
			Mode: "0755",
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
