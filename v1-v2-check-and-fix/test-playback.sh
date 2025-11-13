#!/bin/bash

# Run tests in PLAYBACK mode
# This uses cached responses from recordings/ - no external API calls
# Use this after recording to run tests faster with deterministic data
# Usage: ./test-playback.sh [config_file]
#   config_file: optional (default: config.person-lookup.json)

CONFIG_FILE="${1:-config.person-lookup.json}"

echo "▶️  PLAYBACK MODE - Using cached API responses"
echo "Config: $CONFIG_FILE"
echo ""

./run-tests.sh playback "$CONFIG_FILE"
