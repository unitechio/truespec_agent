#!/bin/bash

# Complete test script - runs both server and agent

echo "🧪 Complete Bootstrap Flow Test"
echo "================================"
echo ""

# Clean previous data
echo "🧹 Cleaning previous test data..."
rm -rf C:/ProgramData/univertech/Agent
echo "   ✓ Cleaned"
echo ""

# Start mock server in background
echo "🚀 Starting mock API server..."
go run tests/mock_server.go > /tmp/mock_server.log 2>&1 &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "   ❌ Server failed to start"
    exit 1
fi

echo "   ✓ Server running on http://localhost:8080"
echo ""

# Set environment variables
export ORG_ID=test-org
export INSTALL_TOKEN=test-token-123
export BOOTSTRAP_URL=http://localhost:8080/api/v1/agents/bootstrap

echo "📝 Environment variables:"
echo "   ORG_ID=$ORG_ID"
echo "   INSTALL_TOKEN=$INSTALL_TOKEN"
echo "   BOOTSTRAP_URL=$BOOTSTRAP_URL"
echo ""

echo "🤖 Starting agent..."
echo "   (Will run for 15 seconds then stop)"
echo ""

# Run agent with timeout
timeout 15s ./build/agent.exe 2>&1 | tee /tmp/agent.log || true

echo ""
echo "📊 Checking results..."
echo ""

# Check if config was created
if [ -f "C:/ProgramData/univertech/Agent/config.json" ]; then
    echo "✅ Config file created"
    echo ""
    echo "Config contents:"
    cat C:/ProgramData/univertech/Agent/config.json
    echo ""
else
    echo "❌ Config file not found"
fi

# Check certificates
if [ -f "C:/ProgramData/univertech/Agent/certs/agent.crt" ]; then
    echo "✅ Agent certificate created"
else
    echo "❌ Agent certificate not found"
fi

if [ -f "C:/ProgramData/univertech/Agent/certs/agent.key" ]; then
    echo "✅ Agent private key created"
else
    echo "❌ Agent private key not found"
fi

if [ -f "C:/ProgramData/univertech/Agent/certs/ca.crt" ]; then
    echo "✅ CA certificate created"
else
    echo "❌ CA certificate not found"
fi

echo ""
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "✅ Test complete!"
echo ""
echo "📋 Logs saved to:"
echo "   - /tmp/mock_server.log"
echo "   - /tmp/agent.log"
