package config

import (
	"fmt"
	"net/url"
	"strconv"
)

// ConfigHelper provides utility methods for working with configuration
type ConfigHelper struct {
	cfg *Config
}

// NewHelper creates a new configuration helper
func NewHelper(cfg *Config) *ConfigHelper {
	return &ConfigHelper{cfg: cfg}
}

// GetPrometheusURL returns a parsed Prometheus URL
func (h *ConfigHelper) GetPrometheusURL() (*url.URL, error) {
	return url.Parse(h.cfg.Prometheus.Address)
}

// GetLLMURL returns a parsed LLM URL
func (h *ConfigHelper) GetLLMURL() (*url.URL, error) {
	return url.Parse(h.cfg.LLM.BaseURL)
}

// GetQdrantAddress returns the Qdrant address in host:port format
func (h *ConfigHelper) GetQdrantAddress() string {
	return fmt.Sprintf("%s:%d", h.cfg.VectorDB.QdrantHost, h.cfg.VectorDB.QdrantPort)
}

// GetServerPortInt returns the server port as an integer
func (h *ConfigHelper) GetServerPortInt() (int, error) {
	return strconv.Atoi(h.cfg.Server.Port)
}

// IsLocalPrometheus checks if Prometheus is running locally
func (h *ConfigHelper) IsLocalPrometheus() bool {
	prometheusURL, err := h.GetPrometheusURL()
	if err != nil {
		return false
	}
	return prometheusURL.Hostname() == "localhost" || prometheusURL.Hostname() == "127.0.0.1"
}

// IsLocalLLM checks if LLM is running locally
func (h *ConfigHelper) IsLocalLLM() bool {
	llmURL, err := h.GetLLMURL()
	if err != nil {
		return false
	}
	return llmURL.Hostname() == "localhost" || llmURL.Hostname() == "127.0.0.1"
}

// GetConfigSummary returns a summary of the configuration (without sensitive data)
func (h *ConfigHelper) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"debug":               h.cfg.Debug,
		"server_address":      h.cfg.GetServerAddress(),
		"prometheus_address":  h.cfg.Prometheus.Address,
		"prometheus_refresh":  h.cfg.Prometheus.RefreshRateMinutes,
		"vectordb_provider":   h.cfg.VectorDB.Provider,
		"vectordb_collection": h.cfg.VectorDB.Collection,
		"llm_model":           h.cfg.LLM.Model,
		"llm_has_api_key":     h.cfg.LLM.APIKey != "",
	}
}

// ValidateURLs validates that all URLs in the configuration are valid
func (h *ConfigHelper) ValidateURLs() error {
	if _, err := h.GetPrometheusURL(); err != nil {
		return fmt.Errorf("invalid prometheus URL: %w", err)
	}

	if _, err := h.GetLLMURL(); err != nil {
		return fmt.Errorf("invalid LLM URL: %w", err)
	}

	return nil
}
