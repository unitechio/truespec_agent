package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Bootstrapped bool `json:"bootstrapped"`

	// Agent identity
	OrgID        string `json:"org_id,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	InstallToken string `json:"install_token,omitempty"`

	APIBaseURL string    `json:"api_base_url,omitempty"`
	TLSConfig  TLSConfig `json:"tls,omitempty"`

	// Data collection
	CollectionInterval time.Duration `json:"collection_interval,omitempty"`
	BatchSize          int           `json:"batch_size,omitempty"`

	// Buffering
	MaxBufferSize int64  `json:"max_buffer_size,omitempty"` // bytes
	BufferDir     string `json:"buffer_dir,omitempty"`

	// Health & monitoring
	HeartbeatInterval time.Duration `json:"heartbeat_interval,omitempty"`

	// Logging
	LogLevel string `json:"log_level,omitempty"`
	LogFile  string `json:"log_file,omitempty"`

	// Update configuration
	UpdateEnabled       bool          `json:"update_enabled,omitempty"`
	UpdateCheckInterval time.Duration `json:"update_check_interval,omitempty"`
}

// TLSConfig holds TLS/mTLS configuration
type TLSConfig struct {
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"` // Only for testing
}

// Load reads configuration from a JSON file
// Returns ErrConfigNotFound if the file doesn't exist
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override empty values with environment variables
	if cfg.OrgID == "" {
		if envOrgID := os.Getenv("ORG_ID"); envOrgID != "" {
			cfg.OrgID = envOrgID
		}
	}
	if cfg.InstallToken == "" {
		if envToken := os.Getenv("INSTALL_TOKEN"); envToken != "" {
			cfg.InstallToken = envToken
		}
	}

	return &cfg, nil
}

// Save writes configuration to a JSON file
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateBootstrap validates the minimal configuration needed for bootstrap
func (c *Config) ValidateBootstrap() error {
	if c.OrgID == "" {
		return fmt.Errorf("%w: org_id is required", ErrInvalidBootstrap)
	}
	if c.InstallToken == "" {
		return fmt.Errorf("%w: install_token is required", ErrInvalidBootstrap)
	}
	return nil
}

// ValidateRuntime validates the full configuration needed for normal operation
func (c *Config) ValidateRuntime() error {
	if !c.Bootstrapped {
		return ErrNotBootstrapped
	}
	if c.OrgID == "" {
		return fmt.Errorf("%w: org_id is required", ErrInvalidRuntime)
	}
	if c.AgentID == "" {
		return fmt.Errorf("%w: agent_id is required", ErrInvalidRuntime)
	}
	if c.APIBaseURL == "" {
		return fmt.Errorf("%w: api_base_url is required", ErrInvalidRuntime)
	}
	if c.CollectionInterval < 10*time.Second {
		return fmt.Errorf("%w: collection_interval must be at least 10 seconds", ErrInvalidRuntime)
	}
	if c.BatchSize < 1 || c.BatchSize > 1000 {
		return fmt.Errorf("%w: batch_size must be between 1 and 1000", ErrInvalidRuntime)
	}
	return nil
}

// MarkBootstrapped marks the configuration as bootstrapped and sets required runtime fields
func (c *Config) MarkBootstrapped(agentID, apiBaseURL string) {
	c.Bootstrapped = true
	c.AgentID = agentID
	c.APIBaseURL = apiBaseURL
	c.InstallToken = "" // Clear install token after successful bootstrap
}

// NewBootstrapConfig creates a minimal configuration for bootstrap from environment variables
func NewBootstrapConfig() *Config {
	return &Config{
		Bootstrapped:        false,
		OrgID:               os.Getenv("ORG_ID"),
		InstallToken:        os.Getenv("INSTALL_TOKEN"),
		CollectionInterval:  60 * time.Second,
		BatchSize:           100,
		MaxBufferSize:       100 * 1024 * 1024, // 100 MB
		BufferDir:           getDefaultBufferDir(),
		HeartbeatInterval:   5 * time.Minute,
		LogLevel:            "info",
		LogFile:             getDefaultLogFile(),
		UpdateEnabled:       true,
		UpdateCheckInterval: 1 * time.Hour,
		TLSConfig: TLSConfig{
			CertFile:           getDefaultCertPath(),
			KeyFile:            getDefaultKeyPath(),
			CAFile:             getDefaultCAPath(),
			InsecureSkipVerify: false,
		},
	}
}

// DefaultConfig returns a fully populated configuration with sensible defaults
// This is primarily used for testing
func DefaultConfig() *Config {
	cfg := NewBootstrapConfig()
	cfg.Bootstrapped = true
	cfg.AgentID = "test-agent-id"
	cfg.APIBaseURL = "https://api.unitechio.space"
	return cfg
}

func getDefaultBufferDir() string {
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\buffer`
	}
	return "/var/lib/your-agent/buffer"
}

func getDefaultLogFile() string {
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\logs\agent.log`
	}
	return "/var/log/your-agent/agent.log"
}

func getDefaultCertPath() string {
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\certs\agent.crt`
	}
	return "/var/lib/your-agent/certs/agent.crt"
}

func getDefaultKeyPath() string {
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\certs\agent.key`
	}
	return "/var/lib/your-agent/certs/agent.key"
}

func getDefaultCAPath() string {
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\certs\ca.crt`
	}
	return "/var/lib/your-agent/certs/ca.crt"
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
