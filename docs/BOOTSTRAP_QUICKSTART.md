# Quick Start Guide - Refactored Bootstrap Flow

## For Developers

### Running the Agent Locally

1. **Start the Mock API Server:**
```bash
cd t:/OWNER_DAT/CODE/OWNER/agent
go run tests/mock_server.go
```

2. **In a new terminal, run the agent:**
```bash
export ORG_ID=test-org
export INSTALL_TOKEN=test-token-123
export BOOTSTRAP_URL=http://localhost:8080/api/v1/agents/bootstrap

# Build and run
go build -o build/agent.exe ./cmd/agent
./build/agent.exe
```

3. **Expected Output:**
```
[AGENT] Starting enterprise-agent v1.0.0
[AGENT] No configuration found, starting bootstrap process...
[BOOTSTRAP] Calling bootstrap API: http://localhost:8080/api/v1/agents/bootstrap
[BOOTSTRAP] Bootstrap successful. Agent ID: agent-1735549934, API: http://localhost:8080
[AGENT] Bootstrap successful, configuration saved
[AGENT] Configuration loaded successfully (Agent ID: agent-1735549934)
[AGENT] Identity verified: Agent ID = agent-1735549934
[AGENT] Agent running successfully
```

### Testing Different Scenarios

#### Test Fresh Install
```bash
# Clean previous data
rm -rf C:/ProgramData/univertech/Agent

# Run agent (will bootstrap)
export ORG_ID=test-org
export INSTALL_TOKEN=test-token-123
export BOOTSTRAP_URL=http://localhost:8080/api/v1/agents/bootstrap
./build/agent.exe
```

#### Test Existing Config
```bash
# Run agent again (will skip bootstrap)
./build/agent.exe
```

#### Test Invalid Credentials
```bash
export ORG_ID=wrong-org
export INSTALL_TOKEN=wrong-token
./build/agent.exe
# Expected: Clear error message
```

#### Test Network Failure
```bash
# Stop mock server
# Run agent
./build/agent.exe
# Expected: Retry with exponential backoff
```

---

## For Production Deployment

### Environment Variables

**Required for Bootstrap:**
- `ORG_ID` - Your organization ID
- `INSTALL_TOKEN` - Installation token from admin portal

**Optional:**
- `BOOTSTRAP_URL` - Override bootstrap endpoint (default: https://api.univertech.space/api/v1/agents/bootstrap)
- `AGENT_CONFIG` - Override config file path

### Installation Steps

1. **Set environment variables:**
```bash
export ORG_ID=your-org-id
export INSTALL_TOKEN=your-install-token
```

2. **Run the agent:**
```bash
./agent
```

3. **Verify bootstrap:**
```bash
# Check config was created
cat C:/ProgramData/univertech/Agent/config.json

# Check certificates
ls -la C:/ProgramData/univertech/Agent/certs/
```

### Troubleshooting

#### Bootstrap Fails with "org_id is required"
- Ensure `ORG_ID` environment variable is set
- Check: `echo $ORG_ID`

#### Bootstrap Fails with "install_token is required"
- Ensure `INSTALL_TOKEN` environment variable is set
- Check: `echo $INSTALL_TOKEN`

#### Bootstrap Fails with "connection refused"
- Check network connectivity
- Verify bootstrap URL is correct
- Check firewall settings

#### Bootstrap Fails with "invalid organization ID"
- Verify `ORG_ID` is correct
- Contact your administrator

#### Bootstrap Fails with "invalid install token"
- Verify `INSTALL_TOKEN` is correct
- Token may have expired - generate a new one

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Agent Start                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │  Try Load Config     │
              └──────────┬───────────┘
                         │
                ┌────────┴────────┐
                │                 │
         Config Exists      Config Not Found
                │                 │
                ▼                 ▼
      ┌─────────────────┐  ┌──────────────────────┐
      │ Validate Runtime│  │ Create Bootstrap Cfg │
      └────────┬────────┘  └──────────┬───────────┘
               │                      │
               │                      ▼
               │           ┌──────────────────────┐
               │           │ Validate Bootstrap   │
               │           └──────────┬───────────┘
               │                      │
               │                      ▼
               │           ┌──────────────────────┐
               │           │  Call Bootstrap API  │
               │           │  (with retry logic)  │
               │           └──────────┬───────────┘
               │                      │
               │                      ▼
               │           ┌──────────────────────┐
               │           │  Save Certificates   │
               │           └──────────┬───────────┘
               │                      │
               │                      ▼
               │           ┌──────────────────────┐
               │           │ Mark as Bootstrapped │
               │           └──────────┬───────────┘
               │                      │
               │                      ▼
               │           ┌──────────────────────┐
               │           │    Save Config       │
               │           └──────────┬───────────┘
               │                      │
               └──────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │ Check Re-Bootstrap?  │
              └──────────┬───────────┘
                         │
                ┌────────┴────────┐
                │                 │
          Cert Valid        Cert Expired
                │                 │
                │                 ▼
                │      ┌──────────────────────┐
                │      │   Re-Bootstrap       │
                │      └──────────┬───────────┘
                │                 │
                └─────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   Agent Running      │
              └──────────────────────┘
```

---

## Key Files

| File | Purpose |
|------|---------|
| `internal/config/config.go` | Configuration management with state tracking |
| `internal/config/errors.go` | Centralized error definitions |
| `internal/identity/bootstrap.go` | Bootstrap flow and certificate management |
| `internal/identity/retry.go` | Exponential backoff retry logic |
| `cmd/agent/main.go` | Main application entry point |
| `tests/mock_server.go` | Mock API server for testing |

---

## Configuration States

### 1. Unbootstrapped
- No config file exists
- Agent will enter bootstrap mode
- Requires: `ORG_ID`, `INSTALL_TOKEN`

### 2. Bootstrapped
- Config file exists with `bootstrapped: true`
- Agent has `AgentID` and certificates
- Agent will skip bootstrap and start normally

### 3. Re-Bootstrap Needed
- Config exists but certificates expired
- Agent will automatically re-bootstrap
- Seamless cert renewal

---

## API Endpoints

### Bootstrap Endpoint
```
POST /api/v1/agents/bootstrap
Content-Type: application/json

Request:
{
  "org_id": "test-org",
  "install_token": "test-token-123",
  "hostname": "agent-host",
  "os": "windows",
  "arch": "amd64",
  "agent_version": "1.0.0"
}

Response:
{
  "agent_id": "agent-1735549934",
  "api_base_url": "https://api.univertech.space",
  "certificate": "-----BEGIN CERTIFICATE-----\n...",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
  "ca_cert": "-----BEGIN CERTIFICATE-----\n...",
  "policy": "{\"version\": \"1.0\"}",
  "expires_at": "2026-12-30T16:32:14+07:00"
}
```

---

## Retry Configuration

Default retry settings:
- **Initial Delay:** 5 seconds
- **Max Delay:** 5 minutes
- **Max Attempts:** 10
- **Backoff Multiplier:** 2.0
- **Jitter:** Enabled (±25%)

Retry sequence:
```
Attempt 1: ~5 seconds
Attempt 2: ~10 seconds
Attempt 3: ~20 seconds
Attempt 4: ~40 seconds
Attempt 5: ~80 seconds (1.3 minutes)
Attempt 6: ~160 seconds (2.7 minutes)
Attempt 7: ~300 seconds (5 minutes, capped)
Attempt 8: ~300 seconds (5 minutes, capped)
Attempt 9: ~300 seconds (5 minutes, capped)
Attempt 10: ~300 seconds (5 minutes, capped)
```

Total retry time: ~45 minutes

---

## Security Notes

1. **Install Token:** Cleared from config after successful bootstrap
2. **Certificates:** Stored with 0600 permissions (owner read/write only)
3. **Config File:** Stored with 0600 permissions
4. **No Secrets in Logs:** Tokens and keys are never logged

---

## Common Commands

```bash
# Build agent
go build -o build/agent.exe ./cmd/agent

# Run mock server
go run tests/mock_server.go

# Clean test data
rm -rf C:/ProgramData/univertech/Agent

# View config
cat C:/ProgramData/univertech/Agent/config.json

# View certificates
openssl x509 -in C:/ProgramData/univertech/Agent/certs/agent.crt -text -noout

# Check cert expiration
openssl x509 -in C:/ProgramData/univertech/Agent/certs/agent.crt -enddate -noout
```
