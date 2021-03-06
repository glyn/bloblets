package bloblet

import (
	"io/ioutil"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/glyn/bloblets/bloblet/filehash"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	It("should correctly scan an empty directory", func() {
		d, err := ioutil.TempDir("", "empty")
		Expect(err).NotTo(HaveOccurred())

		dir, err := Scan(d)
		Expect(err).NotTo(HaveOccurred())
		Expect(dir.path).To(Equal("."))
		Expect(dir.hash).To(Equal(filehash.Zero()))
		Expect(dir.size).To(Equal(int64(0)))
		Expect(dir.files).To(BeEmpty())
		Expect(dir.children).To(BeEmpty())
		Expect(dir.bloblet).To(BeNil())
	})

	It("should correctly scan a directory with no children", func() {
		dir, err := Scan("./test/nochildren")
		Expect(err).NotTo(HaveOccurred())
		Expect(dir.path).To(Equal("."))
		Expect(dir.hash.String()).To(Equal("6fbf71631414af6ab45c4508bc4bc030ae12f891"))
		Expect(dir.size).To(Equal(int64(3218)))
		Expect(hashes(dir.files)).To(Equal(map[string]string{"test/nochildren/file1": filehash.New("./test/nochildren/file1").String(),
			"test/nochildren/file2": filehash.New("./test/nochildren/file2").String()}))
		Expect(dir.children).To(BeEmpty())
		Expect(dir.bloblet).To(BeNil())
	})

	It("should correctly scan a directory with children", func() {
		dir, err := Scan("./test/withchildren")
		Expect(err).NotTo(HaveOccurred())
		Expect(dir.path).To(Equal("."))
		Expect(dir.hash.String()).To(Equal("6fbf71631414af6ab45c4508bc4bc030ae12f891"))
		Expect(dir.size).To(Equal(int64(3218)))
		Expect(hashes(dir.files)).To(Equal(map[string]string{"test/withchildren/file1": filehash.New("./test/withchildren/file1").String(),
			"test/withchildren/file2": filehash.New("./test/withchildren/file2").String()}))
		Expect(len(dir.children)).To(Equal(1))
		_, ok := dir.children["test/withchildren/child1"]
		Expect(ok).To(BeTrue())
		Expect(dir.bloblet).To(BeNil())
	})
})

func hashes(files map[string]*models.AppFileFields) map[string]string {
	h := make(map[string]string, len(files))
	for f, a := range files {
		h[f] = a.Sha1
	}
	return h
}
