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

# Step 0: Build all required binaries and dependencies
echo "🔨 Building dependencies..."
echo ""

# Build reporter binary
echo "📦 Building reporter binary..."
REPORTER_DIR="$(dirname "$REPORTER_BIN")"
if [ -f "$REPORTER_DIR/go.mod" ]; then
    (cd "$REPORTER_DIR" && go build -o reporter "$REPORTER_BUILD_PATH")
    if [ $? -eq 0 ]; then
        echo "   ✅ Reporter binary built successfully"
    else
        echo "   ❌ Failed to build reporter binary"
        exit 1
    fi
else
    echo "   ❌ Reporter source not found at $REPORTER_DIR"
    exit 1
fi

# Build proxy binary
echo "📦 Building proxy binary..."
PROXY_BIN_DIR="$PRROXY_BASE/proxy"
if [ -f "$PROXY_BIN_DIR/go.mod" ]; then
    (cd "$PROXY_BIN_DIR" && go build -o proxy-bin "$PROXY_BUILD_PATH")
    if [ $? -eq 0 ]; then
        echo "   ✅ Proxy binary built successfully"
    else
        echo "   ❌ Failed to build proxy binary"
        exit 1
    fi
else
    echo "   ⚠️  Proxy source not found at $PROXY_BIN_DIR (skipping)"
fi

# Build rest-v2 binary
echo "📦 Building rest-v2 binary..."
if [ -f "$REST_V2_DIR/go.mod" ]; then
    (cd "$REST_V2_DIR" && go build -o rest-v2 "$REST_V2_BUILD_PATH")
    if [ $? -eq 0 ]; then
        echo "   ✅ REST v2 binary built successfully"
    else
        echo "   ❌ Failed to build rest-v2 binary"
        exit 1
    fi
else
    echo "   ⚠️  REST v2 source not found at $REST_V2_DIR (skipping)"
fi

# Build rest-external-user binary
echo "📦 Building rest-external-user binary..."
if [ -f "$REST_EXTERNAL_USER_DIR/go.mod" ]; then
    (cd "$REST_EXTERNAL_USER_DIR" && go build -o rest-external-user "$REST_EXTERNAL_USER_BUILD_PATH")
    if [ $? -eq 0 ]; then
        echo "   ✅ REST external-user binary built successfully"
    else
        echo "   ❌ Failed to build rest-external-user binary"
        exit 1
    fi
else
    echo "   ⚠️  REST external-user source not found at $REST_EXTERNAL_USER_DIR (skipping)"
fi

# Install Node.js dependencies for rest-v1
echo "📦 Installing Node.js dependencies for rest-v1..."
if [ -f "$REST_V1_DIR/package.json" ]; then
    (cd "$REST_V1_DIR" && npm install --silent)
    if [ $? -eq 0 ]; then
        echo "   ✅ REST v1 dependencies installed"
    else
        echo "   ❌ Failed to install rest-v1 dependencies"
        exit 1
    fi
else
    echo "   ⚠️  REST v1 package.json not found at $REST_V1_DIR (skipping)"
fi

echo ""
echo "✅ All dependencies built successfully!"
echo ""

# Step 1: Always clean up for a fresh start
echo "🧹 Cleaning up for fresh initialization..."
echo "   Removing tmp/ folder..."
rm -rf "$TMP_DIR"
echo "   Removing reports/ folder..."
rm -rf "$REPORTS_DIR"
echo "   Removing recordings/ folder..."
rm -rf "$RECORDINGS_DIR"
echo ""

# Step 2: Start services in record mode to capture v1 behavior
echo "📹 Starting recording phase to capture v1 API behavior..."
echo "   This will make real API calls to build our 'ground truth'"
echo ""

echo "🔴 Starting services in RECORD mode..."
PROXY_MODE=record ./start.sh

# Wait for services to be ready
echo "⏳ Waiting for services to initialize..."
sleep 5

# Step 3: Run the comprehensive test to capture behavior
echo ""
echo "🧪 Running comprehensive tests to capture v1 behavior..."
echo "   (This may take a few minutes)"
./run-reporter.sh "$CONFIG_FILE" --max-failures 0

# Step 4: Verify recordings were created
echo ""
echo "📊 Verifying recordings were captured..."
RECORDING_COUNT=$(find "$RECORDINGS_DIR" -name "*.json" 2>/dev/null | wc -l)
if [ "$RECORDING_COUNT" -gt 0 ]; then
    echo "✅ Successfully captured $RECORDING_COUNT recordings!"
    echo ""
    echo "🎉 Initialization complete! You can now run tests in playback mode."
    echo "   Next step: ./run-reporter.sh $CONFIG_FILE"
else
    echo "❌ Warning: No recordings were captured."
    echo "   Check the logs in $TMP_DIR folder for issues."
    echo "   You may need to investigate why the recording didn't work."
    exit 1
fi

echo ""
echo "🎯 Ready for development! The workflow is now initialized."
echo "   All v1 behavior has been recorded and you can start fixing v2 endpoints."