package sqlite3

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
)

// Config holds the configuration for the SQLite3 client
type Config struct {
	DBPath         string
	CollectionName string
	Encoder        embeddings.Encoder
}

type sqlite3DB struct {
	db             *sql.DB
	collectionName string
	encoder        embeddings.Encoder
	validator      *SQLIdentifierValidator
}

// New creates a new SQLite3 vector database client
func New(cfg Config) (*sqlite3DB, error) {
	sqlite_vec.Auto()

	// Initialize the SQL identifier validator
	validator := NewSQLIdentifierValidator()

	// Validate collection name for security
	if err := validator.ValidateIdentifier(cfg.CollectionName); err != nil {
		return nil, fmt.Errorf("invalid collection name: %w", err)
	}

	// Ensure the directory for the database file exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	log.Info().Msgf("opening sqlite3 db at %s", cfg.DBPath)
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite3 db: %w", err)
	}

	v := &sqlite3DB{
		db:             db,
		encoder:        cfg.Encoder,
		collectionName: cfg.CollectionName,
		validator:      validator,
	}

	if err := v.CreateCollection(); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return v, nil
}

func (v *sqlite3DB) CreateCollection() error {
	// Use secure identifier escaping for table name
	safeTableName, err := v.validator.SafeIdentifier(v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to validate collection name: %w", err)
	}

	// Create the table for storing metric metadata with vector embeddings
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			help TEXT,
			type TEXT,
			labels TEXT,
			embedding BLOB
		)
	`, safeTableName)

	_, err = v.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create collection table: %w", err)
	}

	// Create index on name for faster lookups
	safeIndexName, err := v.validator.SafeIdentifier("idx_" + v.collectionName + "_name")
	if err != nil {
		return fmt.Errorf("failed to validate index name: %w", err)
	}

	createIndexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s ON %s(name)
	`, safeIndexName, safeTableName)

	_, err = v.db.Exec(createIndexSQL)
	if err != nil {
		return fmt.Errorf("failed to create name index: %w", err)
	}

	log.Info().Msgf("created collection table: %s", v.collectionName)
	return nil
}

func (v *sqlite3DB) DeleteCollection() error {
	// Use secure identifier escaping for table name
	safeTableName, err := v.validator.SafeIdentifier(v.collectionName)
	if err != nil {
		return fmt.Errorf("failed to validate collection name: %w", err)
	}

	dropTableSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, safeTableName)

	_, err = v.db.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to delete collection table: %w", err)
	}

	log.Info().Msgf("deleted collection table: %s", v.collectionName)
	return nil
}

func (v *sqlite3DB) Close() error {
	return v.db.Close()
}

// Helper methods

func (v *sqlite3DB) createDeterministicID(name string) string {
	hash := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", hash[:16]) // Use first 16 bytes for shorter ID
}

func (v *sqlite3DB) joinLabels(labels []string) string {
	return strings.Join(labels, ", ")
}

func (v *sqlite3DB) splitLabels(labels string) []string {
	if labels == "" {
		return []string{}
	}
	return strings.Split(labels, ", ")
}

func (v *sqlite3DB) encodeEmbedding(embedding []float32) ([]byte, error) {
	buf := make([]byte, len(embedding)*4)
	for i, val := range embedding {
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(val))
	}
	return buf, nil
}

func (v *sqlite3DB) decodeEmbedding(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding data length")
	}

	embedding := make([]float32, len(data)/4)
	for i := 0; i < len(embedding); i++ {
		bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		embedding[i] = math.Float32frombits(bits)
	}
	return embedding, nil
}

func (v *sqlite3DB) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
