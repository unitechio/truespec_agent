package identity

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for exponential backoff retry logic
type RetryConfig struct {
	InitialDelay time.Duration // Initial delay before first retry
	MaxDelay     time.Duration // Maximum delay between retries
	MaxAttempts  int           // Maximum number of retry attempts
	Multiplier   float64       // Backoff multiplier (e.g., 2.0 for doubling)
	Jitter       bool          // Add random jitter to prevent thundering herd
}

// DefaultRetryConfig returns sensible defaults for bootstrap retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 5 * time.Second,
		MaxDelay:     5 * time.Minute,
		MaxAttempts:  10,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Try the operation
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(cfg, attempt)

		// Check if context is cancelled before sleeping
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// calculateDelay computes the delay for a given attempt with exponential backoff
func calculateDelay(cfg RetryConfig, attempt int) time.Duration {
	// Calculate exponential delay: InitialDelay * (Multiplier ^ attempt)
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	// Add jitter if enabled (±25% random variation)
	if cfg.Jitter {
		jitter := delay * 0.25 * (rand.Float64()*2 - 1) // Random value between -0.25 and +0.25
		delay += jitter
	}

	return time.Duration(delay)
}

// IsRetryable determines if an error is worth retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	// DNS errors are retryable (might be temporary)
	// 5xx server errors are retryable
	// 429 rate limit errors are retryable
	// 401/403 auth errors are NOT retryable
	// 400 bad request is NOT retryable

	// This is a simple implementation - in production you'd check specific error types
	return true
}
