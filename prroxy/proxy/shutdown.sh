#!/bin/bash
# Shutdown script for proxy

echo "Stopping proxy on port 8080..."
pkill -f "PORT=8080.*proxy" || pkill -f "PROXY.*8080" || true

# Wait a moment and check if port is still in use
sleep 1
if lsof -i :8080 -t >/dev/null 2>&1; then
    echo "Port 8080 still in use, force killing..."
    lsof -i :8080 -t | xargs -r kill -9
    sleep 1
fi

echo "Proxy stopped"
