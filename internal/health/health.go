package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/identity"
)

type Monitor struct {
	cfg       *config.Config
	identity  *identity.Manager
	logger    *log.Logger
	stopCh    chan struct{}
	startTime time.Time
}

type HealthStatus struct {
	AgentID       string    `json:"agent_id"`
	Version       string    `json:"version"`
	Status        string    `json:"status"` // healthy, degraded, unhealthy
	Uptime        int64     `json:"uptime_seconds"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	MemoryUsageMB float64   `json:"memory_usage_mb"`
	Goroutines    int       `json:"goroutines"`
	Errors        []string  `json:"errors,omitempty"`
}

func NewMonitor(cfg *config.Config, identityMgr *identity.Manager, logger *log.Logger) *Monitor {
	return &Monitor{
		cfg:       cfg,
		identity:  identityMgr,
		logger:    logger,
		stopCh:    make(chan struct{}),
		startTime: time.Now(),
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.logger.Println("Starting health monitor...")

	go func() {
		ticker := time.NewTicker(m.cfg.HeartbeatInterval)
		defer ticker.Stop()

		// Send initial heartbeat
		m.sendHeartbeat(ctx)

		for {
			select {
			case <-ticker.C:
				m.sendHeartbeat(ctx)
			case <-m.stopCh:
				m.logger.Println("Health monitor stopped")
				return
			case <-ctx.Done():
				m.logger.Println("Health monitor context cancelled")
				return
			}
		}
	}()
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}

func (m *Monitor) sendHeartbeat(ctx context.Context) {
	status := m.getHealthStatus()

	client, err := m.identity.GetHTTPClient()
	if err != nil {
		m.logger.Printf("Failed to create HTTP client for heartbeat: %v", err)
		return
	}

	url := m.cfg.APIBaseURL + "/api/v1/heartbeat"

	body, err := json.Marshal(status)
	if err != nil {
		m.logger.Printf("Failed to marshal heartbeat: %v", err)
		return
	}

	req, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		m.logger.Printf("Failed to send heartbeat: %v", err)
		return
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		m.logger.Printf("Heartbeat failed with status %d", req.StatusCode)
		return
	}

	m.logger.Printf("Heartbeat sent successfully (status: %s)", status.Status)
}

func (m *Monitor) getHealthStatus() *HealthStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &HealthStatus{
		AgentID:       m.identity.GetAgentID(),
		Version:       "1.0.0",
		Status:        "healthy",
		Uptime:        int64(time.Since(m.startTime).Seconds()),
		LastHeartbeat: time.Now(),
		MemoryUsageMB: float64(memStats.Alloc) / 1024 / 1024,
		Goroutines:    runtime.NumGoroutine(),
	}
}

// CheckHealth performs a health check
func (m *Monitor) CheckHealth() error {
	status := m.getHealthStatus()

	if status.MemoryUsageMB > 500 {
		return fmt.Errorf("high memory usage: %.2f MB", status.MemoryUsageMB)
	}

	// Check goroutine count (alert if > 1000)
	if status.Goroutines > 1000 {
		return fmt.Errorf("high goroutine count: %d", status.Goroutines)
	}

	return nil
}
