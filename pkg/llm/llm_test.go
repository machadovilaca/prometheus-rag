package llm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
	"github.com/machadovilaca/prometheus-rag/tests/mocks"
)

var _ = Describe("LLM", func() {
	const (
		baseURL = "http://127.0.0.1:1234/v1/"
		apiKey  = "test-api-key"
		model   = "gemma-3-27b-it"
	)

	var (
		llmClient llm.Client
		dbClient  vectordb.Client
		err       error
	)

	BeforeEach(func() {
		mockDB := mocks.NewVectorDBMock()
		mockDB.SearchMetricsFunc = func(query string, limit uint64) ([]*prometheus.MetricMetadata, error) {
			return []*prometheus.MetricMetadata{
				{Name: "up", Help: "Whether the instance is up", Type: "gauge", Labels: []string{"instance", "job"}},
				{Name: "kube_pod_status_phase", Help: "Pod status phase", Type: "gauge", Labels: []string{"pod", "namespace", "phase"}},
			}, nil
		}
		dbClient = mockDB

		llmClient, err = llm.New(llm.Config{
			BaseURL:        baseURL,
			APIKey:         apiKey,
			Model:          model,
			VectorDBClient: dbClient,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("New", func() {
		It("should fail with empty base URL", func() {
			llmClient, err = llm.New(llm.Config{
				BaseURL:        "",
				APIKey:         apiKey,
				Model:          model,
				VectorDBClient: dbClient,
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Run", func() {
		It("should successfully run query", func() {
			query := "Number of up pods"

			response, err := llmClient.Run(query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeEmpty())
		})

		It("should fail with invalid base URL", func() {
			llmClient, err = llm.New(llm.Config{
				BaseURL:        "http://127.0.0.1:9999/v1/",
				APIKey:         apiKey,
				Model:          model,
				VectorDBClient: dbClient,
			})
			Expect(err).NotTo(HaveOccurred())

			response, err := llmClient.Run("test query")
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeEmpty())
		})
	})
})
