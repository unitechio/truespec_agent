#!/bin/bash
# Build .rpm package for Enterprise Agent

set -e

VERSION="1.0.0"
RELEASE="1"
ARCH="x86_64"
PACKAGE_NAME="your-agent"
BUILD_DIR="build/rpm"

echo "Building .rpm package for ${PACKAGE_NAME} v${VERSION}..."

# Clean previous build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Create source tarball
TARBALL="${PACKAGE_NAME}-${VERSION}.tar.gz"
mkdir -p ${BUILD_DIR}/SOURCES/${PACKAGE_NAME}-${VERSION}
cp ../../build/agent-linux-${ARCH} ${BUILD_DIR}/SOURCES/${PACKAGE_NAME}-${VERSION}/agent
cp ../../assets/linux/your-agent.service ${BUILD_DIR}/SOURCES/${PACKAGE_NAME}-${VERSION}/
tar -czf ${BUILD_DIR}/SOURCES/${TARBALL} -C ${BUILD_DIR}/SOURCES ${PACKAGE_NAME}-${VERSION}

# Create spec file
cat > ${BUILD_DIR}/SPECS/${PACKAGE_NAME}.spec <<EOF
Name:           ${PACKAGE_NAME}
Version:        ${VERSION}
Release:        ${RELEASE}%{?dist}
Summary:        Enterprise Endpoint Agent
License:        Proprietary
URL:            https://yourcompany.com/agent
Source0:        %{name}-%{version}.tar.gz

BuildArch:      ${ARCH}
Requires:       systemd

%description
Production-grade endpoint monitoring agent with zero-trust security,
mTLS authentication, policy-driven data collection, and auto-update
capabilities.

%prep
%setup -q

%install
rm -rf \$RPM_BUILD_ROOT

# Create directories
mkdir -p \$RPM_BUILD_ROOT/usr/local/bin
mkdir -p \$RPM_BUILD_ROOT/etc/%{name}
mkdir -p \$RPM_BUILD_ROOT/var/lib/%{name}/{certs,buffer}
mkdir -p \$RPM_BUILD_ROOT/var/log/%{name}
mkdir -p \$RPM_BUILD_ROOT/usr/lib/systemd/system

# Install binary
install -m 0755 agent \$RPM_BUILD_ROOT/usr/local/bin/%{name}

# Install systemd service
install -m 0644 %{name}.service \$RPM_BUILD_ROOT/usr/lib/systemd/system/%{name}.service

# Create config file
cat > \$RPM_BUILD_ROOT/etc/%{name}/config.json <<EOFCONFIG
{
  "org_id": "",
  "install_token": "",
  "api_base_url": "https://api.yourcompany.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "max_buffer_size": 104857600,
  "buffer_dir": "/var/lib/%{name}/buffer",
  "heartbeat_interval": "5m",
  "log_level": "info",
  "log_file": "/var/log/%{name}/agent.log",
  "log_max_size_mb": 100,
  "log_max_backups": 5,
  "update_enabled": true,
  "update_check_interval": "1h",
  "tls": {
    "cert_file": "/var/lib/%{name}/certs/agent.crt",
    "key_file": "/var/lib/%{name}/certs/agent.key",
    "ca_file": "/var/lib/%{name}/certs/ca.crt",
    "insecure_skip_verify": false
  }
}
EOFCONFIG

%files
%defattr(-,root,root,-)
/usr/local/bin/%{name}
/usr/lib/systemd/system/%{name}.service
%config(noreplace) /etc/%{name}/config.json
%dir /var/lib/%{name}
%dir /var/lib/%{name}/certs
%dir /var/lib/%{name}/buffer
%dir /var/log/%{name}

%pre
# Create user
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || useradd -r -g %{name} -s /sbin/nologin -c "Enterprise Agent" %{name}
exit 0

%post
# Set permissions
chown -R %{name}:%{name} /var/lib/%{name}
chown -R %{name}:%{name} /var/log/%{name}
chown -R %{name}:%{name} /etc/%{name}
chmod 600 /etc/%{name}/config.json

# Reload systemd
systemctl daemon-reload

# Enable service
systemctl enable %{name}

echo ""
echo "=========================================="
echo "Enterprise Agent installed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Edit /etc/%{name}/config.json with your org_id and install_token"
echo "2. Start the service: sudo systemctl start %{name}"
echo "3. Check status: sudo systemctl status %{name}"
echo ""

%preun
if [ \$1 -eq 0 ]; then
    # Uninstall
    systemctl stop %{name} || true
    systemctl disable %{name} || true
fi

%postun
if [ \$1 -eq 0 ]; then
    # Uninstall - remove user and data
    userdel %{name} 2>/dev/null || true
    rm -rf /var/lib/%{name}
    rm -rf /var/log/%{name}
fi
systemctl daemon-reload || true

%changelog
* $(date "+%a %b %d %Y") Your Company <support@yourcompany.com> - ${VERSION}-${RELEASE}
- Initial release
EOF

# Build RPM
rpmbuild --define "_topdir ${PWD}/${BUILD_DIR}" -ba ${BUILD_DIR}/SPECS/${PACKAGE_NAME}.spec

# Copy to current directory
cp ${BUILD_DIR}/RPMS/${ARCH}/${PACKAGE_NAME}-${VERSION}-${RELEASE}.*.${ARCH}.rpm .

echo ""
echo "✓ Package built successfully: ${PACKAGE_NAME}-${VERSION}-${RELEASE}.*.${ARCH}.rpm"
echo ""
echo "Install with: sudo rpm -i ${PACKAGE_NAME}-${VERSION}-${RELEASE}.*.${ARCH}.rpm"
