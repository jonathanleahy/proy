#!/bin/bash

FILE="$1"
ACTION="$2"

usage() {
  echo "Usage: $0 <filename> add|remove"
  echo "  add    - Prepends http://localhost:8099/proxy?target= to http(s):// URLs not already proxied"
  echo "  remove - Removes http://localhost:8099/proxy?target= or http://localhost:8099/proxy?target= from URLs"
  echo ""
  echo "Examples:"
  echo "  sh script.sh internal/app/infrastructure/env/env.go add"
  echo "  sh script.sh internal/app/infrastructure/env/env.go remove"
  exit 1
}

if [ -z "$FILE" ] || [ -z "$ACTION" ]; then
  usage
fi

if [ "$ACTION" = "add" ]; then
  # Only add proxy prefix to URLs that don't already have it
  # This prevents creating recursive proxy chains
  perl -pi -e 's#(["'\''`])(https?://(?!localhost:8099/proxy|0\.0\.0\.0:8099/proxy))#\1http://localhost:8099/proxy?target=\2#g' "$FILE"
elif [ "$ACTION" = "remove" ]; then
  # Remove proxy prefix - handles both localhost and 0.0.0.0, with or without https
  # This will remove all layers of proxy wrapping
  perl -pi -e 's#https?://(?:0\.0\.0\.0|localhost):8099/proxy\?target=##g' "$FILE"
else
  usage
fi
