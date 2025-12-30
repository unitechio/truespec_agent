#!/bin/bash

# Simple test script for bootstrap flow

echo "🧪 Testing Bootstrap Flow"
echo "========================="
echo ""

# Clean previous data
echo "🧹 Cleaning previous test data..."
rm -rf C:/ProgramData/univertech/Agent
echo "   ✓ Cleaned"
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

echo "🚀 Starting agent..."
echo "   (Press Ctrl+C to stop after bootstrap completes)"
echo ""

# Run agent
./build/agent.exe
