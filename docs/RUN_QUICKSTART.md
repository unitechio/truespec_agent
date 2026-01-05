đầu tiên build agent file:
    - rm -f build/agent.exe && go build -o build/agent.exe ./cmd/agent && echo "✅ Agent rebuilt successfully"
run server mock:
    - go run tests/mock_server.go
run agent:
    export ORG_ID=test-org
    export INSTALL_TOKEN=test-token-123
    export BOOTSTRAP_URL=http://localhost:8080/api/v1/agents/bootstrap
    - ./build/agent.exe (dùng gitbash/linux terminal)
