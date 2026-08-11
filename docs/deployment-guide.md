# Deployment Guide

## Prerequisites

### System Requirements
- **Kubernetes**: v1.24+
- **Helm**: v3.8+ (optional)
- **Container Runtime**: Docker 20.10+ or Podman 3.0+
- **Resources**:
  - Minimum: 4 CPU cores, 8GB RAM, 50GB storage
  - Recommended: 8 CPU cores, 16GB RAM, 100GB storage

### External Dependencies
- **LLM Service**: OpenAI-compatible API endpoint
- **Prometheus**: v2.40+ (data source)
- **Vector Database**: Qdrant v1.7+ or SQLite with vector extensions

## Quick Start

### 1. Clone and Build

```bash
# Clone repository
git clone https://github.com/machadovilaca/prometheus-rag.git
cd prometheus-rag

# Build Docker image
docker build -f deploy/Dockerfile -t prometheus-rag:latest .

# Or use provided build script
./scripts/build.sh
```

### 2. Configure Environment

```bash
# Create namespace
kubectl create namespace prometheus-rag

# Create secrets
kubectl create secret generic prag-secrets \
  --from-literal=llm-api-key="your-api-key" \
  --namespace=prometheus-rag

# Apply configuration
kubectl apply -f deploy/kubernetes/
```

### 3. Verify Deployment

```bash
# Check pod status
kubectl get pods -n prometheus-rag

# Check service endpoints
kubectl get svc -n prometheus-rag

# View logs
kubectl logs -f deployment/prometheus-rag -n prometheus-rag

# Test health endpoint
kubectl port-forward svc/prometheus-rag 8080:8080 -n prometheus-rag
curl http://localhost:8080/healthz
```

## Production Deployment

### Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │ -> │   Ingress       │ -> │   RAG Service   │
│   (External)    │    │   Controller    │    │   (3 replicas)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   Vector DB     │
                                               │   (Qdrant)      │
                                               └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   Prometheus    │
                                               │   (Monitoring)  │
                                               └─────────────────┘
```

### Step 1: Infrastructure Setup

#### Using Kubernetes

```bash
# Create dedicated namespace
kubectl create namespace prometheus-rag

# Apply RBAC and security policies
kubectl apply -f deploy/kubernetes/rbac.yaml

# Deploy network policies
kubectl apply -f deploy/kubernetes/networking.yaml

# Deploy persistent storage
kubectl apply -f deploy/kubernetes/storage.yaml
```

#### Using Helm

```bash
# Add Helm repository (if available)
helm repo add prometheus-rag https://charts.prometheus-rag.io
helm repo update

# Install with custom values
helm install prometheus-rag prometheus-rag/prometheus-rag \
  --namespace prometheus-rag \
  --create-namespace \
  --values values-production.yaml
```

### Step 2: Vector Database Setup

#### Option A: Qdrant (Recommended for Production)

```bash
# Deploy Qdrant
kubectl apply -f deploy/kubernetes/qdrant.yaml

# Verify Qdrant is running
kubectl wait --for=condition=ready pod -l app=qdrant -n prometheus-rag --timeout=300s

# Create collection (automated during startup)
```

#### Option B: SQLite3 (Development/Small Scale)

```bash
# Create persistent volume for SQLite
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sqlite-storage-pvc
  namespace: prometheus-rag
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF
```

### Step 3: Configuration Management

```yaml
# values-production.yaml
global:
  environment: production
  monitoring:
    enabled: true
  security:
    authentication:
      enabled: true
      method: "api-key"

rag:
  replicas: 3
  image:
    repository: prometheus-rag
    tag: "v1.0.0"
    pullPolicy: IfNotPresent

  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

  config:
    debug: false
    prometheus:
      address: "http://prometheus.monitoring.svc.cluster.local:9090"
      refreshRateMinutes: 10

    vectordb:
      provider: "qdrant"
      qdrant:
        host: "qdrant.prometheus-rag.svc.cluster.local"
        port: 6334
      collection: "prag-metrics"

    llm:
      baseURL: "http://llm-service.ai.svc.cluster.local:8080/v1/"
      model: "granite-3.1-8b-instruct"
      timeout: "30s"

qdrant:
  enabled: true
  replicas: 1
  persistence:
    enabled: true
    size: 50Gi
    storageClass: "fast-ssd"

  resources:
    requests:
      cpu: 1000m
      memory: 2Gi
    limits:
      cpu: 4000m
      memory: 8Gi

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: prometheus-rag.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: prometheus-rag-tls
      hosts:
        - prometheus-rag.your-domain.com

monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s

  grafana:
    dashboards:
      enabled: true

networkPolicy:
  enabled: true
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: frontend
      - namespaceSelector:
          matchLabels:
            name: api-gateway
  egress:
    - to:
      - namespaceSelector:
          matchLabels:
            name: monitoring
    - to:
      - namespaceSelector:
          matchLabels:
            name: ai
```

### Step 4: Security Configuration

```bash
# Create TLS certificates
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@your-domain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# Apply pod security policies
kubectl apply -f deploy/kubernetes/security-policies.yaml

# Configure network policies
kubectl apply -f deploy/kubernetes/networking.yaml
```

### Step 5: Monitoring Setup

```bash
# Deploy ServiceMonitor for Prometheus
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prometheus-rag
  namespace: prometheus-rag
spec:
  selector:
    matchLabels:
      app: prometheus-rag
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
EOF

# Import Grafana dashboard
kubectl create configmap prometheus-rag-dashboard \
  --from-file=deploy/monitoring/grafana-dashboard.json \
  --namespace=monitoring
```

## Environment-Specific Configurations

### Development Environment

```yaml
# values-development.yaml
global:
  environment: development

rag:
  replicas: 1
  config:
    debug: true
    vectordb:
      provider: "sqlite3"
      sqlite3:
        dbPath: "/data/metrics.db"

qdrant:
  enabled: false

ingress:
  enabled: false

monitoring:
  serviceMonitor:
    enabled: false
```

### Staging Environment

```yaml
# values-staging.yaml
global:
  environment: staging

rag:
  replicas: 2
  config:
    debug: false
    prometheus:
      refreshRateMinutes: 5  # More frequent updates for testing

qdrant:
  enabled: true
  replicas: 1
  persistence:
    size: 20Gi

ingress:
  enabled: true
  hosts:
    - host: prometheus-rag-staging.your-domain.com
```

### Production Environment

```yaml
# values-production.yaml (see Step 3 above)
```

## Scaling and High Availability

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: prometheus-rag-hpa
  namespace: prometheus-rag
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: prometheus-rag
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: prometheus-rag-pdb
  namespace: prometheus-rag
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: prometheus-rag
```

### Multi-Region Deployment

```yaml
# Deploy across multiple availability zones
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-rag
spec:
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - prometheus-rag
              topologyKey: topology.kubernetes.io/zone
```

## Backup and Disaster Recovery

### Vector Database Backup

```bash
# Automated backup script
#!/bin/bash
NAMESPACE="prometheus-rag"
BACKUP_DIR="/backup/qdrant/$(date +%Y%m%d_%H%M%S)"

# Create backup
kubectl exec -n $NAMESPACE deployment/qdrant -- \
  tar czf - /qdrant/storage | \
  kubectl cp $NAMESPACE/qdrant-pod:/dev/stdin $BACKUP_DIR/qdrant-backup.tar.gz

# Verify backup
if [ $? -eq 0 ]; then
  echo "Backup completed successfully: $BACKUP_DIR"
else
  echo "Backup failed"
  exit 1
fi
```

### Configuration Backup

```bash
# Backup all configurations
kubectl get configmaps,secrets -n prometheus-rag -o yaml > config-backup.yaml

# Backup persistent volumes
velero backup create prometheus-rag-backup \
  --include-namespaces prometheus-rag \
  --include-resources persistentvolumeclaims,persistentvolumes
```

## Troubleshooting

### Common Issues

1. **Pod Startup Failures**
```bash
# Check pod events
kubectl describe pod -l app=prometheus-rag -n prometheus-rag

# Check resource constraints
kubectl top pods -n prometheus-rag

# Verify configurations
kubectl logs deployment/prometheus-rag -n prometheus-rag
```

2. **Vector Database Connection Issues**
```bash
# Test Qdrant connectivity
kubectl exec -it deployment/prometheus-rag -n prometheus-rag -- \
  curl -s http://qdrant:6333/healthz

# Check network policies
kubectl get networkpolicy -n prometheus-rag
```

3. **LLM Service Issues**
```bash
# Test LLM connectivity
kubectl exec -it deployment/prometheus-rag -n prometheus-rag -- \
  curl -s "http://llm-service.ai.svc.cluster.local:8080/v1/models"
```

### Performance Debugging

```bash
# Enable profiling
kubectl port-forward deployment/prometheus-rag 6060:6060 -n prometheus-rag

# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Collect memory profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Log Analysis

```bash
# Structured log query with jq
kubectl logs deployment/prometheus-rag -n prometheus-rag | \
  jq -r 'select(.level == "ERROR") | .message'

# Export logs to external system
kubectl logs deployment/prometheus-rag -n prometheus-rag | \
  fluent-bit --config fluent-bit.conf
```

## Migration Guide

### From Development to Production

1. **Update Configuration**
```bash
# Switch from SQLite to Qdrant
helm upgrade prometheus-rag prometheus-rag/prometheus-rag \
  --set rag.config.vectordb.provider=qdrant \
  --set qdrant.enabled=true
```

2. **Migrate Data**
```bash
# Export from SQLite
./scripts/export-sqlite.sh > metrics-export.json

# Import to Qdrant
./scripts/import-qdrant.sh metrics-export.json
```

3. **Update DNS and Certificates**
```bash
# Update ingress hostname
kubectl patch ingress prometheus-rag-ingress -n prometheus-rag \
  --type='json' \
  -p='[{"op": "replace", "path": "/spec/rules/0/host", "value": "prometheus-rag.prod.your-domain.com"}]'
```

### Version Upgrades

```bash
# Rolling update
helm upgrade prometheus-rag prometheus-rag/prometheus-rag \
  --set rag.image.tag=v1.1.0 \
  --wait

# Verify upgrade
kubectl rollout status deployment/prometheus-rag -n prometheus-rag

# Rollback if needed
helm rollback prometheus-rag 1
```
