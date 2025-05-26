package sqlite3

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

func (v *sqlite3DB) SearchMetrics(query string, limit uint64) ([]*prometheus.MetricMetadata, error) {
	// Encode the query to a vector
	queryEmbedding, err := v.encoder.EncodeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	// Use secure identifier escaping for table name
	safeTableName, err := v.validator.SafeIdentifier(v.collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to validate collection name: %w", err)
	}

	// Query for similar metrics using cosine similarity
	// We'll calculate similarity in Go since sqlite-vec might need setup
	searchSQL := fmt.Sprintf(`
		SELECT id, name, help, type, labels, embedding
		FROM %s
		ORDER BY name
	`, safeTableName)

	rows, err := v.db.Query(searchSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	type metricWithScore struct {
		metadata *prometheus.MetricMetadata
		score    float64
	}

	var candidates []metricWithScore

	for rows.Next() {
		var id, name, help, metricType, labels string
		var embeddingBytes []byte

		err := rows.Scan(&id, &name, &help, &metricType, &labels, &embeddingBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Decode embedding
		embedding, err := v.decodeEmbedding(embeddingBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to decode embedding, skipping")
			continue
		}

		// Calculate cosine similarity
		similarity := v.cosineSimilarity(queryEmbedding, embedding)

		metadata := &prometheus.MetricMetadata{
			Name:   name,
			Help:   help,
			Type:   metricType,
			Labels: v.splitLabels(labels),
		}

		candidates = append(candidates, metricWithScore{
			metadata: metadata,
			score:    similarity,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Sort by similarity score (descending)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].score < candidates[j].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Apply limit
	maxResults := int(limit)
	if len(candidates) < maxResults {
		maxResults = len(candidates)
	}

	results := make([]*prometheus.MetricMetadata, maxResults)
	for i := 0; i < maxResults; i++ {
		results[i] = candidates[i].metadata
	}

	return results, nil
}
