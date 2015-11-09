package resmatch

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/glyn/bloblets/cliutil"
)

func ResMatchRequest() *http.Request {
	url := "http://localhost:8080/v2/resource_match"
	body := integrityFieldsJSONReader()
	request, err := http.NewRequest("PUT", url, body)
	cliutil.Check(err)

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "bloblets client v0.1 / "+runtime.GOOS)
	return request
}

func ProcessResponse(response []byte) {
	responseFieldsColl := []IntegrityFields{}
	err := json.Unmarshal(response, &responseFieldsColl)
	cliutil.Check(err)
}
