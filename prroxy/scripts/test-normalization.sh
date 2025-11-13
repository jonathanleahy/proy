#!/bin/bash

# Comprehensive test script to verify deep normalization handles property and array ordering

set -e

PASSED=0
FAILED=0

run_test() {
    local test_num=$1
    local test_name=$2
    local json1=$3
    local json2=$4
    local should_match=$5

    normalized1=$(echo "$json1" | jq -f scripts/normalize-json.jq)
    normalized2=$(echo "$json2" | jq -f scripts/normalize-json.jq)

    if [ "$normalized1" = "$normalized2" ]; then
        matches=true
    else
        matches=false
    fi

    echo "Test $test_num: $test_name"

    if [ "$matches" = "$should_match" ]; then
        echo "  Result: ✓ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "  Result: ✗ FAIL"
        echo "  Expected match: $should_match, Got: $matches"
        echo "  Input 1: $json1"
        echo "  Input 2: $json2"
        echo "  Normalized 1: $normalized1"
        echo "  Normalized 2: $normalized2"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

echo "================================================"
echo "  Deep Normalization Test Suite"
echo "================================================"
echo ""

# Test 1: Basic property ordering
run_test 1 "Basic property ordering" \
    '{"name":"John","age":30,"city":"NYC"}' \
    '{"city":"NYC","name":"John","age":30}' \
    true

# Test 2: Simple array ordering
run_test 2 "Simple array ordering" \
    '{"items":["apple","banana","cherry"]}' \
    '{"items":["cherry","apple","banana"]}' \
    true

# Test 3: Complex nested structure
run_test 3 "Complex nested structure" \
    '{"user":{"name":"Alice","id":1},"posts":[{"title":"First","id":1},{"title":"Second","id":2}]}' \
    '{"posts":[{"id":2,"title":"Second"},{"id":1,"title":"First"}],"user":{"id":1,"name":"Alice"}}' \
    true

# Test 4: Deeply nested arrays and objects
run_test 4 "Deeply nested structures" \
    '{"data":{"users":[{"name":"Bob","tags":["admin","user"]},{"name":"Alice","tags":["user","guest"]}]}}' \
    '{"data":{"users":[{"tags":["guest","user"],"name":"Alice"},{"tags":["user","admin"],"name":"Bob"}]}}' \
    true

# Test 5: REST API response format (matching REST v1 and v2)
run_test 5 "REST API response format" \
    '{"userId":1,"userName":"Test","stats":{"total":10,"completed":5}}' \
    '{"stats":{"completed":5,"total":10},"userName":"Test","userId":1}' \
    true

# Test 6: Arrays of primitives
run_test 6 "Arrays of primitives" \
    '{"numbers":[3,1,4,1,5,9],"letters":["c","a","b"]}' \
    '{"letters":["b","c","a"],"numbers":[9,5,4,3,1,1]}' \
    true

# Test 7: Null and boolean values
run_test 7 "Null and boolean values" \
    '{"active":true,"deleted":null,"count":0}' \
    '{"count":0,"active":true,"deleted":null}' \
    true

# Test 8: Empty arrays and objects
run_test 8 "Empty arrays and objects" \
    '{"items":[],"meta":{}}' \
    '{"meta":{},"items":[]}' \
    true

# Test 9: Mixed nesting levels
run_test 9 "Mixed nesting levels" \
    '{"a":{"b":{"c":[{"x":1},{"x":2}]}},"d":[1,2,3]}' \
    '{"d":[3,2,1],"a":{"b":{"c":[{"x":2},{"x":1}]}}}' \
    true

# Test 10: Should NOT match - different values
run_test 10 "Different values (negative test)" \
    '{"name":"John","age":30}' \
    '{"name":"Jane","age":30}' \
    false

# Test 11: Should NOT match - different array contents
run_test 11 "Different array contents (negative test)" \
    '{"items":["apple","banana"]}' \
    '{"items":["apple","cherry"]}' \
    false

# Test 12: Should NOT match - different structure
run_test 12 "Different structure (negative test)" \
    '{"user":{"name":"Test"}}' \
    '{"user":{"name":"Test","age":25}}' \
    false

# Test 13: Unicode and special characters
run_test 13 "Unicode and special characters" \
    '{"message":"Hello 世界","emoji":"🚀"}' \
    '{"emoji":"🚀","message":"Hello 世界"}' \
    true

# Test 14: Numbers as strings vs numbers
run_test 14 "String vs number (negative test)" \
    '{"id":"123"}' \
    '{"id":123}' \
    false

# Test 15: Array with duplicate elements
run_test 15 "Arrays with duplicates" \
    '{"tags":["a","b","a","c","b"]}' \
    '{"tags":["c","b","a","b","a"]}' \
    true

echo "================================================"
echo "  Test Summary"
echo "================================================"
echo "Total tests: $((PASSED + FAILED))"
echo "Passed: $PASSED"
echo "Failed: $FAILED"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo "✓ All tests passed!"
    exit 0
else
    echo ""
    echo "✗ Some tests failed"
    exit 1
fi
