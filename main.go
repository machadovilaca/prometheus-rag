package main

import (
	"log"
	"time"

	"github.com/machadovilaca/prometheus-rag/pkg/rag"
)

func main() {
	rag, err := rag.New()
	if err != nil {
		log.Fatalf("failed to run RAG: %v", err)
	}

	time.Sleep(10 * time.Second)

	metrics, err := rag.Query("how to count the number VMI migrations?")
	if err != nil {
		log.Fatalf("failed to query RAG: %v", err)
	}

	for _, metric := range metrics {
		log.Printf("metric: %s, %s, %s, %v", metric.Name, metric.Help, metric.Type, metric.Labels)
	}
}
