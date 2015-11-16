package bloblet

import "github.com/glyn/bloblets/bloblet/filehash"

func (dir *directory) Condense(minBlobletSize int64) *app {
	for _, child := range dir.children {
		child.Condense(minBlobletSize)
		if child.bloblet == nil {
			addAll(dir.files, child.files)
			dir.size += child.size
			dir.hash.Combine(child.hash)
		}
	}

	if dir.size >= minBlobletSize {
		dir.bloblet = &bloblet{
			path:  dir.path,
			hash:  dir.hash,
			size:  dir.size,
			files: dir.files,
		}
	}
	return nil
}

func addAll(m1, m2 map[string]filehash.Hash) {
	for k, v := range m2 {
		m1[k] = v
	}
}
