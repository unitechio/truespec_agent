package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/unitechio/agent/internal/buffer"
	"github.com/unitechio/agent/internal/collectors"
	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/logging"
)

// TestEndToEndDataCollection tests the full data collection pipeline
func TestEndToEndDataCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Create configuration
	cfg := &config.Config{
		OrgID:              "test-org",
		APIBaseURL:         "https://api.test.com",
		CollectionInterval: 2 * time.Second,
		BatchSize:          10,
		MaxBufferSize:      1024 * 1024,
		BufferDir:          filepath.Join(tmpDir, "buffer"),
		HeartbeatInterval:  5 * time.Second,
		LogLevel:           "debug",
		LogFile:            filepath.Join(tmpDir, "agent.log"),
	}

	// Create standard logger for buffer
	stdLogger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create custom logger for test
	customLogger, err := logging.NewLogger(logging.Config{
		LogPath:    cfg.LogFile,
		Level:      cfg.LogLevel,
		MaxSizeMB:  10,
		MaxBackups: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer customLogger.Close()

	customLogger.Info("Starting end-to-end integration test")

	// Create buffer with standard logger
	buf, err := buffer.New(cfg.BufferDir, cfg.MaxBufferSize, stdLogger)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Test collectors
	collectorList := collectors.NewDefaultCollectors()
	if len(collectorList) == 0 {
		t.Fatal("No collectors available")
	}

	customLogger.Info("Testing %d collectors", len(collectorList))

	// Collect data from each collector
	ctx := context.Background()
	for _, collector := range collectorList {
		customLogger.Debug("Running collector: %s", collector.Name())

		data, err := collector.Collect(ctx)
		if err != nil {
			t.Errorf("Collector %s failed: %v", collector.Name(), err)
			continue
		}

		if data == nil {
			t.Errorf("Collector %s returned nil data", collector.Name())
			continue
		}

		customLogger.Info("Collector %s succeeded", collector.Name())

		// Test buffering
		if err := buf.Write(data); err != nil {
			t.Errorf("Failed to buffer data from %s: %v", collector.Name(), err)
		}
	}

	// Verify buffered data
	batches, err := buf.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read buffer: %v", err)
	}

	if len(batches) != len(collectorList) {
		t.Errorf("Expected %d batches, got %d", len(collectorList), len(batches))
	}

	customLogger.Info("Integration test completed successfully")
	t.Logf("Successfully collected and buffered data from %d collectors", len(collectorList))
}

// TestLogRotation tests log file rotation
func TestLogRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with small max size
	logger, err := logging.NewLogger(logging.Config{
		LogPath:    logPath,
		Level:      "debug",
		MaxSizeMB:  1, // 1 MB
		MaxBackups: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write enough logs to trigger rotation
	for i := 0; i < 10000; i++ {
		logger.Info("Test log message number %d with some additional text to increase size", i)
	}

	// Check if rotation occurred
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	logFileCount := 0
	for _, file := range files {
		if !file.IsDir() {
			logFileCount++
		}
	}

	// Should have at least 2 files (current + rotated)
	if logFileCount < 2 {
		t.Errorf("Expected at least 2 log files, got %d", logFileCount)
	}

	t.Logf("Log rotation test passed: %d log files created", logFileCount)
}

// TestAuditLogging tests audit log functionality
func TestAuditLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.log")

	auditLogger, err := logging.NewAuditLogger(auditPath, "test-agent-123")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	// Log various audit events
	auditLogger.LogBootstrap(true, "test-org", nil)
	auditLogger.LogServiceStart("1.0.0")
	auditLogger.LogPolicyChange("1.0.0", "1.1.0")
	auditLogger.LogUpdate("1.0.0", "1.1.0", true, nil)
	auditLogger.LogAuthFailure("/api/test", "invalid token")
	auditLogger.LogServiceStop("test completed")

	// Read audit log
	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("Audit log is empty")
	}

	// Verify all events are logged
	contentStr := string(content)
	expectedEvents := []string{
		"bootstrap",
		"service_lifecycle",
		"policy_change",
		"agent_update",
		"auth_failure",
	}

	for _, event := range expectedEvents {
		if !contains(contentStr, event) {
			t.Errorf("Audit log missing event: %s", event)
		}
	}

	t.Logf("Audit logging test passed: all %d events logged", len(expectedEvents))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
