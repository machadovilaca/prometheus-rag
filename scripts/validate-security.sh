#!/bin/bash

# Environment validation script for prometheus-rag
# Checks for common security misconfigurations and credential exposures

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ISSUES_FOUND=0

print_status() {
    local status="$1"
    local message="$2"

    case "$status" in
        "pass")
            echo -e "${GREEN}✅ PASS${NC}: $message"
            ;;
        "fail")
            echo -e "${RED}❌ FAIL${NC}: $message"
            ((ISSUES_FOUND++))
            ;;
        "warn")
            echo -e "${YELLOW}⚠️  WARN${NC}: $message"
            ;;
        "info")
            echo -e "ℹ️  INFO: $message"
            ;;
    esac
}

echo "🔍 Prometheus RAG Security Validation"
echo "====================================="

# Check 1: .env file in .gitignore
echo
echo "📋 Checking environment file security..."

if grep -q "^\.env$" .gitignore 2>/dev/null; then
    print_status "pass" ".env is properly ignored in .gitignore"
else
    print_status "fail" ".env is not in .gitignore - add it to prevent credential exposure"
fi

# Check 2: .env file not tracked in git
if git ls-files --error-unmatch .env &>/dev/null; then
    print_status "fail" ".env file is tracked in git - remove it and add to .gitignore"
else
    print_status "pass" ".env file is not tracked in git"
fi

# Check 3: Check for hardcoded API keys in source code
echo
echo "📋 Scanning for hardcoded credentials in source code..."

HARDCODED_PATTERNS=(
    "sk-[a-zA-Z0-9]{48,}"  # OpenAI API keys
    "AIza[0-9A-Za-z_-]{35}"  # Google API keys
    "AKIA[0-9A-Z]{16}"  # AWS Access Key IDs
    "['\"][a-zA-Z0-9]{32,}['\"].*api[_-]?key"  # Generic API keys
    "password\s*[:=]\s*['\"][^'\"]{8,}['\"]"  # Hardcoded passwords
)

FOUND_HARDCODED=false
for pattern in "${HARDCODED_PATTERNS[@]}"; do
    if git grep -E "$pattern" -- '*.go' '*.yaml' '*.yml' '*.json' '*.md' 2>/dev/null | grep -v vendor/ | grep -v ".example" | head -5; then
        FOUND_HARDCODED=true
    fi
done

if [ "$FOUND_HARDCODED" = true ]; then
    print_status "fail" "Potential hardcoded credentials found in source code"
else
    print_status "pass" "No obvious hardcoded credentials found in source code"
fi

# Check 4: Kubernetes secrets validation
echo
echo "📋 Checking Kubernetes deployment security..."

if grep -q 'llm-api-key: ""' deploy/kubernetes/deployment.yaml; then
    print_status "pass" "Kubernetes secret template is properly empty"
else
    print_status "warn" "Kubernetes secret may contain hardcoded values"
fi

# Check 5: Check for exposed secrets in Kubernetes manifests
if grep -E "password|secret|key.*:" deploy/kubernetes/*.yaml | grep -v "secretKeyRef" | grep -v "name:" | grep -v "key:" | grep -v '""'; then
    print_status "fail" "Potential secrets exposed in Kubernetes manifests"
else
    print_status "pass" "No exposed secrets found in Kubernetes manifests"
fi

# Check 6: Environment file validation if it exists
echo
echo "📋 Validating environment configuration..."

if [ -f ".env" ]; then
    # Check for example/placeholder values
    if grep -E "(your-api-key|example|placeholder|changeme)" .env; then
        print_status "warn" "Environment file contains placeholder values - update with real credentials"
    fi

    # Check for actual API key patterns
    if grep -E "sk-[a-zA-Z0-9]{48,}" .env; then
        print_status "warn" "Environment file contains what looks like real API keys - ensure file is secure"
    fi

    # Check file permissions
    PERMS=$(stat -c "%a" .env)
    if [ "$PERMS" = "600" ] || [ "$PERMS" = "400" ]; then
        print_status "pass" ".env file has secure permissions ($PERMS)"
    else
        print_status "warn" ".env file permissions are $PERMS - consider using 600 for better security"
    fi
else
    print_status "info" ".env file not found - using environment variables or defaults"
fi

# Check 7: Docker security
echo
echo "📋 Checking Docker configuration security..."

if [ -f "deploy/Dockerfile" ]; then
    if grep -q "USER.*root" deploy/Dockerfile; then
        print_status "fail" "Dockerfile runs as root user"
    elif grep -q "USER " deploy/Dockerfile; then
        print_status "pass" "Dockerfile uses non-root user"
    else
        print_status "warn" "Dockerfile doesn't specify user - will run as root by default"
    fi
fi

# Check 8: Dependency vulnerabilities (if govulncheck is available)
echo
echo "📋 Checking for known vulnerabilities..."

if command -v govulncheck &> /dev/null; then
    if govulncheck ./... 2>/dev/null | grep -q "No vulnerabilities found"; then
        print_status "pass" "No known vulnerabilities in dependencies"
    else
        print_status "warn" "Potential vulnerabilities found - run 'govulncheck ./...' for details"
    fi
else
    print_status "info" "govulncheck not available - install with 'go install golang.org/x/vuln/cmd/govulncheck@latest'"
fi

# Summary
echo
echo "📊 Security Validation Summary"
echo "=============================="

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}🎉 No critical security issues found!${NC}"
    echo "The application appears to be properly configured for secure deployment."
else
    echo -e "${RED}⚠️  $ISSUES_FOUND security issues found${NC}"
    echo "Please review and fix the issues above before deploying to production."
fi

echo
echo "💡 Additional security recommendations:"
echo "   - Rotate API keys regularly"
echo "   - Use strong authentication for all services"
echo "   - Monitor access logs for suspicious activity"
echo "   - Keep dependencies updated"
echo "   - Use network policies in Kubernetes"
echo "   - Enable audit logging"

exit $ISSUES_FOUND
