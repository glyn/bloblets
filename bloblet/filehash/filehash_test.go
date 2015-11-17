package filehash_test

import (
	"github.com/glyn/bloblets/bloblet/filehash"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hash", func() {
	It("should compute the correct hash of a file", func() {
		Expect(filehash.New("./test/file1").String()).To(Equal("f6e5153f472c30787e548f239c2ee7a2d9534d4f"))
		Expect(filehash.New("./test/file2").String()).To(Equal("995a645c53389f12ca08ca2b206527927741b5de"))
	})

	Describe("Combine", func() {
		It("should correctly combine hashes of files", func() {
			h := filehash.New("./test/file1")
			h.Combine(filehash.New("./test/file2"))
			Expect(h.String()).To(Equal("6fbf71631414af6ab45c4508bc4bc030ae12f891"))
		})

		It("should be commutative", func() {
			h := filehash.New("./test/file1")
			h.Combine(filehash.New("./test/file2"))

			i := filehash.New("./test/file2")
			i.Combine(filehash.New("./test/file1"))
			Expect(h).To(Equal(i))
		})

		It("should leave the original hash unmodified when called with a zero hash", func() {
			h := filehash.New("./test/file1")
			h.Combine(filehash.Zero())
			Expect(h).To(Equal(filehash.New("./test/file1")))
		})
	})

	Describe("Remove", func() {
		It("should invert the effects of Combine", func() {
			h := filehash.New("./test/file1")
			i := filehash.New("./test/file2")
			h.Combine(i)
			h.Remove(i)
			Expect(h).To(Equal(filehash.New("./test/file1")))
		})

	})

	Describe("StringToHash", func() {
		It("should correctly convert a stringified hash", func() {
			i := filehash.New("./test/file1")
			Expect(filehash.StringToHash(i.String())).To(Equal(i))
		})
	})
})
