
# Run locally

## Run development Qdrant VectorDB and Prometheus server

```bash
cd hack && docker compose up -d
```

## Run development LLM server

You should have a LLM server running, and set the environment variable `PRAG_LLM_BASE_URL` and `PRAG_LLM_API_KEY` to the correct values.

## Run RAG

```bash
go run main.go
```

## Environment variables

```bash
PRAG_DEBUG - enable debug mode (default: false, values: true, false)

PRAG_PROMETHEUS_ADDRESS - Prometheus address (default: http://localhost:9090)
PRAG_PROMETHEUS_REFRESH_RATE_MINUTES - Prometheus refresh rate in minutes (default: 10)

PRAG_VECTORDB_HOST - vectorDB host (default: localhost)
PRAG_VECTORDB_PORT - vectorDB port (default: 6334)
PRAG_VECTORDB_COLLECTION - vectorDB collection (default: prag-metrics)
PRAG_VECTORDB_ENCODER_DIR - vectorDB encoder directory (default: ./_models)

PRAG_LLM_BASE_URL - LLM base URL (default: http://localhost:1234/v1/)
PRAG_LLM_API_KEY - LLM API key
PRAG_LLM_MODEL - LLM model (default: granite-3.1-8b-instruct)
```
