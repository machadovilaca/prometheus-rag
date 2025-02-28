package llm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

var _ = Describe("Prompt", func() {
	Context("buildPrompt", func() {
		It("should build prompt with metrics", func() {
			metrics := []*prometheus.MetricMetadata{
				{
					Name:   "http_requests_total",
					Help:   "Total number of HTTP requests",
					Type:   "counter",
					Labels: []string{"method", "status"},
				},
				{
					Name:   "node_memory_usage_bytes",
					Help:   "Current memory usage in bytes",
					Type:   "gauge",
					Labels: []string{"type"},
				},
			}

			prompt, err := llm.BuildPrompt(metrics)
			Expect(err).NotTo(HaveOccurred())
			Expect(prompt).To(ContainSubstring("http_requests_total"))
			Expect(prompt).To(ContainSubstring("Total number of HTTP requests"))
			Expect(prompt).To(ContainSubstring("node_memory_usage_bytes"))
			Expect(prompt).To(ContainSubstring("Current memory usage in bytes"))
		})

		It("should build prompt with empty metrics", func() {
			metrics := []*prometheus.MetricMetadata{}

			prompt, err := llm.BuildPrompt(metrics)
			Expect(err).NotTo(HaveOccurred())
			Expect(prompt).NotTo(BeEmpty())
		})

		It("should build prompt with nil metrics", func() {
			var metrics []*prometheus.MetricMetadata

			prompt, err := llm.BuildPrompt(metrics)
			Expect(err).NotTo(HaveOccurred())
			Expect(prompt).NotTo(BeEmpty())
		})
	})
})
