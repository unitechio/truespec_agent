# Building Installer Packages

This guide covers building installer packages for Windows, Linux, and macOS.

---

## Prerequisites

### All Platforms
- Go 1.21 or later
- Git
- Make

### Windows
- [WiX Toolset 3.11+](https://wixtoolset.org/releases/)
- Visual Studio Build Tools (for signing)
- Code signing certificate (for production)

### Linux
- `dpkg-deb` (for .deb packages)
- `rpmbuild` (for .rpm packages)
- `fakeroot` (recommended)

### macOS
- Xcode Command Line Tools
- Apple Developer ID certificate (for signing)
- `pkgbuild` and `productbuild` (included with Xcode)

---

## Building Binaries

First, build the agent binaries for all platforms:

```bash
# Build for all platforms
make build-all

# Or build for specific platform
make build-windows
make build-linux
make build-darwin
```

This creates binaries in `build/`:
```
build/
├── agent-windows-amd64.exe
├── agent-windows-arm64.exe
├── agent-linux-amd64
├── agent-linux-arm64
├── agent-darwin-amd64
└── agent-darwin-arm64
```

---

## Windows MSI Installer

### 1. Install WiX Toolset

Download and install from: https://wixtoolset.org/releases/

Add to PATH:
```powershell
$env:PATH += ";C:\Program Files (x86)\WiX Toolset v3.11\bin"
```

### 2. Build MSI

```powershell
cd installers/windows

# Compile WiX source
candle.exe agent.wxs -ext WixUtilExtension

# Link and create MSI
light.exe agent.wixobj -ext WixUtilExtension -out YourAgent-1.0.0.msi
```

### 3. Sign MSI (Production)

```powershell
# Sign with Authenticode certificate
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com YourAgent-1.0.0.msi

# Verify signature
signtool verify /pa YourAgent-1.0.0.msi
```

### 4. Test Installation

```powershell
# Install
msiexec /i YourAgent-1.0.0.msi ORG_ID=test-org INSTALL_TOKEN=test-token /l*v install.log

# Verify
Get-Service YourAgentService
Get-Content "C:\ProgramData\YourCompany\Agent\logs\agent.log"

# Uninstall
msiexec /x YourAgent-1.0.0.msi /quiet
```

---

## Linux Packages

### Debian/Ubuntu (.deb)

```bash
cd installers/linux

# Make script executable
chmod +x build-deb.sh

# Build package
./build-deb.sh

# This creates: your-agent_1.0.0_amd64.deb
```

**Test installation:**
```bash
# Install
sudo dpkg -i your-agent_1.0.0_amd64.deb

# Configure
sudo nano /etc/your-agent/config.json
# Add org_id and install_token

# Start service
sudo systemctl start your-agent
sudo systemctl status your-agent

# View logs
sudo journalctl -u your-agent -f

# Uninstall
sudo apt remove your-agent
# or for complete removal:
sudo apt purge your-agent
```

### RHEL/CentOS (.rpm)

```bash
cd installers/linux

# Install rpmbuild if needed
sudo yum install rpm-build

# Make script executable
chmod +x build-rpm.sh

# Build package
./build-rpm.sh

# This creates: your-agent-1.0.0-1.*.x86_64.rpm
```

**Test installation:**
```bash
# Install
sudo rpm -i your-agent-1.0.0-1.*.x86_64.rpm

# Configure
sudo vi /etc/your-agent/config.json

# Start service
sudo systemctl start your-agent
sudo systemctl status your-agent

# Uninstall
sudo rpm -e your-agent
```

### Sign Linux Packages

**For .deb:**
```bash
# Create GPG key if needed
gpg --gen-key

# Sign package
dpkg-sig --sign builder your-agent_1.0.0_amd64.deb

# Verify
dpkg-sig --verify your-agent_1.0.0_amd64.deb
```

**For .rpm:**
```bash
# Import GPG key to RPM
rpm --import /path/to/public-key.asc

# Sign package
rpm --addsign your-agent-1.0.0-1.*.x86_64.rpm

# Verify
rpm --checksig your-agent-1.0.0-1.*.x86_64.rpm
```

---

## macOS Package

### 1. Build Package

```bash
cd installers/macos

# Make script executable
chmod +x build-pkg.sh

# Build package
./build-pkg.sh

# This creates: YourAgent-1.0.0.pkg
```

### 2. Sign Package (Required for Distribution)

```bash
# Sign with Developer ID
productsign --sign "Developer ID Installer: Your Company (TEAMID)" \
  YourAgent-1.0.0.pkg \
  YourAgent-1.0.0-signed.pkg

# Verify signature
pkgutil --check-signature YourAgent-1.0.0-signed.pkg
```

### 3. Notarize (Required for macOS 10.15+)

```bash
# Upload for notarization
xcrun notarytool submit YourAgent-1.0.0-signed.pkg \
  --apple-id "your@email.com" \
  --team-id "TEAMID" \
  --password "app-specific-password" \
  --wait

# Staple notarization ticket
xcrun stapler staple YourAgent-1.0.0-signed.pkg

# Verify
xcrun stapler validate YourAgent-1.0.0-signed.pkg
```

### 4. Test Installation

```bash
# Install
sudo installer -pkg YourAgent-1.0.0-signed.pkg -target /

# Configure
sudo nano /var/lib/your-agent/config.json

# Restart service
sudo launchctl kickstart -k system/com.yourcompany.agent

# View logs
tail -f /var/log/your-agent/agent.log

# Uninstall
sudo launchctl unload /Library/LaunchDaemons/com.yourcompany.agent.plist
sudo rm -rf /usr/local/bin/your-agent
sudo rm -rf /var/lib/your-agent
sudo rm -rf /var/log/your-agent
sudo rm /Library/LaunchDaemons/com.yourcompany.agent.plist
```

---

## Automated Build Pipeline

### GitHub Actions Example

```yaml
name: Build Installers

on:
  push:
    tags:
      - 'v*'

jobs:
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build binary
        run: make build-windows
      - name: Build MSI
        run: |
          cd installers/windows
          candle.exe agent.wxs -ext WixUtilExtension
          light.exe agent.wixobj -ext WixUtilExtension -out YourAgent.msi
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: windows-msi
          path: installers/windows/YourAgent.msi

  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build binary
        run: make build-linux
      - name: Build .deb
        run: |
          cd installers/linux
          chmod +x build-deb.sh
          ./build-deb.sh
      - name: Build .rpm
        run: |
          sudo apt-get install rpm
          cd installers/linux
          chmod +x build-rpm.sh
          ./build-rpm.sh
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: linux-packages
          path: |
            installers/linux/*.deb
            installers/linux/*.rpm

  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build binary
        run: make build-darwin
      - name: Build .pkg
        run: |
          cd installers/macos
          chmod +x build-pkg.sh
          ./build-pkg.sh
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: macos-pkg
          path: installers/macos/YourAgent-*.pkg
```

---

## Distribution

### Hosting Installers

**Option 1: Direct Download**
```
https://downloads.yourcompany.com/agent/
├── windows/
│   └── YourAgent-1.0.0.msi
├── linux/
│   ├── your-agent_1.0.0_amd64.deb
│   └── your-agent-1.0.0-1.x86_64.rpm
└── macos/
    └── YourAgent-1.0.0.pkg
```

**Option 2: Package Repository**

**APT Repository (Debian/Ubuntu):**
```bash
# Create repository
mkdir -p repo/deb
cp your-agent_1.0.0_amd64.deb repo/deb/
cd repo
dpkg-scanpackages deb /dev/null | gzip -9c > deb/Packages.gz

# Users add repository
echo "deb https://repo.yourcompany.com/agent/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/your-agent.list
sudo apt update
sudo apt install your-agent
```

**YUM Repository (RHEL/CentOS):**
```bash
# Create repository
mkdir -p repo/rpm
cp your-agent-1.0.0-1.x86_64.rpm repo/rpm/
createrepo repo/rpm

# Users add repository
sudo tee /etc/yum.repos.d/your-agent.repo <<EOF
[your-agent]
name=Your Agent Repository
baseurl=https://repo.yourcompany.com/agent/rpm
enabled=1
gpgcheck=1
gpgkey=https://repo.yourcompany.com/agent/RPM-GPG-KEY
EOF
sudo yum install your-agent
```

---

## Troubleshooting

### Windows

**Error: "WiX Toolset not found"**
- Install WiX Toolset and add to PATH

**Error: "Failed to install service"**
- Run installer as Administrator
- Check Windows Event Log for details

### Linux

**Error: "dpkg-deb: command not found"**
```bash
sudo apt-get install dpkg-dev
```

**Error: "rpmbuild: command not found"**
```bash
sudo yum install rpm-build
```

**Error: "Permission denied"**
- Make build scripts executable: `chmod +x build-*.sh`

### macOS

**Error: "Developer ID not found"**
- Enroll in Apple Developer Program
- Create Developer ID certificate in Xcode

**Error: "Notarization failed"**
- Ensure binary is signed
- Check for hardened runtime entitlements
- Verify app-specific password

---

## Best Practices

1. **Version Management**
   - Use semantic versioning (MAJOR.MINOR.PATCH)
   - Update version in all build scripts
   - Tag releases in Git

2. **Code Signing**
   - Always sign production installers
   - Use timestamping for long-term validity
   - Store certificates securely (not in repo)

3. **Testing**
   - Test on clean VMs before release
   - Verify upgrade paths
   - Test uninstallation

4. **Documentation**
   - Include release notes
   - Document breaking changes
   - Provide upgrade guides

5. **Distribution**
   - Use HTTPS for downloads
   - Provide checksums (SHA256)
   - Maintain old versions for rollback

---

## Quick Reference

```bash
# Build all installers
make build-all
cd installers/windows && ./build.ps1
cd installers/linux && ./build-deb.sh && ./build-rpm.sh
cd installers/macos && ./build-pkg.sh

# Sign packages
# Windows: signtool sign /f cert.pfx /p pass /t http://timestamp.digicert.com installer.msi
# Linux: dpkg-sig --sign builder package.deb
# macOS: productsign --sign "Developer ID" package.pkg signed.pkg

# Test installation
# Windows: msiexec /i installer.msi /l*v log.txt
# Linux: sudo dpkg -i package.deb
# macOS: sudo installer -pkg package.pkg -target /
```
