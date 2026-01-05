package buffer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Buffer provides persistent storage for telemetry when offline
type Buffer struct {
	dir         string
	maxSize     int64
	logger      *log.Logger
	mu          sync.Mutex
	currentSize int64
}

// New creates a new buffer
func New(dir string, maxSize int64, logger *log.Logger) (*Buffer, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create buffer directory: %w", err)
	}

	b := &Buffer{
		dir:     dir,
		maxSize: maxSize,
		logger:  logger,
	}

	// Calculate current size
	if err := b.calculateSize(); err != nil {
		logger.Printf("Warning: failed to calculate buffer size: %v", err)
	}

	return b, nil
}

// Write adds data to the buffer
func (b *Buffer) Write(data interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Serialize data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Check size limit
	if b.currentSize+int64(len(jsonData)) > b.maxSize {
		return fmt.Errorf("buffer full (current: %d, max: %d)", b.currentSize, b.maxSize)
	}

	// Write to file with timestamp
	filename := fmt.Sprintf("batch_%d.json", time.Now().UnixNano())
	path := filepath.Join(b.dir, filename)

	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write buffer file: %w", err)
	}

	b.currentSize += int64(len(jsonData))
	b.logger.Printf("Buffered %d bytes (total: %d/%d)", len(jsonData), b.currentSize, b.maxSize)

	return nil
}

// ReadAll reads all buffered data
func (b *Buffer) ReadAll() ([][]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	files, err := os.ReadDir(b.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read buffer directory: %w", err)
	}

	var batches [][]byte
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(b.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			b.logger.Printf("Failed to read buffer file %s: %v", file.Name(), err)
			continue
		}

		batches = append(batches, data)
	}

	return batches, nil
}

// Clear removes all buffered data
func (b *Buffer) Clear() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	files, err := os.ReadDir(b.dir)
	if err != nil {
		return fmt.Errorf("failed to read buffer directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(b.dir, file.Name())
		if err := os.Remove(path); err != nil {
			b.logger.Printf("Failed to remove buffer file %s: %v", file.Name(), err)
		}
	}

	b.currentSize = 0
	b.logger.Println("Buffer cleared")

	return nil
}

// calculateSize computes the total size of buffered data
func (b *Buffer) calculateSize() error {
	var total int64

	err := filepath.Walk(b.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})

	if err != nil {
		return err
	}

	b.currentSize = total
	return nil
}

// Size returns the current buffer size
func (b *Buffer) Size() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.currentSize
}

// Prune removes old entries if buffer is too large
func (b *Buffer) Prune() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.currentSize <= b.maxSize {
		return nil
	}

	b.logger.Printf("Buffer size (%d) exceeds limit (%d), pruning...", b.currentSize, b.maxSize)

	// Get all files sorted by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
		size    int64
	}

	var files []fileInfo
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		return fmt.Errorf("failed to read buffer directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(b.dir, entry.Name())
		files = append(files, fileInfo{
			path:    path,
			modTime: info.ModTime(),
			size:    info.Size(),
		})
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files until under limit
	for _, file := range files {
		if b.currentSize <= b.maxSize {
			break
		}

		if err := os.Remove(file.path); err != nil {
			b.logger.Printf("Failed to remove file %s: %v", file.path, err)
			continue
		}

		b.currentSize -= file.size
		b.logger.Printf("Pruned %s (%d bytes)", filepath.Base(file.path), file.size)
	}

	return nil
}
