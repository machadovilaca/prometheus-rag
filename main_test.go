package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machadovilaca/prometheus-rag/pkg/config"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = Describe("Configuration Integration", func() {
	Describe("ConfigurationIntegration", func() {
		var (
			envVarsToCleanup []string
		)

		BeforeEach(func() {
			envVarsToCleanup = []string{}
		})

		AfterEach(func() {
			// Clean up environment variables
			for _, envVar := range envVarsToCleanup {
				_ = os.Unsetenv(envVar)
			}
		})

		setEnvVar := func(key, value string) {
			_ = os.Setenv(key, value)
			envVarsToCleanup = append(envVarsToCleanup, key)
		}

		It("should load configuration correctly with environment variables", func() {
			// Set some test environment variables
			setEnvVar("PRAG_DEBUG", "true")
			setEnvVar("PRAG_HOST", "testhost")
			setEnvVar("PRAG_PORT", "9999")
			setEnvVar("PRAG_VECTORDB_PROVIDER", "sqlite3")
			setEnvVar("PRAG_VECTORDB_SQLITE3_DB_PATH", "/tmp/test.db")

			// Load configuration
			cfg, err := config.Load()
			Expect(err).NotTo(HaveOccurred())

			// Verify configuration is loaded correctly
			Expect(cfg.Debug).To(BeTrue())
			Expect(cfg.Server.Host).To(Equal("testhost"))
			Expect(cfg.Server.Port).To(Equal("9999"))

			// Test configuration adapters
			prometheusConfig := cfg.ToPrometheusConfig()
			Expect(prometheusConfig.Address).To(Equal(cfg.Prometheus.Address))

			vectordbConfig := cfg.ToVectorDBConfig()
			Expect(vectordbConfig.Provider).To(Equal(cfg.VectorDB.Provider))
			Expect(vectordbConfig.Sqlite3DBPath).To(Equal("/tmp/test.db"))

			embeddingsConfig := cfg.ToEmbeddingsConfig()
			Expect(embeddingsConfig.ModelsDir).To(Equal(cfg.VectorDB.EncoderDir))

			// Test utility methods
			Expect(cfg.GetServerAddress()).To(Equal("testhost:9999"))
			Expect(cfg.IsDebugEnabled()).To(BeTrue())
			Expect(cfg.IsSqlite3Provider()).To(BeTrue())
			Expect(cfg.IsQdrantProvider()).To(BeFalse())
		})
	})

	Describe("ConfigurationDefaults", func() {
		var (
			originalEnvVars map[string]string
		)

		BeforeEach(func() {
			// Clear any existing environment variables that might affect defaults
			envVars := []string{
				"PRAG_DEBUG", "PRAG_HOST", "PRAG_PORT",
				"PRAG_PROMETHEUS_ADDRESS", "PRAG_PROMETHEUS_REFRESH_RATE_MINUTES",
				"PRAG_VECTORDB_PROVIDER", "PRAG_VECTORDB_COLLECTION", "PRAG_VECTORDB_ENCODER_DIR",
				"PRAG_VECTORDB_SQLITE3_DB_PATH", "PRAG_VECTORDB_QDRANT_HOST", "PRAG_VECTORDB_QDRANT_PORT",
				"PRAG_LLM_BASE_URL", "PRAG_LLM_API_KEY", "PRAG_LLM_MODEL",
			}

			// Store original values
			originalEnvVars = make(map[string]string)
			for _, envVar := range envVars {
				originalEnvVars[envVar] = os.Getenv(envVar)
				_ = os.Unsetenv(envVar)
			}
		})

		AfterEach(func() {
			// Restore original values after test
			for envVar, val := range originalEnvVars {
				if val != "" {
					_ = os.Setenv(envVar, val)
				}
			}
		})

		It("should load configuration with correct default values", func() {
			cfg, err := config.Load()
			Expect(err).NotTo(HaveOccurred())

			// Check default values
			Expect(cfg.Debug).To(BeFalse())
			Expect(cfg.Server.Host).To(Equal("0.0.0.0"))
			Expect(cfg.Server.Port).To(Equal("8080"))
			Expect(cfg.VectorDB.Provider).To(Equal("sqlite3"))
			Expect(cfg.LLM.Model).To(Equal("granite-3.1-8b-instruct"))
		})
	})
})
