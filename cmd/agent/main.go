package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/health"
	"github.com/unitechio/agent/internal/identity"
	"github.com/unitechio/agent/internal/policy"
	"github.com/unitechio/agent/internal/scheduler"
)

const version = "1.0.0"

func main() {
	configPath := flag.String("config", getDefaultConfigPath(), "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("enterprise-agent v%s\n", version)
		os.Exit(0)
	}

	logger := log.New(os.Stdout, "[AGENT] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("Starting enterprise-agent v%s", version)

	// Context dùng chung cho agent
	ctx, cancel := context.WithCancel(context.Background())

	// Start agent ở background
	go func() {
		if err := run(ctx, *configPath, logger); err != nil {
			logger.Printf("Agent stopped with error: %v", err)
			systray.Quit()
		}
	}()

	// Signal handling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		sig := <-sigChan
		logger.Printf("Received signal: %v, shutting down...", sig)
		cancel()
		systray.Quit()
	}()

	systray.Run(onReady, func() {
		logger.Println("Systray exiting")
		cancel()
	})
}

func run(ctx context.Context, configPath string, logger *log.Logger) error {
	// Step 1: Try to load existing config
	cfg, err := config.Load(configPath)

	if errors.Is(err, config.ErrConfigNotFound) {
		// Step 2: No config exists - enter bootstrap mode
		logger.Println("No configuration found, starting bootstrap process...")

		// Step 3: Create minimal bootstrap config from environment
		cfg = config.NewBootstrapConfig()

		// DEBUG: Log values being used for bootstrap
		logger.Printf("DEBUG: Loaded bootstrap config: OrgID='%s', InstallToken='%s'(len:%d)", cfg.OrgID, cfg.InstallToken, len(cfg.InstallToken))

		if err := cfg.ValidateBootstrap(); err != nil {
			return fmt.Errorf("bootstrap validation failed: %w\nPlease set ORG_ID and INSTALL_TOKEN environment variables", err)
		}

		// Step 4: Run bootstrap with retry
		logger.Println("Starting bootstrap process...")
		cfg, err = identity.RunBootstrap(ctx, cfg)
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		// Step 5: Save the new config
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		logger.Println("Bootstrap successful, configuration saved")

	} else if err != nil {
		// Step 6: Config load failed for other reasons
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Step 7: Validate runtime configuration
	if err := cfg.ValidateRuntime(); err != nil {
		return fmt.Errorf("invalid runtime configuration: %w", err)
	}

	logger.Printf("Configuration loaded successfully (Agent ID: %s)", cfg.AgentID)

	// Step 8: Initialize identity manager
	identityMgr, err := identity.NewManager(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create identity manager: %w", err)
	}

	// Step 9: Check if re-bootstrap is needed (cert expiration)
	if identityMgr.NeedsRebootstrap() {
		logger.Println("Certificate expired or invalid, re-bootstrapping...")

		// Re-bootstrap
		cfg.Bootstrapped = false // Reset bootstrap state
		cfg, err = identity.RunBootstrap(ctx, cfg)
		if err != nil {
			return fmt.Errorf("re-bootstrap failed: %w", err)
		}

		// Save updated config
		if err := cfg.Save(configPath); err != nil {
			logger.Printf("Warning: failed to save config after re-bootstrap: %v", err)
		}

		// Recreate identity manager with new certs
		identityMgr, err = identity.NewManager(cfg, logger)
		if err != nil {
			return fmt.Errorf("failed to recreate identity manager: %w", err)
		}

		logger.Println("Re-bootstrap successful")
	}

	// Step 10: Verify identity
	if err := identityMgr.VerifyIdentity(); err != nil {
		return fmt.Errorf("identity verification failed: %w", err)
	}

	logger.Printf("Identity verified: Agent ID = %s", identityMgr.GetAgentID())

	// Step 11: Initialize policy engine
	policyEngine, err := policy.NewEngine(cfg, identityMgr, logger)
	if err != nil {
		return fmt.Errorf("failed to create policy engine: %w", err)
	}

	// Fetch initial policy
	if err := policyEngine.Refresh(ctx); err != nil {
		logger.Printf("Warning: failed to fetch initial policy, using defaults: %v", err)
	}

	// Step 12: Initialize health monitor
	healthMonitor := health.NewMonitor(cfg, identityMgr, logger)
	healthMonitor.Start(ctx)

	// Step 13: Initialize scheduler
	sched := scheduler.New(cfg, policyEngine, identityMgr, logger)
	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	logger.Println("Agent running successfully")

	// Step 14: Periodically refresh policy and check for updates
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Println("Shutting down agent...")

			// Stop components gracefully
			sched.Stop()
			healthMonitor.Stop()

			return nil

		case <-ticker.C:
			// Periodic policy refresh
			if err := policyEngine.Refresh(ctx); err != nil {
				logger.Printf("Failed to refresh policy: %v", err)
			}
		}
	}
}

func getDefaultConfigPath() string {
	if path := os.Getenv("AGENT_CONFIG"); path != "" {
		return path
	}

	switch runtime.GOOS {
	case "windows":
		return `C:\ProgramData\unitechio\Agent\config.json`
	case "darwin":
		return "/Library/Application Support/unitechio/agent/config.json"
	default: // Linux
		return "/etc/unitechio/agent/config.json"
	}
}

func onReady() {
	systray.SetTitle("My Agent")
	systray.SetTooltip("My Agent is running")

	mOpen := systray.AddMenuItem("Open Dashboard", "Open web UI")
	mQuit := systray.AddMenuItem("Quit", "Exit app")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser("http://localhost:8080")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}

	_ = exec.Command(cmd, args...).Start()
}
func onExit() {}
