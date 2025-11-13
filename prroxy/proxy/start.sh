#!/bin/bash

# Start script for proxy server (Go application)
# Records and replays external API interactions

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PORT=${PORT:-8099}
MODE=${MODE:-playback}

echo "Starting proxy server on port $PORT in $MODE mode..."

# Kill any process using the target port
if lsof -ti:$PORT > /dev/null 2>&1; then
    echo "Killing existing process on port $PORT..."
    lsof -ti:$PORT | xargs -r kill -9
    sleep 1
fi

# Remove old binary and rebuild to ensure latest changes
if [ -f "./proxy" ]; then
    echo "Removing old proxy binary..."
    rm ./proxy
fi

echo "Building proxy binary..."
go build -o proxy ./cmd/proxy

# Run the proxy
# Note: The Go proxy may have environment variables for MODE
# For now, just run the binary
PORT=$PORT ./proxy
