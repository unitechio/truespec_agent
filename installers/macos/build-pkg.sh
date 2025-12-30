#!/bin/bash
# Build .pkg installer for Enterprise Agent on macOS

set -e

VERSION="1.0.0"
PACKAGE_NAME="YourAgent"
IDENTIFIER="com.yourcompany.agent"
BUILD_DIR="build/macos"

echo "Building .pkg installer for ${PACKAGE_NAME} v${VERSION}..."

# Clean previous build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/{root,scripts,resources}

# Create directory structure in root
mkdir -p ${BUILD_DIR}/root/usr/local/bin
mkdir -p ${BUILD_DIR}/root/var/lib/your-agent/{certs,buffer}
mkdir -p ${BUILD_DIR}/root/var/log/your-agent
mkdir -p ${BUILD_DIR}/root/Library/LaunchDaemons

# Copy binary
cp ../../build/agent-darwin-amd64 ${BUILD_DIR}/root/usr/local/bin/your-agent
chmod +x ${BUILD_DIR}/root/usr/local/bin/your-agent

# Copy launchd plist
cp ../../assets/macos/com.yourcompany.agent.plist ${BUILD_DIR}/root/Library/LaunchDaemons/

# Create configuration file
cat > ${BUILD_DIR}/root/var/lib/your-agent/config.json <<EOF
{
  "org_id": "",
  "install_token": "",
  "api_base_url": "https://api.yourcompany.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "max_buffer_size": 104857600,
  "buffer_dir": "/var/lib/your-agent/buffer",
  "heartbeat_interval": "5m",
  "log_level": "info",
  "log_file": "/var/log/your-agent/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5,
  "update_enabled": true,
  "update_check_interval": "1h",
  "tls": {
    "cert_file": "/var/lib/your-agent/certs/agent.crt",
    "key_file": "/var/lib/your-agent/certs/agent.key",
    "ca_file": "/var/lib/your-agent/certs/ca.crt",
    "insecure_skip_verify": false
  }
}
EOF

# Create postinstall script
cat > ${BUILD_DIR}/scripts/postinstall <<'EOF'
#!/bin/bash

# Set permissions
chown -R root:wheel /usr/local/bin/your-agent
chown -R root:wheel /var/lib/your-agent
chown -R root:wheel /var/log/your-agent
chmod 600 /var/lib/your-agent/config.json

# Load launchd service
launchctl load /Library/LaunchDaemons/com.yourcompany.agent.plist

echo ""
echo "=========================================="
echo "Enterprise Agent installed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Edit /var/lib/your-agent/config.json with your org_id and install_token"
echo "2. Restart the service: sudo launchctl kickstart -k system/com.yourcompany.agent"
echo "3. Check logs: tail -f /var/log/your-agent/agent.log"
echo ""

exit 0
EOF

chmod +x ${BUILD_DIR}/scripts/postinstall

# Create preinstall script
cat > ${BUILD_DIR}/scripts/preinstall <<'EOF'
#!/bin/bash

# Stop service if running
launchctl unload /Library/LaunchDaemons/com.yourcompany.agent.plist 2>/dev/null || true

exit 0
EOF

chmod +x ${BUILD_DIR}/scripts/preinstall

# Create Distribution file
cat > ${BUILD_DIR}/Distribution.xml <<EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="1">
    <title>${PACKAGE_NAME}</title>
    <organization>${IDENTIFIER}</organization>
    <domains enable_localSystem="true"/>
    <options customize="never" require-scripts="true" rootVolumeOnly="true" />
    
    <welcome file="welcome.html" mime-type="text/html" />
    <license file="license.txt" mime-type="text/plain" />
    <readme file="readme.html" mime-type="text/html" />
    
    <pkg-ref id="${IDENTIFIER}"/>
    
    <options customize="never" require-scripts="false"/>
    
    <choices-outline>
        <line choice="default">
            <line choice="${IDENTIFIER}"/>
        </line>
    </choices-outline>
    
    <choice id="default"/>
    <choice id="${IDENTIFIER}" visible="false">
        <pkg-ref id="${IDENTIFIER}"/>
    </choice>
    
    <pkg-ref id="${IDENTIFIER}" version="${VERSION}" onConclusion="none">
        ${PACKAGE_NAME}.pkg
    </pkg-ref>
</installer-gui-script>
EOF

# Create welcome message
cat > ${BUILD_DIR}/resources/welcome.html <<EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; }
        h1 { color: #007AFF; }
    </style>
</head>
<body>
    <h1>Welcome to ${PACKAGE_NAME}</h1>
    <p>This installer will install the Enterprise Endpoint Agent on your system.</p>
    <p>The agent provides:</p>
    <ul>
        <li>Zero-trust security with mTLS authentication</li>
        <li>Policy-driven data collection</li>
        <li>Automatic updates with rollback</li>
        <li>Comprehensive logging and audit trails</li>
    </ul>
    <p><strong>Note:</strong> You will need to configure the agent with your organization ID and install token after installation.</p>
</body>
</html>
EOF

# Create readme
cat > ${BUILD_DIR}/resources/readme.html <<EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; }
        code { background: #f5f5f5; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h2>Post-Installation Steps</h2>
    <ol>
        <li>Edit the configuration file:
            <br><code>sudo nano /var/lib/your-agent/config.json</code>
        </li>
        <li>Add your organization ID and install token</li>
        <li>Restart the service:
            <br><code>sudo launchctl kickstart -k system/com.yourcompany.agent</code>
        </li>
        <li>Verify the agent is running:
            <br><code>tail -f /var/log/your-agent/agent.log</code>
        </li>
    </ol>
    
    <h2>Uninstallation</h2>
    <p>To uninstall the agent:</p>
    <pre>
sudo launchctl unload /Library/LaunchDaemons/com.yourcompany.agent.plist
sudo rm -rf /usr/local/bin/your-agent
sudo rm -rf /var/lib/your-agent
sudo rm -rf /var/log/your-agent
sudo rm /Library/LaunchDaemons/com.yourcompany.agent.plist
    </pre>
</body>
</html>
EOF

# Create license
cat > ${BUILD_DIR}/resources/license.txt <<EOF
PROPRIETARY SOFTWARE LICENSE

Copyright (c) $(date +%Y) Your Company. All rights reserved.

This software is proprietary and confidential. Unauthorized copying,
distribution, or use is strictly prohibited.
EOF

# Build component package
pkgbuild --root ${BUILD_DIR}/root \
         --scripts ${BUILD_DIR}/scripts \
         --identifier ${IDENTIFIER} \
         --version ${VERSION} \
         --install-location / \
         ${BUILD_DIR}/${PACKAGE_NAME}.pkg

# Build product package
productbuild --distribution ${BUILD_DIR}/Distribution.xml \
             --resources ${BUILD_DIR}/resources \
             --package-path ${BUILD_DIR} \
             ${PACKAGE_NAME}-${VERSION}.pkg

echo ""
echo "✓ Package built successfully: ${PACKAGE_NAME}-${VERSION}.pkg"
echo ""
echo "Install with: sudo installer -pkg ${PACKAGE_NAME}-${VERSION}.pkg -target /"
echo ""
echo "To sign the package (required for distribution):"
echo "  productsign --sign \"Developer ID Installer: Your Company\" \\"
echo "    ${PACKAGE_NAME}-${VERSION}.pkg ${PACKAGE_NAME}-${VERSION}-signed.pkg"
