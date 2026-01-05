package buffer

import (
	"testing"
)

func TestBufferWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()

	logger := &testLogger{}
	buffer, err := New(tmpDir, 1024*1024, logger) // 1 MB limit
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Write test data
	testData := map[string]interface{}{
		"test":  "data",
		"value": 123,
	}

	if err := buffer.Write(testData); err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	// Read back
	batches, err := buffer.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read from buffer: %v", err)
	}

	if len(batches) != 1 {
		t.Errorf("Expected 1 batch, got %d", len(batches))
	}
}

func TestBufferSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()

	logger := &testLogger{}
	buffer, err := New(tmpDir, 100, logger) // Very small limit
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Write large data that exceeds limit
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeData[string(rune(i))] = "very long string to exceed buffer size"
	}

	err = buffer.Write(largeData)
	if err == nil {
		t.Error("Expected error when exceeding buffer size, got nil")
	}
}

func TestBufferClear(t *testing.T) {
	tmpDir := t.TempDir()

	logger := &testLogger{}
	buffer, err := New(tmpDir, 1024*1024, logger)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Write data
	testData := map[string]interface{}{"test": "data"}
	if err := buffer.Write(testData); err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	// Clear buffer
	if err := buffer.Clear(); err != nil {
		t.Fatalf("Failed to clear buffer: %v", err)
	}

	// Verify empty
	batches, err := buffer.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read from buffer: %v", err)
	}

	if len(batches) != 0 {
		t.Errorf("Expected 0 batches after clear, got %d", len(batches))
	}

	if buffer.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", buffer.Size())
	}
}

func TestBufferPrune(t *testing.T) {
	tmpDir := t.TempDir()

	logger := &testLogger{}
	buffer, err := New(tmpDir, 200, logger) // Small limit
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Write multiple small batches
	for i := 0; i < 5; i++ {
		testData := map[string]interface{}{
			"batch": i,
			"data":  "test",
		}
		buffer.Write(testData)
	}

	// Prune should remove oldest files
	if err := buffer.Prune(); err != nil {
		t.Fatalf("Failed to prune buffer: %v", err)
	}

	// Verify size is under limit
	if buffer.Size() > 200 {
		t.Errorf("Buffer size (%d) still exceeds limit after prune", buffer.Size())
	}
}

// testLogger is a simple logger for testing
type testLogger struct{}

func (l *testLogger) Printf(format string, v ...interface{}) {}
func (l *testLogger) Println(v ...interface{})               {}
