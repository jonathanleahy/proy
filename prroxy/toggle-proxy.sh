#!/bin/bash

# Toggle Proxy Configuration Script
# Usage: ./toggle-proxy.sh [add|remove] [port]
#
# This script toggles proxy configuration for CRM v1 and CRM v2 API env files
# - add: Wraps https:// URLs with proxy format
# - remove: Unwraps proxy format back to direct https:// URLs

set -e

# Configuration
PROXY_PORT="${2:-8099}"  # Default to 8099 if not provided
CRM_V1_ENV="/home/jon/personal/prr/ttt/crm-api/src/main/groovy/com/pismo/crm/config/Env.groovy"
CRM_V2_ENV="/home/jon/personal/prr/ttt/crm-v2-api/internal/app/infrastructure/env/env.go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Usage: $0 [add|remove] [port]"
    echo ""
    echo "Arguments:"
    echo "  add     - Add proxy wrapper to https:// URLs"
    echo "  remove  - Remove proxy wrapper from URLs"
    echo "  port    - Proxy port (default: 8099)"
    echo ""
    echo "Examples:"
    echo "  $0 add          # Add proxy on port 8099 (default)"
    echo "  $0 add 8088     # Add proxy on port 8088 (custom)"
    echo "  $0 remove       # Remove proxy wrapper"
    exit 1
}

# Validate arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Error: Missing argument${NC}"
    usage
fi

MODE=$1

# Validate mode
if [ "$MODE" != "add" ] && [ "$MODE" != "remove" ]; then
    echo -e "${RED}Error: Invalid mode '$MODE'. Must be 'add' or 'remove'${NC}"
    usage
fi

# Check if files exist
if [ ! -f "$CRM_V1_ENV" ]; then
    echo -e "${RED}Error: CRM v1 env file not found: $CRM_V1_ENV${NC}"
    exit 1
fi

if [ ! -f "$CRM_V2_ENV" ]; then
    echo -e "${RED}Error: CRM v2 env file not found: $CRM_V2_ENV${NC}"
    exit 1
fi

# Function to add proxy
add_proxy() {
    local file=$1
    local temp_file="${file}.tmp"

    echo -e "${YELLOW}Processing: $file${NC}"

    # Backup original file
    cp "$file" "${file}.backup"

    # Replace https:// with proxy format (but skip if already wrapped)
    # Pattern matches: 'https://' or "https://" that is NOT already preceded by proxy?target=
    sed -E "s|(['\"])https://([^'\"]*)|\\1http://0.0.0.0:${PROXY_PORT}/proxy?target=https://\\2|g" "$file" > "$temp_file"

    # Count changes
    changes=$(diff "$file" "$temp_file" | grep -c "^[<>]" || true)

    if [ "$changes" -gt 0 ]; then
        mv "$temp_file" "$file"
        echo -e "${GREEN}✓ Added proxy wrapper ($((changes / 2)) URLs modified)${NC}"
    else
        rm "$temp_file"
        echo -e "${YELLOW}⚠ No changes needed (URLs already proxied or no https:// URLs found)${NC}"
    fi
}

# Function to remove proxy
remove_proxy() {
    local file=$1
    local temp_file="${file}.tmp"

    echo -e "${YELLOW}Processing: $file${NC}"

    # Backup original file
    cp "$file" "${file}.backup"

    # Remove proxy wrapper from URLs
    # Pattern matches: http://0.0.0.0:PORT/proxy?target=https://ORIGINAL_URL
    sed -E "s|http://0.0.0.0:[0-9]+/proxy\?target=(https://[^'\"]*)|\\1|g" "$file" > "$temp_file"

    # Count changes
    changes=$(diff "$file" "$temp_file" | grep -c "^[<>]" || true)

    if [ "$changes" -gt 0 ]; then
        mv "$temp_file" "$file"
        echo -e "${GREEN}✓ Removed proxy wrapper ($((changes / 2)) URLs modified)${NC}"
    else
        rm "$temp_file"
        echo -e "${YELLOW}⚠ No changes needed (no proxy URLs found)${NC}"
    fi
}

# Main execution
echo ""
echo "======================================"
echo "  Proxy Configuration Toggle"
echo "======================================"
echo "Mode: $MODE"
echo "Proxy Port: $PROXY_PORT"
echo "======================================"
echo ""

if [ "$MODE" = "add" ]; then
    echo -e "${GREEN}Adding proxy wrapper to URLs...${NC}"
    echo ""
    add_proxy "$CRM_V1_ENV"
    echo ""
    add_proxy "$CRM_V2_ENV"
elif [ "$MODE" = "remove" ]; then
    echo -e "${GREEN}Removing proxy wrapper from URLs...${NC}"
    echo ""
    remove_proxy "$CRM_V1_ENV"
    echo ""
    remove_proxy "$CRM_V2_ENV"
fi

echo ""
echo "======================================"
echo -e "${GREEN}✓ Done!${NC}"
echo "======================================"
echo ""
echo "Backup files created:"
echo "  - ${CRM_V1_ENV}.backup"
echo "  - ${CRM_V2_ENV}.backup"
echo ""
echo "To restore from backup:"
echo "  cp ${CRM_V1_ENV}.backup $CRM_V1_ENV"
echo "  cp ${CRM_V2_ENV}.backup $CRM_V2_ENV"
echo ""
