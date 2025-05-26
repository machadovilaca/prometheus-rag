package config

import (
	"time"

	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

// RAGConfig represents configuration specific to the RAG functionality
type RAGConfig struct {
	PrometheusAddress            string
	PrometheusRefreshRateMinutes int
	VectorDBConfig               vectordb.Config
	LLMConfig                    llm.Config
}

// ToRAGConfig converts the application configuration to RAG-specific configuration
func (c *Config) ToRAGConfig(vectorDBClient vectordb.Client) RAGConfig {
	return RAGConfig{
		PrometheusAddress:            c.Prometheus.Address,
		PrometheusRefreshRateMinutes: c.Prometheus.RefreshRateMinutes,
		VectorDBConfig:               c.ToVectorDBConfig(),
		LLMConfig:                    c.ToLLMConfig(vectorDBClient),
	}
}

// GetPrometheusRefreshInterval returns the prometheus refresh interval as time.Duration
func (r *RAGConfig) GetPrometheusRefreshInterval() time.Duration {
	return time.Duration(r.PrometheusRefreshRateMinutes) * time.Minute
}
