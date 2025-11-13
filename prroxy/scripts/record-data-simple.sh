#!/bin/bash

# Simple script to record JSONPlaceholder test data

set -e

PROXY_URL="${PROXY_URL:-http://0.0.0.0:8080}"
TOTAL=0
SUCCESS=0
FAILED=0

echo "================================================"
echo "  Recording JSONPlaceholder Test Data"
echo "================================================"
echo ""

# Check proxy
echo -n "Checking proxy... "
if curl -s "${PROXY_URL}/health" > /dev/null; then
    echo "✓"
else
    echo "✗ Proxy not running at $PROXY_URL"
    exit 1
fi

# Set record mode
echo -n "Setting record mode... "
curl -s -X POST "${PROXY_URL}/admin/mode" \
    -H "Content-Type: application/json" \
    -d '{"mode":"record"}' > /dev/null
echo "✓"
echo ""

# Record function
record() {
    local url=$1
    local desc=$2
    TOTAL=$((TOTAL + 1))

    if curl -s -m 10 -o /dev/null -w "" "${PROXY_URL}/proxy?target=${url}"; then
        SUCCESS=$((SUCCESS + 1))
        echo "  ✓ $desc"
    else
        FAILED=$((FAILED + 1))
        echo "  ✗ $desc"
    fi
}

# Record users
echo "Recording Users (1-10)..."
for i in {1..10}; do
    record "jsonplaceholder.typicode.com/users/$i" "User $i"
done
echo ""

# Record posts by user
echo "Recording Posts by User (1-10)..."
for i in {1..10}; do
    record "jsonplaceholder.typicode.com/posts?userId=$i" "Posts for User $i"
done
echo ""

# Record todos by user
echo "Recording Todos by User (1-10)..."
for i in {1..10}; do
    record "jsonplaceholder.typicode.com/todos?userId=$i" "Todos for User $i"
done
echo ""

# Record individual posts
echo "Recording Individual Posts (1-100)..."
for i in {1..100}; do
    record "jsonplaceholder.typicode.com/posts/$i" "Post $i" 2>&1 | grep -v "✓" || true
    if [ $((i % 10)) -eq 0 ]; then
        echo "  Progress: $i/100"
    fi
done
echo "  ✓ Completed 100 posts"
echo ""

# Record individual todos
echo "Recording Individual Todos (1-200)..."
for i in {1..200}; do
    record "jsonplaceholder.typicode.com/todos/$i" "Todo $i" 2>&1 | grep -v "✓" || true
    if [ $((i % 20)) -eq 0 ]; then
        echo "  Progress: $i/200"
    fi
done
echo "  ✓ Completed 200 todos"
echo ""

# Summary
echo "================================================"
echo "  Summary"
echo "================================================"
echo "Total requests:  $TOTAL"
echo "Successful:      $SUCCESS"
echo "Failed:          $FAILED"
echo ""

if [ -d "recordings" ]; then
    echo "Recordings saved to: $(pwd)/recordings/"
    echo ""
fi

echo "Next steps:"
echo "  1. View dashboard: http://0.0.0.0:8080/admin/ui"
echo "  2. Switch to playback mode"
echo "  3. Test REST v1 with recorded data"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✓ All recordings completed successfully!"
else
    echo "⚠ Some recordings failed"
    exit 1
fi
