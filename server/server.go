package main

import (
	"net/http"

	"github.com/glyn/bloblets/appbits"
	"github.com/glyn/bloblets/resmatch"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/v2/resource_match", resmatch.ResourceMatchHandler)
	serveMux.HandleFunc("/v2/apps/", appbits.AppHandler)
	http.ListenAndServe("localhost:8080", serveMux)
}
