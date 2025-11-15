#!/bin/bash

# shutdown.sh - Stop REST API v2 server
# This script gracefully stops the running server

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${GREEN}=== REST API v2 - Stopping Server ===${NC}"

# Check if PID file exists
if [ ! -f .server.pid ]; then
    echo -e "${RED}No PID file found. Server may not be running.${NC}"
    exit 1
fi

# Read PID
SERVER_PID=$(cat .server.pid)

# Check if process is running
if ! ps -p $SERVER_PID > /dev/null 2>&1; then
    echo -e "${RED}Server (PID: $SERVER_PID) is not running.${NC}"
    rm -f .server.pid
    exit 1
fi

# Kill the process
echo -e "${GREEN}Stopping server (PID: $SERVER_PID)...${NC}"
kill $SERVER_PID

# Wait for process to terminate
for i in {1..10}; do
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server stopped successfully${NC}"
        rm -f .server.pid
        exit 0
    fi
    sleep 0.5
done

# Force kill if still running
echo -e "${RED}Server did not stop gracefully. Force killing...${NC}"
kill -9 $SERVER_PID 2>/dev/null || true
rm -f .server.pid

echo -e "${GREEN}✓ Server terminated${NC}"
