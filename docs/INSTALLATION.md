# Installation Guide

## Overview

This guide covers installation of the Enterprise Agent on Windows, Linux, and macOS.

---

## Prerequisites

- **Organization ID** and **Install Token** from your admin portal
- Administrator/root access
- Network connectivity to `https://api.yourcompany.com`

---

## Windows Installation

### Method 1: MSI Installer (Recommended)

```powershell
# Download installer
Invoke-WebRequest -Uri "https://downloads.yourcompany.com/agent/windows/agent-1.0.0.msi" -OutFile "agent.msi"

# Install with parameters
msiexec /i agent.msi ORG_ID=your-org-id INSTALL_TOKEN=your-token /quiet /log install.log

# Verify installation
Get-Service YourAgentService
```

### Method 2: Manual Installation

```powershell
# 1. Download binary
Invoke-WebRequest -Uri "https://downloads.yourcompany.com/agent/windows/agent.exe" -OutFile "agent.exe"

# 2. Create directories
New-Item -ItemType Directory -Path "C:\ProgramData\YourCompany\Agent" -Force
New-Item -ItemType Directory -Path "C:\ProgramData\YourCompany\Agent\logs" -Force
New-Item -ItemType Directory -Path "C:\ProgramData\YourCompany\Agent\certs" -Force

# 3. Copy binary
Copy-Item agent.exe "C:\Program Files\YourCompany\Agent\agent.exe"

# 4. Create configuration
@"
{
  "org_id": "your-org-id",
  "install_token": "your-token",
  "api_base_url": "https://api.yourcompany.com",
  "log_level": "info"
}
"@ | Out-File -FilePath "C:\ProgramData\YourCompany\Agent\config.json" -Encoding UTF8

# 5. Install as service
& "C:\Program Files\YourCompany\Agent\agent.exe" install

# 6. Start service
Start-Service YourAgentService
```

### Verify Installation

```powershell
# Check service status
Get-Service YourAgentService

# View logs
Get-EventLog -LogName Application -Source "YourAgent" -Newest 10
Get-Content "C:\ProgramData\YourCompany\Agent\logs\agent.log" -Tail 20
```

---

## Linux Installation

### Debian/Ubuntu (.deb)

```bash
# Download package
wget https://downloads.yourcompany.com/agent/linux/your-agent_1.0.0_amd64.deb

# Install
sudo dpkg -i your-agent_1.0.0_amd64.deb

# Configure (edit with your credentials)
sudo nano /etc/your-agent/config.json

# Start service
sudo systemctl start your-agent
sudo systemctl enable your-agent

# Verify
sudo systemctl status your-agent
```

### RHEL/CentOS (.rpm)

```bash
# Download package
wget https://downloads.yourcompany.com/agent/linux/your-agent-1.0.0.x86_64.rpm

# Install
sudo rpm -i your-agent-1.0.0.x86_64.rpm

# Configure
sudo vi /etc/your-agent/config.json

# Start service
sudo systemctl start your-agent
sudo systemctl enable your-agent

# Verify
sudo systemctl status your-agent
```

### Script Installation

```bash
# Download and run installer
curl -sSL https://downloads.yourcompany.com/agent/install.sh | \
  sudo bash -s -- --org-id your-org-id --token your-token

# Or download first
wget https://downloads.yourcompany.com/agent/install.sh
chmod +x install.sh
sudo ./install.sh --org-id your-org-id --token your-token
```

### Verify Installation

```bash
# Check service status
sudo systemctl status your-agent

# View logs
sudo journalctl -u your-agent -f
tail -f /var/log/your-agent/agent.log

# Check agent version
your-agent --version
```

---

## macOS Installation

### Method 1: Package Installer (.pkg)

```bash
# Download installer
curl -O https://downloads.yourcompany.com/agent/macos/YourAgent-1.0.0.pkg

# Install (requires admin password)
sudo installer -pkg YourAgent-1.0.0.pkg -target /

# Configure
sudo nano /var/lib/your-agent/config.json

# Start service
sudo launchctl load /Library/LaunchDaemons/com.yourcompany.agent.plist

# Verify
sudo launchctl list | grep your-agent
```

### Method 2: Homebrew (if available)

```bash
# Add tap
brew tap yourcompany/agent

# Install
brew install your-agent

# Configure
sudo nano /var/lib/your-agent/config.json

# Start service
brew services start your-agent
```

### Verify Installation

```bash
# Check if running
sudo launchctl list | grep com.yourcompany.agent

# View logs
tail -f /var/log/your-agent/agent.log

# Check version
/usr/local/bin/your-agent --version
```

---

## Configuration

### Minimal Configuration

```json
{
  "org_id": "your-org-id",
  "install_token": "your-install-token",
  "api_base_url": "https://api.yourcompany.com"
}
```

### Full Configuration Example

```json
{
  "org_id": "acme-corp",
  "install_token": "abc123xyz789",
  "api_base_url": "https://api.yourcompany.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "max_buffer_size": 104857600,
  "heartbeat_interval": "5m",
  "log_level": "info",
  "log_file": "/var/log/your-agent/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5,
  "update_enabled": true,
  "update_check_interval": "1h"
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `org_id` | Organization identifier | Required |
| `install_token` | Bootstrap token | Required |
| `api_base_url` | Backend API URL | Required |
| `collection_interval` | Data collection frequency | `60s` |
| `batch_size` | Telemetry batch size | `100` |
| `max_buffer_size` | Offline buffer size (bytes) | `104857600` (100MB) |
| `heartbeat_interval` | Health check frequency | `5m` |
| `log_level` | Logging level (debug/info/warning/error) | `info` |
| `log_max_size_mb` | Log rotation size | `100` |
| `log_max_backups` | Number of rotated logs to keep | `5` |
| `update_enabled` | Enable auto-updates | `true` |
| `update_check_interval` | Update check frequency | `1h` |

---

## Post-Installation

### 1. Verify Bootstrap

Check that the agent successfully registered:

```bash
# Linux
sudo journalctl -u your-agent | grep "Bootstrap"

# Windows
Get-EventLog -LogName Application -Source "YourAgent" | Where-Object {$_.Message -like "*Bootstrap*"}

# macOS
grep "Bootstrap" /var/log/your-agent/agent.log
```

You should see: `Bootstrap completed successfully`

### 2. Check Certificate

Verify that mTLS certificates were issued:

```bash
# Linux/macOS
ls -la /var/lib/your-agent/certs/

# Windows
dir "C:\ProgramData\YourCompany\Agent\certs\"
```

You should see: `agent.crt`, `agent.key`, `ca.crt`

### 3. Monitor Health

Check that heartbeats are being sent:

```bash
# Linux
sudo journalctl -u your-agent | grep "Heartbeat"

# Windows
Get-EventLog -LogName Application -Source "YourAgent" | Where-Object {$_.Message -like "*Heartbeat*"}

# macOS
grep "Heartbeat" /var/log/your-agent/agent.log
```

---

## Troubleshooting

### Agent Won't Start

**Check logs:**
```bash
# Linux
sudo journalctl -u your-agent -n 50

# Windows
Get-EventLog -LogName Application -Source "YourAgent" -Newest 50

# macOS
tail -50 /var/log/your-agent/agent.log
```

**Common issues:**
- Missing configuration file
- Invalid org_id or install_token
- Network connectivity to API
- Permission issues

### Bootstrap Fails

**Error:** `Bootstrap failed: 401 Unauthorized`

**Solution:** Check that install_token is valid and not expired

**Error:** `Bootstrap failed: connection refused`

**Solution:** Verify api_base_url is correct and accessible

### Certificate Errors

**Error:** `Failed to load certificate`

**Solution:** 
1. Check file permissions (should be 0600)
2. Verify certificate files exist
3. Re-run bootstrap if needed

### High CPU/Memory Usage

**Check resource usage:**
```bash
# Linux
top -p $(pgrep your-agent)

# Windows
Get-Process -Name agent | Select-Object CPU,WorkingSet

# macOS
top -pid $(pgrep your-agent)
```

**Solution:**
- Reduce collection_interval
- Decrease batch_size
- Check for collector issues

---

## Uninstallation

### Windows

```powershell
# Stop service
Stop-Service YourAgentService

# Uninstall via Control Panel or:
msiexec /x {PRODUCT-GUID} /quiet

# Or manual removal
& "C:\Program Files\YourCompany\Agent\agent.exe" uninstall
Remove-Item -Recurse "C:\Program Files\YourCompany\Agent"
Remove-Item -Recurse "C:\ProgramData\YourCompany\Agent"
```

### Linux

```bash
# Using script
sudo /opt/your-agent/uninstall.sh

# Or manually
sudo systemctl stop your-agent
sudo systemctl disable your-agent
sudo apt remove your-agent  # Debian/Ubuntu
sudo yum remove your-agent  # RHEL/CentOS
```

### macOS

```bash
# Stop service
sudo launchctl unload /Library/LaunchDaemons/com.yourcompany.agent.plist

# Remove files
sudo rm -rf /usr/local/bin/your-agent
sudo rm -rf /var/lib/your-agent
sudo rm -rf /var/log/your-agent
sudo rm /Library/LaunchDaemons/com.yourcompany.agent.plist
```

---

## Upgrading

### Automatic Updates

The agent automatically checks for updates based on `update_check_interval`. When an update is available:

1. Agent downloads new binary
2. Verifies signature
3. Performs atomic swap
4. Restarts service
5. Runs health check
6. Rolls back if health check fails

### Manual Update

```bash
# Linux
sudo systemctl stop your-agent
sudo dpkg -i your-agent_1.1.0_amd64.deb
sudo systemctl start your-agent

# Windows
Stop-Service YourAgentService
msiexec /i agent-1.1.0.msi /quiet
Start-Service YourAgentService

# macOS
sudo launchctl unload /Library/LaunchDaemons/com.yourcompany.agent.plist
sudo installer -pkg YourAgent-1.1.0.pkg -target /
sudo launchctl load /Library/LaunchDaemons/com.yourcompany.agent.plist
```

---

## Support

For installation issues:
- Email: support@yourcompany.com
- Documentation: https://docs.yourcompany.com/agent
- Status: https://status.yourcompany.com
