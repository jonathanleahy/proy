#!/bin/bash
# Development script to start all services

set -e

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "=== Shutting down existing services ==="
"$SCRIPT_DIR/proxy/shutdown.sh" 2>/dev/null || true
"$SCRIPT_DIR/rest-v1/shutdown.sh" 2>/dev/null || true
"$SCRIPT_DIR/rest-v2/shutdown.sh" 2>/dev/null || true

sleep 2

echo ""
echo "=== Starting services ==="

# Start proxy in playback mode (default)
PROXY_MODE=${PROXY_MODE:-playback}
echo "Starting proxy in $PROXY_MODE mode on port 8080..."
cd "$SCRIPT_DIR/proxy"
PROXY_MODE=$PROXY_MODE PORT=8080 ./proxy > proxy.log 2>&1 &
PROXY_PID=$!
echo "Proxy started (PID: $PROXY_PID)"

sleep 2

# Start REST v1 with proxy
echo "Starting REST v1 on port 3002 (using proxy)..."
cd "$SCRIPT_DIR/rest-v1"
PROXY_URL=http://0.0.0.0:8080/proxy PORT=3002 npm start > rest-v1.log 2>&1 &
REST_V1_PID=$!
echo "REST v1 started (PID: $REST_V1_PID)"

sleep 5

# Start REST v2 with proxy
echo "Starting REST v2 on port 3004 (using proxy)..."
cd "$SCRIPT_DIR/rest-v2"
PROXY_URL=http://0.0.0.0:8080/proxy PORT=3004 ./start.sh > rest-v2.log 2>&1 &
REST_V2_PID=$!
echo "REST v2 started (PID: $REST_V2_PID)"

sleep 2

echo ""
echo "=== All services started ==="
echo "Proxy:    http://0.0.0.0:8080 (mode: $PROXY_MODE)"
echo "REST v1:  http://0.0.0.0:3002"
echo "REST v2:  http://0.0.0.0:3004"
echo ""
echo "Logs:"
echo "  Proxy:   $SCRIPT_DIR/proxy/proxy.log"
echo "  REST v1: $SCRIPT_DIR/rest-v1/rest-v1.log"
echo "  REST v2: $SCRIPT_DIR/rest-v2/rest-v2.log"
echo ""
echo "To stop all services: ./shutdown-all.sh"
