#!/bin/bash

# Generic test runner - used by test-record.sh and test-playback.sh
# Usage: ./run-tests.sh <mode> [config_file]
#   mode: record or playback
#   config_file: optional config file (default: from env CONFIG_FILE)

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from OS-specific env file
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ENV_FILE=""

if [ -f "$SCRIPT_DIR/env.$OS_NAME" ]; then
    ENV_FILE="$SCRIPT_DIR/env.$OS_NAME"
elif [ -f "$SCRIPT_DIR/env" ]; then
    ENV_FILE="$SCRIPT_DIR/env"
fi

if [ -n "$ENV_FILE" ]; then
    set -a  # automatically export all variables
    source "$ENV_FILE"
    set +a  # disable automatic export
fi

MODE=$1
TEST_CONFIG="${2:-$CONFIG_FILE}"

if [ -z "$MODE" ]; then
    echo "Error: Mode not specified"
    echo "Usage: $0 <record|playback> [config_file]"
    exit 1
fi

if [ "$MODE" != "record" ] && [ "$MODE" != "playback" ]; then
    echo "Error: Invalid mode '$MODE'"
    echo "Usage: $0 <record|playback> [config_file]"
    exit 1
fi

echo "=========================================="
echo "Running tests in $MODE mode"
echo "=========================================="
echo ""

# Stop any existing services
echo "Step 1: Cleaning up any existing services..."
./remove.sh
echo ""

# Start services in specified mode
echo "Step 2: Starting services in $MODE mode..."
PROXY_MODE=$MODE ./start.sh
echo ""

# Wait a bit longer for services to fully initialize
echo "Waiting for services to stabilize..."
sleep 3
echo ""

# Run the reporter
echo "Step 3: Running comparison tests..."
echo "Using config: $TEST_CONFIG"
./run-reporter.sh "$TEST_CONFIG"
TEST_RESULT=$?
echo ""

# Show summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Mode: $MODE"
echo "Latest report: $(ls -t reports/*.md | head -1)"
echo ""

# Show quick stats from latest report
LATEST_REPORT=$(ls -t reports/*.md | head -1)
if [ -f "$LATEST_REPORT" ]; then
    echo "Results:"
    grep -E "^\*\*(Total|Matched|Failed)" "$LATEST_REPORT" | head -5
    echo ""
fi

# Keep services running or cleanup
read -p "Keep services running? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleaning up..."
    ./remove.sh
fi

exit $TEST_RESULT
