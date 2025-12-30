#!/bin/bash

# Test script for bootstrap flow
# This script starts the mock API server and tests the agent bootstrap

set -e

echo "🧪 Bootstrap Flow Test"
echo "====================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Clean up previous test artifacts
echo -e "${BLUE}🧹 Cleaning up previous test data...${NC}"
rm -rf C:/ProgramData/univertech/Agent
rm -f build/agent.exe
rm -f tests/mock_server.exe

# Build the mock server
echo -e "${BLUE}🔨 Building mock API server...${NC}"
go build -o tests/mock_server.exe ./tests/mock_server.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build mock server"
    exit 1
fi

# Build the agent
echo -e "${BLUE}🔨 Building agent...${NC}"
go build -o build/agent.exe ./cmd/agent
if [ $? -ne 0 ]; then
    echo "❌ Failed to build agent"
    exit 1
fi

# Start mock server in background
echo -e "${BLUE}🚀 Starting mock API server...${NC}"
./tests/mock_server.exe &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "❌ Mock server failed to start"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Mock server running on http://localhost:8080${NC}"
echo ""

# Set environment variables for bootstrap
export ORG_ID=test-org
export INSTALL_TOKEN=test-token-123
export BOOTSTRAP_URL=http://localhost:8080/api/v1/agents/bootstrap

echo -e "${BLUE}🤖 Starting agent bootstrap...${NC}"
echo "   ORG_ID: $ORG_ID"
echo "   INSTALL_TOKEN: $INSTALL_TOKEN"
echo "   BOOTSTRAP_URL: $BOOTSTRAP_URL"
echo ""

# Run agent (it will bootstrap and then we'll kill it)
timeout 15s ./build/agent.exe || true

echo ""
echo -e "${BLUE}📋 Checking bootstrap results...${NC}"

# Check if config was created
if [ -f "C:/ProgramData/univertech/Agent/config.json" ]; then
    echo -e "${GREEN}✅ Config file created${NC}"
    echo ""
    echo "Config contents:"
    cat C:/ProgramData/univertech/Agent/config.json | jq .
else
    echo -e "${YELLOW}⚠️  Config file not found${NC}"
fi

# Check if certificates were created
if [ -f "C:/ProgramData/univertech/Agent/certs/agent.crt" ]; then
    echo -e "${GREEN}✅ Agent certificate created${NC}"
else
    echo -e "${YELLOW}⚠️  Agent certificate not found${NC}"
fi

if [ -f "C:/ProgramData/univertech/Agent/certs/agent.key" ]; then
    echo -e "${GREEN}✅ Agent private key created${NC}"
else
    echo -e "${YELLOW}⚠️  Agent private key not found${NC}"
fi

if [ -f "C:/ProgramData/univertech/Agent/certs/ca.crt" ]; then
    echo -e "${GREEN}✅ CA certificate created${NC}"
else
    echo -e "${YELLOW}⚠️  CA certificate not found${NC}"
fi

# Clean up
echo ""
echo -e "${BLUE}🧹 Cleaning up...${NC}"
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo -e "${GREEN}✅ Test complete!${NC}"
