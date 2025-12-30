package logging

import (
	"log"
)

// StdLogger wraps our custom Logger to implement standard log.Logger interface
type StdLogger struct {
	logger *Logger
}

// NewStdLogger creates a standard logger wrapper
func NewStdLogger(logger *Logger) *log.Logger {
	return log.New(&logWriter{logger: logger}, "", 0)
}

// logWriter implements io.Writer for standard log compatibility
type logWriter struct {
	logger *Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info(msg)
	return len(p), nil
}

// Printf implements standard logger Printf
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println implements standard logger Println
func (l *Logger) Println(v ...interface{}) {
	l.Info("%v", v...)
}

// Print implements standard logger Print
func (l *Logger) Print(v ...interface{}) {
	l.Info("%v", v...)
}
