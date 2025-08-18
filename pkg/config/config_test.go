package config

import (
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()
	
	// Test default values
	if cfg.IsRoot != (os.Geteuid() == 0) {
		t.Errorf("Expected IsRoot to match current user privileges")
	}
	
	if cfg.Verbose {
		t.Error("Expected Verbose to be false by default")
	}
	
	if cfg.NonInteractive {
		t.Error("Expected NonInteractive to be false by default")
	}
	
	// Test log directory is set
	if cfg.LogDir == "" {
		t.Error("Expected LogDir to be set")
	}
	
	// Test log directory contains expected path elements
	if !strings.Contains(cfg.LogDir, "debian-doctor") {
		t.Error("Expected LogDir to contain 'debian-doctor'")
	}
}

func TestSetVerbose(t *testing.T) {
	cfg := New()
	
	// Test setting verbose to true
	cfg.SetVerbose(true)
	if !cfg.Verbose {
		t.Error("Expected Verbose to be true after SetVerbose(true)")
	}
	
	// Test setting verbose to false
	cfg.SetVerbose(false)
	if cfg.Verbose {
		t.Error("Expected Verbose to be false after SetVerbose(false)")
	}
}

func TestSetNonInteractive(t *testing.T) {
	cfg := New()
	
	// Test setting non-interactive to true
	cfg.SetNonInteractive(true)
	if !cfg.NonInteractive {
		t.Error("Expected NonInteractive to be true after SetNonInteractive(true)")
	}
	
	// Test setting non-interactive to false
	cfg.SetNonInteractive(false)
	if cfg.NonInteractive {
		t.Error("Expected NonInteractive to be false after SetNonInteractive(false)")
	}
}

func TestSetLogDir(t *testing.T) {
	cfg := New()
	originalLogDir := cfg.LogDir
	
	// Test setting custom log directory
	customLogDir := "/tmp/test-debian-doctor"
	cfg.SetLogDir(customLogDir)
	
	if cfg.LogDir != customLogDir {
		t.Errorf("Expected LogDir to be '%s', got '%s'", customLogDir, cfg.LogDir)
	}
	
	// Test that it actually changed from the original
	if cfg.LogDir == originalLogDir {
		t.Error("Expected LogDir to change from original value")
	}
}

func TestLogDirFallback(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	
	// Test with empty HOME to trigger fallback
	os.Setenv("HOME", "")
	defer os.Setenv("HOME", originalHome)
	
	cfg := New()
	
	// Should fallback to /tmp
	if !strings.Contains(cfg.LogDir, "/tmp") {
		t.Errorf("Expected LogDir to contain '/tmp' when HOME is empty, got '%s'", cfg.LogDir)
	}
}

func TestConfigStructure(t *testing.T) {
	cfg := Config{
		LogDir:         "/test/path",
		IsRoot:         true,
		Verbose:        true,
		NonInteractive: true,
	}
	
	if cfg.LogDir != "/test/path" {
		t.Errorf("Expected LogDir '/test/path', got '%s'", cfg.LogDir)
	}
	
	if !cfg.IsRoot {
		t.Error("Expected IsRoot to be true")
	}
	
	if !cfg.Verbose {
		t.Error("Expected Verbose to be true")
	}
	
	if !cfg.NonInteractive {
		t.Error("Expected NonInteractive to be true")
	}
}