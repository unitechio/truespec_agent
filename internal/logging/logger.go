package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARN",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
}

// Logger provides structured logging with rotation
type Logger struct {
	mu          sync.Mutex
	file        *os.File
	logger      *log.Logger
	level       LogLevel
	logPath     string
	maxSize     int64 // bytes
	maxBackups  int
	currentSize int64
}

// Config holds logger configuration
type Config struct {
	LogPath    string
	Level      string
	MaxSizeMB  int
	MaxBackups int
}

// NewLogger creates a new logger with rotation
func NewLogger(cfg Config) (*Logger, error) {
	level := parseLevel(cfg.Level)

	// Create log directory
	logDir := filepath.Dir(cfg.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	// Create multi-writer (file + stdout)
	multiWriter := io.MultiWriter(os.Stdout, file)

	logger := &Logger{
		file:        file,
		logger:      log.New(multiWriter, "", 0), // We'll add our own prefix
		level:       level,
		logPath:     cfg.LogPath,
		maxSize:     int64(cfg.MaxSizeMB) * 1024 * 1024,
		maxBackups:  cfg.MaxBackups,
		currentSize: info.Size(),
	}

	return logger, nil
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	l.log(WARNING, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FATAL, format, v...)
	os.Exit(1)
}

// log writes a log message with the given level
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format: 2024-01-15 10:30:00 [INFO] message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelNames[level]
	message := fmt.Sprintf(format, v...)

	logLine := fmt.Sprintf("%s [%s] %s\n", timestamp, levelStr, message)

	// Write to log
	l.logger.Print(logLine)

	// Update size and rotate if needed
	l.currentSize += int64(len(logLine))
	if l.currentSize >= l.maxSize {
		l.rotate()
	}
}

// rotate rotates the log file
func (l *Logger) rotate() {
	// Close current file
	l.file.Close()

	// Rotate existing backups
	for i := l.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)
		os.Rename(oldPath, newPath)
	}

	// Move current log to .1
	backupPath := fmt.Sprintf("%s.1", l.logPath)
	os.Rename(l.logPath, backupPath)

	// Create new log file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		return
	}

	l.file = file
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.logger.SetOutput(multiWriter)
	l.currentSize = 0

	l.Info("Log file rotated")
}

// Close closes the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// parseLevel converts string to LogLevel
func parseLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warning", "warn":
		return WARNING
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = parseLevel(level)
}
