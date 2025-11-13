#!/bin/bash

# Start script for external user service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PORT=${PORT:-3006}

echo "Starting rest-external-user service on port $PORT..."

# Build if binary doesn't exist
if [ ! -f "./rest-external-user" ]; then
    echo "Building Go application..."
    go build -o rest-external-user ./cmd/server
fi

# Start the server in background
PORT=$PORT ./rest-external-user &
