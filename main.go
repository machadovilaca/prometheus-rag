package main

import (
	"log"

	"github.com/machadovilaca/prometheus-rag/pkg/config"
	"github.com/machadovilaca/prometheus-rag/pkg/server"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	server, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
