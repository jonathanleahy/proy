#!/bin/bash

# Validate test-cases.json structure and content

TEST_CASES_FILE="${TEST_CASES_FILE:-test-data/test-cases.json}"

echo "Validating test cases file: $TEST_CASES_FILE"
echo ""

# Check if file exists
if [ ! -f "$TEST_CASES_FILE" ]; then
    echo "❌ ERROR: Test cases file not found: $TEST_CASES_FILE"
    exit 1
fi

# Check if valid JSON
if ! jq empty "$TEST_CASES_FILE" 2>/dev/null; then
    echo "❌ ERROR: Invalid JSON syntax"
    echo ""
    echo "Common issues:"
    echo "  - Missing comma between objects"
    echo "  - Trailing comma after last item"
    echo "  - Unescaped quotes in strings"
    echo "  - Missing closing bracket"
    exit 1
fi

echo "✅ Valid JSON syntax"

# Check for test_cases array
if ! jq -e '.test_cases' "$TEST_CASES_FILE" > /dev/null 2>&1; then
    echo "❌ ERROR: Missing 'test_cases' array"
    echo ""
    echo "Expected structure:"
    echo '{'
    echo '  "test_cases": [ ... ]'
    echo '}'
    exit 1
fi

echo "✅ Has test_cases array"

# Count test cases
TEST_COUNT=$(jq '.test_cases | length' "$TEST_CASES_FILE")
echo "✅ Found $TEST_COUNT test cases"

if [ "$TEST_COUNT" -eq 0 ]; then
    echo "⚠️  WARNING: No test cases found"
    exit 1
fi

# Validate each test case has required fields
INVALID_COUNT=0
for i in $(seq 0 $((TEST_COUNT - 1))); do
    TEST_CASE=$(jq ".test_cases[$i]" "$TEST_CASES_FILE")

    # Check required fields
    MISSING_FIELDS=""

    if ! echo "$TEST_CASE" | jq -e '.id' > /dev/null 2>&1; then
        MISSING_FIELDS="$MISSING_FIELDS id"
    fi

    if ! echo "$TEST_CASE" | jq -e '.name' > /dev/null 2>&1; then
        MISSING_FIELDS="$MISSING_FIELDS name"
    fi

    if ! echo "$TEST_CASE" | jq -e '.endpoint' > /dev/null 2>&1; then
        MISSING_FIELDS="$MISSING_FIELDS endpoint"
    fi

    if ! echo "$TEST_CASE" | jq -e '.method' > /dev/null 2>&1; then
        MISSING_FIELDS="$MISSING_FIELDS method"
    fi

    if [ -n "$MISSING_FIELDS" ]; then
        TEST_ID=$(echo "$TEST_CASE" | jq -r '.id // "unknown"')
        echo "❌ Test case $i (id: $TEST_ID) missing fields:$MISSING_FIELDS"
        INVALID_COUNT=$((INVALID_COUNT + 1))
    fi
done

if [ "$INVALID_COUNT" -gt 0 ]; then
    echo ""
    echo "❌ $INVALID_COUNT test case(s) have missing required fields"
    echo ""
    echo "Required fields: id, name, endpoint, method"
    exit 1
fi

echo "✅ All test cases have required fields"

# Check for duplicate IDs
DUPLICATE_IDS=$(jq -r '.test_cases[].id' "$TEST_CASES_FILE" | sort | uniq -d)
if [ -n "$DUPLICATE_IDS" ]; then
    echo "❌ ERROR: Duplicate test case IDs found:"
    echo "$DUPLICATE_IDS"
    exit 1
fi

echo "✅ No duplicate test case IDs"

# Validate HTTP methods
INVALID_METHODS=$(jq -r '.test_cases[].method' "$TEST_CASES_FILE" | grep -v -E '^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)$' || true)
if [ -n "$INVALID_METHODS" ]; then
    echo "⚠️  WARNING: Non-standard HTTP methods found:"
    echo "$INVALID_METHODS"
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Validation passed!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test cases: $TEST_COUNT"
echo "File: $TEST_CASES_FILE"
echo ""
echo "Ready to record tests:"
echo "  ./scripts/execute-and-record-v1.sh"
echo ""

exit 0
