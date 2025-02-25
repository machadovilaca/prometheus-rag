package vectordb_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVectordb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vectordb Suite")
}
