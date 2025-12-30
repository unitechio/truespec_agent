#!/bin/bash
# Linux uninstaller script for Enterprise Agent

set -e

AGENT_NAME="your-agent"
AGENT_USER="your-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/${AGENT_NAME}"
DATA_DIR="/var/lib/${AGENT_NAME}"
LOG_DIR="/var/log/${AGENT_NAME}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Enterprise Agent Uninstaller${NC}"
echo "==============================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
   echo -e "${RED}Error: This script must be run as root${NC}"
   exit 1
fi

# Confirm uninstallation
read -p "Are you sure you want to uninstall the agent? This will remove all data. (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Uninstallation cancelled."
    exit 0
fi

echo "Stopping service..."
systemctl stop $AGENT_NAME || true

echo "Disabling service..."
systemctl disable $AGENT_NAME || true

echo "Removing systemd service file..."
rm -f /etc/systemd/system/${AGENT_NAME}.service
systemctl daemon-reload

echo "Removing binary..."
rm -f $INSTALL_DIR/$AGENT_NAME

echo "Removing configuration..."
rm -rf $CONFIG_DIR

echo "Removing data..."
rm -rf $DATA_DIR

echo "Removing logs..."
rm -rf $LOG_DIR

echo "Removing user..."
userdel $AGENT_USER 2>/dev/null || true

echo -e "${GREEN}✓ Uninstallation complete${NC}"
echo ""
echo "The agent has been completely removed from this system."
