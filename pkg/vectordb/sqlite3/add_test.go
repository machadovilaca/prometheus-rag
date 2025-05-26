package sqlite3_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

var _ = Describe("Add", func() {
	var (
		dbClient vectordb.Client
		tempDir  string
		dbPath   string
	)

	BeforeEach(func() {
		var err error
		// Create temporary directory for test database
		tempDir, err = os.MkdirTemp("", "sqlite3_add_test_*")
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

	It("should handle metrics with complex labels", func() {
		metadata := &prometheus.MetricMetadata{
			Name:   "complex_metric",
			Help:   "Metric with many labels",
			Type:   "histogram",
			Labels: []string{"namespace", "pod", "container", "method", "status_code"},
		}

		err := dbClient.AddMetricMetadata(metadata)
		Expect(err).NotTo(HaveOccurred())

		results, err := dbClient.SearchMetrics("complex", 10)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Labels).To(Equal([]string{"namespace", "pod", "container", "method", "status_code"}))
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
})
