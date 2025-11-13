#!/bin/bash
# Shutdown script for REST v2

echo "Stopping REST v2 on port 3004..."
pkill -f "PORT=3004.*rest-v2" || true

# Wait a moment and check if port is still in use
sleep 1
if lsof -i :3004 -t >/dev/null 2>&1; then
    echo "Port 3004 still in use, force killing..."
    lsof -i :3004 -t | xargs -r kill -9
    sleep 1
fi

echo "REST v2 stopped"
