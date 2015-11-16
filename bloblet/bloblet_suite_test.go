package bloblet_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBloblet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bloblet Suite")
}
