#!/bin/bash

# Set default port if not provided
export PORT=${PORT:-3000}

# Set default API key if not provided
export VORTEX_API_KEY=${VORTEX_API_KEY:-demo-api-key}

echo "🚀 Starting Vortex Go SDK Demo"
echo "📱 Port: $PORT"
echo "🔧 API Key: ${VORTEX_API_KEY:0:10}..."
echo ""

# Run the server
cd "$(dirname "$0")"
go run src/server.go src/auth.go