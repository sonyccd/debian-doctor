package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	LogDir     string
	IsRoot     bool
	Verbose    bool
	NonInteractive bool
}

func New() *Config {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".debian-doctor", "logs")
	
	// If home directory is not accessible, use temp
	if homeDir == "" {
		logDir = "/tmp/debian-doctor-logs"
	}

	return &Config{
		LogDir:         logDir,
		IsRoot:         os.Geteuid() == 0,
		Verbose:        false,
		NonInteractive: false,
	}
}

func (c *Config) SetVerbose(verbose bool) {
	c.Verbose = verbose
}

func (c *Config) SetNonInteractive(nonInteractive bool) {
	c.NonInteractive = nonInteractive
}

func (c *Config) SetLogDir(logDir string) {
	c.LogDir = logDir
}