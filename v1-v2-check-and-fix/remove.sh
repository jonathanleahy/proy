#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from env file
if [ -f "$SCRIPT_DIR/env" ]; then
    export $(grep -v '^#' "$SCRIPT_DIR/env" | xargs)
else
    echo "Error: env file not found. Please copy env.example to env and configure."
    exit 1
fi

# Make paths relative to SCRIPT_DIR
TMP_DIR="$SCRIPT_DIR/$TMP_DIR"

# Cleanup script for rest-v1 and rest-v2
echo "Removing proxy configuration from rest-v1 and rest-v2..."

../utils/add-target.sh "$REST_V1_CONFIG" remove 2>/dev/null || true
../utils/add-target.sh "$REST_V2_CONFIG" remove 2>/dev/null || true

# Kill any running processes on the default ports
echo "Killing processes on ports $REST_V1_PORT, $REST_V2_PORT, and $REST_EXTERNAL_USER_PORT..."
lsof -ti:$REST_V1_PORT | xargs kill -9 2>/dev/null || true
lsof -ti:$REST_V2_PORT | xargs kill -9 2>/dev/null || true
lsof -ti:$REST_EXTERNAL_USER_PORT | xargs kill -9 2>/dev/null || true

# Kill processes by PID if files exist
if [ -f "$TMP_DIR/proxy.pid" ]; then
    PROXY_PID=$(cat "$TMP_DIR/proxy.pid")
    kill $PROXY_PID 2>/dev/null || true
    rm "$TMP_DIR/proxy.pid"
fi

if [ -f "$TMP_DIR/rest-v1.pid" ]; then
    REST_V1_PID=$(cat "$TMP_DIR/rest-v1.pid")
    kill $REST_V1_PID 2>/dev/null || true
    rm "$TMP_DIR/rest-v1.pid"
fi

if [ -f "$TMP_DIR/rest-v2.pid" ]; then
    REST_V2_PID=$(cat "$TMP_DIR/rest-v2.pid")
    kill $REST_V2_PID 2>/dev/null || true
    rm "$TMP_DIR/rest-v2.pid"
fi

if [ -f "$TMP_DIR/rest-external-user.pid" ]; then
    REST_EXTERNAL_USER_PID=$(cat "$TMP_DIR/rest-external-user.pid")
    kill $REST_EXTERNAL_USER_PID 2>/dev/null || true
    rm "$TMP_DIR/rest-external-user.pid"
fi

# Also kill proxy on its configured port
lsof -ti:$PROXY_PORT | xargs kill -9 2>/dev/null || true

echo "Cleanup complete."
