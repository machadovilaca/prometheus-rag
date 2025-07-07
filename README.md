# Prometheus RAG: A natural language interface for Prometheus metrics using RAG (Retrieval-Augmented Generation)

Prometheus RAG bridges the gap between natural language and Prometheus metrics by implementing a Retrieval-Augmented Generation (RAG) system. It allows users to query metrics using plain English, automatically finding relevant metrics and generating appropriate PromQL queries.

## 🌟 Key Features

- **Natural Language to PromQL Translation**: Convert plain English queries to PromQL
- **Automatic Metric Metadata Synchronization**: Keeps vector database in sync with Prometheus metrics
- **Vector Similarity Search**: Find relevant metrics using semantic understanding
- **BERT-based Encoding**: Uses LaBSE (Language-agnostic BERT Sentence Embedding) for multilingual support
- **Multiple Vector Database Support**: SQLite3 (default) or Qdrant
- **Modular Architecture**: Reusable packages that can be integrated into other projects
- **OpenAI-Compatible LLM Integration**: Works with any OpenAI-compatible API

## 🏗️ Architecture

The project uses a modular configuration system that allows individual packages to be reused in external projects:

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   User Query    │───▶│   RAG System     │───▶│   PromQL Query  │
│  (Natural Lang) │     │                  │     │   + Context     │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │ Vector Database  │
                        │ (Metrics Search) │
                        └──────────────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │   Prometheus     │
                        │ (Metrics Source) │
                        └──────────────────┘
```

## Prerequisites

- Podman/Docker and Podman/Docker Compose
- Go 1.21+
- Sqlite development package (`sudo dnf install sqlite-devel` on Fedora)
- Access to an LLM server

## 🚀 Quick Start

### 1. Start VectorDB and Prometheus

Launch the development environment using Docker Compose:

```bash
# Start Qdrant VectorDB and Prometheus server
cd hack && docker compose up -d

# Or using Podman
cd hack && podman-compose up -d
```

### 2. Configure Environment

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration. At minimum, set your LLM server details:

```bash
# LLM Configuration (Required)
PRAG_LLM_BASE_URL=http://localhost:1234/v1/
PRAG_LLM_API_KEY=your-api-key-here
PRAG_LLM_MODEL=granite-3.1-8b-instruct

# Optional: Use Qdrant instead of SQLite3
# PRAG_VECTORDB_PROVIDER=qdrant
# PRAG_VECTORDB_QDRANT_HOST=localhost
# PRAG_VECTORDB_QDRANT_PORT=6334
```

### 3. Start RAG Server

```bash
# Load environment and start the server
source .env  # or: export $(cat .env | xargs)
go run main.go
```

### 4. Test the Service

Send a test query to verify everything is working:

```bash
curl -X POST \
  http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the total number of VMs?"}'
```

Example response:
```json
{
  "response": "To get the total number of VMs, you can use: `sum(up{job=\"vm-exporter\"})`"
}
```

## ⚙️ Configuration

The application uses a centralized configuration system that loads settings from environment variables. All packages are designed to be modular and reusable.

### Complete Configuration Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| **Server Configuration** |
| `PRAG_DEBUG` | Enable debug logging | `false` | No |
| `PRAG_HOST` | Server host address | `0.0.0.0` | No |
| `PRAG_PORT` | Server port | `8080` | No |
| **Prometheus Configuration** |
| `PRAG_PROMETHEUS_ADDRESS` | Prometheus server URL | `http://localhost:9090` | No |
| `PRAG_PROMETHEUS_REFRESH_RATE_MINUTES` | Metadata refresh interval (minutes) | `10` | No |
| **Vector Database Configuration** |
| `PRAG_VECTORDB_PROVIDER` | VectorDB provider (`sqlite3` or `qdrant`) | `sqlite3` | No |
| `PRAG_VECTORDB_COLLECTION` | Collection name | `prag-metrics` | No |
| `PRAG_VECTORDB_ENCODER_DIR` | Directory for encoder models | `./_models` | No |
| `PRAG_VECTORDB_SQLITE3_DB_PATH` | SQLite3 database path | `./_data/metrics.db` | If using SQLite3 |
| `PRAG_VECTORDB_QDRANT_HOST` | Qdrant host | `localhost` | If using Qdrant |
| `PRAG_VECTORDB_QDRANT_PORT` | Qdrant port | `6334` | If using Qdrant |
| **LLM Configuration** |
| `PRAG_LLM_BASE_URL` | LLM server base URL | `http://localhost:1234/v1/` | **Yes** |
| `PRAG_LLM_API_KEY` | Authentication key | *(empty)* | **Yes** |
| `PRAG_LLM_MODEL` | Model identifier | `granite-3.1-8b-instruct` | No |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style
- Follow Go best practices and idioms
- Use `gofmt` to format your code
- Add comprehensive tests for new features
- Update documentation as needed

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [LaBSE (Language-agnostic BERT Sentence Embedding)](https://ai.googleblog.com/2020/08/language-agnostic-bert-sentence.html) for multilingual embeddings
- [Qdrant](https://qdrant.tech/) for vector database capabilities
- [Prometheus](https://prometheus.io/) for metrics collection and monitoring
- [Spago](https://github.com/nlpodyssey/spago) for Go-based machine learning

---

For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/machadovilaca/prometheus-rag).
