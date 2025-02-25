package main

import (
	"log"

	"github.com/machadovilaca/prometheus-rag/pkg/rag"
)

func main() {
	err := rag.Run()
	if err != nil {
		log.Fatalf("failed to run RAG: %v", err)
	}
}
