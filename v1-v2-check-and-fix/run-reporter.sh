#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from env file
if [ -f "$SCRIPT_DIR/env" ]; then
    set -a  # automatically export all variables
    source "$SCRIPT_DIR/env"
    set +a  # disable automatic export
else
    echo "Error: env file not found. Please copy env.example to env and configure."
    exit 1
fi

# Configuration
TEST_CONFIG="${1:-$SCRIPT_DIR/$CONFIG_FILE}"
REPORTS_DIR="$SCRIPT_DIR/$REPORTS_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORTS_DIR/report_${TIMESTAMP}.md"

# Create reports directory if it doesn't exist
mkdir -p "$REPORTS_DIR"

# Check if reporter binary exists
if [ ! -f "$REPORTER_BIN" ]; then
    echo "Error: Reporter binary not found at $REPORTER_BIN"
    echo "Please build the reporter first:"
    echo "  cd ../reporter && go build -o reporter ./cmd/reporter"
    exit 1
fi

# Check if config file exists
if [ ! -f "$TEST_CONFIG" ]; then
    echo "Error: Config file not found: $TEST_CONFIG"
    echo "Usage: $0 [config_file]"
    echo ""
    echo "Available configs:"
    echo "  config.user-endpoints.json    - Tests external API calls (default)"
    echo "  config.person-lookup.json     - Tests internal endpoints"
    exit 1
fi

echo "=== Running Reporter ==="
echo "Config: $TEST_CONFIG"
echo "Output: $REPORT_FILE"
echo ""

# Run the reporter with individual endpoint reports enabled
"$REPORTER_BIN" -config "$TEST_CONFIG" -output-dir "$REPORTS_DIR" | tee "$REPORT_FILE"

echo ""
echo "=== Report saved to: $REPORT_FILE ==="
echo ""
echo "Recordings stored in: $SCRIPT_DIR/$RECORDINGS_DIR/"
echo "All reports in: $REPORTS_DIR/"
echo ""

# Check if there are any failures and remind to follow the fix process
FAILURES=$(grep -o "Failing: [0-9]*" "$REPORT_FILE" | grep -o "[0-9]*")
if [ ! -z "$FAILURES" ] && [ "$FAILURES" -gt 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "⚠️  FOUND $FAILURES FAILING ENDPOINT(S)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📋 To fix failing endpoints, READ THIS FIRST:"
    echo "   👉 FIX-PROCESS.md"
    echo ""
    echo "   This document contains the MANDATORY step-by-step"
    echo "   process for fixing endpoints (TDD workflow)."
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi
