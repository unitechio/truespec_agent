đầu tiên build agent file:
    - rm -f build/agent.exe && go build -o build/agent.exe ./cmd/agent && echo "✅ Agent rebuilt successfully"
run server mock:
    - go run tests/mock_server.go
run agent:
    - ./build/agent.exe (dùng gitbash/linux terminal)
