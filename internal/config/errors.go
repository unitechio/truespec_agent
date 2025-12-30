package config

import "errors"

// Sentinel errors for configuration and bootstrap states
var (
	// ErrConfigNotFound is returned when the config file doesn't exist
	ErrConfigNotFound = errors.New("config file not found")

	// ErrNotBootstrapped is returned when agent is not bootstrapped
	ErrNotBootstrapped = errors.New("agent not bootstrapped")

	// ErrInvalidBootstrap is returned when bootstrap configuration is invalid
	ErrInvalidBootstrap = errors.New("invalid bootstrap configuration")

	// ErrInvalidRuntime is returned when runtime configuration is invalid
	ErrInvalidRuntime = errors.New("invalid runtime configuration")

	// ErrCertExpired is returned when agent certificate has expired
	ErrCertExpired = errors.New("certificate expired")

	// ErrBootstrapInProgress is returned when bootstrap is already running
	ErrBootstrapInProgress = errors.New("bootstrap already in progress")
)
