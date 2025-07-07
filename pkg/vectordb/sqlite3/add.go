package sqlite3

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
)

func (v *sqlite3DB) AddMetricMetadata(metadata *prometheus.MetricMetadata) error {
	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid metric metadata: %w", err)
	}

	// Encode the metric metadata to a vector
	embedding, err := v.encoder.EncodeMetricMetadata(*metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metric metadata: %w", err)
	}

	// Convert embedding to bytes
	embeddingBytes, err := v.encodeEmbedding(embedding)
	if err != nil {
		return fmt.Errorf("failed to encode embedding: %w", err)
	}

	// Create deterministic ID based on metric name
	id := v.createDeterministicID(metadata.Name)

	// Use secure identifier escaping for table name
	safeTableName, err := v.validator.SafeIdentifier(v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to validate collection name: %w", err)
	}

	// Insert or replace the metric metadata
	insertSQL := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (id, name, help, type, labels, embedding)
		VALUES (?, ?, ?, ?, ?, ?)
	`, safeTableName)

	_, err = v.db.Exec(insertSQL, id, metadata.Name, metadata.Help, metadata.Type,
		v.joinLabels(metadata.Labels), embeddingBytes)
	if err != nil {
		return fmt.Errorf("failed to insert metric metadata: %w", err)
	}

	return nil
}

func (v *sqlite3DB) BatchAddMetricMetadata(metadataArray []*prometheus.MetricMetadata) error {
	if len(metadataArray) == 0 {
		log.Info().Msg("skipping batch add of metric metadata because there are none")
		return nil
	}

	// Begin transaction for better performance
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Use secure identifier escaping for table name
	safeTableName, err := v.validator.SafeIdentifier(v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to validate collection name: %w", err)
	}

	// Prepare statement
	insertSQL := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (id, name, help, type, labels, embedding)
		VALUES (?, ?, ?, ?, ?, ?)
	`, safeTableName)

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, metadata := range metadataArray {
		if err := metadata.Validate(); err != nil {
			return fmt.Errorf("invalid metric metadata '%s': %w", metadata.Name, err)
		}

		// Encode the metric metadata to a vector
		embedding, err := v.encoder.EncodeMetricMetadata(*metadata)
		if err != nil {
			return fmt.Errorf("failed to encode metric metadata '%s': %w", metadata.Name, err)
		}

		// Convert embedding to bytes
		embeddingBytes, err := v.encodeEmbedding(embedding)
		if err != nil {
			return fmt.Errorf("failed to encode embedding for '%s': %w", metadata.Name, err)
		}

		// Create deterministic ID based on metric name
		id := v.createDeterministicID(metadata.Name)

		// Execute statement
		_, err = stmt.Exec(id, metadata.Name, metadata.Help, metadata.Type,
			v.joinLabels(metadata.Labels), embeddingBytes)
		if err != nil {
			return fmt.Errorf("failed to insert metric metadata '%s': %w", metadata.Name, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Msgf("batch added %d metric metadata entries", len(metadataArray))
	return nil
}
