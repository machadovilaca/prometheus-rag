package main

import (
	"log"

	"github.com/machadovilaca/prometheus-rag/pkg/server"
	"github.com/rs/zerolog"
	"go-simpler.org/env"
)

type config struct {
	Debug bool `env:"PRAG_DEBUG" default:"false"`

	Host string `env:"PRAG_HOST" default:"0.0.0.0"`
	Port string `env:"PRAG_PORT" default:"8080"`
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

	server, err := server.New(cfg.Host, cfg.Port)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
