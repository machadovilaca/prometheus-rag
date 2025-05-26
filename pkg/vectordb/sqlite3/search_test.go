package sqlite3_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("Search", func() {
	var (
		dbClient vectordb.Client
		tempDir  string
		dbPath   string
	)

	BeforeEach(func() {
		var err error
		// Create temporary directory for test database
		tempDir, err = os.MkdirTemp("", "sqlite3_search_test_*")
		Expect(err).NotTo(HaveOccurred())

		dbPath = filepath.Join(tempDir, "test.db")

		// Create SQLite3 client
		dbClient, err = vectordb.New(vectordb.Config{
			Provider:               "sqlite3",
			Sqlite3DBPath:          dbPath,
			CollectionName:         "test-collection",
			EncoderOutputDirectory: "../../../_models",
		})
		Expect(err).NotTo(HaveOccurred())

		// Create collection
		err = dbClient.CreateCollection()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if dbClient != nil {
			err := dbClient.Close()
			Expect(err).NotTo(HaveOccurred())
		}

		// Clean up temporary directory
		if tempDir != "" {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("should return best matching metrics first", func() {
		err := dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name:   "http_requests_total",
			Help:   "Total number of HTTP requests",
			Type:   "counter",
			Labels: []string{"method", "status"},
		})
		Expect(err).NotTo(HaveOccurred())

		err = dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name:   "node_memory_usage",
			Help:   "Memory usage of node",
			Type:   "gauge",
			Labels: []string{"node"},
		})
		Expect(err).NotTo(HaveOccurred())

		// Give some time for processing
		time.Sleep(100 * time.Millisecond)

		results, err := dbClient.SearchMetrics("http requests", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))

		// Check that http_requests_total is in the results
		found := false
		for _, result := range results {
			if result.Name == "http_requests_total" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

		results, err = dbClient.SearchMetrics("memory", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))

		// Check that node_memory_usage is in the results
		found = false
		for _, result := range results {
			if result.Name == "node_memory_usage" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})

	It("should return empty results when no matches found", func() {
		results, err := dbClient.SearchMetrics("does not exist", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	})

	It("should respect the limit parameter", func() {
		err := dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "metric1",
			Help: "Test metric 1",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		err = dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "metric2",
			Help: "Test metric 2",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test metric", 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
	})

	It("should handle search with empty collection", func() {
		results, err := dbClient.SearchMetrics("anything", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	})

	It("should handle search with special characters", func() {
		err := dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name:   "metric_with_special-chars.test",
			Help:   "Metric with special characters: !@#$%^&*()",
			Type:   "histogram",
			Labels: []string{"label-with-dash", "label.with.dots"},
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("special characters", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Name).To(Equal("metric_with_special-chars.test"))
	})

	It("should handle zero limit", func() {
		err := dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "test_metric",
			Help: "Test metric",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test", 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
	})

	It("should handle large limit", func() {
		err := dbClient.AddMetricMetadata(&prometheus.MetricMetadata{
			Name: "test_metric",
			Help: "Test metric",
			Type: "counter",
		})
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("test", 1000)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
	})

	It("should find metrics with complex queries", func() {
		// Add multiple metrics
		metrics := []*prometheus.MetricMetadata{
			{
				Name:   "cpu_usage_percent",
				Help:   "CPU usage percentage by core",
				Type:   "gauge",
				Labels: []string{"cpu", "mode"},
			},
			{
				Name:   "memory_available_bytes",
				Help:   "Available memory in bytes",
				Type:   "gauge",
				Labels: []string{"node"},
			},
			{
				Name:   "disk_io_operations_total",
				Help:   "Total disk I/O operations",
				Type:   "counter",
				Labels: []string{"device", "type"},
			},
		}

		err := dbClient.BatchAddMetricMetadata(metrics)
		Expect(err).NotTo(HaveOccurred())

		// Search for CPU related metrics
		results, err := dbClient.SearchMetrics("cpu usage", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">=", 1))

		// Check that cpu_usage_percent is found
		found := false
		for _, result := range results {
			if result.Name == "cpu_usage_percent" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

		// Search for disk related metrics
		results, err = dbClient.SearchMetrics("disk operations", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">=", 1))

		// Check that disk_io_operations_total is found
		found = false
		for _, result := range results {
			if result.Name == "disk_io_operations_total" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})
})
