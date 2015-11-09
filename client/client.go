package main

import (
	"net/http"

	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
)

func main() {
	client := &http.Client{}

	request := resmatch.ResMatchRequest()

	response := cliutil.Converse(client, request)

	resmatch.ProcessResponse(response)
}
