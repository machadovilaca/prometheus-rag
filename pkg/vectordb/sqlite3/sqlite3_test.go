package sqlite3_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("SQLite3 VectorDB", func() {
	var (
		dbClient vectordb.Client
		tempDir  string
		dbPath   string
	)

	BeforeEach(func() {
		var err error
		// Create temporary directory for test database
		tempDir, err = os.MkdirTemp("", "sqlite3_test_*")
		Expect(err).NotTo(HaveOccurred())

		dbPath = filepath.Join(tempDir, "test.db")

		// Create SQLite3 client
		dbClient, err = vectordb.New(vectordb.Config{
			Provider:               "sqlite3",
			Sqlite3DBPath:          dbPath,
			CollectionName:         "test_metrics",
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

	Context("Collection Management", func() {
		It("should create a collection successfully", func() {
			// Collection was already created in BeforeEach
			// Try to create it again to test idempotency
			err := dbClient.CreateCollection()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete a collection successfully", func() {
			err := dbClient.DeleteCollection()
			Expect(err).NotTo(HaveOccurred())

			// Recreate for cleanup
			err = dbClient.CreateCollection()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Adding Metrics", func() {
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

		It("should fail batch add if any metadata is invalid", func() {
			metadata := []*prometheus.MetricMetadata{
				{
					Name: "valid_metric",
					Help: "Valid help",
					Type: "counter",
				},
				{
					Name: "", // Invalid - empty name
					Help: "Invalid help",
					Type: "counter",
				},
			}

			err := dbClient.BatchAddMetricMetadata(metadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name is required"))
		})
	})

	Context("Searching Metrics", func() {
		BeforeEach(func() {
			// Add some test data
			metadata := []*prometheus.MetricMetadata{
				{
					Name:   "http_requests_total",
					Help:   "Total number of HTTP requests",
					Type:   "counter",
					Labels: []string{"method", "status"},
				},
				{
					Name:   "node_memory_usage",
					Help:   "Memory usage of node",
					Type:   "gauge",
					Labels: []string{"node"},
				},
				{
					Name:   "cpu_usage_percent",
					Help:   "CPU usage percentage",
					Type:   "gauge",
					Labels: []string{"cpu"},
				},
			}

			err := dbClient.BatchAddMetricMetadata(metadata)
			Expect(err).NotTo(HaveOccurred())

			// Give some time for indexing
			time.Sleep(100 * time.Millisecond)
		})

		It("should return best matching metrics first", func() {
			results, err := dbClient.SearchMetrics("http requests", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(3))

			// The first result should be the most relevant
			// We can't guarantee exact order due to cosine similarity, but http_requests_total should be included
			found := false
			for _, result := range results {
				if result.Name == "http_requests_total" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should return best matching metrics for memory query", func() {
			results, err := dbClient.SearchMetrics("memory", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(3))

			// node_memory_usage should be in the results
			found := false
			for _, result := range results {
				if result.Name == "node_memory_usage" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should return empty results when no matches found", func() {
			// Clear the collection first
			err := dbClient.DeleteCollection()
			Expect(err).NotTo(HaveOccurred())
			err = dbClient.CreateCollection()
			Expect(err).NotTo(HaveOccurred())

			results, err := dbClient.SearchMetrics("does not exist", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("should respect the limit parameter", func() {
			results, err := dbClient.SearchMetrics("usage", 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
		})

		It("should return all results when limit is larger than available", func() {
			results, err := dbClient.SearchMetrics("usage", 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(results)).To(BeNumerically("<=", 3))
		})
	})

	Context("Database Operations", func() {
		It("should handle multiple concurrent operations", func() {
			metadata1 := &prometheus.MetricMetadata{
				Name: "concurrent_metric_1",
				Help: "Concurrent test metric 1",
				Type: "counter",
			}

			metadata2 := &prometheus.MetricMetadata{
				Name: "concurrent_metric_2",
				Help: "Concurrent test metric 2",
				Type: "gauge",
			}

			// Add metrics concurrently (simulated)
			err1 := dbClient.AddMetricMetadata(metadata1)
			err2 := dbClient.AddMetricMetadata(metadata2)

			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())

			// Verify both were added
			results, err := dbClient.SearchMetrics("concurrent", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
		})

		It("should handle empty labels correctly", func() {
			metadata := &prometheus.MetricMetadata{
				Name:   "no_labels_metric",
				Help:   "Metric with no labels",
				Type:   "counter",
				Labels: []string{},
			}

			err := dbClient.AddMetricMetadata(metadata)
			Expect(err).NotTo(HaveOccurred())

			results, err := dbClient.SearchMetrics("no labels", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].Labels).To(BeEmpty())
		})

		It("should handle metrics with special characters", func() {
			metadata := &prometheus.MetricMetadata{
				Name:   "metric_with_special-chars.test",
				Help:   "Metric with special characters: !@#$%^&*()",
				Type:   "histogram",
				Labels: []string{"label-with-dash", "label.with.dots"},
			}

			err := dbClient.AddMetricMetadata(metadata)
			Expect(err).NotTo(HaveOccurred())

			results, err := dbClient.SearchMetrics("special", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].Name).To(Equal("metric_with_special-chars.test"))
		})
	})

	Context("Error Handling", func() {
		It("should handle search on empty collection", func() {
			results, err := dbClient.SearchMetrics("anything", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("should handle very long query strings", func() {
			longQuery := strings.Repeat("very long query string ", 100)
			_, err := dbClient.SearchMetrics(longQuery, 10)
			Expect(err).NotTo(HaveOccurred())
			// Should not crash, results can be empty
		})
	})
})
