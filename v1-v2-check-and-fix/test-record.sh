#!/bin/bash

# Run tests in RECORD mode
# This captures external API calls and stores them in recordings/
# Use this the first time or when you want to refresh cached data
# Usage: ./test-record.sh [config_file]
#   config_file: optional (default: config.person-lookup.json)

CONFIG_FILE="${1:-config.person-lookup.json}"

echo "🔴 RECORD MODE - Capturing external API calls"
echo "Config: $CONFIG_FILE"
echo ""

./run-tests.sh record "$CONFIG_FILE"
