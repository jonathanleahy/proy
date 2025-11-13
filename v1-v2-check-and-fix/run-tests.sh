#!/bin/bash

# Generic test runner - used by test-record.sh and test-playback.sh
# Usage: ./run-tests.sh <mode> [config_file]
#   mode: record or playback
#   config_file: optional config file (default: config.person-lookup.json)

MODE=$1
CONFIG_FILE="${2:-config.person-lookup.json}"

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

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

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
echo "Using config: $CONFIG_FILE"
./run-reporter.sh "$CONFIG_FILE"
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
