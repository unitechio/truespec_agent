//go:build darwin
// +build darwin

package service

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallMacOSService installs the agent as a launchd daemon
func InstallMacOSService(serviceName string) error {
	plistPath := "/Library/LaunchDaemons/" + serviceName + ".plist"

	// Check if service already exists
	if _, err := os.Stat(plistPath); err == nil {
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--config</string>
        <string>/var/lib/your-agent/config.json</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    
    <key>StandardOutPath</key>
    <string>/var/log/your-agent/stdout.log</string>
    
    <key>StandardErrorPath</key>
    <string>/var/log/your-agent/stderr.log</string>
    
    <key>ThrottleInterval</key>
    <integer>10</integer>
    
    <key>WorkingDirectory</key>
    <string>/var/lib/your-agent</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
`, serviceName, exePath)

	// Write plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load service
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load service: %w", err)
	}

	return nil
}

// UninstallMacOSService removes the launchd daemon
func UninstallMacOSService(serviceName string) error {
	plistPath := "/Library/LaunchDaemons/" + serviceName + ".plist"

	// Unload service
	exec.Command("launchctl", "unload", plistPath).Run()

	// Remove plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	return nil
}

// StartMacOSService starts the launchd daemon
func StartMacOSService(serviceName string) error {
	return exec.Command("launchctl", "start", serviceName).Run()
}

// StopMacOSService stops the launchd daemon
func StopMacOSService(serviceName string) error {
	return exec.Command("launchctl", "stop", serviceName).Run()
}
