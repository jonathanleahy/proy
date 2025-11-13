#!/bin/bash

# Run tests in PLAYBACK mode
# This uses cached responses from recordings/ - no external API calls
# Use this after recording to run tests faster with deterministic data
# Usage: ./test-playback.sh [config_file]
#   config_file: optional (default: from env CONFIG_FILE)

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from env file
if [ -f "$SCRIPT_DIR/env" ]; then
    set -a  # automatically export all variables
    source "$SCRIPT_DIR/env"
    set +a  # disable automatic export
fi

TEST_CONFIG="${1:-$CONFIG_FILE}"

echo "▶️  PLAYBACK MODE - Using cached API responses"
echo "Config: $TEST_CONFIG"
echo ""

./run-tests.sh playback "$TEST_CONFIG"
