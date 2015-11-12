package resmatch

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/glyn/bloblets/cliutil"
)

func ResMatchRequest(ifs []IntegrityFields) *http.Request {
	url := "http://localhost:8080/v2/resource_match"
	body := integrityFieldsJSONReader(ifs)
	request, err := http.NewRequest("PUT", url, body)
	cliutil.Check(err)

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "bloblets client v0.1 / "+runtime.GOOS)
	return request
}

func ProcessResponse(response []byte) []IntegrityFields {
	responseFieldsColl := []IntegrityFields{}
	err := json.Unmarshal(response, &responseFieldsColl)
	cliutil.Check(err)
	//	fmt.Printf("Returned SHA1s/Sizes: %#v\n", responseFieldsColl)
	return responseFieldsColl
}
