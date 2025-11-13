#!/bin/bash

# Script to compare REST v1 and v2 responses and generate both JSON and Markdown reports

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
COMPARISON_JSON="/home/jon/personal/prroxy/test-data/comparison-report.json"
COMPARISON_MD="/home/jon/personal/prroxy/test-data/comparison-report.md"
DETAILED_DIR="/home/jon/personal/prroxy/test-data/detailed-reports"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  REST v1 vs v2 Comprehensive Comparison${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if test cases file exists
if [ ! -f "$TEST_CASES_FILE" ]; then
    echo -e "${RED}Error: $TEST_CASES_FILE not found${NC}"
    exit 1
fi

# Create detailed reports directory
mkdir -p "$DETAILED_DIR"

# Initialize JSON report
echo '{"timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",' > "$COMPARISON_JSON"
echo '"test_cases": [' >> "$COMPARISON_JSON"

# Initialize Markdown report
cat > "$COMPARISON_MD" << EOF
# REST API v1 vs v2 Comparison Report

**Generated:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")

## Summary

| Metric | Value |
|--------|-------|
| Total Tests | 0 |
| Passing | 0 |
| Failing | 0 |
| Pass Rate | 0% |

## Test Results

| # | Test Name | Endpoint | Method | v1 Status | v2 Status | Result |
|---|-----------|----------|--------|-----------|-----------|--------|
EOF

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

    # Compare responses (normalize by parsing, removing timestamps, and re-encoding)
    # Remove common timestamp fields that will always differ between requests
    # Then apply deep normalization to handle property and array ordering differences
    v1_normalized=$(echo "$v1_response" | jq 'del(.generatedAt, .timestamp, .createdAt, .updatedAt)' | jq -f scripts/normalize-json.jq)
    v2_normalized=$(echo "$v2_response" | jq 'del(.generatedAt, .timestamp, .createdAt, .updatedAt)' | jq -f scripts/normalize-json.jq)

    # Add to JSON report
    if [ "$i" -gt 0 ]; then
        echo "," >> "$COMPARISON_JSON"
    fi

    echo -n "{" >> "$COMPARISON_JSON"
    echo -n "\"id\": \"$test_id\"," >> "$COMPARISON_JSON"
    echo -n "\"name\": \"$test_name\"," >> "$COMPARISON_JSON"
    echo -n "\"endpoint\": \"$endpoint\"," >> "$COMPARISON_JSON"
    echo -n "\"method\": \"$method\"," >> "$COMPARISON_JSON"
    echo -n "\"v1_status\": $v1_status," >> "$COMPARISON_JSON"
    echo -n "\"v2_status\": $v2_status," >> "$COMPARISON_JSON"
    echo -n "\"v1_response\": $v1_response," >> "$COMPARISON_JSON"
    echo -n "\"v2_response\": $v2_response," >> "$COMPARISON_JSON"

    # Check if responses match
    if [ "$v1_normalized" = "$v2_normalized" ] && [ "$v1_status" = "$v2_status" ]; then
        echo -e "${GREEN}✓${NC}"
        echo -n "\"match\": true" >> "$COMPARISON_JSON"
        PASSED=$((PASSED + 1))
        RESULT_EMOJI="✅"
        RESULT_TEXT="PASS"
    else
        echo -e "${RED}✗ (responses differ)${NC}"
        echo -n "\"match\": false" >> "$COMPARISON_JSON"
        FAILED=$((FAILED + 1))
        RESULT_EMOJI="❌"
        RESULT_TEXT="FAIL"
    fi

    echo -n "}" >> "$COMPARISON_JSON"

    # Add to Markdown report
    echo "| $((i+1)) | $test_name | \`$endpoint\` | $method | $v1_status | $v2_status | $RESULT_EMOJI $RESULT_TEXT |" >> "$COMPARISON_MD"

    # Create detailed report for this test
    DETAIL_FILE="$DETAILED_DIR/test-$(printf "%02d" $((i+1)))-$test_id.md"
    cat > "$DETAIL_FILE" << EOF
# Test Case: $test_name

**Test ID:** $test_id
**Endpoint:** \`$endpoint\`
**Method:** $method
**Result:** $RESULT_EMOJI $RESULT_TEXT

## Status Codes
- **REST v1:** $v1_status
- **REST v2:** $v2_status

## REST v1 Response
\`\`\`json
$(echo "$v1_response" | jq '.')
\`\`\`

## REST v2 Response
\`\`\`json
$(echo "$v2_response" | jq '.')
\`\`\`

## Differences
EOF

    if [ "$v1_normalized" != "$v2_normalized" ] || [ "$v1_status" != "$v2_status" ]; then
        echo "### Status Code Difference" >> "$DETAIL_FILE"
        if [ "$v1_status" != "$v2_status" ]; then
            echo "- v1 returned $v1_status, v2 returned $v2_status" >> "$DETAIL_FILE"
        fi

        echo "" >> "$DETAIL_FILE"
        echo "### Response Body Differences" >> "$DETAIL_FILE"

        # Try to identify key differences
        if [ "$v1_status" = "200" ] && [ "$v2_status" = "200" ]; then
            echo "Both services returned success, but response structures differ." >> "$DETAIL_FILE"
        elif [ "$v2_status" = "500" ]; then
            echo "REST v2 returned an error (likely missing proxy recording)." >> "$DETAIL_FILE"
        fi
    else
        echo "Responses match perfectly! ✅" >> "$DETAIL_FILE"
    fi
done

# Close JSON report
echo "]," >> "$COMPARISON_JSON"
echo "\"summary\": {" >> "$COMPARISON_JSON"
echo "\"total\": $num_tests," >> "$COMPARISON_JSON"
echo "\"passed\": $PASSED," >> "$COMPARISON_JSON"
echo "\"failed\": $FAILED," >> "$COMPARISON_JSON"
PASS_RATE=$(echo "scale=2; $PASSED * 100 / $num_tests" | bc)
echo "\"pass_rate\": \"$PASS_RATE%\"" >> "$COMPARISON_JSON"
echo "}}" >> "$COMPARISON_JSON"

# Update Markdown summary
sed -i "s/| Total Tests | 0 |/| Total Tests | $num_tests |/" "$COMPARISON_MD"
sed -i "s/| Passing | 0 |/| Passing | $PASSED |/" "$COMPARISON_MD"
sed -i "s/| Failing | 0 |/| Failing | $FAILED |/" "$COMPARISON_MD"
sed -i "s/| Pass Rate | 0% |/| Pass Rate | $PASS_RATE% |/" "$COMPARISON_MD"

# Add detailed analysis to Markdown
cat >> "$COMPARISON_MD" << EOF

## Detailed Analysis

### Passing Tests (${PASSED})
EOF

# List passing tests
jq -r '.test_cases[] | select(.match == true) | "- **" + .name + "** (" + .endpoint + ")"' "$COMPARISON_JSON" 2>/dev/null >> "$COMPARISON_MD" || echo "None" >> "$COMPARISON_MD"

cat >> "$COMPARISON_MD" << EOF

### Failing Tests (${FAILED})
EOF

# List failing tests with reasons
if [ -f "$COMPARISON_JSON" ]; then
    echo "" >> "$COMPARISON_MD"
    i=1
    jq -r '.test_cases[] | select(.match == false) | .name + "|" + .endpoint + "|" + (.v1_status | tostring) + "|" + (.v2_status | tostring)' "$COMPARISON_JSON" 2>/dev/null | while IFS='|' read -r name endpoint v1_status v2_status; do
        echo "#### $i. $name" >> "$COMPARISON_MD"
        echo "- **Endpoint:** \`$endpoint\`" >> "$COMPARISON_MD"
        echo "- **Status:** v1=$v1_status, v2=$v2_status" >> "$COMPARISON_MD"
        if [ "$v2_status" = "500" ]; then
            echo "- **Likely Cause:** Missing proxy recording for REST v2 request format" >> "$COMPARISON_MD"
        elif [ "$v1_status" = "$v2_status" ]; then
            echo "- **Likely Cause:** Response structure mismatch" >> "$COMPARISON_MD"
        else
            echo "- **Likely Cause:** Different HTTP status codes" >> "$COMPARISON_MD"
        fi
        echo "" >> "$COMPARISON_MD"
        i=$((i+1))
    done
fi

cat >> "$COMPARISON_MD" << EOF

## Next Steps

1. **Fix Missing Recordings:** Record the failing endpoints through REST v2 in record mode
2. **Fix Response Mismatches:** Update REST v2 handlers to match v1 response formats exactly
3. **Verify Todo Truncation:** Check why REST v2 is not returning complete todo lists

## Files Generated

- **JSON Report:** \`$COMPARISON_JSON\`
- **Markdown Report:** \`$COMPARISON_MD\`
- **Detailed Reports:** \`$DETAILED_DIR/\`

---
*Report generated on $(date)*
EOF

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
echo -e "${BLUE}Reports Generated:${NC}"
echo -e "  • JSON Report:     ${GREEN}$COMPARISON_JSON${NC}"
echo -e "  • Markdown Report: ${GREEN}$COMPARISON_MD${NC}"
echo -e "  • Detailed Tests:  ${GREEN}$DETAILED_DIR/${NC}"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Perfect match! REST v2 produces identical responses to REST v1${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠ Some differences found. Check the reports for details.${NC}"
    exit 1
fi