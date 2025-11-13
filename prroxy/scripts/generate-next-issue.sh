#!/bin/bash

# Script to analyze comparison report and generate a focused "next issue" report
# This creates a detailed report for the next problem that needs to be fixed

set -e

# Configuration
COMPARISON_JSON="/home/jon/personal/prroxy/test-data/comparison-report.json"
DETAILED_DIR="/home/jon/personal/prroxy/test-data/detailed-reports"
NEXT_ISSUE_REPORT="/home/jon/personal/prroxy/test-data/next-issue.md"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Generating Next Issue Report${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if comparison report exists
if [ ! -f "$COMPARISON_JSON" ]; then
    echo -e "${RED}Error: Comparison report not found. Run comparison first.${NC}"
    exit 1
fi

# Analyze failures to prioritize issues
echo -e "${YELLOW}Analyzing test failures to identify next issue...${NC}"

# Count different types of failures
STATUS_200_MISMATCHES=$(jq '[.test_cases[] | select(.v1_status == 200 and .v2_status == 200 and .match == false)] | length' "$COMPARISON_JSON")
STATUS_500_ERRORS=$(jq '[.test_cases[] | select(.v2_status == 500)] | length' "$COMPARISON_JSON")
STATUS_404_MISMATCHES=$(jq '[.test_cases[] | select(.v1_status == 404 and .v2_status != 404)] | length' "$COMPARISON_JSON")

echo "- Response mismatches (both 200): $STATUS_200_MISMATCHES"
echo "- REST v2 500 errors: $STATUS_500_ERRORS"
echo "- Status code mismatches: $STATUS_404_MISMATCHES"
echo ""

# Priority: Fix response mismatches first (easier), then 500 errors
if [ $STATUS_200_MISMATCHES -gt 0 ]; then
    echo -e "${GREEN}Next issue: Response structure mismatch${NC}"
    ISSUE_TYPE="response_mismatch"
    # Get first test with response mismatch
    FAILING_TEST=$(jq -r '.test_cases[] | select(.v1_status == 200 and .v2_status == 200 and .match == false) | .id' "$COMPARISON_JSON" | head -1)
elif [ $STATUS_500_ERRORS -gt 0 ]; then
    echo -e "${GREEN}Next issue: REST v2 500 errors (missing proxy recordings)${NC}"
    ISSUE_TYPE="missing_recordings"
    # Get first test with 500 error
    FAILING_TEST=$(jq -r '.test_cases[] | select(.v2_status == 500) | .id' "$COMPARISON_JSON" | head -1)
else
    echo -e "${GREEN}All tests passing!${NC}"
    echo "# All Tests Passing" > "$NEXT_ISSUE_REPORT"
    echo "" >> "$NEXT_ISSUE_REPORT"
    echo "Congratulations! All tests are now passing." >> "$NEXT_ISSUE_REPORT"
    exit 0
fi

# Get test details
TEST_NAME=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .name" "$COMPARISON_JSON")
TEST_ENDPOINT=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .endpoint" "$COMPARISON_JSON")
TEST_METHOD=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .method" "$COMPARISON_JSON")
V1_STATUS=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v1_status" "$COMPARISON_JSON")
V2_STATUS=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v2_status" "$COMPARISON_JSON")

# Generate the next issue report
cat > "$NEXT_ISSUE_REPORT" << EOF
# Next Issue to Fix

**Generated:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Issue Priority:** HIGH

## Problem Summary

**Issue Type:** $([ "$ISSUE_TYPE" = "response_mismatch" ] && echo "Response Structure Mismatch" || echo "Missing Proxy Recording")
**Affected Test:** $TEST_NAME
**Endpoint:** \`$TEST_ENDPOINT\`
**Method:** $TEST_METHOD
**Status Codes:** v1=$V1_STATUS, v2=$V2_STATUS

## Issue Details
EOF

if [ "$ISSUE_TYPE" = "response_mismatch" ]; then
    cat >> "$NEXT_ISSUE_REPORT" << EOF

### Problem
REST v2 is returning a different response structure than REST v1, even though both return HTTP 200.
This indicates the response transformation logic in REST v2 needs adjustment.

### Root Cause Analysis
EOF

    # Check if it's a todo truncation issue
    if [[ "$FAILING_TEST" == *"report"* ]]; then
        V1_TODOS=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v1_response.todos" "$COMPARISON_JSON" 2>/dev/null)
        V2_TODOS=$(jq -r ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v2_response.todos" "$COMPARISON_JSON" 2>/dev/null)

        cat >> "$NEXT_ISSUE_REPORT" << EOF
The report endpoint is not returning the complete list of todos:
- REST v1 returns all todos (pending and completed)
- REST v2 is truncating the todo lists to only 4-5 items

**Location to fix:** \`/home/jon/personal/prroxy/rest-v2/internal/domain/user/service.go\`
- Lines 109-113: Remove todo limiting logic
- The service is incorrectly limiting todos when \`includeAll\` is false
- REST v1 always returns all todos regardless of flags

### Fix Required
\`\`\`go
// Current (WRONG):
if !includeAll && len(todos) > 5 {
    displayTodos = todos[:5]
}

// Should be (CORRECT):
// Don't limit todos - always return all todos to match v1 behavior
displayTodos = todos
\`\`\`
EOF
    fi

elif [ "$ISSUE_TYPE" = "missing_recordings" ]; then
    cat >> "$NEXT_ISSUE_REPORT" << EOF

### Problem
REST v2 is getting HTTP 500 errors because the proxy doesn't have recordings for REST v2's request format.

### Root Cause Analysis
- REST v2 requests: \`/proxy?target=jsonplaceholder.typicode.com/users/5\`
- Proxy has recordings for: \`/users/5\` (recorded via REST v1)
- The proxy cannot match these different request formats

### Fix Options

#### Option 1: Record Missing Endpoints (Recommended)
Run REST v2 in record mode to capture the missing endpoints:
\`\`\`bash
# Start proxy in record mode
cd /home/jon/personal/prroxy/proxy
npm run record

# Make requests through REST v2 to record them
curl http://0.0.0.0:3004$TEST_ENDPOINT
\`\`\`

#### Option 2: Modify REST v2 Client
Update the JSONPlaceholder client to use the same request format as v1:
- File: \`/home/jon/personal/prroxy/rest-v2/internal/adapters/outbound/jsonplaceholder/client.go\`
- Adjust how the proxy URL is constructed to match v1's format
EOF
fi

# Add response comparison
if [ -f "$DETAILED_DIR/test-*-$FAILING_TEST.md" ]; then
    cat >> "$NEXT_ISSUE_REPORT" << EOF

## Response Comparison

### REST v1 Response (Expected)
\`\`\`json
$(jq ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v1_response" "$COMPARISON_JSON" 2>/dev/null | head -20)
...
\`\`\`

### REST v2 Response (Actual)
\`\`\`json
$(jq ".test_cases[] | select(.id == \"$FAILING_TEST\") | .v2_response" "$COMPARISON_JSON" 2>/dev/null | head -20)
...
\`\`\`
EOF
fi

# Add testing instructions
cat >> "$NEXT_ISSUE_REPORT" << EOF

## How to Test Fix

1. Apply the fix to the identified file(s)
2. Rebuild REST v2:
   \`\`\`bash
   cd /home/jon/personal/prroxy/rest-v2
   go build -o rest-v2 ./cmd/server
   \`\`\`

3. Restart REST v2:
   \`\`\`bash
   pkill -f rest-v2
   PROXY_URL=http://0.0.0.0:8080/proxy PORT=3004 ./rest-v2 &
   \`\`\`

4. Run comparison for this specific test:
   \`\`\`bash
   curl -X $TEST_METHOD "http://0.0.0.0:3002$TEST_ENDPOINT" > /tmp/v1.json
   curl -X $TEST_METHOD "http://0.0.0.0:3004$TEST_ENDPOINT" > /tmp/v2.json
   diff /tmp/v1.json /tmp/v2.json
   \`\`\`

5. If fixed, run full comparison:
   \`\`\`bash
   /home/jon/personal/prroxy/scripts/compare-v1-v2-with-reports.sh
   \`\`\`

## Success Criteria
- [ ] REST v2 returns same response structure as v1
- [ ] Status codes match (v1=$V1_STATUS, v2 should also be $V1_STATUS)
- [ ] This test passes in the comparison report
- [ ] No regression in other passing tests

## Files to Review
- Main issue location: See "Root Cause Analysis" section above
- Test details: \`$DETAILED_DIR/test-*-$FAILING_TEST.md\`
- Full comparison: \`/home/jon/personal/prroxy/test-data/comparison-report.md\`

---
*This report was automatically generated to help fix REST v2 compatibility issues*
EOF

echo ""
echo -e "${GREEN}✓ Next issue report generated successfully${NC}"
echo -e "Report location: ${BLUE}$NEXT_ISSUE_REPORT${NC}"
echo ""
echo "To view the report:"
echo "  cat $NEXT_ISSUE_REPORT"