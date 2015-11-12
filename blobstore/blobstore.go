package blobstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var blobDir string
var mu sync.Mutex

func init() {
	var err error
	blobDir, err = ioutil.TempDir("", "blobstore")
	if err != nil {
		panic("Failed to create blobstore: " + err.Error())
	}
}

func Present(key string) bool {
	mu.Lock()
	defer mu.Unlock()
	return present(key)
}

func Add(key, src string) {
	mu.Lock()
	defer mu.Unlock()
	dest := pathOf(key)
	if present(key) {
		err := exec.Command("cmp", src, dest).Run()
		if err, ok := err.(*exec.ExitError); ok {
			panic(fmt.Sprintf("Cannot add file %s: key %s already present in blobstore", src, key))
		} else if err != nil {
			panic("Add failed: " + err.Error())
		}
		// Files are equal
		return
	}
	copyFile(dest, src)
}

func Get(key, dest string) {
	mu.Lock()
	defer mu.Unlock()
	if !present(key) {
		panic("Key not present")
	}
	copyFile(dest, pathOf(key))
}

func pathOf(key string) string {
	return filepath.Join(blobDir, key)
}

func present(key string) bool {
	_, err := os.Lstat(pathOf(key))
	return err == nil
}

func copyFile(dest, src string) {
	err := exec.Command("cp", src, dest).Run()
	if err != nil {
		panic("copyFile failed: " + err.Error())
	}
}
