#!/bin/bash
# Shutdown script for REST v1

echo "Stopping REST v1 on port 3002..."
pkill -f "PORT=3002.*npm" || true

# Wait a moment and check if port is still in use
sleep 1
if lsof -i :3002 -t >/dev/null 2>&1; then
    echo "Port 3002 still in use, force killing..."
    lsof -i :3002 -t | xargs -r kill -9
    sleep 1
fi

echo "REST v1 stopped"
