package config

import (
	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

// ToPrometheusConfig converts the application configuration to prometheus package configuration
func (c *Config) ToPrometheusConfig() prometheus.Config {
	return prometheus.Config{
		Address: c.Prometheus.Address,
	}
}

// ToVectorDBConfig converts the application configuration to vectordb package configuration
func (c *Config) ToVectorDBConfig() vectordb.Config {
	return vectordb.Config{
		Provider:               c.VectorDB.Provider,
		Sqlite3DBPath:          c.VectorDB.Sqlite3DBPath,
		QdrantHost:             c.VectorDB.QdrantHost,
		QdrantPort:             c.VectorDB.QdrantPort,
		CollectionName:         c.VectorDB.Collection,
		EncoderOutputDirectory: c.VectorDB.EncoderDir,
	}
}

// ToLLMConfig converts the application configuration to llm package configuration
func (c *Config) ToLLMConfig(vectorDBClient vectordb.Client) llm.Config {
	return llm.Config{
		BaseURL:        c.LLM.BaseURL,
		APIKey:         c.LLM.APIKey,
		Model:          c.LLM.Model,
		VectorDBClient: vectorDBClient,
	}
}

// ToEmbeddingsConfig converts the application configuration to embeddings package configuration
func (c *Config) ToEmbeddingsConfig() embeddings.Config {
	return embeddings.Config{
		ModelsDir: c.VectorDB.EncoderDir,
	}
}
