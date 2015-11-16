package bloblet

import "github.com/glyn/bloblets/bloblet/filehash"

type app struct {
	bloblets map[string]*bloblet
	extra    map[string]*directory
}

type bloblet struct {
	path  string
	hash  filehash.Hash
	size  int64
	files map[string]filehash.Hash
}

//       Scan               Condense
// path ------> *directory ----------> *app
