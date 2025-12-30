package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/identity"
)

// Engine manages agent policy
type Engine struct {
	cfg      *config.Config
	identity *identity.Manager
	logger   *log.Logger
	mu       sync.RWMutex
	current  *Policy
}

// Policy represents the agent's runtime configuration
type Policy struct {
	Version    string                     `json:"version"`
	UpdatedAt  time.Time                  `json:"updated_at"`
	Collectors map[string]CollectorPolicy `json:"collectors"`
	Update     UpdatePolicy               `json:"update"`
	Telemetry  TelemetryPolicy            `json:"telemetry"`
}

// CollectorPolicy defines settings for a specific collector
type CollectorPolicy struct {
	Enabled  bool                   `json:"enabled"`
	Interval time.Duration          `json:"interval,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// UpdatePolicy defines auto-update settings
type UpdatePolicy struct {
	Enabled       bool          `json:"enabled"`
	Channel       string        `json:"channel"` // stable, beta, dev
	CheckInterval time.Duration `json:"check_interval"`
}

// TelemetryPolicy defines data transmission settings
type TelemetryPolicy struct {
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	Compression   bool          `json:"compression"`
}

// NewEngine creates a new policy engine
func NewEngine(cfg *config.Config, identityMgr *identity.Manager, logger *log.Logger) (*Engine, error) {
	return &Engine{
		cfg:      cfg,
		identity: identityMgr,
		logger:   logger,
		current:  defaultPolicy(),
	}, nil
}

// Refresh fetches the latest policy from the server
func (e *Engine) Refresh(ctx context.Context) error {
	e.logger.Println("Refreshing policy from server...")

	client, err := e.identity.GetHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	url := e.cfg.APIBaseURL + "/api/v1/policy"
	req, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch policy: %w", err)
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		body, _ := io.ReadAll(req.Body)
		return fmt.Errorf("policy fetch failed with status %d: %s", req.StatusCode, string(body))
	}

	var newPolicy Policy
	if err := json.NewDecoder(req.Body).Decode(&newPolicy); err != nil {
		return fmt.Errorf("failed to decode policy: %w", err)
	}

	// Update current policy
	e.mu.Lock()
	oldVersion := e.current.Version
	e.current = &newPolicy
	e.mu.Unlock()

	if oldVersion != newPolicy.Version {
		e.logger.Printf("Policy updated: %s -> %s", oldVersion, newPolicy.Version)
	}

	return nil
}

// Get returns the current policy (thread-safe)
func (e *Engine) Get() *Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current
}

// IsCollectorEnabled checks if a collector is enabled
func (e *Engine) IsCollectorEnabled(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if collectorPolicy, ok := e.current.Collectors[name]; ok {
		return collectorPolicy.Enabled
	}
	return false
}

// GetCollectorInterval returns the interval for a collector
func (e *Engine) GetCollectorInterval(name string) time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if collectorPolicy, ok := e.current.Collectors[name]; ok && collectorPolicy.Interval > 0 {
		return collectorPolicy.Interval
	}
	return e.cfg.CollectionInterval
}

// defaultPolicy returns a safe default policy
func defaultPolicy() *Policy {
	return &Policy{
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
		Collectors: map[string]CollectorPolicy{
			"system": {Enabled: true},
			"cpu":    {Enabled: true, Interval: 60 * time.Second},
			"memory": {Enabled: true, Interval: 60 * time.Second},
			"disk":   {Enabled: true, Interval: 300 * time.Second},
			"network": {
				Enabled:  true,
				Interval: 60 * time.Second,
				Options: map[string]interface{}{
					"collect_mac": false, // Privacy: MAC collection disabled by default
				},
			},
		},
		Update: UpdatePolicy{
			Enabled:       true,
			Channel:       "stable",
			CheckInterval: 1 * time.Hour,
		},
		Telemetry: TelemetryPolicy{
			BatchSize:     100,
			FlushInterval: 5 * time.Minute,
			Compression:   true,
		},
	}
}
