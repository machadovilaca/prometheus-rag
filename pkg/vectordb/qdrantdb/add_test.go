package qdrantdb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("Add", func() {
	var (
		dbClient vectordb.Client
	)

	BeforeEach(func() {
		var err error
		dbClient, err = vectordb.New(vectordb.Config{
			Provider:               "qdrant",
			QdrantHost:             "localhost",
			QdrantPort:             6334,
			CollectionName:         "test-collection",
			EncoderOutputDirectory: "../../../_models",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := dbClient.DeleteCollection()
		Expect(err).NotTo(HaveOccurred())

		err = dbClient.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should successfully add metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name:   "test_metric",
			Help:   "Test metric help text",
			Type:   "counter",
			Labels: []string{"label1", "label2"},
		}

		err := dbClient.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail when adding invalid metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name: "", // Empty name should fail
			Help: "Test help",
		}

		err := dbClient.AddMetricMetadata(metadata)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("name is required"))
	})

	It("should replace existing metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name: "test_metric",
			Help: "Test help",
			Type: "counter",
		}

		err := dbClient.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		metadata.Help = "Updated help"

		err = dbClient.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test_metric", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Help).To(Equal("Updated help"))
	})

	It("should successfully add a batch of metric metadata", func() {
		metadata := []*prometheus.MetricMetadata{
			{
				Name: "test_metric_1",
				Help: "Test help 1",
				Type: "counter",
			},
			{
				Name: "test_metric_2",
				Help: "Test help 2",
				Type: "counter",
			},
		}

		err := dbClient.BatchAddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))

		Expect(results).To(ContainElement(HaveField("Name", "test_metric_1")))
		Expect(results).To(ContainElement(HaveField("Name", "test_metric_2")))
	})

	It("should skip batch add of metric metadata when there are none", func() {
		err := dbClient.BatchAddMetricMetadata([]*prometheus.MetricMetadata{})
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(0))
	})
})
