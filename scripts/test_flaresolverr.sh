#!/bin/bash

# Test script to verify FlareSolverr is working
# Usage: ./scripts/test_flaresolverr.sh

echo "Testing FlareSolverr connection..."

# Check if FlareSolverr is running
if ! curl -s http://localhost:8191/v1 > /dev/null 2>&1; then
    echo "ERROR: FlareSolverr is not running at http://localhost:8191/v1"
    echo "Start FlareSolverr with:"
    echo "  docker run -p 8191:8191 -e LOG_LEVEL=info ghcr.io/flaresolverr/flaresolverr:latest"
    exit 1
fi

echo "FlareSolverr is running!"

# Test FlareSolverr with Cloudflare URL (same as frontend test)
echo ""
echo "Testing FlareSolverr with Cloudflare URL..."
RESPONSE=$(curl -s -X POST http://localhost:8191/v1 \
    -H "Content-Type: application/json" \
    -d '{
        "cmd": "request.get",
        "url": "https://www.cloudflare.com/cdn-cgi/trace",
        "maxTimeout": 30
    }')

if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "SUCCESS: FlareSolverr resolved Cloudflare page successfully"
    # Try to extract trace info if present
    if echo "$RESPONSE" | grep -q "ip="; then
        IP=$(echo "$RESPONSE" | grep -o 'ip=[^\\]*' | head -1 | sed 's/ip=//')
        echo "Your IP via FlareSolverr: $IP"
    fi
else
    echo "ERROR: FlareSolverr request failed"
    echo "Response: $RESPONSE"
fi

echo ""
echo "Test complete!"
