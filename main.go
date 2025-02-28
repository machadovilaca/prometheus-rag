package main

import (
	"log"
	"time"

	"github.com/rs/zerolog"
	"go-simpler.org/env"

	"github.com/machadovilaca/prometheus-rag/pkg/rag"
)

type config struct {
	Debug bool `env:"PRAG_DEBUG" default:"false"`
}

func main() {
	var cfg config
	if err := env.Load(&cfg, nil); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	rag, err := rag.New()
	if err != nil {
		log.Fatalf("failed to run RAG: %v", err)
	}

	time.Sleep(10 * time.Second)

	answer, err := rag.Query("line graph with the increase of the total number of VMs per namespace")
	if err != nil {
		log.Fatalf("failed to query RAG: %v", err)
	}

	log.Printf("answer: %s", answer)
}
