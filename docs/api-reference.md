# API Documentation

## prometheus-rag API Reference

### Base URL
```
https://prometheus-rag.your-domain.com
```

### Authentication
Currently, the API supports API key authentication (when enabled):
```http
Authorization: Bearer YOUR_API_KEY
```

## Endpoints

### Health Check
Check the health status of the RAG service.

**Endpoint:** `GET /healthz`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "components": {
    "vectordb": "healthy",
    "llm": "healthy",
    "prometheus": "healthy"
  }
}
```

**Status Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - One or more components are unhealthy

### Query Natural Language Interface
Process natural language queries and return PromQL queries with context.

**Endpoint:** `POST /query`

**Request Body:**
```json
{
  "query": "What is the CPU usage of my services?",
  "context": {
    "timeRange": "1h",
    "namespace": "production",
    "includeLabels": ["service", "pod"]
  }
}
```

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | Yes | Natural language query about metrics |
| `context.timeRange` | string | No | Time range for metrics (default: "5m") |
| `context.namespace` | string | No | Kubernetes namespace filter |
| `context.includeLabels` | array | No | Labels to include in response |

**Response:**
```json
{
  "response": "To get the CPU usage of your services, you can use this PromQL query:\n\n```promql\nsum by (service) (rate(cpu_usage_seconds_total[5m]))\n```\n\nThis query will show the CPU usage rate per service over the last 5 minutes.",
  "promqlQueries": [
    {
      "query": "sum by (service) (rate(cpu_usage_seconds_total[5m]))",
      "description": "CPU usage rate per service",
      "relevance": 0.95
    }
  ],
  "relatedMetrics": [
    {
      "name": "cpu_usage_seconds_total",
      "description": "Total CPU time consumed",
      "type": "counter"
    },
    {
      "name": "memory_usage_bytes",
      "description": "Memory usage in bytes",
      "type": "gauge"
    }
  ],
  "processingTime": "1.2s",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Query processed successfully
- `400 Bad Request` - Invalid request body or parameters
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Processing error

### Metrics Discovery
Discover available metrics and their metadata.

**Endpoint:** `GET /metrics`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `search` | string | Search term for metric names |
| `type` | string | Filter by metric type (counter, gauge, histogram) |
| `limit` | integer | Number of results to return (default: 50, max: 200) |
| `offset` | integer | Pagination offset (default: 0) |

**Response:**
```json
{
  "metrics": [
    {
      "name": "cpu_usage_seconds_total",
      "description": "Total CPU time consumed by container",
      "type": "counter",
      "labels": ["container", "pod", "namespace"],
      "lastSeen": "2024-01-01T11:59:00Z"
    }
  ],
  "totalCount": 1247,
  "hasMore": true
}
```

### Embeddings Management
Manage vector embeddings for metrics.

**Endpoint:** `POST /embeddings/refresh`

**Description:** Trigger refresh of metric embeddings from Prometheus.

**Response:**
```json
{
  "status": "started",
  "jobId": "refresh-20240101-120000",
  "estimatedDuration": "5m"
}
```

**Endpoint:** `GET /embeddings/status`

**Response:**
```json
{
  "totalEmbeddings": 1247,
  "lastUpdate": "2024-01-01T11:45:00Z",
  "status": "up-to-date",
  "vectorDbProvider": "qdrant",
  "collectionInfo": {
    "name": "prag-metrics",
    "vectorCount": 1247,
    "indexStatus": "ready"
  }
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "INVALID_QUERY",
    "message": "Query parameter is required and cannot be empty",
    "details": {
      "field": "query",
      "provided": "",
      "expected": "non-empty string"
    },
    "timestamp": "2024-01-01T12:00:00Z",
    "requestId": "req-abc123"
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_QUERY` | Query parameter is invalid or missing |
| `PROCESSING_ERROR` | Error occurred during query processing |
| `RATE_LIMIT_EXCEEDED` | Too many requests from client |
| `SERVICE_UNAVAILABLE` | Backend service is temporarily unavailable |
| `AUTHENTICATION_FAILED` | Invalid or missing authentication |
| `INTERNAL_ERROR` | Unexpected server error |

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Default Limit:** 100 requests per minute per IP
- **Burst Limit:** 10 requests per second
- **Headers:** Rate limit information is included in response headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704110460
```

## SDK Examples

### cURL
```bash
# Health check
curl -X GET "https://prometheus-rag.your-domain.com/healthz"

# Query processing
curl -X POST "https://prometheus-rag.your-domain.com/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "query": "Show me memory usage by pod",
    "context": {
      "timeRange": "1h",
      "namespace": "production"
    }
  }'
```

### Python
```python
import requests
import json

class PrometheusRAGClient:
    def __init__(self, base_url, api_key=None):
        self.base_url = base_url
        self.headers = {"Content-Type": "application/json"}
        if api_key:
            self.headers["Authorization"] = f"Bearer {api_key}"

    def query(self, query_text, context=None):
        payload = {"query": query_text}
        if context:
            payload["context"] = context

        response = requests.post(
            f"{self.base_url}/query",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def health_check(self):
        response = requests.get(f"{self.base_url}/healthz")
        return response.status_code == 200

# Usage
client = PrometheusRAGClient(
    "https://prometheus-rag.your-domain.com",
    api_key="your-api-key"
)

result = client.query(
    "What is the error rate for my web services?",
    context={"timeRange": "24h", "namespace": "production"}
)
print(result["response"])
```

### JavaScript/Node.js
```javascript
class PrometheusRAGClient {
    constructor(baseURL, apiKey = null) {
        this.baseURL = baseURL;
        this.headers = {
            'Content-Type': 'application/json'
        };
        if (apiKey) {
            this.headers['Authorization'] = `Bearer ${apiKey}`;
        }
    }

    async query(queryText, context = null) {
        const payload = { query: queryText };
        if (context) {
            payload.context = context;
        }

        const response = await fetch(`${this.baseURL}/query`, {
            method: 'POST',
            headers: this.headers,
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return await response.json();
    }

    async healthCheck() {
        const response = await fetch(`${this.baseURL}/healthz`);
        return response.ok;
    }
}

// Usage
const client = new PrometheusRAGClient(
    'https://prometheus-rag.your-domain.com',
    'your-api-key'
);

client.query('Show me CPU utilization trends')
    .then(result => console.log(result.response))
    .catch(error => console.error('Query failed:', error));
```

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type PrometheusRAGClient struct {
    BaseURL string
    APIKey  string
    Client  *http.Client
}

type QueryRequest struct {
    Query   string                 `json:"query"`
    Context map[string]interface{} `json:"context,omitempty"`
}

type QueryResponse struct {
    Response       string `json:"response"`
    ProcessingTime string `json:"processingTime"`
    Timestamp      string `json:"timestamp"`
}

func (c *PrometheusRAGClient) Query(query string, context map[string]interface{}) (*QueryResponse, error) {
    req := QueryRequest{
        Query:   query,
        Context: context,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/query", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "application/json")
    if c.APIKey != "" {
        httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
    }

    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result QueryResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}

// Usage
func main() {
    client := &PrometheusRAGClient{
        BaseURL: "https://prometheus-rag.your-domain.com",
        APIKey:  "your-api-key",
        Client:  &http.Client{},
    }

    context := map[string]interface{}{
        "timeRange": "1h",
        "namespace": "production",
    }

    result, err := client.Query("What is the memory usage?", context)
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Response)
}
```

## WebSocket API (Future Enhancement)

For real-time query processing and streaming responses:

```javascript
const ws = new WebSocket('wss://prometheus-rag.your-domain.com/ws');

ws.onopen = function() {
    ws.send(JSON.stringify({
        type: 'query',
        data: {
            query: 'Monitor CPU usage in real-time',
            stream: true
        }
    }));
};

ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    if (response.type === 'partial_response') {
        console.log('Partial:', response.data);
    } else if (response.type === 'complete') {
        console.log('Final:', response.data);
    }
};
```
