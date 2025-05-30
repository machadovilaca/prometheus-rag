package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	promAPI "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Client interface for interacting with Prometheus
type Client interface {
	// ListMetricsMetadata lists all metrics metadata from Prometheus
	ListMetricsMetadata() ([]*MetricMetadata, error)
}

// Config represents the configuration for the Prometheus API
type Config struct {
	Address string
}

type api struct {
	client promAPI.Client
}

// New creates a new Prometheus client
func New(cfg Config) (Client, error) {
	client, err := promAPI.NewClient(promAPI.Config{
		Address: cfg.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &api{
		client: client,
	}, nil
}

// ListMetricsMetadata lists all metrics metadata from Prometheus
func (p *api) ListMetricsMetadata() ([]*MetricMetadata, error) {
	v1api := promv1.NewAPI(p.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := v1api.Metadata(ctx, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics metadata: %w", err)
	}

	return p.convertMetadata(results), nil
}

func (p *api) convertMetadata(results map[string][]promv1.Metadata) []*MetricMetadata {
	metrics := make([]*MetricMetadata, 0, len(results))

	for metric, metadata := range results {
		metrics = append(metrics, &MetricMetadata{
			Name:   metric,
			Help:   metadata[0].Help,
			Type:   string(metadata[0].Type),
			Labels: p.getMetricLabels(metric),
		})
	}

	return metrics
}

func (p *api) getMetricLabels(metric string) []string {
	v1api := promv1.NewAPI(p.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, _, err := v1api.LabelNames(ctx, []string{metric}, time.Time{}, time.Time{})
	if err != nil {
		log.Error().Err(err).Msg("failed to get metric labels")
		return []string{}
	}

	return results
}
