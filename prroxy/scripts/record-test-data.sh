#!/bin/bash

# Script to record comprehensive test data from JSONPlaceholder API
# This creates recordings that REST v1 endpoints depend on

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
PROXY_URL="${PROXY_URL:-http://0.0.0.0:8080}"
JSONPLACEHOLDER_BASE="jsonplaceholder.typicode.com"

# Counters
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  JSONPlaceholder Test Data Recording Script${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Function to check if proxy is running
check_proxy() {
    echo -e "${YELLOW}Checking if proxy is running...${NC}"
    if curl -s "${PROXY_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Proxy is running at ${PROXY_URL}${NC}"
        return 0
    else
        echo -e "${RED}✗ Proxy is not running at ${PROXY_URL}${NC}"
        echo ""
        echo "Please start the proxy first:"
        echo "  cd proxy && make run"
        echo "  OR"
        echo "  ./build/proxy"
        exit 1
    fi
}

# Function to set proxy to record mode
set_record_mode() {
    echo -e "${YELLOW}Setting proxy to record mode...${NC}"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${PROXY_URL}/admin/mode" \
        -H "Content-Type: application/json" \
        -d '{"mode":"record"}')

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Proxy set to record mode${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to set record mode (HTTP $HTTP_CODE)${NC}"
        exit 1
    fi
}

# Function to make a proxied request
record_request() {
    local target=$1
    local description=$2

    ((TOTAL_REQUESTS++))

    echo -n "  Recording: $description... "

    # Make the request and capture the HTTP code
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY_URL}/proxy?target=${target}")

    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓${NC}"
        ((SUCCESSFUL_REQUESTS++))
        return 0
    else
        echo -e "${RED}✗ (HTTP $HTTP_CODE)${NC}"
        ((FAILED_REQUESTS++))
        return 0  # Don't fail the script, just count it
    fi
}

# Function to show progress
show_progress() {
    local current=$1
    local total=$2
    local description=$3
    echo -e "${BLUE}[${current}/${total}] ${description}${NC}"
}

# Main recording process
main() {
    check_proxy
    echo ""
    set_record_mode
    echo ""

    # Record Users (1-10)
    echo -e "${BLUE}Recording Users (1-10)...${NC}"
    for user_id in {1..10}; do
        record_request "${JSONPLACEHOLDER_BASE}/users/${user_id}" "User ${user_id}"
    done
    echo ""

    # Record Posts by User (needed for /summary endpoint)
    echo -e "${BLUE}Recording Posts by User (1-10)...${NC}"
    for user_id in {1..10}; do
        record_request "${JSONPLACEHOLDER_BASE}/posts?userId=${user_id}" "Posts for User ${user_id}"
    done
    echo ""

    # Record Todos by User (needed for /report endpoint)
    echo -e "${BLUE}Recording Todos by User (1-10)...${NC}"
    for user_id in {1..10}; do
        record_request "${JSONPLACEHOLDER_BASE}/todos?userId=${user_id}" "Todos for User ${user_id}"
    done
    echo ""

    # Record individual posts (1-100) for comprehensive coverage
    echo -e "${BLUE}Recording Individual Posts (1-100)...${NC}"
    for post_id in {1..100}; do
        if [ $((post_id % 10)) -eq 1 ] || [ $post_id -eq 100 ]; then
            show_progress $post_id 100 "Posts"
        fi
        record_request "${JSONPLACEHOLDER_BASE}/posts/${post_id}" "Post ${post_id}" > /dev/null
    done
    echo -e "${GREEN}✓ Completed 100 posts${NC}"
    echo ""

    # Record individual todos (1-200) for comprehensive coverage
    echo -e "${BLUE}Recording Individual Todos (1-200)...${NC}"
    for todo_id in {1..200}; do
        if [ $((todo_id % 20)) -eq 1 ] || [ $todo_id -eq 200 ]; then
            show_progress $todo_id 200 "Todos"
        fi
        record_request "${JSONPLACEHOLDER_BASE}/todos/${todo_id}" "Todo ${todo_id}" > /dev/null
    done
    echo -e "${GREEN}✓ Completed 200 todos${NC}"
    echo ""

    # Show summary
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}  Recording Summary${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo ""
    echo -e "Total requests:      ${TOTAL_REQUESTS}"
    echo -e "${GREEN}Successful:          ${SUCCESSFUL_REQUESTS}${NC}"
    if [ $FAILED_REQUESTS -gt 0 ]; then
        echo -e "${RED}Failed:              ${FAILED_REQUESTS}${NC}"
    else
        echo -e "Failed:              ${FAILED_REQUESTS}"
    fi
    echo ""

    # Check recordings directory
    if [ -d "recordings" ]; then
        echo -e "${YELLOW}Recordings saved to:${NC}"
        echo "  $(pwd)/recordings/"
        echo ""
        echo -e "${YELLOW}Recorded services:${NC}"
        ls -1 recordings/ 2>/dev/null || echo "  (none yet)"
    fi
    echo ""

    # Provide next steps
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}  Next Steps${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo ""
    echo "1. View recordings in web dashboard:"
    echo -e "   ${GREEN}http://0.0.0.0:8080/admin/ui${NC}"
    echo ""
    echo "2. Switch to playback mode:"
    echo -e "   ${GREEN}curl -X POST http://0.0.0.0:8080/admin/mode \\${NC}"
    echo -e "   ${GREEN}     -H \"Content-Type: application/json\" \\${NC}"
    echo -e "   ${GREEN}     -d '{\"mode\":\"playback\"}'${NC}"
    echo ""
    echo "3. Test REST v1 with recorded data:"
    echo -e "   ${GREEN}cd rest-v1 && npm start${NC}"
    echo -e "   ${GREEN}curl http://0.0.0.0:3000/api/user/1${NC}"
    echo ""

    if [ $FAILED_REQUESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All recordings completed successfully!${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠ Some recordings failed. Check the output above.${NC}"
        exit 1
    fi
}

# Run the script
main
