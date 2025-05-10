package rag

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go-simpler.org/env"

	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

// Client is the main client for the RAG
type Client struct {
	cfg config

	vectorDBClient   vectordb.Client
	prometheusClient prometheus.Client
	llmClient        llm.Client
	metricsMetadata  []*prometheus.MetricMetadata
}

type config struct {
	PrometheusAddress            string `env:"PRAG_PROMETHEUS_ADDRESS" default:"http://localhost:9090"`
	PrometheusRefreshRateMinutes int    `env:"PRAG_PROMETHEUS_REFRESH_RATE_MINUTES" default:"10"`

	VectorDBProvider   string `env:"PRAG_VECTORDB_PROVIDER" default:"sqlite3"`
	Sqlite3DBPath      string `env:"PRAG_VECTORDB_SQLITE3_DB_PATH" default:"./_data/metrics.db"`
	VectorDBQdrantHost string `env:"PRAG_VECTORDB_QDRANT_HOST" default:"localhost"`
	VectorDBQdrantPort int    `env:"PRAG_VECTORDB_QDRANT_PORT" default:"6334"`
	VectorDBCollection string `env:"PRAG_VECTORDB_COLLECTION" default:"prag-metrics"`
	VectorDBEncoderDir string `env:"PRAG_VECTORDB_ENCODER_DIR" default:"./_models"`

	LLMBaseURL string `env:"PRAG_LLM_BASE_URL" default:"http://localhost:1234/v1/"`
	LLMApiKey  string `env:"PRAG_LLM_API_KEY"`
	LLMModel   string `env:"PRAG_LLM_MODEL" default:"granite-3.1-8b-instruct"`
}

// New creates a new RAG client
func New() (*Client, error) {
	log.Info().Msg("starting RAG")

	var err error
	r := &Client{}

	if err := env.Load(&r.cfg, nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	r.vectorDBClient, err = r.connectToVectorDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vectorDB: %w", err)
	}

	log.Info().Msg("starting LLM client")
	r.llmClient, err = llm.New(llm.Config{
		BaseURL:        r.cfg.LLMBaseURL,
		APIKey:         r.cfg.LLMApiKey,
		Model:          r.cfg.LLMModel,
		VectorDBClient: r.vectorDBClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	err = r.startPrometheusSync()
	if err != nil {
		return nil, fmt.Errorf("failed to start prometheus sync: %w", err)
	}

	return r, nil
}

func (r *Client) Query(query string) (string, error) {
	response, err := r.llmClient.Run(query)
	if err != nil {
		return "", fmt.Errorf("failed to run LLM: %w", err)
	}

	return response, nil
}

func (r *Client) connectToVectorDB() (vectordb.Client, error) {
	log.Info().Msg("starting VectorDB client")
	vectordbConfig := vectordb.Config{
		Provider:               r.cfg.VectorDBProvider,
		Sqlite3DBPath:          r.cfg.Sqlite3DBPath,
		QdrantHost:             r.cfg.VectorDBQdrantHost,
		QdrantPort:             r.cfg.VectorDBQdrantPort,
		CollectionName:         r.cfg.VectorDBCollection,
		EncoderOutputDirectory: r.cfg.VectorDBEncoderDir,
	}

	vectordbAPI, err := vectordb.New(vectordbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vectordb API: %w", err)
	}

	return vectordbAPI, nil
}

func (r *Client) startPrometheusSync() error {
	log.Info().Msg("starting Prometheus client")
	var err error

	prometheusConfig := prometheus.Config{
		Address: r.cfg.PrometheusAddress,
	}

	r.prometheusClient, err = prometheus.New(prometheusConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus API: %w", err)
	}

	ticker := time.NewTicker(time.Duration(r.cfg.PrometheusRefreshRateMinutes) * time.Minute)
	go func() {
		r.listMetricsMetadata()

		for range ticker.C {
			r.listMetricsMetadata()
		}
	}()

	return nil
}

func (r *Client) listMetricsMetadata() {
	var err error

	log.Info().Msg("listing metrics metadata from Prometheus")

	r.metricsMetadata, err = r.prometheusClient.ListMetricsMetadata()
	if err != nil {
		log.Error().Err(err).Msg("failed to list metrics metadata")
		return
	}

	log.Info().Msgf("found %d metrics metadata", len(r.metricsMetadata))

	err = r.vectorDBClient.BatchAddMetricMetadata(r.metricsMetadata)
	if err != nil {
		log.Error().Err(err).Msg("failed to add metrics metadata to vectorDB")
		return
	}

	log.Info().Msg("metrics metadata added to vectorDB")
}
