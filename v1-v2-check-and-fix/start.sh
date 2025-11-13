#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Preserve CLI overrides so we can re-apply after sourcing env
CLI_PROXY_MODE="${PROXY_MODE:-}"
CLI_CLEAN_RECORDINGS="${CLEAN_RECORDINGS:-}"

# Load environment variables from env file
if [ -f "$SCRIPT_DIR/env" ]; then
    set -a  # automatically export all variables
    source "$SCRIPT_DIR/env"
    set +a  # disable automatic export
else
    echo "Error: env file not found. Please copy env.example to env and configure."
    exit 1
fi

# Make paths relative to SCRIPT_DIR
RECORDINGS_DIR="$SCRIPT_DIR/$RECORDINGS_DIR"
REPORTS_DIR="$SCRIPT_DIR/$REPORTS_DIR"
TMP_DIR="$SCRIPT_DIR/$TMP_DIR"

# Force playback mode by default for start.sh (CLI can override)
if [ -n "$CLI_PROXY_MODE" ]; then
    PROXY_MODE="$CLI_PROXY_MODE"
else
    PROXY_MODE="playback"
fi

# Re-apply CLI override for clean recordings if provided
if [ -n "$CLI_CLEAN_RECORDINGS" ]; then
    CLEAN_RECORDINGS="$CLI_CLEAN_RECORDINGS"
fi

# Build all required binaries and dependencies
echo "=== Building dependencies ==="

# Build reporter binary (needed for testing)
if [ ! -f "$REPORTER_BIN" ] || [ "$REPORTER_BIN" -ot "$(dirname "$REPORTER_BIN")/cmd/reporter/main.go" ]; then
    echo "Building reporter binary..."
    REPORTER_DIR="$(dirname "$REPORTER_BIN")"
    if [ -f "$REPORTER_DIR/go.mod" ]; then
        (cd "$REPORTER_DIR" && go build -o reporter ./cmd/reporter)
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build reporter binary"
            exit 1
        fi
        echo "✅ Reporter binary built"
    fi
fi

# Build proxy binary (if not already built or source changed)
PROXY_BIN_DIR="$PRROXY_BASE/proxy"
PROXY_BIN="$PROXY_BIN_DIR/proxy-bin"
if [ ! -f "$PROXY_BIN" ] || [ "$PROXY_BIN" -ot "$PROXY_BIN_DIR/cmd/proxy/main.go" ]; then
    echo "Building proxy binary..."
    if [ -f "$PROXY_BIN_DIR/go.mod" ]; then
        (cd "$PROXY_BIN_DIR" && go build -o proxy-bin ./cmd/proxy)
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build proxy binary"
            exit 1
        fi
        echo "✅ Proxy binary built"
    fi
fi

# Install Node.js dependencies if needed
if [ -f "$REST_V1_DIR/package.json" ]; then
    if [ ! -d "$REST_V1_DIR/node_modules" ]; then
        echo "Installing Node.js dependencies for rest-v1..."
        (cd "$REST_V1_DIR" && npm install --silent)
        if [ $? -ne 0 ]; then
            echo "❌ Failed to install rest-v1 dependencies"
            exit 1
        fi
        echo "✅ REST v1 dependencies installed"
    fi
fi

echo ""

# Clean up reports for fresh test results (keep recordings for reuse)
echo "=== Cleaning up old data ==="
if [ -d "$REPORTS_DIR" ]; then
    echo "Removing old reports..."
    rm -rf "$REPORTS_DIR"
fi

# Create fresh directories (preserve existing recordings)
mkdir -p "$RECORDINGS_DIR"
mkdir -p "$REPORTS_DIR"
mkdir -p "$TMP_DIR"

# Set clean recordings flag (default false)
CLEAN_RECORDINGS="${CLEAN_RECORDINGS:-false}"

# Clean up recordings if explicitly requested
if [ "$CLEAN_RECORDINGS" = "true" ]; then
    echo "=== Removing old recordings (CLEAN_RECORDINGS=true) ==="
    if [ -d "$RECORDINGS_DIR" ]; then
        rm -rf "$RECORDINGS_DIR"
        mkdir -p "$RECORDINGS_DIR"
    fi
fi

echo "=== Starting services ==="
echo "Proxy mode: $PROXY_MODE (playback by default, use PROXY_MODE=record ./start.sh to record)"
echo "Starting with fresh reports, preserving existing recordings"

./remove.sh

# point rest-v1 and rest-v2 to the 'proxy' target URLs
../utils/add-target.sh "$REST_V1_CONFIG" add
../utils/add-target.sh "$REST_V2_CONFIG" add

# Rebuild rest-v1 after modifying source
echo "Rebuilding rest-v1 with proxy configuration..."
cd "$REST_V1_DIR"
npm run build > /dev/null 2>&1
cd "$SCRIPT_DIR"

# Rebuild rest-v2 if needed
echo "Rebuilding rest-v2 with proxy configuration..."
cd "$REST_V2_DIR"
go build -o rest-v2 ./cmd/server > /dev/null 2>&1
cd "$SCRIPT_DIR"

# Convert relative paths to absolute paths
PROXY_DIR_ABS="$(realpath "$PROXY_DIR")"
REST_V1_DIR_ABS="$(realpath "$REST_V1_DIR")"
REST_V2_DIR_ABS="$(realpath "$REST_V2_DIR")"
REST_EXTERNAL_USER_DIR_ABS="$(realpath "$REST_EXTERNAL_USER_DIR")"
RECORDINGS_DIR_ABS="$(realpath "$RECORDINGS_DIR")"

# Start rest-external-user first (needed by rest-v1)
echo "Starting rest-external-user on port $REST_EXTERNAL_USER_PORT..."
cd "$REST_EXTERNAL_USER_DIR_ABS"
PORT=$REST_EXTERNAL_USER_PORT ./start.sh > "$TMP_DIR/rest-external-user.log" 2>&1 &
REST_EXTERNAL_USER_PID=$!
echo "REST external-user started (PID: $REST_EXTERNAL_USER_PID)"

# Wait a moment for external service to start
sleep 2

# Start proxy in background
echo "Starting proxy in $PROXY_MODE mode with recordings in $RECORDINGS_DIR_ABS..."
cd "$PROXY_DIR_ABS"
MODE=$PROXY_MODE ./proxy-bin -recordings-dir="$RECORDINGS_DIR_ABS" > "$TMP_DIR/proxy.log" 2>&1 &
PROXY_PID=$!
echo "Proxy started (PID: $PROXY_PID)"

# Wait a moment for proxy to start
sleep 2

# Start rest-v1 in background
echo "Starting rest-v1 on port $REST_V1_PORT..."
cd "$REST_V1_DIR_ABS"
PORT=$REST_V1_PORT ./start.sh > "$TMP_DIR/rest-v1.log" 2>&1 &
REST_V1_PID=$!
echo "REST v1 started (PID: $REST_V1_PID)"

# Wait a moment
sleep 2

# Start rest-v2 in background
echo "Starting rest-v2 on port $REST_V2_PORT..."
cd "$REST_V2_DIR_ABS"
PORT=$REST_V2_PORT ./start.sh > "$TMP_DIR/rest-v2.log" 2>&1 &
REST_V2_PID=$!
echo "REST v2 started (PID: $REST_V2_PID)"

# Save PIDs
cd "$SCRIPT_DIR"
echo "$PROXY_PID" > "$TMP_DIR/proxy.pid"
echo "$REST_V1_PID" > "$TMP_DIR/rest-v1.pid"
echo "$REST_V2_PID" > "$TMP_DIR/rest-v2.pid"
echo "$REST_EXTERNAL_USER_PID" > "$TMP_DIR/rest-external-user.pid"

echo ""
echo "=== All services started ==="
echo "External: PID $REST_EXTERNAL_USER_PID (log: tmp/rest-external-user.log)"
echo "Proxy:    PID $PROXY_PID (log: tmp/proxy.log)"
echo "REST v1:  PID $REST_V1_PID (log: tmp/rest-v1.log)"
echo "REST v2:  PID $REST_V2_PID (log: tmp/rest-v2.log)"
echo ""
echo "Waiting 5 seconds for services to initialize..."
sleep 5
echo "Ready to run tests!"
