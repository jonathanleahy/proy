#!/bin/bash

# Start script for source-of-truth API (REST v1)
# This API serves as the reference implementation

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PORT=${PORT:-3002}

echo "Starting source-of-truth API on port $PORT..."

# Check if dependencies are installed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Build if needed
if [ ! -d "dist" ]; then
    echo "Building TypeScript..."
    npm run build
fi

# Start the server
PORT=$PORT npm start