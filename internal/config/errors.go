package config

import "errors"

var (
	ErrConfigNotFound      = errors.New("config file not found")
	ErrNotBootstrapped     = errors.New("agent not bootstrapped")
	ErrInvalidBootstrap    = errors.New("invalid bootstrap configuration")
	ErrInvalidRuntime      = errors.New("invalid runtime configuration")
	ErrCertExpired         = errors.New("certificate expired")
	ErrBootstrapInProgress = errors.New("bootstrap already in progress")
)
