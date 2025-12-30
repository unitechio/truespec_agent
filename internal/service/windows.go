//go:build windows
// +build windows

package service

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/unitechio/agent/internal/config"
)

// windowsService implements the Windows Service interface
type windowsService struct {
	cfg    *config.Config
	logger *log.Logger
	stopCh chan struct{}
}

// Execute is called by the Windows Service Manager
func (ws *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	changes <- svc.Status{State: svc.StartPending}
	ws.logger.Println("Windows Service starting...")

	// Start the agent in a goroutine
	// Note: The actual agent logic should be implemented here
	// For now, we just keep the service running
	go func() {
		ws.logger.Println("Agent running as Windows Service")
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	ws.logger.Println("Windows Service running")

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				ws.logger.Println("Windows Service stopping...")
				changes <- svc.Status{State: svc.StopPending}
				close(ws.stopCh)
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				ws.logger.Printf("Unexpected control request #%d", c)
			}
		}
	}

	changes <- svc.Status{State: svc.Stopped}
	ws.logger.Println("Windows Service stopped")
	return
}

// RunAsService runs the agent as a Windows Service
func RunAsService(cfg *config.Config, logger *log.Logger) error {
	ws := &windowsService{
		cfg:    cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	return svc.Run("YourAgentService", ws)
}

// IsWindowsService checks if the process is running as a Windows Service
func IsWindowsService() (bool, error) {
	return svc.IsWindowsService()
}

// InstallWindowsService installs the agent as a Windows Service
func InstallWindowsService(serviceName, displayName, description string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	config := mgr.Config{
		DisplayName:      displayName,
		Description:      description,
		StartType:        mgr.StartAutomatic,
		ServiceStartName: `NT AUTHORITY\SYSTEM`,
	}

	s, err = m.CreateService(serviceName, exePath, config)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	// Set recovery actions (restart on failure)
	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
	}, 86400) // Reset failure count after 24 hours
	if err != nil {
		return fmt.Errorf("failed to set recovery actions: %w", err)
	}

	return nil
}

// UninstallWindowsService removes the Windows Service
func UninstallWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer s.Close()

	// Stop the service if running
	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("failed to query service status: %w", err)
	}

	if status.State != svc.Stopped {
		_, err = s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		// Wait for service to stop
		timeout := time.Now().Add(30 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("timeout waiting for service to stop")
			}
			time.Sleep(300 * time.Millisecond)
			status, err = s.Query()
			if err != nil {
				return fmt.Errorf("failed to query service status: %w", err)
			}
		}
	}

	// Delete the service
	err = s.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// StartWindowsService starts the Windows Service
func StartWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}
