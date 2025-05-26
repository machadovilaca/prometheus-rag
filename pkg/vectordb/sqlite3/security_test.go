package sqlite3

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/machadovilaca/prometheus-rag/pkg/embeddings"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSecurityValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SQLite3 Security Suite")
}

var _ = Describe("SQLite3 Security Validation", func() {
	var (
		tempDir                  string
		encoder                  embeddings.Encoder
		maliciousCollectionNames []string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "sqlite3_security_test")
		Expect(err).NotTo(HaveOccurred())

		// Create a mock encoder for testing
		encoder = &mockEncoder{}

		// Common SQL injection payloads for testing
		maliciousCollectionNames = []string{
			// SQL injection attempts
			"'; DROP TABLE users; --",
			"collection\"; DROP TABLE metrics; --",
			"test'; INSERT INTO admin VALUES('hacker','password'); --",
			"collection\\\"; PRAGMA foreign_keys=OFF; --",
			// Invalid characters
			"collection with spaces",
			"collection@#$%",
			"collection<script>alert('xss')</script>",
			// Path traversal attempts
			"../../../etc/passwd",
			"..\\..\\windows\\system32",
			// Null byte injection
			"collection\x00malicious",
			// Unicode/encoding attacks
			"collection\u0000",
			"collection\uffff",
			// Empty/special cases
			"",
			".",
			"..",
			// SQL keywords
			"SELECT",
			"DROP",
			"DELETE",
			"INSERT",
			"UPDATE",
			"CREATE",
			"ALTER",
			"TABLE",
			// Too long identifier
			"this_is_a_very_long_collection_name_that_exceeds_the_maximum_allowed_length_for_sql_identifiers_and_should_be_rejected_by_the_validator",
		}
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Collection Name Validation", func() {
		It("should reject malicious collection names", func() {
			for _, maliciousName := range maliciousCollectionNames {
				dbPath := filepath.Join(tempDir, "test.db")
				cfg := Config{
					DBPath:         dbPath,
					CollectionName: maliciousName,
					Encoder:        encoder,
				}

				_, err := New(cfg)
				Expect(err).To(HaveOccurred(), "Expected error for malicious collection name: %s", maliciousName)
				Expect(err.Error()).To(ContainSubstring("invalid collection name"),
					"Error should indicate invalid collection name for: %s", maliciousName)
			}
		})

		It("should accept valid collection names", func() {
			validNames := []string{
				"metrics",
				"test_collection",
				"collection123",
				"my_test_collection_v2",
				"prometheus_metrics",
				"_private_collection",
			}

			for _, validName := range validNames {
				dbPath := filepath.Join(tempDir, "test_"+validName+".db")
				cfg := Config{
					DBPath:         dbPath,
					CollectionName: validName,
					Encoder:        encoder,
				}

				db, err := New(cfg)
				Expect(err).NotTo(HaveOccurred(), "Valid collection name should be accepted: %s", validName)

				// Clean up
				if db != nil {
					db.Close()
				}
			}
		})
	})

	Describe("SQL Identifier Validator", func() {
		var validator *SQLIdentifierValidator

		BeforeEach(func() {
			validator = NewSQLIdentifierValidator()
		})

		It("should validate safe identifiers", func() {
			safeIdentifiers := []string{
				"metrics",
				"test_table",
				"table123",
				"_private_table",
				"prometheus_metrics_v2",
			}

			for _, identifier := range safeIdentifiers {
				err := validator.ValidateIdentifier(identifier)
				Expect(err).NotTo(HaveOccurred(), "Safe identifier should be valid: %s", identifier)
			}
		})

		It("should reject unsafe identifiers", func() {
			unsafeIdentifiers := []string{
				"'; DROP TABLE test; --",
				"table with spaces",
				"table@special",
				"123_starts_with_number",
				"",
				"SELECT",
				"DROP",
			}

			for _, identifier := range unsafeIdentifiers {
				err := validator.ValidateIdentifier(identifier)
				Expect(err).To(HaveOccurred(), "Unsafe identifier should be rejected: %s", identifier)
			}
		})

		It("should properly escape valid identifiers", func() {
			testCases := map[string]string{
				"metrics":          `"metrics"`,
				"test_table":       `"test_table"`,
				"table_with_quote": `"table_with_quote"`,
			}

			for input, expected := range testCases {
				escaped := validator.EscapeIdentifier(input)
				Expect(escaped).To(Equal(expected), "Identifier should be properly escaped")
			}
		})

		It("should escape identifiers containing quotes", func() {
			identifier := `table"with"quotes`
			expected := `"table""with""quotes"`
			escaped := validator.EscapeIdentifier(identifier)
			Expect(escaped).To(Equal(expected), "Quotes should be escaped by doubling")
		})
	})
})

// mockEncoder implements embeddings.Encoder for testing
type mockEncoder struct{}

func (m *mockEncoder) GetDimension() (int, error) {
	return 5, nil
}

func (m *mockEncoder) EncodeQuery(query string) ([]float32, error) {
	// Return a dummy embedding for testing
	return []float32{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

func (m *mockEncoder) EncodeMetricMetadata(metadata prometheus.MetricMetadata) ([]float32, error) {
	// Return a dummy embedding for testing
	return []float32{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}
