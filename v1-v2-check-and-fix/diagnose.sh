#!/bin/bash

echo "=== Diagnostic Report ==="
echo ""

echo "1. Environment Check:"
echo "   OS: $(uname -s)"
echo "   Working directory: $(pwd)"
echo ""

echo "2. Services Status:"
lsof -i:3002 -i:3004 -i:3006 -i:8099 2>/dev/null || echo "   No services running on expected ports"
echo ""

echo "3. Service Health Check Links:"
echo "   Click these links to verify services are running:"
echo ""
echo "   REST v1 (port 3002):"
echo "   → http://localhost:3002/api/user/1"
echo "   → http://localhost:3002/api/person?surname=Thompson&dob=1985-03-15"
echo ""
echo "   REST v2 (port 3004):"
echo "   → http://localhost:3004/api/user/1"
echo "   → http://localhost:3004/api/person?surname=Thompson&dob=1985-03-15"
echo ""
echo "   REST External User (port 3006) - Person Data Service:"
echo "   → http://localhost:3006/health"
echo "   → http://localhost:3006/person?surname=Thompson&dob=1985-03-15"
echo ""
echo "   Proxy (port 8099):"
echo "   → http://localhost:8099/proxy?target=http%3A%2F%2F0.0.0.0%3A3006%2Fperson%3Fsurname%3DThompson%26dob%3D1985-03-15"
echo ""

echo "4. Recordings Check:"
RECORDING_COUNT=$(find recordings -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
echo "   Total recordings: $RECORDING_COUNT"
if [ "$RECORDING_COUNT" -gt 0 ]; then
    echo "   Sample recordings:"
    find recordings -name "*.json" 2>/dev/null | head -5
fi
echo ""

echo "5. Recent Report Summary:"
if [ -f reports/report_*.md ]; then
    LATEST_REPORT=$(ls -t reports/report_*.md | head -1)
    echo "   Latest report: $LATEST_REPORT"
    echo ""
    grep -E "^(- Total:|- Passing:|- Failing:)" "$LATEST_REPORT" 2>/dev/null || echo "   Could not parse report"
    echo ""
    echo "   Failing endpoints:"
    grep "❌" "$LATEST_REPORT" 2>/dev/null | head -10 || echo "   No failures found or report format unexpected"
else
    echo "   No reports found"
fi
echo ""

echo "6. Log File Errors (last 10):"
echo "   REST v1 errors:"
grep -i "error" tmp/rest-v1.log 2>/dev/null | tail -5 || echo "   No errors or log not found"
echo ""
echo "   REST v2 errors:"
grep -i "error" tmp/rest-v2.log 2>/dev/null | tail -5 || echo "   No errors or log not found"
echo ""
echo "   Proxy errors:"
grep -i "error" tmp/proxy.log 2>/dev/null | tail -5 || echo "   No errors or log not found"
echo ""

echo "7. Binary Check:"
echo "   Reporter binary: $([ -f ../reporter/reporter ] && echo "EXISTS" || echo "MISSING")"
echo "   Proxy binary: $([ -f ../prroxy/proxy/proxy-bin ] && echo "EXISTS" || echo "MISSING")"
echo "   REST v2 binary: $([ -f ../prroxy/rest-v2/rest-v2 ] && echo "EXISTS" || echo "MISSING")"
echo ""

echo "8. Test endpoints via curl:"
echo "   Testing GET /api/person?surname=Thompson&dob=1985-03-15"
echo "   V1 Response:"
curl -s "http://0.0.0.0:3002/api/person?surname=Thompson&dob=1985-03-15" 2>/dev/null | head -c 200
echo ""
echo "   V2 Response:"
curl -s "http://0.0.0.0:3004/api/person?surname=Thompson&dob=1985-03-15" 2>/dev/null | head -c 200
echo ""
echo ""

echo "=== End Diagnostic Report ==="
