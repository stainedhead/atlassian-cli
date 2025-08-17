package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Event represents an audit log entry
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// Logger handles audit logging
type Logger struct {
	file *os.File
}

// NewLogger creates a new audit logger
func NewLogger() (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, ".atlassian-cli", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "audit.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	return &Logger{file: file}, nil
}

// Log records an audit event
func (l *Logger) Log(user, command string, args []string, success bool, err error) {
	event := Event{
		Timestamp: time.Now(),
		User:      user,
		Command:   command,
		Args:      args,
		Success:   success,
	}

	if err != nil {
		event.Error = err.Error()
	}

	data, _ := json.Marshal(event)
	l.file.WriteString(string(data) + "\n")
}

// Close closes the audit logger
func (l *Logger) Close() error {
	return l.file.Close()
}