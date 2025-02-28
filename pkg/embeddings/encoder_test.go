package embeddings_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

const bertDimension = 768

var _ = Describe("Encoder", func() {
	var (
		encoder embeddings.Encoder
	)

	BeforeEach(func() {
		var err error

		encoder, err = embeddings.NewEncoder(embeddings.Config{
			ModelsDir: "../../_models",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("GetDimension", func() {
		It("should return the dimension of the model", func() {
			dimension, err := encoder.GetDimension()
			Expect(err).NotTo(HaveOccurred())
			Expect(dimension).To(Equal(bertDimension))
		})
	})

	Context("Encode", func() {
		It("should encode query", func() {
			query := "test query"

			vector, err := encoder.EncodeQuery(query)
			Expect(err).NotTo(HaveOccurred())
			Expect(vector).NotTo(BeNil())
			Expect(len(vector)).To(Equal(bertDimension))
		})

		It("should encode metric metadata", func() {
			metadata := prometheus.MetricMetadata{
				Name:   "test_metric",
				Help:   "test help",
				Type:   "counter",
				Labels: []string{"label1", "label2"},
			}

			vector, err := encoder.EncodeMetricMetadata(metadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(vector).NotTo(BeNil())
		})
	})
})
