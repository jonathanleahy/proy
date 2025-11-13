#!/bin/bash

# Script to execute test cases against REST v1 and record responses

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REST_V1_URL="http://0.0.0.0:3002"
PROXY_URL="http://0.0.0.0:8080"
TEST_CASES_FILE="test-data/test-cases.json"
TEST_CASES_WITH_RESPONSES="test-data/test-cases-with-responses.json"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  REST v1 Test Execution and Recording${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed${NC}"
    echo "Install with: sudo apt-get install jq"
    exit 1
fi

# Check if proxy is running
echo -n "Checking proxy... "
if curl -s "${PROXY_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "Proxy not running. Please start it first."
    exit 1
fi

# Set proxy to record mode
echo -n "Setting proxy to record mode... "
curl -s -X POST "${PROXY_URL}/admin/mode" \
    -H "Content-Type: application/json" \
    -d '{"mode":"record"}' > /dev/null
echo -e "${GREEN}✓${NC}"

# Check if REST v1 is running
echo -n "Checking REST v1... "
if curl -s "${REST_V1_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${YELLOW}REST v1 not running. Please start it in another terminal:${NC}"
    echo "  cd rest-v1"
    echo "  npm install"
    echo "  npm start"
    exit 1
fi

echo ""
echo -e "${BLUE}Executing test cases...${NC}"
echo ""

# Read test cases
if [ ! -f "$TEST_CASES_FILE" ]; then
    echo -e "${RED}Error: $TEST_CASES_FILE not found${NC}"
    exit 1
fi

# Create a copy of test cases to store responses
cp "$TEST_CASES_FILE" "$TEST_CASES_WITH_RESPONSES"

# Get number of test cases
num_tests=$(jq '.test_cases | length' "$TEST_CASES_FILE")

SUCCESS=0
FAILED=0

# Process each test case
for i in $(seq 0 $((num_tests - 1))); do
    # Extract test case details
    test=$(jq ".test_cases[$i]" "$TEST_CASES_FILE")
    test_id=$(echo "$test" | jq -r '.id')
    test_name=$(echo "$test" | jq -r '.name')
    endpoint=$(echo "$test" | jq -r '.endpoint')
    method=$(echo "$test" | jq -r '.method')

    echo -n "  [$((i+1))/$num_tests] $test_name... "

    # Build curl command based on method
    if [ "$method" = "GET" ]; then
        # Execute GET request
        response=$(curl -s -X GET "${REST_V1_URL}${endpoint}" 2>/dev/null || echo '{"error":"Request failed"}')
        status_code=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${REST_V1_URL}${endpoint}")
    elif [ "$method" = "POST" ]; then
        # Get request body
        body=$(echo "$test" | jq -c '.request_body')
        headers=$(echo "$test" | jq -r '.headers | to_entries | map("-H \"\(.key): \(.value)\"") | join(" ")')

        # Execute POST request
        response=$(curl -s -X POST "${REST_V1_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body" 2>/dev/null || echo '{"error":"Request failed"}')
        status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${REST_V1_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$body")
    fi

    # Check if request was successful
    if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
        echo -e "${GREEN}✓ (HTTP $status_code)${NC}"
        SUCCESS=$((SUCCESS + 1))

        # Update test case with response
        jq ".test_cases[$i].expected_response = $response | .test_cases[$i].status_code = $status_code" \
            "$TEST_CASES_WITH_RESPONSES" > tmp.json && mv tmp.json "$TEST_CASES_WITH_RESPONSES"
    else
        echo -e "${RED}✗ (HTTP $status_code)${NC}"
        FAILED=$((FAILED + 1))

        # Store error response
        jq ".test_cases[$i].expected_response = $response | .test_cases[$i].status_code = $status_code | .test_cases[$i].error = true" \
            "$TEST_CASES_WITH_RESPONSES" > tmp.json && mv tmp.json "$TEST_CASES_WITH_RESPONSES"
    fi

    # Small delay to avoid overwhelming the server
    sleep 0.1
done

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "Total tests:     $num_tests"
echo -e "${GREEN}Successful:      $SUCCESS${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed:          $FAILED${NC}"
else
    echo -e "Failed:          $FAILED"
fi
echo ""
echo -e "Test cases with responses saved to: ${GREEN}$TEST_CASES_WITH_RESPONSES${NC}"
echo ""

# Get proxy recording statistics
echo -e "${BLUE}Proxy Recording Statistics:${NC}"
curl -s "${PROXY_URL}/admin/status" | jq '.'
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests completed successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Switch proxy to playback mode"
    echo "  2. Start REST v2"
    echo "  3. Run comparison script"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed. Check the responses.${NC}"
    exit 1
fi