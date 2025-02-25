package rag

import (
	"fmt"

	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

func Run() error {
	fmt.Println("starting RAG")

	fmt.Println("loading prometheus API")
	prometheusConfig := prometheus.Config{
		Address: "http://localhost:9090",
	}
	prometheusAPI, err := prometheus.New(prometheusConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus API: %w", err)
	}

	fmt.Println("loading vectordb API")
	vectordbConfig := vectordb.Config{
		Host:                   "localhost",
		Port:                   6334,
		EncoderOutputDirectory: "./_models",
	}
	vectordbAPI, err := vectordb.New(vectordbConfig)
	if err != nil {
		return fmt.Errorf("failed to create vectordb API: %w", err)
	}

	fmt.Println("listing metrics metadata")
	metricsMetadata, err := prometheusAPI.ListMetricsMetadata()
	if err != nil {
		return fmt.Errorf("failed to list metrics metadata: %w", err)
	}

	for _, metadata := range metricsMetadata {
		fmt.Printf("adding metric metadata: %s, %s, %s, %s, %v\n", metadata.Name, metadata.Help, metadata.Type, metadata.Unit, metadata.Labels)
	}

	fmt.Println("adding metrics metadata to vectordb")
	err = vectordbAPI.BatchAddMetricMetadata(metricsMetadata)
	if err != nil {
		return fmt.Errorf("failed to add metrics metadata to vectordb: %w", err)
	}

	fmt.Println("searching metrics")
	query := "how to count the number VMI migrations?"
	metrics, err := vectordbAPI.SearchMetrics(query, 10)
	if err != nil {
		return fmt.Errorf("failed to search metrics: %w", err)
	}

	fmt.Println("metrics found:")
	for i, metric := range metrics {
		fmt.Printf("%d) metric: %s, %s, %s, %s, %v\n", i, metric.Name, metric.Help, metric.Type, metric.Unit, metric.Labels)
	}

	fmt.Println("done")
	return nil
}
