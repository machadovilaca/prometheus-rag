package vectordb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("Add", func() {
	var (
		db vectordb.VectorDB
	)

	BeforeEach(func() {
		var err error
		db, err = vectordb.New(vectordb.Config{
			Host: "localhost",
			Port: 6334,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := db.DeleteCollection()
		Expect(err).NotTo(HaveOccurred())

		err = db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should successfully add metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name:   "test_metric",
			Help:   "Test metric help text",
			Type:   "counter",
			Unit:   "seconds",
			Labels: []string{"label1", "label2"},
		}

		err := db.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail when adding invalid metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name: "", // Empty name should fail
			Help: "Test help",
		}

		err := db.AddMetricMetadata(metadata)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("name, help, and type are required"))
	})

	It("should replace existing metric metadata", func() {
		metadata := &prometheus.MetricMetadata{
			Name: "test_metric",
			Help: "Test help",
			Type: "counter",
		}

		err := db.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		metadata.Help = "Updated help"

		err = db.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		results, err := db.SearchMetrics("test_metric", 10)
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

		err := db.BatchAddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		results, err := db.SearchMetrics("test", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))

		Expect(results).To(ContainElement(HaveField("Name", "test_metric_1")))
		Expect(results).To(ContainElement(HaveField("Name", "test_metric_2")))
	})
})
