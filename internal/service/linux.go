//go:build linux
// +build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallLinuxService installs the agent as a systemd service
func InstallLinuxService(serviceName string) error {
	// Copy service file to systemd directory
	serviceFile := "/etc/systemd/system/" + serviceName + ".service"

	// Check if service already exists
	if _, err := os.Stat(serviceFile); err == nil {
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=Enterprise Agent
Documentation=https://docs.unitechio.com/agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=your-agent
Group=your-agent
ExecStart=%s --config /etc/your-agent/config.json
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/your-agent /var/log/your-agent

# Resource limits
LimitNOFILE=65536
MemoryMax=512M
CPUQuota=50%%

[Install]
WantedBy=multi-user.target
`, exePath)

	// Write service file
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// UninstallLinuxService removes the systemd service
func UninstallLinuxService(serviceName string) error {
	// Stop service
	exec.Command("systemctl", "stop", serviceName).Run()

	// Disable service
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove service file
	serviceFile := "/etc/systemd/system/" + serviceName + ".service"
	if err := os.Remove(serviceFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// StartLinuxService starts the systemd service
func StartLinuxService(serviceName string) error {
	return exec.Command("systemctl", "start", serviceName).Run()
}

// StopLinuxService stops the systemd service
func StopLinuxService(serviceName string) error {
	return exec.Command("systemctl", "stop", serviceName).Run()
}
