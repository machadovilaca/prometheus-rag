package qdrantdb

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
)

// Config holds the configuration for the Qdrant client
type Config struct {
	QdrantHost     string
	QdrantPort     int
	CollectionName string
	Encoder        embeddings.Encoder
}

type qdrantDB struct {
	client  *qdrant.Client
	encoder embeddings.Encoder

	collectionName string
}

// New creates a new Qdrant client connection
func New(cfg Config) (*qdrantDB, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.QdrantHost,
		Port: cfg.QdrantPort,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	if cfg.CollectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	v := &qdrantDB{client: client, encoder: cfg.Encoder, collectionName: cfg.CollectionName}

	if err := v.CreateCollection(); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return v, nil
}

func (v *qdrantDB) CreateCollection() error {
	log.Info().Msgf("creating collection %s", v.collectionName)

	exists, err := v.client.CollectionExists(context.Background(), v.collectionName)
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
		CollectionName: v.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(encodingDimension),
			Distance: qdrant.Distance_Cosine,
		}),
	}); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

func (v *qdrantDB) DeleteCollection() error {
	return v.client.DeleteCollection(context.Background(), v.collectionName)
}

func (v *qdrantDB) Close() error {
	return v.client.Close()
}
