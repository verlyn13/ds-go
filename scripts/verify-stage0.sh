#!/usr/bin/env bash
# Stage 0 Verification Script for Agent B (ds-go)
# Validates all Stage 0 requirements are met

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${DS_BASE_URL:-http://127.0.0.1:7777}"
TOKEN="${DS_TOKEN:-}"

echo "==================================="
echo "Stage 0 Verification for Agent B"
echo "==================================="
echo "Base URL: $BASE_URL"
echo "Token: ${TOKEN:+[SET]}"
echo ""

# Helper function for API calls
api_call() {
    local endpoint="$1"
    local description="$2"

    if [ -n "$TOKEN" ]; then
        curl -sS -H "Authorization: Bearer $TOKEN" "$BASE_URL$endpoint"
    else
        curl -sS "$BASE_URL$endpoint"
    fi
}

# Check function
check() {
    local test_name="$1"
    local command="$2"

    echo -n "Checking $test_name... "
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# Track failures
FAILURES=0

echo "1. Server Health Check"
echo "----------------------"
if HEALTH=$(api_call "/v1/health" "health check" 2>/dev/null); then
    echo -e "${GREEN}✓${NC} Server is running"
    if echo "$HEALTH" | jq -e '.schema_version == "ds.v1"' >/dev/null; then
        echo -e "${GREEN}✓${NC} Health endpoint has schema_version: ds.v1"
    else
        echo -e "${RED}✗${NC} Health endpoint missing proper schema_version"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} Server is not responding"
    echo "Please start the server with: DS_TOKEN=<token> DS_CORS=1 go run ./cmd/ds serve --addr 127.0.0.1:7777"
    exit 1
fi
echo ""

echo "2. Discovery Endpoints"
echo "----------------------"

# Check well-known
if WELLKNOWN=$(api_call "/.well-known/obs-bridge.json" "well-known" 2>/dev/null); then
    echo -e "${GREEN}✓${NC} /.well-known/obs-bridge.json is accessible"
    if echo "$WELLKNOWN" | jq -e '.contractVersion' >/dev/null; then
        echo -e "${GREEN}✓${NC} Well-known has valid structure"
    else
        echo -e "${RED}✗${NC} Well-known missing required fields"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} /.well-known/obs-bridge.json not accessible"
    ((FAILURES++))
fi

# Check discovery services
if DISCOVERY=$(api_call "/api/discovery/services" "discovery" 2>/dev/null); then
    echo -e "${GREEN}✓${NC} /api/discovery/services is accessible"
    if echo "$DISCOVERY" | jq -e '.ds and .ts' >/dev/null; then
        echo -e "${GREEN}✓${NC} Discovery has ds object and ts timestamp"
    else
        echo -e "${RED}✗${NC} Discovery missing required fields"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} /api/discovery/services not accessible"
    ((FAILURES++))
fi
echo ""

echo "3. Self-Status Endpoint"
echo "-----------------------"
if SELFSTATUS=$(api_call "/api/self-status" "self-status" 2>/dev/null); then
    echo -e "${GREEN}✓${NC} /api/self-status is accessible"

    # Check for nowMs
    if echo "$SELFSTATUS" | jq -e '.nowMs' >/dev/null; then
        NOWMS=$(echo "$SELFSTATUS" | jq -r '.nowMs')
        echo -e "${GREEN}✓${NC} Self-status includes nowMs: $NOWMS"
    else
        echo -e "${RED}✗${NC} Self-status missing nowMs"
        ((FAILURES++))
    fi

    # Check for schema_version
    if echo "$SELFSTATUS" | jq -e '.schema_version == "ds.v1"' >/dev/null; then
        echo -e "${GREEN}✓${NC} Self-status has schema_version: ds.v1"
    else
        echo -e "${RED}✗${NC} Self-status missing proper schema_version"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} /api/self-status not accessible"
    ((FAILURES++))
fi
echo ""

echo "4. Capabilities Endpoint"
echo "------------------------"
if CAPS=$(api_call "/v1/capabilities" "capabilities" 2>/dev/null); then
    echo -e "${GREEN}✓${NC} /v1/capabilities is accessible"
    if echo "$CAPS" | jq -e '.schema_version == "ds.v1"' >/dev/null; then
        echo -e "${GREEN}✓${NC} Capabilities has schema_version: ds.v1"
    else
        echo -e "${RED}✗${NC} Capabilities missing proper schema_version"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} /v1/capabilities not accessible"
    ((FAILURES++))
fi
echo ""

echo "5. Go Client Package"
echo "--------------------"
if [ -f "pkg/dsclient/client.go" ]; then
    echo -e "${GREEN}✓${NC} pkg/dsclient/client.go exists"

    # Check for required methods
    for method in "SelfStatus" "Discovery" "Capabilities"; do
        if grep -q "func.*$method" pkg/dsclient/client.go; then
            echo -e "${GREEN}✓${NC} Client has $method method"
        else
            echo -e "${RED}✗${NC} Client missing $method method"
            ((FAILURES++))
        fi
    done
else
    echo -e "${RED}✗${NC} pkg/dsclient/client.go not found"
    ((FAILURES++))
fi
echo ""

echo "6. Example & Tests"
echo "------------------"
if [ -f "examples/go-client/main.go" ]; then
    echo -e "${GREEN}✓${NC} examples/go-client/main.go exists"
else
    echo -e "${RED}✗${NC} examples/go-client/main.go not found"
    ((FAILURES++))
fi

if [ -f "pkg/dsclient/client_test.go" ]; then
    echo -e "${GREEN}✓${NC} pkg/dsclient/client_test.go exists"

    # Run tests
    echo -n "Running client tests... "
    if go test ./pkg/dsclient/... >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        ((FAILURES++))
    fi
else
    echo -e "${RED}✗${NC} pkg/dsclient/client_test.go not found"
    ((FAILURES++))
fi
echo ""

echo "7. Documentation"
echo "----------------"
if [ -f "docs/integration-checklist.md" ]; then
    echo -e "${GREEN}✓${NC} docs/integration-checklist.md exists"
    if grep -q "Stage 0 Requirements" docs/integration-checklist.md; then
        echo -e "${GREEN}✓${NC} Documentation includes Stage 0 requirements"
    else
        echo -e "${YELLOW}⚠${NC} Documentation may need Stage 0 requirements update"
    fi
else
    echo -e "${RED}✗${NC} docs/integration-checklist.md not found"
    ((FAILURES++))
fi
echo ""

echo "==================================="
echo "Stage 0 Verification Summary"
echo "==================================="
if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✓ All Stage 0 requirements PASSED!${NC}"
    echo ""
    echo "Agent B (ds-go) is fully compliant with Stage 0 requirements:"
    echo "  ✓ schema_version: 'ds.v1' on core endpoints"
    echo "  ✓ /api/self-status includes nowMs"
    echo "  ✓ Discovery endpoints implemented"
    echo "  ✓ Go client package with example and tests"
    exit 0
else
    echo -e "${RED}✗ Stage 0 verification FAILED with $FAILURES issues${NC}"
    echo ""
    echo "Please address the issues above to complete Stage 0."
    exit 1
fi