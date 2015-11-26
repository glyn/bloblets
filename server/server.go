package main

import (
	"net/http"
	"os"

	"github.com/glyn/bloblets/appbits"
)

func main() {
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/v2/apps/", appbits.AppHandler)
	http.ListenAndServe(":"+os.Getenv("PORT"), serveMux)
}
