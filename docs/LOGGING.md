# Logging System Documentation

## Overview

The Enterprise Agent includes a comprehensive logging system with:
- **Structured logging** with multiple log levels
- **Automatic log rotation** to prevent disk space issues
- **Audit logging** for security-relevant events
- **Dual output** (file + stdout) for easy debugging

---

## Application Logging

### Log Levels

| Level | Usage |
|-------|-------|
| **DEBUG** | Detailed diagnostic information for troubleshooting |
| **INFO** | General informational messages about agent operation |
| **WARNING** | Warning messages for potentially problematic situations |
| **ERROR** | Error messages for failures that don't stop the agent |
| **FATAL** | Critical errors that cause agent termination |

### Configuration

```json
{
  "log_level": "info",
  "log_file": "/var/log/your-agent/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5
}
```

### Log Format

```
2024-01-15 10:30:00 [INFO] Agent starting v1.0.0
2024-01-15 10:30:01 [INFO] Bootstrap completed successfully
2024-01-15 10:30:02 [DEBUG] Collector 'cpu' started with interval 60s
2024-01-15 10:30:03 [WARNING] Failed to send heartbeat, retrying...
2024-01-15 10:30:04 [ERROR] Policy refresh failed: connection timeout
```

### Log Rotation

Logs are automatically rotated when they reach the configured size:

```
agent.log       # Current log
agent.log.1     # Previous log
agent.log.2     # Older log
agent.log.3     # Oldest log
```

Old logs beyond `log_max_backups` are automatically deleted.

---

## Audit Logging

### Purpose

Audit logs capture **security-relevant events** for compliance and forensics:
- Agent registration (bootstrap)
- Authentication failures
- Policy changes
- Certificate rotation
- Agent updates
- Service start/stop

### Format

Audit logs use **structured JSON** for easy parsing:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "bootstrap",
  "severity": "INFO",
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "action": "agent_registration",
  "resource": "acme-corp",
  "result": "success"
}
```

### Audit Events

| Event Type | Description |
|------------|-------------|
| `bootstrap` | Agent initial registration |
| `auth_failure` | Failed authentication attempt |
| `policy_change` | Policy version update |
| `cert_rotation` | Certificate renewal |
| `agent_update` | Binary update |
| `service_lifecycle` | Service start/stop |

### Storage

- **Location**: `/var/log/your-agent/audit.log`
- **Permissions**: `0600` (owner read/write only)
- **Retention**: 90 days (configurable)
- **Format**: One JSON object per line (JSONL)

---

## Usage Examples

### In Code

```go
import "github.com/unitechio/agent/internal/logging"

// Create logger
logger, err := logging.NewLogger(logging.Config{
    LogPath:    "/var/log/your-agent/agent.log",
    Level:      "info",
    MaxSizeMB:  100,
    MaxBackups: 5,
})
if err != nil {
    log.Fatal(err)
}
defer logger.Close()

// Log messages
logger.Info("Agent started successfully")
logger.Debug("Collector data: %+v", data)
logger.Warning("Retry attempt %d/%d", attempt, maxRetries)
logger.Error("Failed to send telemetry: %v", err)

// Create audit logger
auditLogger, err := logging.NewAuditLogger(
    "/var/log/your-agent/audit.log",
    agentID,
)
if err != nil {
    log.Fatal(err)
}
defer auditLogger.Close()

// Log audit events
auditLogger.LogBootstrap(true, "acme-corp", nil)
auditLogger.LogPolicyChange("1.0.0", "1.1.0")
auditLogger.LogUpdate("1.0.0", "1.1.1", true, nil)
```

---

## Viewing Logs

### Linux (systemd)

```bash
# View application logs
journalctl -u your-agent -f

# View log file directly
tail -f /var/log/your-agent/agent.log

# View audit logs
cat /var/log/your-agent/audit.log | jq .
```

### Windows

```powershell
# View Event Log
Get-EventLog -LogName Application -Source "YourAgent" -Newest 50

# View log file
Get-Content C:\ProgramData\YourCompany\Agent\logs\agent.log -Tail 50 -Wait

# View audit logs
Get-Content C:\ProgramData\YourCompany\Agent\logs\audit.log | ConvertFrom-Json
```

### macOS

```bash
# View logs
tail -f /var/log/your-agent/agent.log

# View audit logs
cat /var/log/your-agent/audit.log | jq .
```

---

## Log Analysis

### Search for Errors

```bash
# Linux
grep "ERROR" /var/log/your-agent/agent.log

# Windows
Select-String -Path "C:\ProgramData\YourCompany\Agent\logs\agent.log" -Pattern "ERROR"
```

### Analyze Audit Events

```bash
# Count events by type
cat audit.log | jq -r '.event_type' | sort | uniq -c

# Find failed authentications
cat audit.log | jq 'select(.event_type == "auth_failure")'

# Find policy changes in last 24h
cat audit.log | jq 'select(.event_type == "policy_change" and .timestamp > "'$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)'")'
```

---

## Troubleshooting

### High Log Volume

If logs are growing too fast:

1. **Increase rotation size**: Set `log_max_size_mb` to a larger value
2. **Reduce log level**: Change from `debug` to `info` or `warning`
3. **Increase backup count**: Keep more rotated logs for analysis

### Missing Logs

Check:
- Log directory permissions
- Disk space availability
- Service is running
- Log level is not set too high (e.g., `fatal` only)

### Audit Log Compliance

For compliance requirements:
- Set retention to 90+ days
- Forward to SIEM system (Splunk, ELK, etc.)
- Enable log integrity checking
- Restrict file permissions to `0600`

---

## Best Practices

1. **Use appropriate log levels**:
   - `DEBUG`: Development only
   - `INFO`: Production default
   - `WARNING`: Potential issues
   - `ERROR`: Failures
   - `FATAL`: Critical errors

2. **Include context**:
   ```go
   logger.Error("Failed to send telemetry to %s: %v", endpoint, err)
   ```

3. **Avoid logging sensitive data**:
   - ❌ Passwords, tokens, API keys
   - ❌ Full certificate contents
   - ✅ Redacted identifiers
   - ✅ Error messages (without secrets)

4. **Monitor log size**:
   - Set up alerts for rapid log growth
   - Review rotation settings periodically

5. **Centralize audit logs**:
   - Forward to SIEM for correlation
   - Enable real-time alerting
   - Maintain tamper-proof copies
