#!/bin/bash

# start.sh - Start REST API v2 server
# This script builds and starts the REST API v2 server

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default port
PORT="${PORT:-3004}"

echo -e "${BLUE}=== REST API v2 - Starting Server ===${NC}"

# Build the server
echo -e "${GREEN}Building server...${NC}"
go build -o bin/server ./cmd/server

# Start the server in background
echo -e "${GREEN}Starting server on port $PORT...${NC}"
PORT=$PORT ./bin/server &

# Save PID
SERVER_PID=$!
echo $SERVER_PID > .server.pid

echo -e "${GREEN}✓ Server started (PID: $SERVER_PID)${NC}"
echo -e "${BLUE}Server is running on http://0.0.0.0:$PORT${NC}"
echo -e "  Health: http://0.0.0.0:$PORT/health"
echo -e "  User:   http://0.0.0.0:$PORT/api/user/1"
echo -e "  Person: http://0.0.0.0:$PORT/api/person?surname=Thompson&dob=1985-03-15"
echo -e ""
echo -e "To stop: ./shutdown.sh"
