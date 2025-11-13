#!/bin/bash

# Start script for new implementation API (REST v2)
# This API is being migrated to match source-of-truth

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PORT=${PORT:-3004}

echo "Starting new implementation API on port $PORT..."

# Build if binary doesn't exist
if [ ! -f "./rest-v2" ]; then
    echo "Building Go application..."
    go build -o rest-v2 ./cmd/server
fi

# Start the server
PORT=$PORT ./rest-v2