package cliutil

import (
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
)

func Check(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatalf("Error: %s", err)
	}
}

func Converse(client *http.Client, request *http.Request) []byte {
	response, err := client.Do(request)
	Check(err)
	defer response.Body.Close()

	responseBytes, err := ioutil.ReadAll(response.Body)
	Check(err)

	if response.StatusCode != 200 {
		log.Fatalf("Received status %d: %s", response.StatusCode, string(responseBytes))
	}
	return responseBytes
}
