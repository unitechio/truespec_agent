package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unitechio/agent/internal/logging"
)

func main() {
	fmt.Println("=== Logging System Test ===")
	fmt.Println()

	// Create test directory
	testDir := filepath.Join(".", "test_logs")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create test directory: %v\n", err)
		return
	}

	// Run all tests
	passed := 0
	failed := 0

	type testCase struct {
		name string
		fn   func(string) error
	}

	testCases := []testCase{
		{"Basic Application Logging", testBasicLogging},
		{"Log Rotation", testLogRotation},
		{"Audit Logging (JSON)", testAuditLogging},
		{"Log Level Filtering", testLogLevels},
	}

	for i, test := range testCases {
		fmt.Printf("Test %d: %s\n", i+1, test.name)
		fmt.Println(strings.Repeat("-", 60))

		if err := test.fn(testDir); err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			failed++
		} else {
			fmt.Println("✅ PASSED")
			passed++
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Results: %d/%d tests passed\n", passed, len(testCases))
	fmt.Println()
	fmt.Println("📁 Log files location: ./test_logs/")
	fmt.Println()
	fmt.Println("You can inspect the following files:")
	fmt.Println("  - test_logs/basic.log       - Basic logging output")
	fmt.Println("  - test_logs/rotation.log*   - Rotated log files")
	fmt.Println("  - test_logs/audit.log       - JSON audit logs")
	fmt.Println("  - test_logs/levels.log      - Log level filtering")
}

func testBasicLogging(testDir string) error {
	logPath := filepath.Join(testDir, "basic.log")

	// Remove old log file
	os.Remove(logPath)

	// Create logger
	logger, err := logging.NewLogger(logging.Config{
		LogPath:    logPath,
		Level:      "debug",
		MaxSizeMB:  1,
		MaxBackups: 3,
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Write different log levels
	logger.Debug("This is a debug message with data: %v", map[string]int{"count": 42})
	logger.Info("Agent started successfully")
	logger.Warning("Retry attempt %d/%d", 1, 3)
	logger.Error("Failed to connect: %v", "connection timeout")

	// Close logger to flush
	if err := logger.Close(); err != nil {
		return fmt.Errorf("failed to close logger: %w", err)
	}

	// Wait a bit for file system
	time.Sleep(100 * time.Millisecond)

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		return fmt.Errorf("log file is empty")
	}

	// Verify all log levels are present
	checks := []struct {
		level string
		desc  string
	}{
		{"[DEBUG]", "DEBUG messages"},
		{"[INFO]", "INFO messages"},
		{"[WARN]", "WARNING messages"},
		{"[ERROR]", "ERROR messages"},
	}

	for _, check := range checks {
		if strings.Contains(contentStr, check.level) {
			fmt.Printf("  ✓ %s written\n", check.desc)
		} else {
			return fmt.Errorf("%s not found in log", check.desc)
		}
	}

	fmt.Printf("  ✓ Log file size: %d bytes\n", len(content))
	return nil
}

func testLogRotation(testDir string) error {
	logPath := filepath.Join(testDir, "rotation.log")

	// Remove old log files
	os.Remove(logPath)
	for i := 1; i <= 5; i++ {
		os.Remove(fmt.Sprintf("%s.%d", logPath, i))
	}

	// Create logger with very small max size to trigger rotation
	logger, err := logging.NewLogger(logging.Config{
		LogPath:    logPath,
		Level:      "info",
		MaxSizeMB:  0, // 0 MB will trigger rotation frequently
		MaxBackups: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Write many logs to trigger rotation
	fmt.Println("  Writing 500 log entries to trigger rotation...")
	for i := 0; i < 500; i++ {
		logger.Info("Log entry #%04d - Testing log rotation with a reasonably long message to fill up the file quickly", i)
	}

	logger.Close()
	time.Sleep(200 * time.Millisecond)

	// Check for rotated files
	rotatedCount := 0
	for i := 1; i <= 5; i++ {
		rotatedPath := fmt.Sprintf("%s.%d", logPath, i)
		if info, err := os.Stat(rotatedPath); err == nil {
			rotatedCount++
			fmt.Printf("  ✓ Found %s (%d bytes)\n", filepath.Base(rotatedPath), info.Size())
		}
	}

	// Check current log file
	if info, err := os.Stat(logPath); err == nil {
		fmt.Printf("  ✓ Current log: %s (%d bytes)\n", filepath.Base(logPath), info.Size())
	}

	if rotatedCount == 0 {
		return fmt.Errorf("no rotated files found - rotation may not be working")
	}

	fmt.Printf("  ✓ Log rotation working! Found %d rotated files\n", rotatedCount)
	return nil
}

func testAuditLogging(testDir string) error {
	auditPath := filepath.Join(testDir, "audit.log")
	agentID := "test-agent-550e8400-e29b-41d4-a716-446655440000"

	// Remove old audit log
	os.Remove(auditPath)

	// Create audit logger
	auditLogger, err := logging.NewAuditLogger(auditPath, agentID)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Log various audit events
	events := []struct {
		name string
		fn   func()
	}{
		{"Service Start", func() { auditLogger.LogServiceStart("1.0.0") }},
		{"Bootstrap", func() { auditLogger.LogBootstrap(true, "test-org-123", nil) }},
		{"Policy Change", func() { auditLogger.LogPolicyChange("1.0.0", "1.1.0") }},
		{"Agent Update", func() { auditLogger.LogUpdate("1.0.0", "1.0.1", true, nil) }},
		{"Auth Failure", func() { auditLogger.LogAuthFailure("https://api.example.com", "invalid token") }},
		{"Cert Rotation", func() { auditLogger.LogCertRotation(true, time.Now().Add(90*24*time.Hour), nil) }},
		{"Service Stop", func() { auditLogger.LogServiceStop("test completed") }},
	}

	for _, event := range events {
		event.fn()
		time.Sleep(10 * time.Millisecond)
	}

	auditLogger.Close()
	time.Sleep(100 * time.Millisecond)

	// Check if file exists
	if _, err := os.Stat(auditPath); os.IsNotExist(err) {
		return fmt.Errorf("audit log file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(auditPath)
	if err != nil {
		return fmt.Errorf("failed to read audit log: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("audit log file is empty")
	}

	// Parse and verify JSON format
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	validJSON := 0

	for i, line := range lines {
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("line %d is not valid JSON: %w", i+1, err)
		}
		validJSON++

		// Verify required fields
		requiredFields := []string{"timestamp", "event_type", "severity", "action", "result"}
		for _, field := range requiredFields {
			if _, ok := event[field]; !ok {
				return fmt.Errorf("line %d missing required field: %s", i+1, field)
			}
		}
	}

	fmt.Printf("  ✓ Created %d audit events\n", validJSON)
	fmt.Println("  ✓ All events are valid JSON")
	fmt.Println("  ✓ All required fields present")
	fmt.Printf("  ✓ File size: %d bytes\n", len(content))

	return nil
}

func testLogLevels(testDir string) error {
	logPath := filepath.Join(testDir, "levels.log")

	// Remove old log file
	os.Remove(logPath)

	// Test with INFO level (should filter out DEBUG)
	logger, err := logging.NewLogger(logging.Config{
		LogPath:    logPath,
		Level:      "info",
		MaxSizeMB:  1,
		MaxBackups: 3,
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Debug("This DEBUG message should NOT appear")
	logger.Info("This INFO message should appear")
	logger.Warning("This WARNING message should appear")
	logger.Error("This ERROR message should appear")

	logger.Close()
	time.Sleep(100 * time.Millisecond)

	// Read content
	content, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		return fmt.Errorf("log file is empty")
	}

	// Verify filtering
	hasDebug := strings.Contains(contentStr, "[DEBUG]")
	hasInfo := strings.Contains(contentStr, "[INFO]")
	hasWarning := strings.Contains(contentStr, "[WARN]")
	hasError := strings.Contains(contentStr, "[ERROR]")

	if hasDebug {
		return fmt.Errorf("DEBUG messages should be filtered out at INFO level")
	}
	fmt.Println("  ✓ DEBUG messages correctly filtered out")

	if !hasInfo {
		return fmt.Errorf("INFO messages should be present")
	}
	fmt.Println("  ✓ INFO messages present")

	if !hasWarning {
		return fmt.Errorf("WARNING messages should be present")
	}
	fmt.Println("  ✓ WARNING messages present")

	if !hasError {
		return fmt.Errorf("ERROR messages should be present")
	}
	fmt.Println("  ✓ ERROR messages present")

	return nil
}
