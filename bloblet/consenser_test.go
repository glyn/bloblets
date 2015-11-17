package bloblet

import (
	"fmt"
	"strings"

	"io/ioutil"
	"os"
	"path/filepath"

	"math"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consenser", func() {
	It("should condense small directories", func() {
		minCompressedSize := int64(65536)
		minSize := int64(4 * minCompressedSize)
		dir, err := Scan("./test/fasterxml")
		Expect(err).NotTo(HaveOccurred())

		condensate := dir.Condense(minSize, minCompressedSize)

		n, t, minc := dumpBloblets(dir, 0)
		nb := dumpNonBloblets(dir, 0)

		fmt.Printf("Min size %d produced %d bloblets of average size %d, minimum compressed size %d, and %d non-bloblets\n",
			minSize, n, t/int64(n), minc, nb)

		Expect(dir.children["test/fasterxml/jackson"].
			children["test/fasterxml/jackson/core"].
			bloblet.size).To(Equal(int64(443846)))

		Expect(dir.children["test/fasterxml/jackson"].
			children["test/fasterxml/jackson/databind"].
			children["test/fasterxml/jackson/databind/deser"].
			bloblet.size).To(Equal(int64(479705)))

		Expect(len(condensate)).To(Equal(n + 1)) // there is one large file in the test data
	})
})

func dumpBloblets(dir *directory, indent int) (int, int64, int64) {
	var numBloblets int = 0
	var totalSize int64 = 0
	var minCompressedSize int64 = math.MaxInt64
	if dir.bloblet != nil {
		cr := compressedSize(dir.bloblet)
		minCompressedSize = cr
		fmt.Printf(strings.Repeat(" ", indent))
		spew.Printf("Bloblet @ %s of size %d, compressed size %d\n", dir.bloblet.path, dir.bloblet.size, cr)
		totalSize += dir.bloblet.size
		numBloblets++

	}

	for _, child := range dir.children {
		n, t, c := dumpBloblets(child, indent+2)
		numBloblets += n
		totalSize += t
		if c < minCompressedSize {
			minCompressedSize = c
		}
	}

	return numBloblets, totalSize, minCompressedSize
}

func compressedSize(b *bloblet) int64 {
	d, err := ioutil.TempDir("", "compress")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(d)

	err = b.Compress(d)
	if err != nil {
		panic(err)
	}

	fi, err := os.Lstat(filepath.Join(d, "bloblet.zip"))
	if err != nil {
		panic(err)
	}

	return fi.Size()
}

func dumpNonBloblets(dir *directory, indent int) int {
	var num int = 0
	if dir.bloblet == nil && dir.size > 0 {
		//		fmt.Printf(strings.Repeat(" ", indent))
		//		spew.Printf("Non-bloblet @ %s of size %d consisting of %d files\n", dir.path, dir.size, len(dir.files))
		num++
	}

	for _, child := range dir.children {
		num += dumpNonBloblets(child, indent+2)
	}

	return num
}
