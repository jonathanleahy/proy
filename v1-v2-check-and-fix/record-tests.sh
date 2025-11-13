#!/bin/bash

# Initialize Workflow - Ensures recordings are captured before testing
# This script handles the initial setup and recording phase

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Clean up services on Ctrl+C
trap 'echo ""; echo "🛑 Interrupted! Cleaning up services..."; ./remove.sh; exit 1' INT TERM

# Load environment variables from OS-specific env file
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ENV_FILE=""

if [ -f "$SCRIPT_DIR/env.$OS_NAME" ]; then
    ENV_FILE="$SCRIPT_DIR/env.$OS_NAME"
elif [ -f "$SCRIPT_DIR/env" ]; then
    ENV_FILE="$SCRIPT_DIR/env"
else
    echo "Error: No env file found for $OS_NAME"
    echo "Please create one of: env.$OS_NAME, env (copy from env.example)"
    exit 1
fi

set -a  # automatically export all variables
source "$ENV_FILE"
set +a  # disable automatic export

echo "🚀 Initializing API Validation Workflow"
echo "========================================"

# Make paths absolute
RECORDINGS_DIR="$SCRIPT_DIR/$RECORDINGS_DIR"
REPORTS_DIR="$SCRIPT_DIR/$REPORTS_DIR"
TMP_DIR="$SCRIPT_DIR/$TMP_DIR"

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

# Recreate directories
mkdir -p "$TMP_DIR"
mkdir -p "$REPORTS_DIR"
mkdir -p "$RECORDINGS_DIR"
echo ""

# Step 2: Start services in record mode to capture v1 behavior
echo "📹 Starting recording phase to capture v1 API behavior..."
echo "   This will make real API calls to build our 'ground truth'"
echo ""

# Force record mode
PROXY_MODE="record"

echo "🔴 Starting services in RECORD mode..."
./remove.sh

# point rest-v1 and rest-v2 to the 'proxy' target URLs
../utils/add-target.sh "$REST_V1_CONFIG" add
../utils/add-target.sh "$REST_V2_CONFIG" add

# Convert relative paths to absolute paths
PROXY_DIR_ABS="$(realpath "$PROXY_DIR")"
REST_V1_DIR_ABS="$(realpath "$REST_V1_DIR")"
REST_V2_DIR_ABS="$(realpath "$REST_V2_DIR")"
REST_EXTERNAL_USER_DIR_ABS="$(realpath "$REST_EXTERNAL_USER_DIR")"
RECORDINGS_DIR_ABS="$(realpath "$RECORDINGS_DIR")"

# Start rest-external-user first (needed by rest-v1)
if [ "${SKIP_REST_EXTERNAL_USER:-false}" = "true" ]; then
    echo "⏭️  Skipping REST external-user (SKIP_REST_EXTERNAL_USER=true)"
    REST_EXTERNAL_USER_PID=""
else
    echo "Starting rest-external-user on port $REST_EXTERNAL_USER_PORT..."
    cd "$REST_EXTERNAL_USER_DIR_ABS"
    PORT=$REST_EXTERNAL_USER_PORT ./start.sh > "$TMP_DIR/rest-external-user.log" 2>&1 &
    REST_EXTERNAL_USER_PID=$!
    echo "REST external-user started (PID: $REST_EXTERNAL_USER_PID)"
    sleep 2
fi

# Start proxy in background (RECORD MODE)
if [ "${SKIP_PROXY:-false}" = "true" ]; then
    echo "⏭️  Skipping Proxy (SKIP_PROXY=true)"
    PROXY_PID=""
else
    echo "Starting proxy in RECORD mode with recordings in $RECORDINGS_DIR_ABS..."
    cd "$PROXY_DIR_ABS"
    MODE=$PROXY_MODE ./proxy-bin -recordings-dir="$RECORDINGS_DIR_ABS" > "$TMP_DIR/proxy.log" 2>&1 &
    PROXY_PID=$!
    echo "Proxy started (PID: $PROXY_PID)"
    sleep 2
fi

# Start rest-v1 in background
if [ "${SKIP_REST_V1:-false}" = "true" ]; then
    echo "⏭️  Skipping REST v1 (SKIP_REST_V1=true)"
    REST_V1_PID=""
else
    echo "Starting rest-v1 on port $REST_V1_PORT..."
    cd "$REST_V1_DIR_ABS"
    if [ -n "$REST_V1_START_COMMAND" ]; then
        echo "Using custom start command: $REST_V1_START_COMMAND"
        PORT=$REST_V1_PORT eval $REST_V1_START_COMMAND > "$TMP_DIR/rest-v1.log" 2>&1 &
    else
        PORT=$REST_V1_PORT ./start.sh > "$TMP_DIR/rest-v1.log" 2>&1 &
    fi
    REST_V1_PID=$!
    echo "REST v1 started (PID: $REST_V1_PID)"
    sleep 2
fi

# Start rest-v2 in background
if [ "${SKIP_REST_V2:-false}" = "true" ]; then
    echo "⏭️  Skipping REST v2 (SKIP_REST_V2=true)"
    REST_V2_PID=""
else
    echo "Starting rest-v2 on port $REST_V2_PORT..."
    cd "$REST_V2_DIR_ABS"
    if [ -n "$REST_V2_START_COMMAND" ]; then
        echo "Using custom start command: $REST_V2_START_COMMAND"
        PORT=$REST_V2_PORT eval $REST_V2_START_COMMAND > "$TMP_DIR/rest-v2.log" 2>&1 &
    else
        PORT=$REST_V2_PORT ./start.sh > "$TMP_DIR/rest-v2.log" 2>&1 &
    fi
    REST_V2_PID=$!
    echo "REST v2 started (PID: $REST_V2_PID)"
fi

# Save PIDs
cd "$SCRIPT_DIR"
echo "$PROXY_PID" > "$TMP_DIR/proxy.pid"
echo "$REST_V1_PID" > "$TMP_DIR/rest-v1.pid"
echo "$REST_V2_PID" > "$TMP_DIR/rest-v2.pid"
echo "$REST_EXTERNAL_USER_PID" > "$TMP_DIR/rest-external-user.pid"

# Wait for services to be ready
echo "⏳ Waiting for services to initialize..."
sleep 5

# Step 3: Run the comprehensive test to capture behavior
echo ""
echo "🧪 Running comprehensive tests to capture v1 behavior..."
echo "   (This may take a few minutes)"
cd "$SCRIPT_DIR"
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
echo "   Services are still running - use ./remove.sh to stop them"
echo ""