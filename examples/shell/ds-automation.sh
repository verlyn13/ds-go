#!/usr/bin/env bash
# ds-go Shell Automation Examples
# Collection of useful shell scripts for repository management

set -euo pipefail

# Configuration
API_BASE="${DS_API_URL:-http://127.0.0.1:7777/v1}"
JQ_AVAILABLE=$(command -v jq >/dev/null 2>&1 && echo "yes" || echo "no")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function for API calls
api_call() {
    local method="$1"
    local endpoint="$2"
    shift 2
    curl -s -X "$method" "${API_BASE}${endpoint}" "$@"
}

# Pretty print JSON if jq is available
json_print() {
    if [ "$JQ_AVAILABLE" = "yes" ]; then
        jq '.'
    else
        cat
    fi
}

# Check if server is running
check_server() {
    if ! api_call GET "/capabilities" >/dev/null 2>&1; then
        echo -e "${RED}Error: ds server not running${NC}"
        echo "Start it with: ds serve --addr 127.0.0.1:7777"
        exit 1
    fi
}

# Function: Show repository status summary
show_summary() {
    echo -e "${BLUE}=== Repository Status Summary ===${NC}"
    local status=$(api_call GET "/status")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        local total=$(echo "$status" | jq -r '.data.summary.total')
        local clean=$(echo "$status" | jq -r '.data.summary.clean')
        local dirty=$(echo "$status" | jq -r '.data.summary.dirty')
        local ahead=$(echo "$status" | jq -r '.data.summary.ahead')
        local behind=$(echo "$status" | jq -r '.data.summary.behind')

        echo "Total Repositories: $total"
        echo -e "Clean: ${GREEN}$clean${NC}"
        echo -e "Dirty: ${YELLOW}$dirty${NC}"
        echo -e "Ahead: ${BLUE}$ahead${NC}"
        echo -e "Behind: ${RED}$behind${NC}"
    else
        echo "$status"
    fi
}

# Function: List dirty repositories
list_dirty_repos() {
    echo -e "${YELLOW}=== Dirty Repositories ===${NC}"
    local repos=$(api_call GET "/status?dirty=true")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        echo "$repos" | jq -r '.data.repositories[] | "\(.name): \(.uncommitted_files) files"'
    else
        echo "$repos"
    fi
}

# Function: Run command on all repos
exec_on_all() {
    local cmd="${1:-echo 'No command provided'}"
    echo -e "${BLUE}=== Executing: $cmd ===${NC}"

    local result=$(api_call POST "/exec" \
        -H "Content-Type: application/json" \
        -d "{\"cmd\": \"$cmd\"}")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        local success=$(echo "$result" | jq -r '.data.summary.success')
        local failed=$(echo "$result" | jq -r '.data.summary.failed')
        echo -e "Success: ${GREEN}$success${NC}, Failed: ${RED}$failed${NC}"

        # Show failures
        echo "$result" | jq -r '.data.results[] | select(.success == false) | "\(.repo): \(.stderr)"'
    else
        echo "$result"
    fi
}

# Function: Fetch all repositories
fetch_all() {
    echo -e "${BLUE}=== Fetching All Repositories ===${NC}"
    local result=$(api_call GET "/fetch")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        local success=$(echo "$result" | jq -r '.data.summary.success')
        local failed=$(echo "$result" | jq -r '.data.summary.failed')
        echo -e "Fetched: ${GREEN}$success${NC}, Failed: ${RED}$failed${NC}"
    else
        echo "$result"
    fi
}

# Function: Check policy compliance
check_policy() {
    echo -e "${BLUE}=== Policy Compliance Check ===${NC}"
    local result=$(api_call GET "/policy/check?fail_on=critical")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        local total=$(echo "$result" | jq -r '.data.summary.total')
        local passed=$(echo "$result" | jq -r '.data.summary.passed')
        local failed=$(echo "$result" | jq -r '.data.summary.failed')
        local threshold=$(echo "$result" | jq -r '.data.failed_threshold')

        echo "Total Checks: $total"
        echo -e "Passed: ${GREEN}$passed${NC}"
        echo -e "Failed: ${RED}$failed${NC}"

        if [ "$threshold" = "true" ]; then
            echo -e "${RED}⚠ Critical failures detected!${NC}"
        fi

        # Show failures
        echo "$result" | jq -r '.data.results[] | select(.passed == false) | "✗ \(.severity): \(.name)"'
    else
        echo "$result"
    fi
}

# Function: Organize repositories
organize_repos() {
    echo -e "${BLUE}=== Repository Organization ===${NC}"

    # First, show the plan
    local plan=$(api_call GET "/organize/plan")

    if [ "$JQ_AVAILABLE" = "yes" ]; then
        local moves=$(echo "$plan" | jq -r '.data.summary.total_moves')
        echo "Planned moves: $moves"

        if [ "$moves" -gt 0 ]; then
            echo "$plan" | jq -r '.data.moves[] | "  \(.name): \(.old_path) -> \(.new_path)"'

            read -p "Apply changes? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                local result=$(api_call POST "/organize/apply?require_clean=true")
                local moved=$(echo "$result" | jq -r '.data.summary.moved')
                echo -e "${GREEN}Moved $moved repositories${NC}"
            fi
        else
            echo -e "${GREEN}All repositories are organized correctly${NC}"
        fi
    else
        echo "$plan"
    fi
}

# Function: Daily maintenance routine
daily_maintenance() {
    echo -e "${BLUE}=== Daily Maintenance Routine ===${NC}"

    # 1. Scan for new repos
    echo "1. Scanning for new repositories..."
    api_call GET "/scan" | json_print

    # 2. Fetch remotes
    echo "2. Fetching remote information..."
    fetch_all

    # 3. Check status
    echo "3. Checking repository status..."
    show_summary

    # 4. Check policy
    echo "4. Checking compliance..."
    check_policy

    # 5. Report issues
    local dirty=$(api_call GET "/status" | jq -r '.data.summary.dirty')
    local behind=$(api_call GET "/status" | jq -r '.data.summary.behind')

    if [ "$dirty" -gt 0 ] || [ "$behind" -gt 0 ]; then
        echo -e "${YELLOW}⚠ Attention needed:${NC}"
        [ "$dirty" -gt 0 ] && echo "  - $dirty repositories have uncommitted changes"
        [ "$behind" -gt 0 ] && echo "  - $behind repositories are behind remote"
    else
        echo -e "${GREEN}✓ All repositories are up to date${NC}"
    fi
}

# Function: Stream fetch progress
stream_fetch() {
    echo -e "${BLUE}=== Streaming Fetch Progress ===${NC}"
    echo "Fetching repositories (press Ctrl+C to stop)..."

    curl -N -H "Accept: text/event-stream" "${API_BASE}/fetch/sse" | while IFS= read -r line; do
        if [[ $line == data:* ]]; then
            data="${line#data: }"
            if [ "$JQ_AVAILABLE" = "yes" ]; then
                echo "$data" | jq -r '"\(.repo): \(.status)"' 2>/dev/null || true
            else
                echo "$data"
            fi
        fi
    done
}

# Main menu
show_menu() {
    echo -e "${BLUE}=== ds-go Automation Menu ===${NC}"
    echo "1) Show repository summary"
    echo "2) List dirty repositories"
    echo "3) Fetch all repositories"
    echo "4) Check policy compliance"
    echo "5) Organize repositories"
    echo "6) Run daily maintenance"
    echo "7) Execute command on all repos"
    echo "8) Stream fetch progress"
    echo "q) Quit"
    echo -n "Select option: "
}

# Main script
main() {
    check_server

    if [ $# -eq 0 ]; then
        # Interactive mode
        while true; do
            show_menu
            read -r option
            echo

            case $option in
                1) show_summary ;;
                2) list_dirty_repos ;;
                3) fetch_all ;;
                4) check_policy ;;
                5) organize_repos ;;
                6) daily_maintenance ;;
                7)
                    read -p "Enter command to execute: " cmd
                    exec_on_all "$cmd"
                    ;;
                8) stream_fetch ;;
                q) exit 0 ;;
                *) echo "Invalid option" ;;
            esac

            echo
            read -p "Press Enter to continue..."
            clear
        done
    else
        # Command mode
        case "$1" in
            summary) show_summary ;;
            dirty) list_dirty_repos ;;
            fetch) fetch_all ;;
            policy) check_policy ;;
            organize) organize_repos ;;
            daily) daily_maintenance ;;
            exec) shift; exec_on_all "$*" ;;
            stream) stream_fetch ;;
            *)
                echo "Usage: $0 [summary|dirty|fetch|policy|organize|daily|exec <cmd>|stream]"
                echo "   or: $0 (for interactive menu)"
                exit 1
                ;;
        esac
    fi
}

# Run main function
main "$@"