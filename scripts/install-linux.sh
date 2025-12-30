#!/bin/bash
# Linux installer script for Enterprise Agent

set -e

AGENT_NAME="your-agent"
AGENT_USER="your-agent"
AGENT_GROUP="your-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/${AGENT_NAME}"
DATA_DIR="/var/lib/${AGENT_NAME}"
LOG_DIR="/var/log/${AGENT_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Enterprise Agent Installer${NC}"
echo "=============================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
   echo -e "${RED}Error: This script must be run as root${NC}"
   exit 1
fi

# Parse command-line arguments
ORG_ID=""
INSTALL_TOKEN=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --org-id)
      ORG_ID="$2"
      shift 2
      ;;
    --token)
      INSTALL_TOKEN="$2"
      shift 2
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
  esac
done

if [ -z "$ORG_ID" ] || [ -z "$INSTALL_TOKEN" ]; then
    echo -e "${RED}Error: --org-id and --token are required${NC}"
    echo "Usage: $0 --org-id YOUR_ORG_ID --token YOUR_TOKEN"
    exit 1
fi

echo -e "${YELLOW}Installing Enterprise Agent...${NC}"

# Create dedicated user and group
if ! id -u $AGENT_USER > /dev/null 2>&1; then
    echo "Creating user $AGENT_USER..."
    useradd --system --no-create-home --shell /usr/sbin/nologin $AGENT_USER
fi

# Create directories
echo "Creating directories..."
mkdir -p $CONFIG_DIR
mkdir -p $DATA_DIR/{certs,buffer}
mkdir -p $LOG_DIR

# Copy binary
echo "Installing agent binary..."
cp ./agent $INSTALL_DIR/$AGENT_NAME
chmod +x $INSTALL_DIR/$AGENT_NAME

# Create configuration file
echo "Creating configuration..."
cat > $CONFIG_DIR/config.json <<EOF
{
  "org_id": "$ORG_ID",
  "install_token": "$INSTALL_TOKEN",
  "api_base_url": "https://api.yourcompany.com",
  "collection_interval": "60s",
  "batch_size": 100,
  "max_buffer_size": 104857600,
  "buffer_dir": "$DATA_DIR/buffer",
  "heartbeat_interval": "5m",
  "log_level": "info",
  "log_file": "$LOG_DIR/agent.log",
  "update_enabled": true,
  "update_check_interval": "1h",
  "tls": {
    "cert_file": "$DATA_DIR/certs/agent.crt",
    "key_file": "$DATA_DIR/certs/agent.key",
    "ca_file": "$DATA_DIR/certs/ca.crt",
    "insecure_skip_verify": false
  }
}
EOF

# Set permissions
echo "Setting permissions..."
chown -R $AGENT_USER:$AGENT_GROUP $CONFIG_DIR
chown -R $AGENT_USER:$AGENT_GROUP $DATA_DIR
chown -R $AGENT_USER:$AGENT_GROUP $LOG_DIR
chmod 600 $CONFIG_DIR/config.json

# Install systemd service
echo "Installing systemd service..."
cat > /etc/systemd/system/${AGENT_NAME}.service <<EOF
[Unit]
Description=Enterprise Agent
Documentation=https://docs.yourcompany.com/agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$AGENT_USER
Group=$AGENT_GROUP
ExecStart=$INSTALL_DIR/$AGENT_NAME --config $CONFIG_DIR/config.json
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Resource limits
LimitNOFILE=65536
MemoryMax=512M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Enable service
echo "Enabling service..."
systemctl enable $AGENT_NAME

# Bootstrap agent
echo "Bootstrapping agent..."
sudo -u $AGENT_USER $INSTALL_DIR/$AGENT_NAME --config $CONFIG_DIR/config.json --bootstrap

# Start service
echo "Starting service..."
systemctl start $AGENT_NAME

# Check status
sleep 2
if systemctl is-active --quiet $AGENT_NAME; then
    echo -e "${GREEN}✓ Installation successful!${NC}"
    echo ""
    echo "Service status:"
    systemctl status $AGENT_NAME --no-pager
    echo ""
    echo "Useful commands:"
    echo "  View logs:    journalctl -u $AGENT_NAME -f"
    echo "  Stop service: systemctl stop $AGENT_NAME"
    echo "  Restart:      systemctl restart $AGENT_NAME"
else
    echo -e "${RED}✗ Service failed to start${NC}"
    echo "Check logs with: journalctl -u $AGENT_NAME -n 50"
    exit 1
fi
