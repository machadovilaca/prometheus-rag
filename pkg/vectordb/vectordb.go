package vectordb

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

const (
	collectionName = "metrics"
)

// VectorDB represents the interface for interacting with the vector database
type VectorDB interface {
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
	Host string
	Port int

	EncoderOutputDirectory string
}

type vectorDB struct {
	client  *qdrant.Client
	encoder embeddings.Encoder
}

// New creates a new Qdrant client connection
func New(cfg Config) (VectorDB, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	encoder, err := embeddings.NewEncoder(embeddings.Config{
		ModelsDir: cfg.EncoderOutputDirectory,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	v := &vectorDB{client: client, encoder: encoder}

	if err := v.CreateCollection(); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return v, nil
}

func (v *vectorDB) CreateCollection() error {
	exists, err := v.client.CollectionExists(context.Background(), collectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %w", err)
	}

	if exists {
		return nil
	}

	encodingDimension, err := v.encoder.GetDimension()
	if err != nil {
		return fmt.Errorf("failed to get encoding dimension: %w", err)
	}

	if err = v.client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(encodingDimension),
			Distance: qdrant.Distance_Cosine,
		}),
	}); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

func (v *vectorDB) DeleteCollection() error {
	return v.client.DeleteCollection(context.Background(), collectionName)
}

func (v *vectorDB) Close() error {
	return v.client.Close()
}
