package bloblet

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Consenser", func() {
	It("should consense small directories", func() {
		dir, err := Scan("./test/fasterxml")
		Expect(err).NotTo(HaveOccurred())

		dir.Condense(65536 * 3)

		//		dumpBloblets(dir, 0)

		Expect(dir.children["test/fasterxml/jackson"].
			children["test/fasterxml/jackson/core"].
			bloblet.size).To(Equal(int64(443846)))

		Expect(dir.children["test/fasterxml/jackson"].
			children["test/fasterxml/jackson/databind"].
			children["test/fasterxml/jackson/databind/deser"].
			bloblet.size).To(Equal(int64(479705)))
	})
})

func dumpBloblets(dir *directory, indent int) {
	if dir.bloblet != nil {
		fmt.Printf(strings.Repeat(" ", indent))
		spew.Printf("Bloblet @ %s of size %d\n", dir.bloblet.path, dir.bloblet.size)
	}

	for _, child := range dir.children {
		dumpBloblets(child, indent+2)
	}
}
