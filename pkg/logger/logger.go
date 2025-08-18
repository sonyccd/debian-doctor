package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/debian-doctor/debian-doctor/pkg/config"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
}

// NewFromConfig creates a new logger using configuration
func NewFromConfig(cfg *config.Config) (*Logger, error) {
	return New(cfg.LogDir)
}

func New(logDir string) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("debian-doctor_%s.log", timestamp))
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "", log.LstdFlags)

	return &Logger{
		file:   file,
		logger: logger,
	}, nil
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.logger.Printf("[WARNING] "+format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}

func (l *Logger) Close() error {
	if l.file != nil {
		err := l.file.Close()
		l.file = nil // Prevent double close
		return err
	}
	return nil
}

// GetLogPath returns the path to the current log file
func (l *Logger) GetLogPath() string {
	if l.file != nil {
		return l.file.Name()
	}
	return ""
}