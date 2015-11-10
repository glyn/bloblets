package resmatch

type matcher map[string]int64

func NewMatcher() matcher {
	return make(map[string]int64, 100)
}

func (m matcher) matchResources(ifs []IntegrityFields) ([]IntegrityFields, error) {
	matched := make([]IntegrityFields, 0, len(m))
	for _, f := range ifs {
		if f.Size > 65535 /* TODO: m[f.Sha1] == f.Size*/ {
			matched = append(matched, f)
		}
	}

	return matched, nil
}
