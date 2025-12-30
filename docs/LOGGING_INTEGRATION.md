# Logging Integration Example

This document shows how to integrate the logging system into your agent components.

## Basic Usage

```go
package main

import (
    "github.com/unitechio/agent/internal/logging"
)

func main() {
    // Create logger
    logger, err := logging.NewLogger(logging.Config{
        LogPath:    "/var/log/your-agent/agent.log",
        Level:      "info",
        MaxSizeMB:  100,
        MaxBackups: 5,
    })
    if err != nil {
        panic(err)
    }
    defer logger.Close()

    // Use logger
    logger.Info("Agent starting...")
    logger.Debug("Debug information: %+v", someData)
    logger.Warning("Potential issue detected")
    logger.Error("Failed to connect: %v", err)
}
```

## Integration with Existing Components

### 1. Replace log.Logger with logging.Logger

**Before:**
```go
import "log"

type Component struct {
    logger *log.Logger
}

func NewComponent(logger *log.Logger) *Component {
    return &Component{logger: logger}
}
```

**After:**
```go
import "github.com/unitechio/agent/internal/logging"

type Component struct {
    logger *logging.Logger
}

func NewComponent(logger *logging.Logger) *Component {
    return &Component{logger: logger}
}
```

### 2. For Components Requiring *log.Logger

Use the compatibility wrapper:

```go
import (
    "log"
    "github.com/unitechio/agent/internal/logging"
)

// Create custom logger
customLogger, _ := logging.NewLogger(logging.Config{...})

// Convert to standard logger for components that need it
stdLogger := logging.NewStdLogger(customLogger)

// Use with components expecting *log.Logger
buffer := buffer.New(dir, size, stdLogger)
```

## Audit Logging Integration

```go
import "github.com/unitechio/agent/internal/logging"

// Create audit logger
auditLogger, err := logging.NewAuditLogger(
    "/var/log/your-agent/audit.log",
    agentID,
)
if err != nil {
    panic(err)
}
defer auditLogger.Close()

// Log security events
auditLogger.LogBootstrap(true, orgID, nil)
auditLogger.LogPolicyChange(oldVer, newVer)
auditLogger.LogUpdate(oldVer, newVer, true, nil)
```

## Main Application Integration

```go
package main

import (
    "github.com/unitechio/agent/internal/config"
    "github.com/unitechio/agent/internal/logging"
    "github.com/unitechio/agent/internal/identity"
    "github.com/unitechio/agent/internal/collectors"
    "github.com/unitechio/agent/internal/scheduler"
)

func main() {
    // Load config
    cfg, _ := config.Load("config.json")

    // Create application logger
    appLogger, err := logging.NewLogger(logging.Config{
        LogPath:    cfg.LogFile,
        Level:      cfg.LogLevel,
        MaxSizeMB:  cfg.LogMaxSizeMB,
        MaxBackups: cfg.LogMaxBackups,
    })
    if err != nil {
        panic(err)
    }
    defer appLogger.Close()

    appLogger.Info("Agent starting v1.0.0")

    // Create audit logger
    auditLogger, _ := logging.NewAuditLogger(
        cfg.AuditLogFile,
        cfg.AgentID,
    )
    defer auditLogger.Close()

    // Bootstrap identity
    identityMgr, _ := identity.NewManager(cfg, appLogger)
    if err := identityMgr.Bootstrap(ctx); err != nil {
        appLogger.Error("Bootstrap failed: %v", err)
        auditLogger.LogBootstrap(false, cfg.OrgID, err)
        return
    }
    appLogger.Info("Bootstrap successful")
    auditLogger.LogBootstrap(true, cfg.OrgID, nil)

    // Create scheduler with logger
    sched := scheduler.New(cfg, policyEngine, identityMgr, appLogger)
    sched.Start(ctx)

    appLogger.Info("Agent running")
}
```

## Configuration

Add logging configuration to your config.json:

```json
{
  "log_level": "info",
  "log_file": "/var/log/your-agent/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5,
  "audit_log_file": "/var/log/your-agent/audit.log"
}
```

## Best Practices

1. **Use appropriate log levels**:
   - DEBUG: Development/troubleshooting only
   - INFO: Normal operations
   - WARNING: Potential issues
   - ERROR: Failures
   - FATAL: Critical errors (exits program)

2. **Include context in messages**:
   ```go
   logger.Error("Failed to send telemetry to %s: %v", endpoint, err)
   ```

3. **Log security events to audit log**:
   ```go
   auditLogger.LogAuthFailure(endpoint, reason)
   ```

4. **Close loggers on shutdown**:
   ```go
   defer logger.Close()
   defer auditLogger.Close()
   ```

5. **Don't log sensitive data**:
   - ❌ Passwords, tokens, API keys
   - ✅ Redacted identifiers, error messages
