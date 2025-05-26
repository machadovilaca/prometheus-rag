package vectordb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb/qdrantdb"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb/sqlite3"
)

// Client interface for interacting with the VectorDB
type Client interface {
	// CreateCollection creates the collection in the vector database
	CreateCollection() error

	// DeleteCollection deletes the collection from the vector database
	DeleteCollection() error

	// AddMetricMetadata adds a metric metadata entry to the vector database
	AddMetricMetadata(metadata *prometheus.MetricMetadata) error

	// BatchAddMetricMetadata adds a batch of metric metadata entries to the vector database
	BatchAddMetricMetadata(metadata []*prometheus.MetricMetadata) error

	// SearchMetrics searches for relevant metrics based on a natural language query
	// Returns a list of metric metadata entries sorted by relevance
	SearchMetrics(query string, limit uint64) ([]*prometheus.MetricMetadata, error)

	// Close closes the connection to the vector database
	Close() error
}

// Config represents the configuration for the vector database
type Config struct {
	Provider string

	Sqlite3DBPath string

	QdrantHost string
	QdrantPort int

	CollectionName         string
	EncoderOutputDirectory string
}

// ErrUnsupportedProvider is returned when an unsupported provider is specified
var ErrUnsupportedProvider = errors.New("unsupported provider")

func New(cfg Config) (Client, error) {
	log.Info().Msgf("creating encoder with output directory %s", cfg.EncoderOutputDirectory)
	encoder, err := embeddings.NewEncoder(embeddings.Config{
		ModelsDir: cfg.EncoderOutputDirectory,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	switch strings.ToLower(cfg.Provider) {
	case "qdrant":
		log.Info().Msg("starting Qdrant client")
		return qdrantdb.New(qdrantdb.Config{
			QdrantHost:     cfg.QdrantHost,
			QdrantPort:     cfg.QdrantPort,
			CollectionName: cfg.CollectionName,
			Encoder:        encoder,
		})
	case "sqlite3":
		log.Info().Msg("starting SQLite3 client")
		return sqlite3.New(sqlite3.Config{
			DBPath:         cfg.Sqlite3DBPath,
			CollectionName: cfg.CollectionName,
			Encoder:        encoder,
		})
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, cfg.Provider)
	}
}
