package resmatch

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/glyn/bloblets/cliutil"
)

func integrityFieldsJSONReader() io.Reader {
	integrityFieldsJson, err := json.Marshal(integrityFields())
	cliutil.Check(err)
	return bytes.NewReader(integrityFieldsJson)
}

func integrityFieldsJSON() ([]byte, error) {
	integrityFieldsJson, err := json.Marshal(integrityFields())
	if err != nil {
		return nil, err
	}
	return integrityFieldsJson, nil
}

func integrityFields() []IntegrityFields {
	return []IntegrityFields{}
}

type IntegrityFields struct {
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}
