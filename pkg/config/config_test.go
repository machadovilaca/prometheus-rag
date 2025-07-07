package config

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("Load", func() {
		Context("with default values", func() {
			It("should load configuration with correct defaults", func() {
				cfg, err := Load()
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Server.Host).To(Equal("0.0.0.0"))
				Expect(cfg.Server.Port).To(Equal("8080"))
				Expect(cfg.VectorDB.Provider).To(Equal("sqlite3"))
			})
		})

		Context("with environment variables", func() {
			var envVarsToCleanup []string

			BeforeEach(func() {
				envVarsToCleanup = []string{}
			})

			AfterEach(func() {
				for _, envVar := range envVarsToCleanup {
					_ = os.Unsetenv(envVar)
				}
			})

			setEnvVar := func(key, value string) {
				_ = os.Setenv(key, value)
				envVarsToCleanup = append(envVarsToCleanup, key)
			}

			It("should load configuration from environment variables", func() {
				setEnvVar("PRAG_DEBUG", "true")
				setEnvVar("PRAG_HOST", "localhost")
				setEnvVar("PRAG_PORT", "9000")
				setEnvVar("PRAG_VECTORDB_PROVIDER", "qdrant")

				cfg, err := Load()
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Debug).To(BeTrue())
				Expect(cfg.Server.Host).To(Equal("localhost"))
				Expect(cfg.Server.Port).To(Equal("9000"))
				Expect(cfg.VectorDB.Provider).To(Equal("qdrant"))
			})
		})
	})

	Describe("Validate", func() {
		It("should return error for invalid configuration", func() {
			cfg := &Config{
				Server: ServerConfig{
					Host: "", // Invalid: empty host
					Port: "8080",
				},
				Prometheus: PrometheusConfig{
					Address:            "http://localhost:9090",
					RefreshRateMinutes: 10,
				},
				VectorDB: VectorDBConfig{
					Provider:      "sqlite3",
					Collection:    "test",
					EncoderDir:    "./_models",
					Sqlite3DBPath: "./_data/test.db",
				},
				LLM: LLMConfig{
					BaseURL: "http://localhost:1234/v1/",
					Model:   "test-model",
				},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Utility Methods", func() {
		var cfg *Config

		BeforeEach(func() {
			cfg = &Config{
				Debug: true,
				Server: ServerConfig{
					Host: "localhost",
					Port: "9000",
				},
				VectorDB: VectorDBConfig{
					Provider: "QDRANT", // Test case insensitive
				},
			}
		})

		Context("IsDebugEnabled", func() {
			It("should return true when debug is enabled", func() {
				Expect(cfg.IsDebugEnabled()).To(BeTrue())
			})
		})

		Context("GetServerAddress", func() {
			It("should return correct server address", func() {
				Expect(cfg.GetServerAddress()).To(Equal("localhost:9000"))
			})
		})

		Context("Provider detection", func() {
			It("should detect qdrant provider case insensitively", func() {
				Expect(cfg.IsQdrantProvider()).To(BeTrue())
				Expect(cfg.IsSqlite3Provider()).To(BeFalse())
			})
		})
	})

	Describe("Adapters", func() {
		var cfg *Config

		BeforeEach(func() {
			cfg = &Config{
				Prometheus: PrometheusConfig{
					Address: "http://localhost:9090",
				},
				VectorDB: VectorDBConfig{
					Provider:      "sqlite3",
					Collection:    "test-collection",
					EncoderDir:    "./_models",
					Sqlite3DBPath: "./_data/test.db",
				},
				LLM: LLMConfig{
					BaseURL: "http://localhost:1234/v1/",
					APIKey:  "test-key",
					Model:   "test-model",
				},
			}
		})

		Context("ToPrometheusConfig", func() {
			It("should convert to prometheus config correctly", func() {
				promConfig := cfg.ToPrometheusConfig()
				Expect(promConfig.Address).To(Equal("http://localhost:9090"))
			})
		})

		Context("ToVectorDBConfig", func() {
			It("should convert to vectordb config correctly", func() {
				vectorConfig := cfg.ToVectorDBConfig()
				Expect(vectorConfig.Provider).To(Equal("sqlite3"))
				Expect(vectorConfig.CollectionName).To(Equal("test-collection"))
			})
		})

		Context("ToEmbeddingsConfig", func() {
			It("should convert to embeddings config correctly", func() {
				embeddingsConfig := cfg.ToEmbeddingsConfig()
				Expect(embeddingsConfig.ModelsDir).To(Equal("./_models"))
			})
		})
	})
})
