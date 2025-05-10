# Prometheus RAG: A natural language interface for Prometheus metrics using RAG (Retrieval-Augmented Generation)

Prometheus RAG bridges the gap between natural language and Prometheus metrics by implementing a Retrieval-Augmented Generation (RAG) system. It allows users to query metrics using plain English, automatically finding relevant metrics and generating appropriate PromQL queries.

Key features:

- Natural language to PromQL translation

- Automatic metric metadata synchronization

- Vector similarity search for relevant metrics

- BERT-based encoding for semantic understanding

- Integration with vector database

This guide will help you set up and run Prometheus RAG locally for development.

## Prerequisites

- Podman/Docker and Podman/Docker Compose
- Go 1.21+
- Access to an LLM server

## Quick Start

1. **Start VectorDB and Prometheus**

   Launch the development Qdrant VectorDB and Prometheus server:

   ```bash
   # You can use podman or docker
   cd hack && docker compose up -d
   ```

2. **Configure LLM Server**

   Ensure you have a running LLM server and set these required environment variables:

   ```bash
   export PRAG_LLM_BASE_URL="your-llm-server-url"
   export PRAG_LLM_API_KEY="your-api-key"
   ```

3. **Start RAG Server**

   Launch the RAG application:

   ```bash
   go run main.go
   ```

4. **Test the Service**

   Send a test query:

   ```bash
   curl -X POST \
     http://localhost:8080/query \
     -H "Content-Type: application/json" \
     -d '{"query": "What is the total number of VMs?"}'
   ```

## Configuration

The application can be configured using the following environment variables:

### Server Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PRAG_DEBUG` | Enable debug logging | `false` |
| `PRAG_HOST` | Server host address | `0.0.0.0` |
| `PRAG_PORT` | Server port | `8080` |

### Prometheus Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PRAG_PROMETHEUS_ADDRESS` | Prometheus server URL | `http://localhost:9090` |
| `PRAG_PROMETHEUS_REFRESH_RATE_MINUTES` | Metadata refresh interval | `10` |

### VectorDB Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PRAG_VECTORDB_PROVIDER` | VectorDB provider (sqlite3, qdrant) | `qdrant` |
| `PRAG_VECTORDB_SQLITE3_DB_PATH` | SQLite3 database path (Required only if SQLite3 provider) | `./_data/prag.db` |
| `PRAG_VECTORDB_QDRANT_HOST` | Qdrant host (Required only if Qdrant provider)  | `localhost` |
| `PRAG_VECTORDB_QDRANT_PORT` | Qdrant port (Required only if Qdrant provider) | `6333` |
| `PRAG_VECTORDB_COLLECTION` | Collection name | `prag-metrics` |
| `PRAG_VECTORDB_ENCODER_DIR` | Encoder models directory | `./_models` |

### OpenAI-compatible LLM Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PRAG_LLM_BASE_URL` | LLM server base URL | `http://localhost:1234/v1/` |
| `PRAG_LLM_API_KEY` | Authentication key |  |
| `PRAG_LLM_MODEL` | Model identifier | `granite-3.1-8b-instruct` |
