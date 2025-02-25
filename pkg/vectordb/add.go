package vectordb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

func (v *vectorDB) AddMetricMetadata(metadata *prometheus.MetricMetadata) error {
	pointStruct, err := v.newPointStruct(metadata)
	if err != nil {
		return fmt.Errorf("failed to create point struct: %w", err)
	}

	_, err = v.client.Upsert(
		context.Background(),
		&qdrant.UpsertPoints{
			CollectionName: collectionName,
			Points:         []*qdrant.PointStruct{pointStruct},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upsert metric metadata: %w", err)
	}

	return nil
}

func (v *vectorDB) BatchAddMetricMetadata(metadata []*prometheus.MetricMetadata) error {
	points := make([]*qdrant.PointStruct, len(metadata))

	for i, m := range metadata {
		pointStruct, err := v.newPointStruct(m)
		if err != nil {
			return fmt.Errorf("failed to create point struct: %w", err)
		}

		points[i] = pointStruct
	}

	_, err := v.client.Upsert(
		context.Background(),
		&qdrant.UpsertPoints{
			CollectionName: collectionName,
			Points:         points,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upsert metric metadata: %w", err)
	}

	return nil
}

func (v *vectorDB) newPointStruct(metadata *prometheus.MetricMetadata) (*qdrant.PointStruct, error) {
	encodedMetadata, err := v.encoder.EncodeMetricMetadata(*metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to encode metric metadata: %w", err)
	}

	deterministicUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(metadata.Name))

	return &qdrant.PointStruct{
		Id:      qdrant.NewID(deterministicUUID.String()),
		Vectors: qdrant.NewVectorsDense(encodedMetadata),
		Payload: qdrant.NewValueMap(metadata.ToMap()),
	}, nil
}
