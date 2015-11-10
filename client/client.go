package main

import (
	"net/http"
	"os"

	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
	"github.com/glyn/bloblets/scanner"
)

func main() {
	client := &http.Client{}

	ifs := scanner.Scan(os.Args[1])

	request := resmatch.ResMatchRequest(ifs)

	response := cliutil.Converse(client, request)

	resmatch.ProcessResponse(response)
}
