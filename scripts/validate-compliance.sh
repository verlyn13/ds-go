#!/usr/bin/env bash
# ds-go System Compliance Validation Script
# Ensures ds-go aligns with system configuration policies

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Helper functions
check() {
    local description="$1"
    local command="$2"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "Checking $description... "
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo -e "${RED}✗${NC}"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

warn_check() {
    local description="$1"
    local command="$2"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "Checking $description... "
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo -e "${YELLOW}⚠${NC} (warning)"
        WARNINGS=$((WARNINGS + 1))
        return 1
    fi
}

echo "================================================"
echo "ds-go System Compliance Validation"
echo "================================================"
echo ""

# 1. Directory Structure Compliance
echo "📁 Directory Structure Compliance"
echo "--------------------------------"
check "~/Development exists" "[ -d ~/Development ]"
check "~/Development/personal exists" "[ -d ~/Development/personal ]"
check "~/Development/work exists" "[ -d ~/Development/work ]"
warn_check "~/Development/business exists" "[ -d ~/Development/business ]"
warn_check "~/Development/business-org exists" "[ -d ~/Development/business-org ]"
warn_check "~/Development/hubofwyn exists" "[ -d ~/Development/hubofwyn ]"
check "ds-go in correct location" "[ -d ~/Development/personal/ds-go ]"
echo ""

# 2. Tool Configuration
echo "🛠️  Tool Configuration"
echo "--------------------"
check "mise installed" "command -v mise"
check ".mise.toml exists" "[ -f .mise.toml ]"
check ".envrc exists" "[ -f .envrc ]"
check "direnv installed" "command -v direnv"
check "Go version 1.20+" "go version | grep -E 'go1\.(2[0-9]|[3-9][0-9])'"
check "golangci-lint installed" "command -v golangci-lint"
echo ""

# 3. SSH Configuration
echo "🔐 SSH Configuration"
echo "-------------------"
check "~/.ssh/config exists" "[ -f ~/.ssh/config ]"
check "github-work host configured" "grep -q 'Host github-work' ~/.ssh/config"
warn_check "github-business host configured" "grep -q 'Host github-business' ~/.ssh/config"
warn_check "github-business-org host configured" "grep -q 'Host github-business-org' ~/.ssh/config"
echo ""

# 4. ds Configuration
echo "⚙️  ds Configuration"
echo "------------------"
check "ds config directory exists" "[ -d ~/.config/ds ]"
if [ -f ~/.config/ds/config.yaml ]; then
    check "ds config file exists" "[ -f ~/.config/ds/config.yaml ]"
    check "verlyn13 account configured" "grep -q 'verlyn13:' ~/.config/ds/config.yaml"
    check "jjohnson-47 account configured" "grep -q 'jjohnson-47:' ~/.config/ds/config.yaml"
    warn_check "happy-patterns account configured" "grep -q 'happy-patterns:' ~/.config/ds/config.yaml"
else
    echo -e "ds config file exists... ${YELLOW}⚠${NC} (Run 'ds init' to create)"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 5. Project Standards
echo "📋 Project Standards"
echo "------------------"
check ".gitignore exists" "[ -f .gitignore ]"
check "README.md exists" "[ -f README.md ]"
check "CLAUDE.md exists" "[ -f CLAUDE.md ]"
check "SYSTEM_INTEGRATION.md exists" "[ -f SYSTEM_INTEGRATION.md ]"
check ".golangci.yml exists" "[ -f .golangci.yml ]"
check "Makefile exists" "[ -f Makefile ]"
echo ""

# 6. Security Compliance
echo "🔒 Security Compliance"
echo "---------------------"
check "No .env files in repo" "! find . -name '.env' -type f | grep -q '.env'"
check "No *.pem files in repo" "! find . -name '*.pem' -type f | grep -q '.pem'"
check "No *.key files in repo" "! find . -name '*.key' -type f | grep -q '.key'"
check ".env in .gitignore" "grep -q '^\\.env' .gitignore"
echo ""

# 7. Build and Test
echo "🔨 Build and Test"
echo "----------------"
if check "Project builds successfully" "mise run build 2>/dev/null"; then
    check "Binary created" "[ -f ./ds ]"
fi
warn_check "Tests pass" "mise run test 2>/dev/null"
echo ""

# Summary
echo "================================================"
echo "Validation Summary"
echo "================================================"
echo -e "Total Checks: $TOTAL_CHECKS"
echo -e "Passed: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "Failed: ${RED}$FAILED_CHECKS${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"

# Calculate compliance percentage
COMPLIANCE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
echo ""
echo -n "Compliance Score: "
if [ $COMPLIANCE -ge 90 ]; then
    echo -e "${GREEN}${COMPLIANCE}%${NC} ✅"
elif [ $COMPLIANCE -ge 70 ]; then
    echo -e "${YELLOW}${COMPLIANCE}%${NC} ⚠️"
else
    echo -e "${RED}${COMPLIANCE}%${NC} ❌"
fi

# Exit code based on critical failures
if [ $FAILED_CHECKS -gt 0 ]; then
    echo ""
    echo "❌ Validation failed with $FAILED_CHECKS critical issues"
    exit 1
else
    echo ""
    echo "✅ Validation passed!"
    exit 0
fi