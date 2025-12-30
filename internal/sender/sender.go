package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/unitechio/agent/internal/buffer"
	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/identity"
)

// Sender handles secure transmission of telemetry data
type Sender struct {
	cfg      *config.Config
	identity *identity.Manager
	logger   *log.Logger
	buffer   *buffer.Buffer
	client   *http.Client
	mu       sync.Mutex
	queue    []TelemetryBatch
}

// TelemetryBatch represents a batch of telemetry data
type TelemetryBatch struct {
	AgentID   string                   `json:"agent_id"`
	Timestamp time.Time                `json:"timestamp"`
	Data      []map[string]interface{} `json:"data"`
}

// NewSender creates a new sender
func NewSender(cfg *config.Config, identityMgr *identity.Manager, logger *log.Logger) (*Sender, error) {
	// Create buffer for offline storage
	buf, err := buffer.New(cfg.BufferDir, cfg.MaxBufferSize, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create buffer: %w", err)
	}

	// Get mTLS HTTP client
	client, err := identityMgr.GetHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Sender{
		cfg:      cfg,
		identity: identityMgr,
		logger:   logger,
		buffer:   buf,
		client:   client,
		queue:    make([]TelemetryBatch, 0),
	}, nil
}

// Send queues data for transmission
func (s *Sender) Send(ctx context.Context, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to current batch
	batch := TelemetryBatch{
		AgentID:   s.identity.GetAgentID(),
		Timestamp: time.Now(),
		Data:      []map[string]interface{}{data},
	}

	s.queue = append(s.queue, batch)

	// Flush if batch size reached
	if len(s.queue) >= s.cfg.BatchSize {
		return s.flush(ctx)
	}

	return nil
}

// Flush sends all queued data
func (s *Sender) flush(ctx context.Context) error {
	if len(s.queue) == 0 {
		return nil
	}

	s.logger.Printf("Flushing %d batches to server...", len(s.queue))

	// Combine all batches
	combined := TelemetryBatch{
		AgentID:   s.identity.GetAgentID(),
		Timestamp: time.Now(),
		Data:      make([]map[string]interface{}, 0),
	}

	for _, batch := range s.queue {
		combined.Data = append(combined.Data, batch.Data...)
	}

	// Try to send with retry
	err := s.sendWithRetry(ctx, combined)
	if err != nil {
		// If send fails, save to buffer
		s.logger.Printf("Failed to send telemetry, buffering: %v", err)
		if bufErr := s.buffer.Write(combined); bufErr != nil {
			s.logger.Printf("Failed to buffer data: %v", bufErr)
			return fmt.Errorf("send failed and buffer failed: %w", bufErr)
		}
		return err
	}

	// Clear queue on success
	s.queue = make([]TelemetryBatch, 0)
	s.logger.Printf("Successfully sent %d data points", len(combined.Data))

	// Try to flush buffered data
	s.flushBuffer(ctx)

	return nil
}

// sendWithRetry sends data with exponential backoff retry
func (s *Sender) sendWithRetry(ctx context.Context, batch TelemetryBatch) error {
	maxRetries := 5
	baseDelay := 1 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, 16s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			s.logger.Printf("Retry attempt %d/%d after %v", attempt+1, maxRetries, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := s.sendBatch(ctx, batch)
		if err == nil {
			return nil
		}

		lastErr = err
		s.logger.Printf("Send attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// sendBatch sends a single batch to the server
func (s *Sender) sendBatch(ctx context.Context, batch TelemetryBatch) error {
	url := s.cfg.APIBaseURL + "/api/v1/telemetry"

	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// flushBuffer attempts to send buffered data
func (s *Sender) flushBuffer(ctx context.Context) {
	batches, err := s.buffer.ReadAll()
	if err != nil {
		s.logger.Printf("Failed to read buffer: %v", err)
		return
	}

	if len(batches) == 0 {
		return
	}

	s.logger.Printf("Attempting to flush %d buffered batches", len(batches))

	for _, data := range batches {
		var batch TelemetryBatch
		if err := json.Unmarshal(data, &batch); err != nil {
			s.logger.Printf("Failed to unmarshal buffered batch: %v", err)
			continue
		}

		if err := s.sendBatch(ctx, batch); err != nil {
			s.logger.Printf("Failed to send buffered batch: %v", err)
			// Stop trying if one fails (network still down)
			return
		}
	}

	// Clear buffer on success
	if err := s.buffer.Clear(); err != nil {
		s.logger.Printf("Failed to clear buffer: %v", err)
	} else {
		s.logger.Println("Successfully flushed all buffered data")
	}
}

// Start begins periodic flushing
func (s *Sender) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				s.flush(ctx)
				s.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}
