#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Preserve CLI overrides so we can re-apply after sourcing env
CLI_PROXY_MODE="${PROXY_MODE:-}"
CLI_CLEAN_RECORDINGS="${CLEAN_RECORDINGS:-}"

# Load environment variables from OS-specific env file
# Tries in order: env.<os> (e.g., env.linux, env.darwin), then env, then env.example
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ENV_FILE=""

if [ -f "$SCRIPT_DIR/env.$OS_NAME" ]; then
    ENV_FILE="$SCRIPT_DIR/env.$OS_NAME"
    echo "Using OS-specific config: env.$OS_NAME"
elif [ -f "$SCRIPT_DIR/env" ]; then
    ENV_FILE="$SCRIPT_DIR/env"
    echo "Using default config: env"
else
    echo "Error: No env file found for $OS_NAME"
    echo "Please create one of: env.$OS_NAME, env (copy from env.example)"
    exit 1
fi

set -a  # automatically export all variables
source "$ENV_FILE"
set +a  # disable automatic export

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
        (cd "$REPORTER_DIR" && go build -o reporter "$REPORTER_BUILD_PATH")
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
        (cd "$PROXY_BIN_DIR" && go build -o proxy-bin "$PROXY_BUILD_PATH")
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build proxy binary"
            exit 1
        fi
        echo "✅ Proxy binary built"
    fi
fi

# Build rest-external-user binary (if not already built or source changed)
REST_EXTERNAL_USER_BIN="$REST_EXTERNAL_USER_DIR/rest-external-user"
if [ ! -f "$REST_EXTERNAL_USER_BIN" ] || [ "$REST_EXTERNAL_USER_BIN" -ot "$REST_EXTERNAL_USER_DIR/cmd/server/main.go" ]; then
    echo "Building rest-external-user binary..."
    if [ -f "$REST_EXTERNAL_USER_DIR/go.mod" ]; then
        (cd "$REST_EXTERNAL_USER_DIR" && go build -o rest-external-user "$REST_EXTERNAL_USER_BUILD_PATH")
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build rest-external-user binary"
            exit 1
        fi
        echo "✅ REST external-user binary built"
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

# Rebuild rest-v1 after modifying source (unless skip flag is set)
if [ "${SKIP_BUILD_REST_V1:-false}" = "true" ]; then
    echo "⏭️  Skipping REST v1 build (SKIP_BUILD_REST_V1=true)"
else
    echo "Rebuilding rest-v1 with proxy configuration..."
    cd "$REST_V1_DIR"
    if eval $REST_V1_BUILD_COMMAND 2>&1 | tee "$TMP_DIR/rest-v1-build.log"; then
        echo "✅ REST v1 build successful"
    else
        echo "❌ REST v1 build failed! Check tmp/rest-v1-build.log for details"
        tail -20 "$TMP_DIR/rest-v1-build.log"
        exit 1
    fi
    cd "$SCRIPT_DIR"
fi

# Rebuild rest-v2 if needed (unless skip flag is set)
if [ "${SKIP_BUILD_REST_V2:-false}" = "true" ]; then
    echo "⏭️  Skipping REST v2 build (SKIP_BUILD_REST_V2=true)"
else
    echo "Rebuilding rest-v2 with proxy configuration..."
    cd "$REST_V2_DIR"
    if go build -o rest-v2 "$REST_V2_BUILD_PATH" 2>&1 | tee "$TMP_DIR/rest-v2-build.log"; then
        echo "✅ REST v2 build successful"
    else
        echo "❌ REST v2 build failed! Check tmp/rest-v2-build.log for details"
        tail -20 "$TMP_DIR/rest-v2-build.log"
        exit 1
    fi
    cd "$SCRIPT_DIR"
fi

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
if [ -n "$REST_V1_START_COMMAND" ]; then
    echo "Using custom start command: $REST_V1_START_COMMAND"
    PORT=$REST_V1_PORT eval $REST_V1_START_COMMAND > "$TMP_DIR/rest-v1.log" 2>&1 &
else
    PORT=$REST_V1_PORT ./start.sh > "$TMP_DIR/rest-v1.log" 2>&1 &
fi
REST_V1_PID=$!
echo "REST v1 started (PID: $REST_V1_PID)"

# Wait a moment
sleep 2

# Start rest-v2 in background
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

# Function to check if port is listening
check_port() {
    local port=$1
    local service=$2
    if lsof -i:$port -sTCP:LISTEN >/dev/null 2>&1; then
        echo "✅ $service is listening on port $port"
        return 0
    else
        echo "❌ $service is NOT listening on port $port"
        return 1
    fi
}

# Function to check if process is still running
check_process() {
    local pid=$1
    local service=$2
    if ps -p $pid > /dev/null 2>&1; then
        echo "✅ $service process is running (PID: $pid)"
        return 0
    else
        echo "❌ $service process has died (PID: $pid)"
        return 1
    fi
}

echo ""
echo "=== Verifying Services ==="

# Check processes
ALL_RUNNING=true
check_process $REST_EXTERNAL_USER_PID "REST external-user" || ALL_RUNNING=false
check_process $PROXY_PID "Proxy" || ALL_RUNNING=false
check_process $REST_V1_PID "REST v1" || ALL_RUNNING=false
check_process $REST_V2_PID "REST v2" || ALL_RUNNING=false

echo ""

# Check ports
check_port $REST_EXTERNAL_USER_PORT "REST external-user" || ALL_RUNNING=false
check_port $PROXY_PORT "Proxy" || ALL_RUNNING=false
check_port $REST_V1_PORT "REST v1" || ALL_RUNNING=false
check_port $REST_V2_PORT "REST v2" || ALL_RUNNING=false

echo ""

if [ "$ALL_RUNNING" = false ]; then
    echo "⚠️  WARNING: Some services failed to start properly!"
    echo ""
    echo "Check the logs for details:"
    echo "  tail -50 tmp/rest-external-user.log"
    echo "  tail -50 tmp/proxy.log"
    echo "  tail -50 tmp/rest-v1.log"
    echo "  tail -50 tmp/rest-v2.log"
    echo ""
    echo "Showing last 10 lines from each log:"
    echo ""
    echo "=== REST external-user log ==="
    tail -10 tmp/rest-external-user.log 2>/dev/null || echo "No log available"
    echo ""
    echo "=== Proxy log ==="
    tail -10 tmp/proxy.log 2>/dev/null || echo "No log available"
    echo ""
    echo "=== REST v1 log ==="
    tail -10 tmp/rest-v1.log 2>/dev/null || echo "No log available"
    echo ""
    echo "=== REST v2 log ==="
    tail -10 tmp/rest-v2.log 2>/dev/null || echo "No log available"
    echo ""

    # Check if we should exit or continue
    if [ "${STRICT_SERVICE_CHECK:-true}" = "true" ]; then
        echo "❌ Exiting due to service failures (set STRICT_SERVICE_CHECK=false in env to continue anyway)"
        exit 1
    else
        echo "⚠️  Continuing anyway (STRICT_SERVICE_CHECK=false)"
        echo "   Services that are running will work, others may fail during tests"
    fi
fi

echo "🎉 All services started successfully!"
echo ""
echo "=== Service URLs ==="
echo "REST v1:         http://localhost:$REST_V1_PORT"
echo "REST v2:         http://localhost:$REST_V2_PORT"
echo "External User:   http://localhost:$REST_EXTERNAL_USER_PORT"
echo "Proxy:           http://localhost:$PROXY_PORT (mode: $PROXY_MODE)"
echo "Proxy Admin UI:  http://localhost:$PROXY_PORT/admin/ui"
echo ""
echo "=== Example Endpoints ==="
echo "REST v1 health:  curl http://localhost:$REST_V1_PORT/health"
echo "REST v2 health:  curl http://localhost:$REST_V2_PORT/health"
echo "REST v1 user:    curl http://localhost:$REST_V1_PORT/api/user/1"
echo "REST v2 user:    curl http://localhost:$REST_V2_PORT/api/user/1"
echo ""
echo "Ready to run tests!"
echo ""
echo "ℹ️  Services are running in the background"
echo "   To stop all services: ./remove.sh"
echo "   To view logs: tail -f tmp/*.log"
echo ""
echo "   Press Ctrl+C to exit (services will keep running)"
echo ""

# Keep script running to prevent shell from killing background processes
# User can Ctrl+C to exit, services will continue running
trap 'echo ""; echo "Services still running. Use ./remove.sh to stop them."; exit 0' INT TERM

# Wait indefinitely (or until user presses Ctrl+C)
wait
