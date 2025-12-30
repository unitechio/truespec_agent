package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		LogPath:    logPath,
		Level:      "debug",
		MaxSizeMB:  1,
		MaxBackups: 3,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test different log levels
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warning("This is a warning message")
	logger.Error("This is an error message")

	// Verify log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Verify log file has content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file is empty")
	}
}

func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		LogPath:    logPath,
		Level:      "warning",
		MaxSizeMB:  1,
		MaxBackups: 3,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Debug and Info should not be logged
	logger.Debug("This should not appear")
	logger.Info("This should not appear")

	// Warning and Error should be logged
	logger.Warning("This should appear")
	logger.Error("This should appear")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Log file is empty")
	}

	// Should not contain debug/info messages
	if contains(contentStr, "should not appear") {
		t.Error("Debug/Info messages were logged when level is WARNING")
	}

	// Should contain warning/error messages
	if !contains(contentStr, "should appear") {
		t.Error("Warning/Error messages were not logged")
	}
}

func TestAuditLogger(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.log")

	auditLogger, err := NewAuditLogger(auditPath, "test-agent-123")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	// Log various events
	auditLogger.LogBootstrap(true, "test-org", nil)
	auditLogger.LogPolicyChange("1.0.0", "1.1.0")
	auditLogger.LogServiceStart("1.0.0")

	// Verify audit log exists
	if _, err := os.Stat(auditPath); os.IsNotExist(err) {
		t.Error("Audit log file was not created")
	}

	// Verify audit log has content
	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(content) == 0 {
		t.Error("Audit log is empty")
	}

	contentStr := string(content)
	if !contains(contentStr, "bootstrap") {
		t.Error("Bootstrap event not found in audit log")
	}
	if !contains(contentStr, "policy_change") {
		t.Error("Policy change event not found in audit log")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
