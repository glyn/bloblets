package resmatch

import "github.com/glyn/bloblets/blobstore"

type matcher map[string]int64

func NewMatcher() matcher {
	return make(map[string]int64, 100)
}

func (m matcher) matchResources(ifs []IntegrityFields) ([]IntegrityFields, error) {
	matched := make([]IntegrityFields, 0, len(m))
	for _, f := range ifs {

		if f.Size > 65535 && blobstore.Present(f.Sha1) {
			matched = append(matched, f)
		}
	}

	return matched, nil
}
