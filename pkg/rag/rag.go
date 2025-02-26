package rag

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go-simpler.org/env"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

type rag struct {
	cfg config

	vectorDBAPI vectordb.VectorDB

	prometheusAPI   prometheus.API
	metricsMetadata []*prometheus.MetricMetadata
}

type config struct {
	PrometheusAddress            string `env:"PRAG_PROMETHEUS_ADDRESS" default:"http://localhost:9090"`
	PrometheusRefreshRateMinutes int    `env:"PRAG_PROMETHEUS_REFRESH_RATE_MINUTES" default:"10"`

	VectorDBHost       string `env:"PRAG_VECTORDB_HOST" default:"localhost"`
	VectorDBPort       int    `env:"PRAG_VECTORDB_PORT" default:"6334"`
	VectorDBCollection string `env:"PRAG_VECTORDB_COLLECTION" default:"prag-metrics"`
	VectorDBEncoderDir string `env:"PRAG_VECTORDB_ENCODER_DIR" default:"./_models"`
}

func New() (*rag, error) {
	log.Info().Msg("starting RAG")

	var err error
	r := &rag{}

	if err := env.Load(&r.cfg, nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	r.vectorDBAPI, err = r.connectToVectorDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vectorDB: %w", err)
	}

	err = r.startPrometheusSync()
	if err != nil {
		return nil, fmt.Errorf("failed to start prometheus sync: %w", err)
	}

	return r, nil
}

func (r *rag) Query(query string) ([]*prometheus.MetricMetadata, error) {
	return r.vectorDBAPI.SearchMetrics(query, 10)
}

func (r *rag) connectToVectorDB() (vectordb.VectorDB, error) {
	log.Info().Msg("starting VectorDB client")
	vectordbConfig := vectordb.Config{
		Host:                   r.cfg.VectorDBHost,
		Port:                   r.cfg.VectorDBPort,
		CollectionName:         r.cfg.VectorDBCollection,
		EncoderOutputDirectory: r.cfg.VectorDBEncoderDir,
	}

	vectordbAPI, err := vectordb.New(vectordbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vectordb API: %w", err)
	}

	return vectordbAPI, nil
}

func (r *rag) startPrometheusSync() error {
	log.Info().Msg("starting Prometheus client")
	var err error

	prometheusConfig := prometheus.Config{
		Address: r.cfg.PrometheusAddress,
	}

	r.prometheusAPI, err = prometheus.New(prometheusConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus API: %w", err)
	}

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		r.listMetricsMetadata()

		for {
			select {
			case <-ticker.C:
				r.listMetricsMetadata()
			}
		}
	}()

	return nil
}

func (r *rag) listMetricsMetadata() {
	var err error

	log.Info().Msg("listing metrics metadata from Prometheus")

	r.metricsMetadata, err = r.prometheusAPI.ListMetricsMetadata()
	if err != nil {
		log.Error().Err(err).Msg("failed to list metrics metadata")
		return
	}

	log.Info().Msgf("found %d metrics metadata", len(r.metricsMetadata))

	err = r.vectorDBAPI.BatchAddMetricMetadata(r.metricsMetadata)
	if err != nil {
		log.Error().Err(err).Msg("failed to add metrics metadata to vectorDB")
		return
	}

	log.Info().Msg("metrics metadata added to vectorDB")
}
