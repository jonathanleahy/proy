#!/bin/bash

# Script to compare REST v1 and v2 responses for all test cases

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REST_V1_URL="http://0.0.0.0:3002"
REST_V2_URL="http://0.0.0.0:3004"
TEST_CASES_FILE="/home/jon/personal/prroxy/test-data/test-cases-with-responses.json"
COMPARISON_REPORT="/home/jon/personal/prroxy/test-data/comparison-report.json"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  REST v1 vs v2 Comprehensive Comparison${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if test cases file exists
if [ ! -f "$TEST_CASES_FILE" ]; then
    echo -e "${RED}Error: $TEST_CASES_FILE not found${NC}"
    exit 1
fi

# Initialize report
echo '{"timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",' > "$COMPARISON_REPORT"
echo '"test_cases": [' >> "$COMPARISON_REPORT"

# Get number of test cases
num_tests=$(jq '.test_cases | length' "$TEST_CASES_FILE")

PASSED=0
FAILED=0

echo -e "${BLUE}Running comparison tests...${NC}"
echo ""

# Process each test case
for i in $(seq 0 $((num_tests - 1))); do
    # Extract test case details
    test=$(jq ".test_cases[$i]" "$TEST_CASES_FILE")
    test_id=$(echo "$test" | jq -r '.id')
    test_name=$(echo "$test" | jq -r '.name')
    endpoint=$(echo "$test" | jq -r '.endpoint')
    method=$(echo "$test" | jq -r '.method')
    expected_response=$(echo "$test" | jq -c '.expected_response')

    echo -n "[$((i+1))/$num_tests] $test_name... "

    # Build request based on method
    if [ "$method" = "GET" ]; then
        # Execute GET request for v1
        v1_response=$(curl -s -X GET "${REST_V1_URL}${endpoint}" 2>/dev/null || echo '{"error":"Request failed"}')
        v1_status=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${REST_V1_URL}${endpoint}")

        # Execute GET request for v2
        v2_response=$(curl -s -X GET "${REST_V2_URL}${endpoint}" 2>/dev/null || echo '{"error":"Request failed"}')
        v2_status=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${REST_V2_URL}${endpoint}")
    elif [ "$method" = "POST" ]; then
        # Get request body
        body=$(echo "$test" | jq -c '.request_body')

        # Execute POST request for v1
        v1_response=$(curl -s -X POST "${REST_V1_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body" 2>/dev/null || echo '{"error":"Request failed"}')
        v1_status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${REST_V1_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body")

        # Execute POST request for v2
        v2_response=$(curl -s -X POST "${REST_V2_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body" 2>/dev/null || echo '{"error":"Request failed"}')
        v2_status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${REST_V2_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body")
    fi

    # Compare responses (normalize by parsing and re-encoding)
    v1_normalized=$(echo "$v1_response" | jq -S '.')
    v2_normalized=$(echo "$v2_response" | jq -S '.')

    # Create test case result
    if [ "$i" -gt 0 ]; then
        echo "," >> "$COMPARISON_REPORT"
    fi

    echo -n "{" >> "$COMPARISON_REPORT"
    echo -n "\"id\": \"$test_id\"," >> "$COMPARISON_REPORT"
    echo -n "\"name\": \"$test_name\"," >> "$COMPARISON_REPORT"
    echo -n "\"endpoint\": \"$endpoint\"," >> "$COMPARISON_REPORT"
    echo -n "\"method\": \"$method\"," >> "$COMPARISON_REPORT"
    echo -n "\"v1_status\": $v1_status," >> "$COMPARISON_REPORT"
    echo -n "\"v2_status\": $v2_status," >> "$COMPARISON_REPORT"
    echo -n "\"v1_response\": $v1_response," >> "$COMPARISON_REPORT"
    echo -n "\"v2_response\": $v2_response," >> "$COMPARISON_REPORT"

    # Check if responses match
    if [ "$v1_normalized" = "$v2_normalized" ] && [ "$v1_status" = "$v2_status" ]; then
        echo -e "${GREEN}✓${NC}"
        echo -n "\"match\": true" >> "$COMPARISON_REPORT"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ (responses differ)${NC}"
        echo -n "\"match\": false" >> "$COMPARISON_REPORT"
        FAILED=$((FAILED + 1))
    fi

    echo -n "}" >> "$COMPARISON_REPORT"
done

# Close report
echo "]," >> "$COMPARISON_REPORT"
echo "\"summary\": {" >> "$COMPARISON_REPORT"
echo "\"total\": $num_tests," >> "$COMPARISON_REPORT"
echo "\"passed\": $PASSED," >> "$COMPARISON_REPORT"
echo "\"failed\": $FAILED," >> "$COMPARISON_REPORT"
echo "\"pass_rate\": $(echo "scale=2; $PASSED * 100 / $num_tests" | bc)%" >> "$COMPARISON_REPORT"
echo "}}" >> "$COMPARISON_REPORT"

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "Total tests:     $num_tests"
echo -e "${GREEN}Matching:        $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Differences:     $FAILED${NC}"
else
    echo -e "Differences:     $FAILED"
fi

echo ""
echo -e "Detailed report saved to: ${GREEN}$COMPARISON_REPORT${NC}"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Perfect match! REST v2 produces identical responses to REST v1${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠ Some differences found. Check the report for details.${NC}"
    exit 1
fi