package qdrantdb_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQdrantdb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Qdrantdb Suite")
}
