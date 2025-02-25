package embeddings_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEmbeddings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Embeddings Suite")
}
