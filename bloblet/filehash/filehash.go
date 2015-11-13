package filehash

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/cloudfoundry/gofileutils/fileutils"
)

type Hash []byte

func Zero() Hash {
	return Hash(make([]byte, sha1.Size))
}

// New returns a Hash of the contents of the specified file.
func New(filePath string) Hash {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		panic(err)
	}

	if fileInfo.IsDir() {
		panic("cannot compute hash of directory")
	} else {
		hash := sha1.New()
		err = fileutils.CopyPathToWriter(filePath, hash)
		if err != nil {
			panic(err)
		}
		return Hash(hash.Sum(nil))
	}
}

func (k Hash) String() string {
	return fmt.Sprintf("%x", string(k))
}

func (h1 Hash) Combine(h2 Hash) {
	if len(h1) != len(h2) {
		panic("Invalid hash length")
	}
	for i, b := range h1 {
		h1[i] = b ^ h2[i]
	}
}
