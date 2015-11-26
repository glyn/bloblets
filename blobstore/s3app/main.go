package main

import (
	"os"

	"github.com/glyn/bloblets/blobstore"
)

func main() {
	defer blobstore.Terminate()

	blobstore.Add("key name", os.Args[1])

	present := blobstore.Present("key name")
	if !present {
		panic("Head request failed")
	}

	blobstore.Get("key name", os.Args[2])
}
