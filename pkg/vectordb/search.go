package vectordb

import (
	"context"
	"fmt"
	"strings"

	"github.com/qdrant/go-client/qdrant"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

func (v *vectorDB) SearchMetrics(query string, limit uint64) ([]*prometheus.MetricMetadata, error) {
	encodedQuery, err := v.encoder.EncodeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	searchResults, err := v.client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(encodedQuery),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayloadEnable(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search metrics: %w", err)
	}

	return convertSearchResults(searchResults), nil
}

func convertSearchResults(results []*qdrant.ScoredPoint) []*prometheus.MetricMetadata {
	var metrics []*prometheus.MetricMetadata
	for _, result := range results {
		metrics = append(metrics, fromQdrantMap(result.Payload))
	}
	return metrics
}

func fromQdrantMap(m map[string]*qdrant.Value) *prometheus.MetricMetadata {
	return &prometheus.MetricMetadata{
		Name:   m["name"].GetStringValue(),
		Help:   m["help"].GetStringValue(),
		Type:   m["type"].GetStringValue(),
		Unit:   m["unit"].GetStringValue(),
		Labels: strings.Split(m["labels"].GetStringValue(), ", "),
	}
}
