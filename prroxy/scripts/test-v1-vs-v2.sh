#!/bin/bash

# Script to test and compare REST v1 vs v2

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REST_V1_URL="http://0.0.0.0:3002"
REST_V2_URL="http://0.0.0.0:3004"
PROXY_URL="http://0.0.0.0:8080"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  REST v1 vs v2 Comparison Test${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check services
echo -e "${BLUE}Checking services...${NC}"
echo ""

# Check proxy
echo -n "Proxy status: "
PROXY_STATUS=$(curl -s "${PROXY_URL}/admin/status")
MODE=$(echo "$PROXY_STATUS" | jq -r '.mode')
if [ "$MODE" = "playback" ]; then
    echo -e "${GREEN}✓ (playback mode)${NC}"
else
    echo -e "${RED}✗ (mode: $MODE)${NC}"
fi

# Check REST v1
echo -n "REST v1 health: "
if curl -s "${REST_V1_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

# Check REST v2
echo -n "REST v2 health: "
if curl -s "${REST_V2_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

echo ""
echo -e "${BLUE}Testing endpoints...${NC}"
echo ""

# Test user endpoint
echo "Testing GET /api/user/1:"
echo "------------------------"

echo "REST v1 response:"
V1_RESPONSE=$(curl -s "${REST_V1_URL}/api/user/1")
echo "$V1_RESPONSE" | jq '{id, name, username}' 2>/dev/null || echo "$V1_RESPONSE"

echo ""
echo "REST v2 response:"
V2_RESPONSE=$(curl -s "${REST_V2_URL}/api/user/1")
echo "$V2_RESPONSE" | jq '{id, name, username}' 2>/dev/null || echo "$V2_RESPONSE"

echo ""
echo -e "${BLUE}Direct proxy test:${NC}"
echo "Testing proxy directly with same URL format:"
DIRECT_TEST=$(curl -s "${PROXY_URL}/proxy?target=jsonplaceholder.typicode.com/users/1")
echo "$DIRECT_TEST" | jq '{id, name, username}' 2>/dev/null || echo "$DIRECT_TEST"

echo ""
echo -e "${BLUE}Proxy Statistics:${NC}"
curl -s "${PROXY_URL}/admin/status" | jq '{mode, playback_hits, playback_misses, record_count}'