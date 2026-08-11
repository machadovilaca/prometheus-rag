#!/bin/bash

# Script to help setup secrets securely for prometheus-rag deployment
# This script should be run by administrators to set up secrets properly

set -euo pipefail

NAMESPACE="prometheus-rag"
SECRET_NAME="prag-secrets"

echo "🔐 Prometheus RAG Secrets Setup"
echo "================================="

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo "📝 Creating namespace: $NAMESPACE"
    kubectl create namespace "$NAMESPACE"
fi

# Function to prompt for secret input
prompt_secret() {
    local var_name="$1"
    local description="$2"
    local required="$3"

    echo
    echo "📋 $description"
    if [[ "$required" == "true" ]]; then
        echo "   (Required)"
    else
        echo "   (Optional - press Enter to skip)"
    fi

    read -s -p "   Enter value: " value
    echo

    if [[ -z "$value" && "$required" == "true" ]]; then
        echo "❌ This field is required"
        return 1
    fi

    echo -n "$value"
}

# Collect secrets
echo "Please provide the following secrets:"

# LLM API Key
while true; do
    LLM_API_KEY=$(prompt_secret "LLM_API_KEY" "LLM API Key (for OpenAI, Anthropic, etc.)" "false")
    if [[ $? -eq 0 ]]; then
        break
    fi
done

# Prometheus Token (optional)
PROMETHEUS_TOKEN=$(prompt_secret "PROMETHEUS_TOKEN" "Prometheus Authentication Token" "false")

# Validate API key format if provided
if [[ -n "$LLM_API_KEY" ]]; then
    if [[ "$LLM_API_KEY" =~ ^sk-[a-zA-Z0-9]{48,}$ ]] || [[ "$LLM_API_KEY" =~ ^[a-zA-Z0-9_-]{32,}$ ]]; then
        echo "✅ API key format looks valid"
    else
        echo "⚠️  Warning: API key format doesn't match common patterns"
        read -p "   Continue anyway? (y/N): " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            echo "❌ Aborted"
            exit 1
        fi
    fi
fi

# Check if secret already exists
if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo
    echo "⚠️  Secret '$SECRET_NAME' already exists in namespace '$NAMESPACE'"
    read -p "   Do you want to update it? (y/N): " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "❌ Aborted"
        exit 1
    fi

    echo "🗑️  Deleting existing secret..."
    kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"
fi

# Create the secret
echo
echo "🔧 Creating Kubernetes secret..."

kubectl_args=("create" "secret" "generic" "$SECRET_NAME" "-n" "$NAMESPACE")

if [[ -n "$LLM_API_KEY" ]]; then
    kubectl_args+=("--from-literal=llm-api-key=$LLM_API_KEY")
fi

if [[ -n "$PROMETHEUS_TOKEN" ]]; then
    kubectl_args+=("--from-literal=prometheus-token=$PROMETHEUS_TOKEN")
fi

if kubectl "${kubectl_args[@]}"; then
    echo "✅ Secret created successfully!"
else
    echo "❌ Failed to create secret"
    exit 1
fi

# Set proper labels
echo "🏷️  Adding labels to secret..."
kubectl label secret "$SECRET_NAME" -n "$NAMESPACE" \
    app=prometheus-rag \
    component=secrets \
    managed-by=setup-script

echo
echo "🎉 Secrets setup completed!"
echo
echo "📋 Next steps:"
echo "   1. Deploy the application: kubectl apply -f deploy/kubernetes/"
echo "   2. Monitor the deployment: kubectl get pods -n $NAMESPACE"
echo "   3. Check logs: kubectl logs -f deployment/prometheus-rag -n $NAMESPACE"
echo
echo "🔒 Security reminders:"
echo "   - Keep your API keys secure and rotate them regularly"
echo "   - Monitor API usage for any suspicious activity"
echo "   - Use least-privilege access for service accounts"
