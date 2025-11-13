#!/bin/bash
# Shutdown all services

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "=== Shutting down all services ==="
"$SCRIPT_DIR/proxy/shutdown.sh"
"$SCRIPT_DIR/rest-v1/shutdown.sh"
"$SCRIPT_DIR/rest-v2/shutdown.sh"
echo "All services stopped"
