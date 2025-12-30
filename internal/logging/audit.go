package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent represents a security-relevant event
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Severity  string                 `json:"severity"`
	AgentID   string                 `json:"agent_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource,omitempty"`
	Result    string                 `json:"result"` // success, failure
	Details   map[string]interface{} `json:"details,omitempty"`
	SourceIP  string                 `json:"source_ip,omitempty"`
	ErrorMsg  string                 `json:"error_msg,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	agentID  string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string, agentID string) (*AuditLogger, error) {
	// Create audit log directory
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Open audit log file (append mode)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	return &AuditLogger{
		file:     file,
		filePath: logPath,
		agentID:  agentID,
	}, nil
}

// LogEvent logs an audit event
func (a *AuditLogger) LogEvent(event AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Set timestamp and agent ID
	event.Timestamp = time.Now()
	if event.AgentID == "" {
		event.AgentID = a.agentID
	}

	// Serialize to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Write to file
	_, err = a.file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Sync to disk (important for audit logs)
	return a.file.Sync()
}

// LogBootstrap logs agent bootstrap event
func (a *AuditLogger) LogBootstrap(success bool, orgID string, err error) {
	event := AuditEvent{
		EventType: "bootstrap",
		Severity:  "INFO",
		Action:    "agent_registration",
		Resource:  orgID,
		Result:    "success",
	}

	if !success {
		event.Result = "failure"
		event.Severity = "ERROR"
		if err != nil {
			event.ErrorMsg = err.Error()
		}
	}

	a.LogEvent(event)
}

// LogPolicyChange logs policy update event
func (a *AuditLogger) LogPolicyChange(oldVersion, newVersion string) {
	a.LogEvent(AuditEvent{
		EventType: "policy_change",
		Severity:  "INFO",
		Action:    "policy_update",
		Result:    "success",
		Details: map[string]interface{}{
			"old_version": oldVersion,
			"new_version": newVersion,
		},
	})
}

// LogUpdate logs agent update event
func (a *AuditLogger) LogUpdate(oldVersion, newVersion string, success bool, err error) {
	event := AuditEvent{
		EventType: "agent_update",
		Severity:  "INFO",
		Action:    "binary_update",
		Result:    "success",
		Details: map[string]interface{}{
			"old_version": oldVersion,
			"new_version": newVersion,
		},
	}

	if !success {
		event.Result = "failure"
		event.Severity = "ERROR"
		if err != nil {
			event.ErrorMsg = err.Error()
		}
	}

	a.LogEvent(event)
}

// LogAuthFailure logs authentication failure
func (a *AuditLogger) LogAuthFailure(endpoint string, reason string) {
	a.LogEvent(AuditEvent{
		EventType: "auth_failure",
		Severity:  "WARNING",
		Action:    "authentication",
		Resource:  endpoint,
		Result:    "failure",
		ErrorMsg:  reason,
	})
}

// LogCertRotation logs certificate rotation event
func (a *AuditLogger) LogCertRotation(success bool, expiryDate time.Time, err error) {
	event := AuditEvent{
		EventType: "cert_rotation",
		Severity:  "INFO",
		Action:    "certificate_renewal",
		Result:    "success",
		Details: map[string]interface{}{
			"new_expiry": expiryDate.Format(time.RFC3339),
		},
	}

	if !success {
		event.Result = "failure"
		event.Severity = "ERROR"
		if err != nil {
			event.ErrorMsg = err.Error()
		}
	}

	a.LogEvent(event)
}

// LogServiceStart logs service start event
func (a *AuditLogger) LogServiceStart(version string) {
	a.LogEvent(AuditEvent{
		EventType: "service_lifecycle",
		Severity:  "INFO",
		Action:    "service_start",
		Result:    "success",
		Details: map[string]interface{}{
			"version": version,
		},
	})
}

// LogServiceStop logs service stop event
func (a *AuditLogger) LogServiceStop(reason string) {
	a.LogEvent(AuditEvent{
		EventType: "service_lifecycle",
		Severity:  "INFO",
		Action:    "service_stop",
		Result:    "success",
		Details: map[string]interface{}{
			"reason": reason,
		},
	})
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		return a.file.Close()
	}
	return nil
}
