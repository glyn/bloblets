package resmatch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/glyn/bloblets/cliutil"
)

func ResMatchRequest(server string, ifs []IntegrityFields) *http.Request {
	url := fmt.Sprintf("http://%s/v2/resource_match", server)
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
