#!/bin/bash

echo "=== Proxy Diagnostics ==="
echo ""

echo "1. Checking directory structure..."
ls -la web/ 2>/dev/null || echo "ERROR: web/ directory not found!"
echo ""

echo "2. Checking embed.go..."
cat web/embed.go 2>/dev/null || echo "ERROR: web/embed.go not found!"
echo ""

echo "3. Checking dashboard.html..."
ls -lh web/dashboard.html 2>/dev/null || echo "ERROR: web/dashboard.html not found!"
echo ""

echo "4. Testing build..."
go build -o proxy-diagnostic-test ./cmd/proxy 2>&1
BUILD_EXIT=$?
if [ $BUILD_EXIT -eq 0 ]; then
    echo "✓ Build successful"
    rm -f proxy-diagnostic-test
else
    echo "✗ Build failed with exit code $BUILD_EXIT"
fi
echo ""

echo "5. Checking if proxy is running..."
if pgrep -f "./proxy" > /dev/null; then
    echo "✓ Proxy process is running"
    echo "   PID(s):" $(pgrep -f "./proxy")
else
    echo "✗ No proxy process found"
fi
echo ""

echo "6. Testing endpoints..."
PORT=${PORT:-8099}
if curl -s -o /dev/null -w "%{http_code}" http://0.0.0.0:$PORT/health 2>/dev/null | grep -q "200"; then
    echo "✓ Health endpoint: OK"
else
    echo "✗ Health endpoint: Failed"
fi

if curl -s -o /dev/null -w "%{http_code}" http://0.0.0.0:$PORT/admin/ui 2>/dev/null | grep -q "200"; then
    echo "✓ UI endpoint: OK"
else
    echo "✗ UI endpoint: Failed"
    echo ""
    echo "   Response:"
    curl -s http://0.0.0.0:$PORT/admin/ui 2>&1 | head -5
fi
echo ""

echo "7. Checking logs (last 20 lines)..."
echo "   Run the proxy in foreground to see live logs:"
echo "   ./start.sh"
echo ""

echo "=== End Diagnostics ==="
