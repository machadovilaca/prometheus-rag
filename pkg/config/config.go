// Package config provides centralized configuration management for the prometheus-rag application.
// It supports loading configuration from environment variables with sensible defaults.
package config

import (
	"fmt"
	"strings"

	"go-simpler.org/env"
)

// Config represents the complete application configuration
type Config struct {
	// Debug enables debug logging
	Debug bool `env:"PRAG_DEBUG" default:"false"`

	// Server configuration
	Server ServerConfig

	// Prometheus configuration
	Prometheus PrometheusConfig

	// VectorDB configuration
	VectorDB VectorDBConfig

	// LLM configuration
	LLM LLMConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host string `env:"PRAG_HOST" default:"0.0.0.0"`
	Port string `env:"PRAG_PORT" default:"8080"`
}

// PrometheusConfig holds Prometheus-specific configuration
type PrometheusConfig struct {
	Address            string `env:"PRAG_PROMETHEUS_ADDRESS" default:"http://localhost:9090"`
	RefreshRateMinutes int    `env:"PRAG_PROMETHEUS_REFRESH_RATE_MINUTES" default:"10"`
}

// VectorDBConfig holds vector database configuration
type VectorDBConfig struct {
	Provider   string `env:"PRAG_VECTORDB_PROVIDER" default:"sqlite3"`
	Collection string `env:"PRAG_VECTORDB_COLLECTION" default:"prag-metrics"`
	EncoderDir string `env:"PRAG_VECTORDB_ENCODER_DIR" default:"./_models"`

	// SQLite3 specific
	Sqlite3DBPath string `env:"PRAG_VECTORDB_SQLITE3_DB_PATH" default:"./_data/metrics.db"`

	// Qdrant specific
	QdrantHost string `env:"PRAG_VECTORDB_QDRANT_HOST" default:"localhost"`
	QdrantPort int    `env:"PRAG_VECTORDB_QDRANT_PORT" default:"6334"`
}

// LLMConfig holds LLM-specific configuration
type LLMConfig struct {
	BaseURL string `env:"PRAG_LLM_BASE_URL" default:"http://localhost:1234/v1/"`
	APIKey  string `env:"PRAG_LLM_API_KEY"`
	Model   string `env:"PRAG_LLM_MODEL" default:"granite-3.1-8b-instruct"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := env.Load(&cfg, nil); err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadWithDefaults loads configuration with custom defaults
func LoadWithDefaults(defaults *Config) (*Config, error) {
	var cfg Config
	if defaults != nil {
		cfg = *defaults
	}

	if err := env.Load(&cfg, nil); err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a file (for testing)
func LoadFromFile(filepath string) (*Config, error) {
	// This could be extended to support JSON/YAML config files
	// For now, we just use the standard env loading
	return Load()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if c.Prometheus.Address == "" {
		return fmt.Errorf("prometheus address cannot be empty")
	}

	if c.Prometheus.RefreshRateMinutes <= 0 {
		return fmt.Errorf("prometheus refresh rate must be greater than 0")
	}

	if c.VectorDB.Provider == "" {
		return fmt.Errorf("vectordb provider cannot be empty")
	}

	if err := ValidateProvider(c.VectorDB.Provider); err != nil {
		return err
	}

	if c.VectorDB.Collection == "" {
		return fmt.Errorf("vectordb collection cannot be empty")
	}

	if c.VectorDB.EncoderDir == "" {
		return fmt.Errorf("vectordb encoder directory cannot be empty")
	}

	// Validate provider-specific configurations
	switch c.VectorDB.Provider {
	case "sqlite3":
		if c.VectorDB.Sqlite3DBPath == "" {
			return fmt.Errorf("sqlite3 db path cannot be empty when using sqlite3 provider")
		}
	case "qdrant":
		if c.VectorDB.QdrantHost == "" {
			return fmt.Errorf("qdrant host cannot be empty when using qdrant provider")
		}
		if c.VectorDB.QdrantPort <= 0 {
			return fmt.Errorf("qdrant port must be greater than 0")
		}
	default:
		return fmt.Errorf("unsupported vectordb provider: %s", c.VectorDB.Provider)
	}

	if c.LLM.BaseURL == "" {
		return fmt.Errorf("llm base URL cannot be empty")
	}

	if c.LLM.Model == "" {
		return fmt.Errorf("llm model cannot be empty")
	}

	return nil
}

// ValidateProvider validates if the vectordb provider is supported
func ValidateProvider(provider string) error {
	supportedProviders := []string{"sqlite3", "qdrant"}
	provider = strings.ToLower(provider)

	for _, supported := range supportedProviders {
		if provider == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported vectordb provider '%s', supported providers: %v", provider, supportedProviders)
}

// GetServerAddress returns the server address in host:port format
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsDebugEnabled returns whether debug mode is enabled
func (c *Config) IsDebugEnabled() bool {
	return c.Debug
}

// GetVectorDBProvider returns the vector database provider
func (c *Config) GetVectorDBProvider() string {
	return strings.ToLower(c.VectorDB.Provider)
}

// IsQdrantProvider returns true if using Qdrant as vector database
func (c *Config) IsQdrantProvider() bool {
	return c.GetVectorDBProvider() == "qdrant"
}

// IsSqlite3Provider returns true if using SQLite3 as vector database
func (c *Config) IsSqlite3Provider() bool {
	return c.GetVectorDBProvider() == "sqlite3"
}
