# Enterprise Endpoint Agent

A production-grade endpoint monitoring agent written in Go, designed for enterprise deployment across Windows, Linux, and macOS.

## 🎯 Overview

This agent is built following enterprise security standards with:
- ✅ **Zero-trust security** - mTLS certificate-based authentication
- ✅ **Transparent operation** - No stealth behavior, clean install/uninstall
- ✅ **Policy-driven** - Server-controlled data collection
- ✅ **Auto-update** - Cryptographically signed updates with rollback
- ✅ **Production-ready** - Service-based, graceful shutdown, health monitoring

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│           Enterprise Agent                   │
├─────────────────────────────────────────────┤
│  Identity Manager (mTLS Bootstrap)          │
│  Policy Engine (Hot-reload)                 │
│  Scheduler (Jitter, Graceful Shutdown)      │
│  Collectors (System, CPU, Memory, Disk, Net)│
│  Health Monitor (Heartbeat)                 │
│  Auto-Updater (Signature Verification)      │
└─────────────────────────────────────────────┘
         │                    │
         │ mTLS               │ mTLS
         ▼                    ▼
┌──────────────┐      ┌──────────────┐
│ Control Plane│      │  Data Plane  │
│ (Auth/Policy)│      │ (Telemetry)  │
└──────────────┘      └──────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Windows 10+, Linux (systemd), or macOS 10.15+

### Build

```bash
# Build for current platform
make build

# Cross-compile for all platforms
make build-all
```

### Installation

#### Windows
```powershell
# Install as Windows Service
.\agent.exe install --org-id YOUR_ORG_ID --token YOUR_INSTALL_TOKEN

# Start service
sc start YourAgentService
```

#### Linux
```bash
# Install .deb package
sudo dpkg -i your-agent_1.0.0_amd64.deb

# Or install .rpm package
sudo rpm -i your-agent-1.0.0.x86_64.rpm

# Start service
sudo systemctl start your-agent
sudo systemctl enable your-agent
```

#### macOS
```bash
# Install .pkg
sudo installer -pkg YourAgent.pkg -target /

# Start service
sudo launchctl load /Library/LaunchDaemons/com.unitechio.agent.plist
```

## 📋 Configuration

Configuration file location:
- **Windows**: `C:\ProgramData\unitechio\Agent\config.json`
- **Linux**: `/etc/your-agent/config.json`
- **macOS**: `/var/lib/your-agent/config.json`

Example configuration:

```json
{
  "org_id": "your-org-id",
  "api_base_url": "https://api.unitechio.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "heartbeat_interval": "5m",
  "log_level": "info",
  "update_enabled": true
}
```

## 🔒 Security

### Zero-Trust Identity

1. **Bootstrap**: Agent exchanges `INSTALL_TOKEN` for X.509 certificate
2. **mTLS**: All communications use mutual TLS authentication
3. **Certificate Rotation**: Automatic renewal before expiration
4. **Revocation**: Server-side certificate revocation support

### Data Privacy

- **Minimal Collection**: Only system metadata by default
- **Policy-Controlled**: All collectors can be disabled server-side
- **No PII**: No user activity, file contents, or process lists
- **Transparent**: Full documentation of collected data

### Secure Storage

| OS      | Method                          | Location                                      |
|---------|---------------------------------|-----------------------------------------------|
| Windows | DPAPI encryption                | `%ProgramData%\unitechio\Agent\certs\`      |
| Linux   | File permissions (0600)         | `/var/lib/your-agent/certs/`                  |
| macOS   | File permissions (0600)         | `/var/lib/your-agent/certs/`                  |

## 📊 Data Collection

Default collectors (all configurable):

- **System**: OS version, hostname, architecture
- **CPU**: Model, cores, usage percentage
- **Memory**: Total, used, available
- **Disk**: Mount points, usage
- **Network**: Interfaces, IP addresses (MAC optional)

## 📝 Logging

Comprehensive logging system with:

- **Multiple log levels**: DEBUG, INFO, WARNING, ERROR, FATAL
- **Automatic rotation**: Prevents disk space issues
- **Dual output**: File + stdout for easy debugging
- **Audit logging**: Security events in structured JSON format

**Application Logs:**
```
2024-01-15 10:30:00 [INFO] Agent starting v1.0.0
2024-01-15 10:30:01 [INFO] Bootstrap completed successfully
2024-01-15 10:30:02 [DEBUG] Collector 'cpu' started
```

**Audit Logs (JSON):**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "bootstrap",
  "action": "agent_registration",
  "result": "success"
}
```

See [LOGGING.md](docs/LOGGING.md) for details.

## 🔄 Auto-Update

The agent supports secure auto-updates:

1. Check for updates (configurable interval)
2. Download new binary
3. Verify cryptographic signature
4. Atomic swap with old binary
5. Restart service
6. Health check (rollback on failure)

## 🛠️ Development

### Project Structure

```
agent/
├── cmd/agent/              # Main entry point
├── internal/
│   ├── identity/           # mTLS & bootstrap
│   ├── collectors/         # Data collectors
│   ├── scheduler/          # Job scheduling
│   ├── policy/             # Policy engine
│   ├── health/             # Health monitoring
│   ├── config/             # Configuration
│   └── service/            # OS service integration
├── pkg/api/proto/          # Protobuf definitions
├── assets/                 # Service templates
└── installers/             # Installer scripts
```

### Running Tests

```bash
# Unit tests
go test ./internal/... -cover

# Integration tests
go test ./tests/integration/... -tags=integration

# Race detection
go test ./... -race
```

### Building Installers

```bash
# Windows MSI
cd installers/windows
candle Product.wxs
light -out agent.msi Product.wixobj

# Linux .deb
cd installers/linux
dpkg-deb --build your-agent

# macOS .pkg
cd installers/macos
pkgbuild --root ./root --identifier com.unitechio.agent YourAgent.pkg
```

## 📝 API Endpoints

The agent communicates with these backend endpoints:

- `POST /api/v1/agents/bootstrap` - Initial registration
- `GET /api/v1/policy` - Fetch policy
- `POST /api/v1/telemetry` - Send collected data
- `POST /api/v1/heartbeat` - Health check
- `GET /api/v1/updates/metadata` - Check for updates

## 🔍 Monitoring

### Logs

- **Windows**: Windows Event Log + `C:\ProgramData\unitechio\Agent\logs\`
- **Linux**: journald (`journalctl -u your-agent`)
- **macOS**: `/var/log/your-agent/`

### Health Checks

The agent exposes health metrics:
- Uptime
- Memory usage
- Goroutine count
- Last successful heartbeat

## 📜 License

Copyright © 2024 Your Company. All rights reserved.

## 🤝 Support

For support, please contact: support@unitechio.com
