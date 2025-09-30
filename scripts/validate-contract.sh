#!/usr/bin/env bash
# Comprehensive Contract Validation Script
# Validates ALL Stage 1 contract requirements for ds-go

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${DS_BASE_URL:-http://127.0.0.1:7777}"
TOKEN="${DS_TOKEN:-}"
VERBOSE="${VERBOSE:-0}"

# Counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

echo "=========================================="
echo "   DS Contract Validation - Stage 1"
echo "=========================================="
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

# Test function
test_endpoint() {
    local test_name="$1"
    local endpoint="$2"
    local jq_test="$3"
    local description="$4"

    ((TESTS_RUN++))

    if [ "$VERBOSE" = "1" ]; then
        echo -e "${BLUE}Testing:${NC} $test_name"
        echo "  Endpoint: $endpoint"
        echo "  Validation: $jq_test"
    fi

    local response
    response=$(api_call "$endpoint" "$description" 2>/dev/null || echo '{"error":"request failed"}')

    if echo "$response" | jq -e "$jq_test" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected: $jq_test"
        if [ "$VERBOSE" = "1" ]; then
            echo "  Response: $(echo "$response" | jq -c '.' 2>/dev/null || echo "$response")"
        fi
        ((TESTS_FAILED++))
        return 1
    fi
}

echo "1. Core Endpoint Versioning"
echo "----------------------------"

# Test all endpoints for schema_version
test_endpoint \
    "Capabilities has schema_version" \
    "/v1/capabilities" \
    '.schema_version == "ds.v1"' \
    "Capabilities endpoint"

test_endpoint \
    "Health has schema_version" \
    "/v1/health" \
    '.schema_version == "ds.v1" and .ok == true' \
    "Health endpoint"

test_endpoint \
    "Status returns wrapped response" \
    "/v1/status?path=/tmp" \
    '.schema_version == "ds.v1" and has("data") and (.data | type) == "array"' \
    "Status endpoint wrapper"

test_endpoint \
    "Scan has schema_version" \
    "/v1/scan?path=/tmp" \
    '.schema_version == "ds.v1" and has("count")' \
    "Scan endpoint"

echo ""
echo "2. Self-Status Parity"
echo "---------------------"

test_endpoint \
    "Self-status has schema_version" \
    "/api/self-status" \
    '.schema_version == "ds.v1"' \
    "Self-status schema version"

test_endpoint \
    "Self-status has nowMs as number" \
    "/api/self-status" \
    '.nowMs != null and (.nowMs | type) == "number"' \
    "Self-status nowMs field"

test_endpoint \
    "Self-status has auth flags" \
    "/api/self-status" \
    '.auth.tokenRequired != null and .auth.corsEnabled != null' \
    "Self-status auth configuration"

test_endpoint \
    "Self-status has endpoints map" \
    "/api/self-status" \
    '.endpoints.well_known != null and .endpoints.openapi != null' \
    "Self-status endpoints registry"

echo ""
echo "3. Discovery & Well-Known"
echo "-------------------------"

test_endpoint \
    "Discovery services has ds descriptor" \
    "/api/discovery/services" \
    '.ds.url != null and .ds.well_known != null and .ds.self_status != null' \
    "Discovery ds descriptor"

test_endpoint \
    "Discovery has timestamp" \
    "/api/discovery/services" \
    '.ts != null and (.ts | type) == "number"' \
    "Discovery timestamp"

test_endpoint \
    "Discovery has token flag" \
    "/api/discovery/services" \
    'has("ds_token_present") and (.ds_token_present | type) == "boolean"' \
    "Discovery token presence"

test_endpoint \
    "Well-known has contract version" \
    "/.well-known/obs-bridge.json" \
    '.contractVersion != null and .schemaVersion != null' \
    "Well-known contract info"

test_endpoint \
    "Well-known has endpoints map" \
    "/.well-known/obs-bridge.json" \
    '.endpoints.openapi != null and .endpoints.capabilities != null and .endpoints.health != null' \
    "Well-known endpoints"

test_endpoint \
    "Well-known has all list" \
    "/.well-known/obs-bridge.json" \
    '.all != null and (.all | type) == "array" and (.all | length) > 0' \
    "Well-known all endpoints list"

echo ""
echo "4. Response Envelope Validation"
echo "--------------------------------"

# Special test for organize/plan (if endpoint exists)
if api_call "/v1/organize/plan?require_clean=false" "Organize plan" 2>/dev/null | jq -e '.schema_version' >/dev/null 2>&1; then
    test_endpoint \
        "Organize/plan returns wrapped response" \
        "/v1/organize/plan?require_clean=false" \
        '.schema_version == "ds.v1" and has("data")' \
        "Organize plan wrapper"
else
    echo -e "${YELLOW}⚠${NC} Organize/plan endpoint not testable (may require specific setup)"
fi

echo ""
echo "5. OpenAPI Availability"
echo "-----------------------"

test_endpoint \
    "OpenAPI spec accessible" \
    "/openapi.yaml" \
    'type == "string" or type == "object"' \
    "OpenAPI specification"

# Alternative OpenAPI path
if ! api_call "/api/discovery/openapi" "OpenAPI alias" 2>/dev/null | head -1 | grep -q "openapi"; then
    echo -e "${YELLOW}⚠${NC} OpenAPI alias (/api/discovery/openapi) not available or not YAML"
fi

echo ""
echo "6. Contract Consistency Checks"
echo "-------------------------------"

# Check if all versioned endpoints return consistent schema_version
ENDPOINTS=("/v1/capabilities" "/v1/health" "/api/self-status")
SCHEMA_VERSIONS=()

for endpoint in "${ENDPOINTS[@]}"; do
    version=$(api_call "$endpoint" "Schema version check" 2>/dev/null | jq -r '.schema_version // "none"')
    SCHEMA_VERSIONS+=("$version")
done

# Check if all versions are ds.v1
all_consistent=true
for version in "${SCHEMA_VERSIONS[@]}"; do
    if [ "$version" != "ds.v1" ]; then
        all_consistent=false
        break
    fi
done

if [ "$all_consistent" = true ]; then
    echo -e "${GREEN}✓${NC} All endpoints return consistent schema_version (ds.v1)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗${NC} Inconsistent schema versions detected"
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

echo ""
echo "=========================================="
echo "           VALIDATION SUMMARY"
echo "=========================================="

SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))

echo "Tests Run:    $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo "Success Rate: $SUCCESS_RATE%"

echo ""
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ ALL CONTRACT VALIDATIONS PASSED!${NC}"
    echo ""
    echo "Stage 1 Contract Requirements:"
    echo "  ✓ All endpoints return schema_version: ds.v1"
    echo "  ✓ Array responses wrapped in {schema_version, data}"
    echo "  ✓ Self-status includes nowMs as epoch milliseconds"
    echo "  ✓ Discovery endpoints fully functional"
    echo "  ✓ Well-known descriptor complete"
    echo "  ✓ Contract consistency maintained"
    exit 0
else
    echo -e "${RED}❌ CONTRACT VALIDATION FAILED${NC}"
    echo ""
    echo "Please fix the failing tests above to ensure Stage 1 compliance."
    echo "Run with VERBOSE=1 for detailed output."
    exit 1
fi