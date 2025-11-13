#!/bin/bash

# Run tests in RECORD mode
# This captures external API calls and stores them in recordings/
# Use this the first time or when you want to refresh cached data
# Usage: ./test-record.sh [config_file]
#   config_file: optional (default: from env CONFIG_FILE)

# Get the directory where this script is located
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

TEST_CONFIG="${1:-$CONFIG_FILE}"

echo "🔴 RECORD MODE - Capturing external API calls"
echo "Config: $TEST_CONFIG"
echo ""

./run-tests.sh record "$TEST_CONFIG"
