package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "debian-doctor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Test successful logger creation
	logger, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	if logger.file == nil {
		t.Error("Expected file to be set")
	}
	
	if logger.logger == nil {
		t.Error("Expected logger to be set")
	}
	
	// Test that log file was created
	logPath := logger.GetLogPath()
	if logPath == "" {
		t.Error("Expected log path to be set")
	}
	
	if !strings.Contains(logPath, "debian-doctor") {
		t.Error("Expected log path to contain 'debian-doctor'")
	}
	
	// Test that file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Expected log file to exist at %s", logPath)
	}
}

func TestNewWithInvalidDir(t *testing.T) {
	// Test with non-existent directory
	invalidDir := "/non/existent/directory"
	logger, err := New(invalidDir)
	
	if err == nil {
		t.Error("Expected error when creating logger with invalid directory")
		if logger != nil {
			logger.Close()
		}
	}
}

func TestLoggingMethods(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "debian-doctor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	logger, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// Test different logging methods
	logger.Info("Test info message")
	logger.Warning("Test warning message")
	logger.Error("Test error message")
	logger.Debug("Test debug message")
	
	// Give it a moment to write
	time.Sleep(100 * time.Millisecond)
	
	// Read the log file
	content, err := ioutil.ReadFile(logger.GetLogPath())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	logContent := string(content)
	
	// Check that all messages were logged
	if !strings.Contains(logContent, "[INFO] Test info message") {
		t.Error("Expected info message in log file")
	}
	
	if !strings.Contains(logContent, "[WARNING] Test warning message") {
		t.Error("Expected warning message in log file")
	}
	
	if !strings.Contains(logContent, "[ERROR] Test error message") {
		t.Error("Expected error message in log file")
	}
	
	if !strings.Contains(logContent, "[DEBUG] Test debug message") {
		t.Error("Expected debug message in log file")
	}
}

func TestFormattedLogging(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "debian-doctor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	logger, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// Test formatted logging
	testValue := 42
	testString := "test"
	logger.Info("Test %s with value %d", testString, testValue)
	
	// Give it a moment to write
	time.Sleep(100 * time.Millisecond)
	
	// Read the log file
	content, err := ioutil.ReadFile(logger.GetLogPath())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	logContent := string(content)
	expectedMessage := "Test test with value 42"
	
	if !strings.Contains(logContent, expectedMessage) {
		t.Errorf("Expected formatted message '%s' in log file", expectedMessage)
	}
}

func TestClose(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "debian-doctor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	logger, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	// Test closing
	err = logger.Close()
	if err != nil {
		t.Errorf("Expected no error when closing logger, got: %v", err)
	}
	
	// Test closing again (should not error)
	err = logger.Close()
	if err != nil {
		t.Errorf("Expected no error when closing already closed logger, got: %v", err)
	}
}

func TestGetLogPath(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := ioutil.TempDir("", "debian-doctor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	logger, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	logPath := logger.GetLogPath()
	
	// Test that path is in the expected directory
	if !strings.HasPrefix(logPath, tmpDir) {
		t.Errorf("Expected log path to start with %s, got %s", tmpDir, logPath)
	}
	
	// Test that filename contains expected elements
	filename := filepath.Base(logPath)
	if !strings.Contains(filename, "debian-doctor") {
		t.Error("Expected filename to contain 'debian-doctor'")
	}
	
	if !strings.HasSuffix(filename, ".log") {
		t.Error("Expected filename to end with '.log'")
	}
}

func TestLoggerStructure(t *testing.T) {
	// Test that Logger struct has expected fields
	logger := &Logger{}
	
	// Test that we can set fields (basic struct validation)
	logger.file = nil
	logger.logger = nil
	
	// Test GetLogPath with nil file
	path := logger.GetLogPath()
	if path != "" {
		t.Errorf("Expected empty path for nil file, got '%s'", path)
	}
}