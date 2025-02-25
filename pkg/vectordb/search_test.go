package vectordb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("Search", func() {
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

	It("should return best matching metrics first", func() {
		err := db.AddMetricMetadata(&prometheus.MetricMetadata{
			Name:   "http_requests_total",
			Help:   "Total number of HTTP requests",
			Type:   "counter",
			Unit:   "requests",
			Labels: []string{"method", "status"},
		})
		Expect(err).NotTo(HaveOccurred())

		err = db.AddMetricMetadata(&prometheus.MetricMetadata{
			Name:   "node_memory_usage",
			Help:   "Memory usage of node",
			Type:   "gauge",
			Unit:   "bytes",
			Labels: []string{"node"},
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := db.SearchMetrics("http requests", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[0].Name).To(Equal("http_requests_total"))
		Expect(results[1].Name).To(Equal("node_memory_usage"))

		results, err = db.SearchMetrics("memory", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[0].Name).To(Equal("node_memory_usage"))
		Expect(results[1].Name).To(Equal("http_requests_total"))
	})

	It("should return empty results when no matches found", func() {
		results, err := db.SearchMetrics("does not exist", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	})

	It("should respect the limit parameter", func() {
		err := db.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "metric1",
			Help: "Test metric 1",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		err = db.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "metric2",
			Help: "Test metric 2",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := db.SearchMetrics("test metric", 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
	})
})
