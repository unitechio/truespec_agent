#!/bin/bash
# Build .deb package for Enterprise Agent

set -e

VERSION="1.0.0"
ARCH="amd64"
PACKAGE_NAME="your-agent"
BUILD_DIR="build/deb"

echo "Building .deb package for ${PACKAGE_NAME} v${VERSION}..."

# Clean previous build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Create directory structure
mkdir -p ${BUILD_DIR}/DEBIAN
mkdir -p ${BUILD_DIR}/usr/local/bin
mkdir -p ${BUILD_DIR}/etc/${PACKAGE_NAME}
mkdir -p ${BUILD_DIR}/var/lib/${PACKAGE_NAME}/{certs,buffer}
mkdir -p ${BUILD_DIR}/var/log/${PACKAGE_NAME}
mkdir -p ${BUILD_DIR}/lib/systemd/system

# Copy binary
cp ../../build/agent-linux-${ARCH} ${BUILD_DIR}/usr/local/bin/${PACKAGE_NAME}
chmod +x ${BUILD_DIR}/usr/local/bin/${PACKAGE_NAME}

# Copy configuration template
cat > ${BUILD_DIR}/etc/${PACKAGE_NAME}/config.json <<EOF
{
  "org_id": "",
  "install_token": "",
  "api_base_url": "https://api.yourcompany.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "max_buffer_size": 104857600,
  "buffer_dir": "/var/lib/${PACKAGE_NAME}/buffer",
  "heartbeat_interval": "5m",
  "log_level": "info",
  "log_file": "/var/log/${PACKAGE_NAME}/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5,
  "update_enabled": true,
  "update_check_interval": "1h",
  "tls": {
    "cert_file": "/var/lib/${PACKAGE_NAME}/certs/agent.crt",
    "key_file": "/var/lib/${PACKAGE_NAME}/certs/agent.key",
    "ca_file": "/var/lib/${PACKAGE_NAME}/certs/ca.crt",
    "insecure_skip_verify": false
  }
}
EOF

# Copy systemd service file
cp ../../assets/linux/your-agent.service ${BUILD_DIR}/lib/systemd/system/${PACKAGE_NAME}.service

# Create control file
cat > ${BUILD_DIR}/DEBIAN/control <<EOF
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Section: admin
Priority: optional
Architecture: ${ARCH}
Maintainer: Your Company <support@yourcompany.com>
Description: Enterprise Endpoint Agent
 Production-grade endpoint monitoring agent with zero-trust security,
 mTLS authentication, policy-driven data collection, and auto-update
 capabilities.
Homepage: https://yourcompany.com/agent
EOF

# Create postinst script
cat > ${BUILD_DIR}/DEBIAN/postinst <<'EOF'
#!/bin/bash
set -e

PACKAGE_NAME="your-agent"

# Create dedicated user
if ! id -u ${PACKAGE_NAME} > /dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/sbin/nologin ${PACKAGE_NAME}
fi

# Set permissions
chown -R ${PACKAGE_NAME}:${PACKAGE_NAME} /var/lib/${PACKAGE_NAME}
chown -R ${PACKAGE_NAME}:${PACKAGE_NAME} /var/log/${PACKAGE_NAME}
chown -R ${PACKAGE_NAME}:${PACKAGE_NAME} /etc/${PACKAGE_NAME}
chmod 600 /etc/${PACKAGE_NAME}/config.json

# Reload systemd
systemctl daemon-reload

# Enable service (but don't start yet - user needs to configure)
systemctl enable ${PACKAGE_NAME}

echo ""
echo "=========================================="
echo "Enterprise Agent installed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Edit /etc/${PACKAGE_NAME}/config.json with your org_id and install_token"
echo "2. Start the service: sudo systemctl start ${PACKAGE_NAME}"
echo "3. Check status: sudo systemctl status ${PACKAGE_NAME}"
echo ""
EOF

chmod +x ${BUILD_DIR}/DEBIAN/postinst

# Create prerm script
cat > ${BUILD_DIR}/DEBIAN/prerm <<'EOF'
#!/bin/bash
set -e

PACKAGE_NAME="your-agent"

# Stop service
if systemctl is-active --quiet ${PACKAGE_NAME}; then
    systemctl stop ${PACKAGE_NAME}
fi

# Disable service
systemctl disable ${PACKAGE_NAME} || true

exit 0
EOF

chmod +x ${BUILD_DIR}/DEBIAN/prerm

# Create postrm script
cat > ${BUILD_DIR}/DEBIAN/postrm <<'EOF'
#!/bin/bash
set -e

PACKAGE_NAME="your-agent"

if [ "$1" = "purge" ]; then
    # Remove user
    userdel ${PACKAGE_NAME} 2>/dev/null || true
    
    # Remove data directories
    rm -rf /var/lib/${PACKAGE_NAME}
    rm -rf /var/log/${PACKAGE_NAME}
    rm -rf /etc/${PACKAGE_NAME}
fi

# Reload systemd
systemctl daemon-reload || true

exit 0
EOF

chmod +x ${BUILD_DIR}/DEBIAN/postrm

# Build package
dpkg-deb --build ${BUILD_DIR} ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb

echo ""
echo "✓ Package built successfully: ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "Install with: sudo dpkg -i ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
