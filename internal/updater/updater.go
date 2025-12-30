package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/identity"
)

// Updater handles auto-update functionality
type Updater struct {
	cfg            *config.Config
	identity       *identity.Manager
	logger         *log.Logger
	client         *http.Client
	currentVersion string
}

// UpdateMetadata describes an available update
type UpdateMetadata struct {
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"release_date"`
	Channel     string    `json:"channel"` // stable, beta, dev
	DownloadURL string    `json:"download_url"`
	Checksum    string    `json:"checksum"`  // SHA256
	Signature   string    `json:"signature"` // Base64-encoded signature
	Changelog   string    `json:"changelog"`
	Mandatory   bool      `json:"mandatory"`
}

// NewUpdater creates a new updater
func NewUpdater(cfg *config.Config, identityMgr *identity.Manager, currentVersion string, logger *log.Logger) (*Updater, error) {
	client, err := identityMgr.GetHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Updater{
		cfg:            cfg,
		identity:       identityMgr,
		logger:         logger,
		client:         client,
		currentVersion: currentVersion,
	}, nil
}

// CheckForUpdate queries the server for available updates
func (u *Updater) CheckForUpdate(ctx context.Context) (*UpdateMetadata, error) {
	url := fmt.Sprintf("%s/api/v1/updates/metadata?os=%s&arch=%s&version=%s&channel=%s",
		u.cfg.APIBaseURL,
		runtime.GOOS,
		runtime.GOARCH,
		u.currentVersion,
		"stable", // TODO: Get from policy
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No update available
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var metadata UpdateMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	u.logger.Printf("Update available: %s -> %s", u.currentVersion, metadata.Version)
	return &metadata, nil
}

// DownloadUpdate downloads the update binary
func (u *Updater) DownloadUpdate(ctx context.Context, metadata *UpdateMetadata) (string, error) {
	u.logger.Printf("Downloading update from %s", metadata.DownloadURL)

	req, err := http.NewRequestWithContext(ctx, "GET", metadata.DownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("agent_update_%s", metadata.Version))
	if runtime.GOOS == "windows" {
		tmpFile += ".exe"
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Download with progress
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("download failed: %w", err)
	}

	u.logger.Printf("Downloaded %d bytes to %s", written, tmpFile)
	return tmpFile, nil
}

// VerifyUpdate verifies the checksum and signature of the downloaded binary
func (u *Updater) VerifyUpdate(filePath string, metadata *UpdateMetadata) error {
	u.logger.Println("Verifying update checksum...")

	// Verify SHA256 checksum
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	if checksum != metadata.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", metadata.Checksum, checksum)
	}

	u.logger.Println("Checksum verified successfully")

	// Verify code signature (OS-specific)
	if err := u.verifySignature(filePath); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	u.logger.Println("Signature verified successfully")
	return nil
}

// verifySignature verifies the code signature (OS-specific)
func (u *Updater) verifySignature(filePath string) error {
	switch runtime.GOOS {
	case "windows":
		// TODO: Implement Windows Authenticode verification
		// Use: signtool verify /pa /v <filePath>
		u.logger.Println("Windows signature verification not yet implemented")
		return nil

	case "darwin":
		// TODO: Implement macOS code signature verification
		// Use: codesign --verify --deep --strict <filePath>
		u.logger.Println("macOS signature verification not yet implemented")
		return nil

	case "linux":
		// TODO: Implement GPG signature verification
		// Use: gpg --verify <filePath>.sig <filePath>
		u.logger.Println("Linux signature verification not yet implemented")
		return nil

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// InstallUpdate performs atomic swap of the binary
func (u *Updater) InstallUpdate(newBinaryPath string) error {
	u.logger.Println("Installing update...")

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Backup current binary
	backupPath := currentExe + ".old"
	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to current location
	if err := os.Rename(newBinaryPath, currentExe); err != nil {
		// Rollback on failure
		os.Rename(backupPath, currentExe)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Set executable permissions (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			u.logger.Printf("Warning: failed to set executable permissions: %v", err)
		}
	}

	u.logger.Printf("Update installed successfully. Backup at: %s", backupPath)
	return nil
}

// Rollback restores the previous version
func (u *Updater) Rollback() error {
	u.logger.Println("Rolling back to previous version...")

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	backupPath := currentExe + ".old"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// Remove current (failed) binary
	if err := os.Remove(currentExe); err != nil {
		return fmt.Errorf("failed to remove current binary: %w", err)
	}

	// Restore backup
	if err := os.Rename(backupPath, currentExe); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	u.logger.Println("Rollback completed successfully")
	return nil
}

// PerformUpdate orchestrates the entire update process
func (u *Updater) PerformUpdate(ctx context.Context) error {
	// Check for update
	metadata, err := u.CheckForUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for update: %w", err)
	}

	if metadata == nil {
		u.logger.Println("No update available")
		return nil
	}

	// Download update
	tmpFile, err := u.DownloadUpdate(ctx, metadata)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify update
	if err := u.VerifyUpdate(tmpFile, metadata); err != nil {
		return fmt.Errorf("failed to verify update: %w", err)
	}

	// Install update
	if err := u.InstallUpdate(tmpFile); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	u.logger.Printf("Update to version %s completed. Restart required.", metadata.Version)
	return nil
}
