//go:build !windows
// +build !windows

package service

import (
	"log"

	"github.com/unitechio/agent/internal/config"
)

// RunAsService runs the agent as an OS service (non-Windows stub)
func RunAsService(cfg *config.Config, logger *log.Logger) error {
	logger.Println("RunAsService called on non-Windows platform")
	return nil
}

// IsWindowsService checks if running as Windows Service (non-Windows stub)
func IsWindowsService() (bool, error) {
	return false, nil
}
