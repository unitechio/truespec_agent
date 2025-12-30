package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"
)

// BootstrapRequest matches the agent's request structure
type BootstrapRequest struct {
	OrgID        string `json:"org_id"`
	InstallToken string `json:"install_token"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	AgentVersion string `json:"agent_version"`
}

// BootstrapResponse matches the agent's expected response
type BootstrapResponse struct {
	AgentID     string `json:"agent_id"`
	APIBaseURL  string `json:"api_base_url"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
	CACert      string `json:"ca_cert"`
	Policy      string `json:"policy"`
	ExpiresAt   string `json:"expires_at"`
}

func main() {
	http.HandleFunc("/api/v1/agents/bootstrap", handleBootstrap)
	http.HandleFunc("/health", handleHealth)

	addr := ":8080"
	log.Printf("🚀 Mock API Server starting on %s", addr)
	log.Printf("📡 Bootstrap endpoint: http://localhost%s/api/v1/agents/bootstrap", addr)
	log.Printf("💚 Health endpoint: http://localhost%s/health", addr)
	log.Println()
	log.Println("Valid credentials:")
	log.Println("  ORG_ID: test-org")
	log.Println("  INSTALL_TOKEN: test-token-123")
	log.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleBootstrap(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Received bootstrap request from %s", r.RemoteAddr)

	// Only accept POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req BootstrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Failed to parse request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("   OrgID: %s", req.OrgID)
	log.Printf("   InstallToken: %s", req.InstallToken)
	log.Printf("   Hostname: %s", req.Hostname)
	log.Printf("   OS: %s", req.OS)
	log.Printf("   Arch: %s", req.Arch)

	// Validate credentials
	if req.OrgID != "test-org" {
		log.Printf("❌ Invalid OrgID: %s", req.OrgID)
		http.Error(w, "Invalid organization ID", http.StatusUnauthorized)
		return
	}

	if req.InstallToken != "test-token-123" {
		log.Printf("❌ Invalid InstallToken: %s", req.InstallToken)
		http.Error(w, "Invalid install token", http.StatusUnauthorized)
		return
	}

	// Generate agent ID
	agentID := fmt.Sprintf("agent-%d", time.Now().Unix())

	// Generate mock certificates
	cert, key, caCert, err := generateMockCertificates(agentID)
	if err != nil {
		log.Printf("❌ Failed to generate certificates: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create response
	expiresAt := time.Now().Add(365 * 24 * time.Hour) // 1 year
	resp := BootstrapResponse{
		AgentID:     agentID,
		APIBaseURL:  "http://localhost:8080", // Point back to this server
		Certificate: cert,
		PrivateKey:  key,
		CACert:      caCert,
		Policy:      `{"version": "1.0", "rules": []}`,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}

	log.Printf("✅ Bootstrap successful!")
	log.Printf("   AgentID: %s", agentID)
	log.Printf("   Expires: %s", expiresAt.Format(time.RFC3339))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// generateMockCertificates creates self-signed certificates for testing
func generateMockCertificates(agentID string) (certPEM, keyPEM, caCertPEM string, err error) {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	// Create CA certificate
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Univertech Mock CA"},
			CommonName:   "Univertech Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	// Generate agent private key
	agentKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	// Create agent certificate
	agentTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Univertech Agent"},
			CommonName:   agentID,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	agentCertBytes, err := x509.CreateCertificate(rand.Reader, &agentTemplate, &caTemplate, &agentKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	// Encode to PEM
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: agentCertBytes}))
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(agentKey)}))
	caCertPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes}))

	return certPEM, keyPEM, caCertPEM, nil
}
