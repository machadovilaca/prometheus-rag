# Configuration Package

This package provides centralized configuration management for the prometheus-rag application. It supports loading configuration from environment variables with sensible defaults and provides adapters to convert the central configuration to package-specific configurations.

## Features

- **Centralized Configuration**: All configuration is managed through a single `Config` struct
- **Environment Variable Support**: Configuration values can be set via environment variables
- **Validation**: Built-in validation ensures configuration integrity
- **Modular Design**: Package-specific adapters allow easy integration without tight coupling
- **Reusable Packages**: Individual packages can be used in external projects with their own configuration

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/machadovilaca/prometheus-rag/pkg/config"
)

func main() {
    // Load configuration from environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load configuration: %v", err)
    }

    // Use configuration
    fmt.Printf("Server will run on %s\n", cfg.GetServerAddress())
}
```

### Package Integration

```go
// For vectordb package
vectordbConfig := cfg.ToVectorDBConfig()
vectordbClient, err := vectordb.New(vectordbConfig)

// For prometheus package
prometheusConfig := cfg.ToPrometheusConfig()
prometheusClient, err := prometheus.New(prometheusConfig)

// For LLM package
llmConfig := cfg.ToLLMConfig(vectordbClient)
llmClient, err := llm.New(llmConfig)
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PRAG_DEBUG` | Enable debug logging | `false` |
| `PRAG_HOST` | Server host | `0.0.0.0` |
| `PRAG_PORT` | Server port | `8080` |
| `PRAG_PROMETHEUS_ADDRESS` | Prometheus server address | `http://localhost:9090` |
| `PRAG_PROMETHEUS_REFRESH_RATE_MINUTES` | Metrics refresh interval | `10` |
| `PRAG_VECTORDB_PROVIDER` | Vector database provider (`sqlite3` or `qdrant`) | `sqlite3` |
| `PRAG_VECTORDB_COLLECTION` | Vector database collection name | `prag-metrics` |
| `PRAG_VECTORDB_ENCODER_DIR` | Directory for encoder models | `./_models` |
| `PRAG_VECTORDB_SQLITE3_DB_PATH` | SQLite3 database path | `./_data/metrics.db` |
| `PRAG_VECTORDB_QDRANT_HOST` | Qdrant host | `localhost` |
| `PRAG_VECTORDB_QDRANT_PORT` | Qdrant port | `6334` |
| `PRAG_LLM_BASE_URL` | LLM API base URL | `http://localhost:1234/v1/` |
| `PRAG_LLM_API_KEY` | LLM API key | `` |
| `PRAG_LLM_MODEL` | LLM model name | `granite-3.1-8b-instruct` |

## Architecture

### Configuration Flow

```
Environment Variables
         ↓
    config.Load()
         ↓
    Central Config
         ↓
   Package Adapters
         ↓
Package-specific Configs
         ↓
   Package Instances
```

### Package Adapters

The configuration package provides adapters that convert the central configuration to package-specific configurations:

- `ToPrometheusConfig()` - For prometheus package
- `ToVectorDBConfig()` - For vectordb package
- `ToLLMConfig()` - For llm package
- `ToEmbeddingsConfig()` - For embeddings package
- `ToRAGConfig()` - For RAG-specific configuration

This design allows packages to remain independent and reusable while providing a centralized configuration experience for the main application.

## Validation

The configuration includes comprehensive validation:

- Required fields are checked
- Provider-specific configurations are validated
- Port numbers and URLs are validated
- Dependencies between configurations are checked

## Benefits

1. **Modularity**: Packages can be used independently with their own configuration
2. **Centralization**: Main application has a single configuration source
3. **Reusability**: Packages like `vectordb` can be used in external projects
4. **Maintainability**: Configuration changes are centralized and type-safe
5. **Testability**: Each package can be tested with specific configurations
