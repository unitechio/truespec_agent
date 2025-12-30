package identity

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/unitechio/agent/internal/config"
)

// Manager handles agent identity and mTLS certificates
type Manager struct {
	cfg      *config.Config
	logger   *log.Logger
	agentID  string
	certPath string
	keyPath  string
	caPath   string
}

// BootstrapRequest is sent to the server during initial registration
type BootstrapRequest struct {
	OrgID        string `json:"org_id"`
	InstallToken string `json:"install_token"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	AgentVersion string `json:"agent_version"`
}

// BootstrapResponse contains the agent identity and certificates
type BootstrapResponse struct {
	AgentID     string `json:"agent_id"`
	APIBaseURL  string `json:"api_base_url"` // Server-provided API endpoint
	Certificate string `json:"certificate"`  // PEM-encoded X.509 certificate
	PrivateKey  string `json:"private_key"`  // PEM-encoded private key
	CACert      string `json:"ca_cert"`      // PEM-encoded CA certificate
	Policy      string `json:"policy"`       // Initial policy JSON
	ExpiresAt   string `json:"expires_at"`   // Certificate expiration timestamp (RFC3339)
}

// NewManager creates a new identity manager
func NewManager(cfg *config.Config, logger *log.Logger) (*Manager, error) {
	certDir := filepath.Dir(cfg.TLSConfig.CertFile)
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	return &Manager{
		cfg:      cfg,
		logger:   logger,
		certPath: cfg.TLSConfig.CertFile,
		keyPath:  cfg.TLSConfig.KeyFile,
		caPath:   cfg.TLSConfig.CAFile,
	}, nil
}

// HasIdentity checks if the agent has been bootstrapped
func (m *Manager) HasIdentity() bool {
	// Check if certificate and key files exist
	if _, err := os.Stat(m.certPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(m.keyPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(m.caPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// NeedsRebootstrap checks if the agent needs to re-bootstrap
// This can happen if certificates are expired or corrupted
func (m *Manager) NeedsRebootstrap() bool {
	// If no identity exists, definitely need bootstrap
	if !m.HasIdentity() {
		return true
	}

	// Try to verify the identity
	if err := m.VerifyIdentity(); err != nil {
		m.logger.Printf("Identity verification failed: %v", err)
		return true
	}

	// Check certificate expiration
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return true
	}

	cert, err := tls.X509KeyPair(certPEM, nil)
	if err != nil {
		return true
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return true
	}

	// Check if cert expires within 24 hours
	if time.Until(x509Cert.NotAfter) < 24*time.Hour {
		m.logger.Printf("Certificate expires soon: %v", x509Cert.NotAfter)
		return true
	}

	return false
}

// RunBootstrap performs the complete bootstrap flow and returns a fully configured Config
// This is the main entry point for bootstrapping from scratch
func RunBootstrap(ctx context.Context, cfg *config.Config) (*config.Config, error) {
	if err := cfg.ValidateBootstrap(); err != nil {
		return nil, fmt.Errorf("bootstrap validation failed: %w", err)
	}

	// Create identity manager
	logger := log.New(os.Stdout, "[BOOTSTRAP] ", log.LstdFlags|log.Lshortfile)
	manager, err := NewManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity manager: %w", err)
	}

	// Perform bootstrap with retry
	var resp *BootstrapResponse
	retryConfig := DefaultRetryConfig()
	err = RetryWithBackoff(ctx, retryConfig, func() error {
		var retryErr error
		resp, retryErr = manager.callBootstrapAPI(ctx, BootstrapRequest{
			OrgID:        cfg.OrgID,
			InstallToken: cfg.InstallToken,
			Hostname:     getHostname(),
			OS:           getOS(),
			Arch:         getArch(),
			AgentVersion: "1.0.0",
		})
		return retryErr
	})
	if err != nil {
		return nil, fmt.Errorf("bootstrap failed after retries: %w", err)
	}

	// Save certificates
	if err := manager.saveCertificates(resp); err != nil {
		return nil, fmt.Errorf("failed to save certificates: %w", err)
	}

	// Mark config as bootstrapped and populate runtime fields
	cfg.MarkBootstrapped(resp.AgentID, resp.APIBaseURL)

	logger.Printf("Bootstrap successful. Agent ID: %s, API: %s", resp.AgentID, resp.APIBaseURL)
	return cfg, nil
}

// Bootstrap performs the initial agent registration
// Deprecated: Use RunBootstrap instead
func (m *Manager) Bootstrap(ctx context.Context) error {
	if m.cfg.InstallToken == "" {
		return fmt.Errorf("install_token is required for bootstrap")
	}

	hostname := getHostname()

	req := BootstrapRequest{
		OrgID:        m.cfg.OrgID,
		InstallToken: m.cfg.InstallToken,
		Hostname:     hostname,
		OS:           getOS(),
		Arch:         getArch(),
		AgentVersion: "1.0.0",
	}

	m.logger.Printf("Bootstrapping agent for org %s...", m.cfg.OrgID)

	// Call bootstrap API with retry
	var resp *BootstrapResponse
	retryConfig := DefaultRetryConfig()
	err := RetryWithBackoff(ctx, retryConfig, func() error {
		var retryErr error
		resp, retryErr = m.callBootstrapAPI(ctx, req)
		return retryErr
	})
	if err != nil {
		return fmt.Errorf("bootstrap API call failed: %w", err)
	}

	// Save agent ID to config
	m.agentID = resp.AgentID
	m.cfg.MarkBootstrapped(resp.AgentID, resp.APIBaseURL)

	// Save certificates
	if err := m.saveCertificates(resp); err != nil {
		return fmt.Errorf("failed to save certificates: %w", err)
	}

	// Update config file
	configPath := getConfigPath()
	if err := m.cfg.Save(configPath); err != nil {
		m.logger.Printf("Warning: failed to save config: %v", err)
	}

	m.logger.Printf("Bootstrap successful. Agent ID: %s", m.agentID)
	return nil
}

// callBootstrapAPI sends the bootstrap request to the server
func (m *Manager) callBootstrapAPI(ctx context.Context, req BootstrapRequest) (*BootstrapResponse, error) {
	// Use hardcoded bootstrap endpoint (not from config, as config might not have it yet)
	bootstrapURL := "https://api.unitechio.space/api/v1/agents/bootstrap"
	if envURL := os.Getenv("BOOTSTRAP_URL"); envURL != "" {
		bootstrapURL = envURL
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use standard HTTPS (no mTLS yet, as we don't have certs)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	m.logger.Printf("Calling bootstrap API: %s", bootstrapURL)

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("bootstrap failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp BootstrapResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If server doesn't provide APIBaseURL, use the bootstrap URL's base
	if resp.APIBaseURL == "" {
		resp.APIBaseURL = "https://api.unitechio.space"
	}

	return &resp, nil
}

// saveCertificates writes certificates to disk with proper permissions
func (m *Manager) saveCertificates(resp *BootstrapResponse) error {
	// Save certificate
	if err := os.WriteFile(m.certPath, []byte(resp.Certificate), 0600); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Save private key (most sensitive, 0600 permissions)
	if err := os.WriteFile(m.keyPath, []byte(resp.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Save CA certificate
	if err := os.WriteFile(m.caPath, []byte(resp.CACert), 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	m.logger.Println("Certificates saved successfully")
	return nil
}

// VerifyIdentity checks if the stored certificates are valid
func (m *Manager) VerifyIdentity() error {
	// Load certificate
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	// Load private key
	keyPEM, err := os.ReadFile(m.keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse certificate and key
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse certificate/key: %w", err)
	}

	// Parse X.509 certificate to extract agent ID
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	// Extract agent ID from certificate Common Name
	m.agentID = x509Cert.Subject.CommonName
	m.logger.Printf("Identity verified: Agent ID = %s", m.agentID)

	return nil
}

// GetAgentID returns the agent's unique identifier
func (m *Manager) GetAgentID() string {
	return m.agentID
}

// GetTLSConfig returns a TLS configuration for mTLS
func (m *Manager) GetTLSConfig() (*tls.Config, error) {
	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(m.certPath, m.keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate/key: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(m.caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// GetHTTPClient returns an HTTP client configured for mTLS
func (m *Manager) GetHTTPClient() (*http.Client, error) {
	tlsConfig, err := m.GetTLSConfig()
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 30 * 1000000000, // 30 seconds
	}, nil
}

// Helper functions
func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

func getOS() string {
	return os.Getenv("GOOS")
}

func getArch() string {
	return os.Getenv("GOARCH")
}

func getConfigPath() string {
	if path := os.Getenv("AGENT_CONFIG"); path != "" {
		return path
	}
	if isWindows() {
		return `C:\ProgramData\unitechio\Agent\config.json`
	}
	return "/etc/your-agent/config.json"
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
