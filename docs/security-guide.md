# Security Configuration Guide

This guide provides comprehensive instructions for securely configuring and deploying prometheus-rag.

## 🔐 Credentials Management

### Environment Variables

**Never commit actual credentials to version control.** Use environment variables to provide sensitive configuration:

```bash
# Required for LLM integration
export PRAG_LLM_API_KEY="your-actual-api-key"

# Optional for Prometheus authentication
export PRAG_PROMETHEUS_TOKEN="your-prometheus-token"
```

### Local Development

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your actual credentials:
   ```bash
   # DO NOT commit this file
   PRAG_LLM_API_KEY=sk-your-actual-openai-key
   ```

3. Ensure proper file permissions:
   ```bash
   chmod 600 .env
   ```

## 🚀 Production Deployment

### Kubernetes Secrets

Use the provided script to securely set up secrets:

```bash
./scripts/setup-secrets.sh
```

Or manually create secrets:

```bash
# Create namespace
kubectl create namespace prometheus-rag

# Create secret with your actual API key
kubectl create secret generic prag-secrets \
  --from-literal=llm-api-key="your-actual-api-key" \
  --from-literal=prometheus-token="your-prometheus-token" \
  -n prometheus-rag
```

### Docker Environment

For Docker deployments, use environment variables:

```bash
docker run -e PRAG_LLM_API_KEY="your-api-key" prometheus-rag:latest
```

## 🛡️ Security Best Practices

### API Key Security

1. **Use strong, unique API keys** for each environment
2. **Rotate keys regularly** (every 90 days recommended)
3. **Monitor API usage** for suspicious activity
4. **Use least-privilege access** when possible
5. **Store keys in secure secret management systems**

### Access Control

1. **Enable authentication** on all production endpoints
2. **Use TLS/HTTPS** for all communications
3. **Implement rate limiting** to prevent abuse
4. **Monitor and log** access patterns

### Infrastructure Security

1. **Run containers as non-root users**
2. **Use read-only filesystems** where possible
3. **Enable network policies** in Kubernetes
4. **Keep dependencies updated**
5. **Scan for vulnerabilities** regularly

## 🔍 Security Validation

Run the security validation script to check your configuration:

```bash
./scripts/validate-security.sh
```

This script will check for:
- Credential exposure in git
- Hardcoded secrets in code
- Insecure file permissions
- Common misconfigurations

## 🚨 Incident Response

If credentials are accidentally exposed:

1. **Immediately rotate** all exposed credentials
2. **Check access logs** for unauthorized usage
3. **Review git history** for any committed secrets
4. **Update security practices** to prevent recurrence

### Git History Cleanup

If secrets were committed to git:

```bash
# Remove sensitive file from history (USE WITH CAUTION)
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch .env' \
  --prune-empty --tag-name-filter cat -- --all

# Force push to remote (if necessary)
git push origin --force --all
```

## 📋 Security Checklist

Before deployment, ensure:

- [ ] No hardcoded credentials in source code
- [ ] `.env` files are in `.gitignore`
- [ ] Secrets are properly configured in deployment
- [ ] API keys are rotated and secure
- [ ] Authentication is enabled
- [ ] TLS is configured
- [ ] Dependencies are up-to-date
- [ ] Security scanning is in place
- [ ] Monitoring and alerting is configured
- [ ] Incident response plan is documented

## 🔗 Additional Resources

- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [Container Security Guide](https://www.nist.gov/publications/application-container-security-guide)
- [Secret Management Best Practices](https://cloud.google.com/secret-manager/docs/best-practices)
