package rag

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/config"
	"github.com/machadovilaca/prometheus-rag/pkg/llm"
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

// Client is the main client for the RAG
type Client struct {
	cfg config.RAGConfig

	vectorDBClient   vectordb.Client
	prometheusClient prometheus.Client
	llmClient        llm.Client
	metricsMetadata  []*prometheus.MetricMetadata
}

// New creates a new RAG client
func New(cfg *config.Config) (*Client, error) {
	log.Info().Msg("starting RAG")

	var err error
	r := &Client{}

	r.vectorDBClient, err = r.connectToVectorDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vectorDB: %w", err)
	}

	// Create RAG-specific configuration
	r.cfg = cfg.ToRAGConfig(r.vectorDBClient)

	log.Info().Msg("starting LLM client")
	r.llmClient, err = llm.New(r.cfg.LLMConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	err = r.startPrometheusSync(cfg)
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

func (r *Client) connectToVectorDB(cfg *config.Config) (vectordb.Client, error) {
	log.Info().Msg("starting VectorDB client")
	vectordbConfig := cfg.ToVectorDBConfig()

	vectordbAPI, err := vectordb.New(vectordbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vectordb API: %w", err)
	}

	return vectordbAPI, nil
}

func (r *Client) startPrometheusSync(cfg *config.Config) error {
	log.Info().Msg("starting Prometheus client")
	var err error

	prometheusConfig := cfg.ToPrometheusConfig()

	r.prometheusClient, err = prometheus.New(prometheusConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus API: %w", err)
	}

	ticker := time.NewTicker(r.cfg.GetPrometheusRefreshInterval())
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
