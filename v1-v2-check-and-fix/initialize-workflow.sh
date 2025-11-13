#!/bin/bash

# Initialize Workflow - Ensures recordings are captured before testing
# This script handles the initial setup and recording phase

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

echo "🚀 Initializing API Validation Workflow"
echo "========================================"

# Step 0: Always clean up for a fresh start
echo "🧹 Cleaning up for fresh initialization..."
echo "   Removing tmp/ folder..."
rm -rf "$TMP_DIR"
echo "   Removing reports/ folder..."
rm -rf "$REPORTS_DIR"
echo "   Removing recordings/ folder..."
rm -rf "$RECORDINGS_DIR"
echo ""

# Step 1: Start services in record mode to capture v1 behavior
echo "📹 Starting recording phase to capture v1 API behavior..."
echo "   This will make real API calls to build our 'ground truth'"
echo ""

echo "🔴 Starting services in RECORD mode..."
PROXY_MODE=record ./start.sh

# Wait for services to be ready
echo "⏳ Waiting for services to initialize..."
sleep 5

# Step 2: Run the comprehensive test to capture behavior
echo ""
echo "🧪 Running comprehensive tests to capture v1 behavior..."
echo "   (This may take a few minutes)"
./run-reporter.sh config.comprehensive.json --max-failures 0

# Step 3: Verify recordings were created
echo ""
echo "📊 Verifying recordings were captured..."
RECORDING_COUNT=$(find "$RECORDINGS_DIR" -name "*.json" 2>/dev/null | wc -l)
if [ "$RECORDING_COUNT" -gt 0 ]; then
    echo "✅ Successfully captured $RECORDING_COUNT recordings!"
    echo ""
    echo "🎉 Initialization complete! You can now run tests in playback mode."
    echo "   Next step: ./run-reporter.sh config.comprehensive.json"
else
    echo "❌ Warning: No recordings were captured."
    echo "   Check the logs in $TMP_DIR folder for issues."
    echo "   You may need to investigate why the recording didn't work."
    exit 1
fi

echo ""
echo "🎯 Ready for development! The workflow is now initialized."
echo "   All v1 behavior has been recorded and you can start fixing v2 endpoints."