# Debian Doctor

[![Release](https://img.shields.io/github/v/release/sonyccd/debian-doctor)](https://github.com/sonyccd/debian-doctor/releases/latest)
[![Build Status](https://github.com/sonyccd/debian-doctor/workflows/Release/badge.svg)](https://github.com/sonyccd/debian-doctor/actions)
[![Snap Store](https://snapcraft.io/debian-doctor/badge.svg)](https://snapcraft.io/debian-doctor)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonyccd/debian-doctor)](https://goreportcard.com/report/github.com/sonyccd/debian-doctor)
[![License](https://img.shields.io/github/license/sonyccd/debian-doctor)](LICENSE)

A comprehensive system diagnostic and troubleshooting tool specifically designed for Debian-based Linux systems (Debian, Ubuntu, Mint, etc.). Features automatic system health checks and interactive problem diagnosis with fix suggestions.

## 🎯 Quick Start

```bash
# Install via Snap (recommended)
sudo snap install debian-doctor

# Run system diagnosis
debian-doctor
```

## Features

### 🔍 System Checks
- **System Information**: OS version, kernel, hostname, uptime
- **Disk Space Analysis**: Usage monitoring with configurable thresholds (85%/95% warnings)
- **Memory Usage**: RAM and swap monitoring with utilization metrics
- **Network Configuration**: Interface status, IP addresses, DNS resolution
- **System Services**: Critical service health monitoring (requires root)
- **Filesystem Health**: Mount point validation and disk errors
- **Package System**: APT integrity and broken package detection
- **Log Analysis**: System error log scanning and reporting

### 🩺 Interactive Diagnosis
- **Boot Issues**: System startup problems and service failures
- **Performance Issues**: CPU, memory, and load analysis with optimization tips
- **Network Issues**: Connectivity troubleshooting and DNS resolution
- **Disk Issues**: Storage problems, cleanup suggestions, and filesystem errors
- **Service Issues**: Service management problems and dependency resolution
- **Display Issues**: Graphics, X11, and display manager problems
- **Package Issues**: APT package system problems and repository health
- **Permission Issues**: File access problems and security analysis

### 💻 Interface Features
- **80s Retro Design**: Classic monospace terminal aesthetic
- **Universal Compatibility**: Simple text-based interface works everywhere
- **Real-time Progress**: ASCII progress bars during system checks
- **Interactive Menus**: Numbered options for easy navigation
- **Color-coded Output**: Status indicators (✓ ⚠ ✗ ℹ) for quick visual feedback
- **Comprehensive Logging**: Dual output to file and console
- **Permission Awareness**: Adapts functionality based on user privileges

## Installation

### 📦 Snap Package (Recommended)

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/debian-doctor)

```bash
# Install stable version
sudo snap install debian-doctor

# Install development version (latest from main branch)
sudo snap install debian-doctor --edge

# Run the application
debian-doctor

# Check version
debian-doctor --version
```

**Benefits of Snap installation:**
- ✅ Automatic updates
- ✅ Sandboxed security with classic confinement
- ✅ Works on all major Linux distributions
- ✅ No dependency conflicts

### 📋 Debian Package

```bash
# Download and install latest .deb package
wget https://github.com/sonyccd/debian-doctor/releases/latest/download/debian-doctor_1.0.0-1_amd64.deb
sudo dpkg -i debian-doctor_1.0.0-1_amd64.deb

# Install missing dependencies if needed
sudo apt-get install -f

# Verify installation
debian-doctor --version
man debian-doctor
```

### ⚡ Quick Install (Binary Release)

```bash
# Linux AMD64
wget -O- https://github.com/sonyccd/debian-doctor/releases/latest/download/debian-doctor-linux-amd64.tar.gz | tar xz
sudo mv debian-doctor-linux-amd64 /usr/local/bin/debian-doctor

# Linux ARM64 (Raspberry Pi, etc.)
wget -O- https://github.com/sonyccd/debian-doctor/releases/latest/download/debian-doctor-linux-arm64.tar.gz | tar xz
sudo mv debian-doctor-linux-arm64 /usr/local/bin/debian-doctor

# Linux ARMv7 (32-bit ARM devices)
wget -O- https://github.com/sonyccd/debian-doctor/releases/latest/download/debian-doctor-linux-armv7.tar.gz | tar xz
sudo mv debian-doctor-linux-armv7 /usr/local/bin/debian-doctor
```

### Build from Source

Requirements:
- Go 1.21 or higher

```bash
git clone https://github.com/sonyccd/debian-doctor.git
cd debian-doctor
go build -o debian-doctor .
sudo mv debian-doctor /usr/local/bin/
```

## Usage

### Interactive Mode (Default)

```bash
debian-doctor

# Or with root privileges for full diagnostics
sudo debian-doctor
```

### Command Line Mode

```bash
# Diagnose specific issue types
debian-doctor --diagnose disk        # Disk space issues
debian-doctor --diagnose network     # Network problems
debian-doctor --diagnose services    # Service issues
debian-doctor --diagnose packages    # Package management
debian-doctor --diagnose permissions # File permissions

# Analyze file/directory permissions
debian-doctor --check /path/to/file

# Generate system summary
debian-doctor --summary

# Show help and version
debian-doctor --help
debian-doctor --version
```

### Interactive Menu Options

1. **>>> RUN SYSTEM CHECK** - Execute full diagnostic scan
2. **>>> INTERACTIVE DIAGNOSIS** - Diagnose specific issues
3. **>>> VIEW SYSTEM LOGS** - Display diagnostic history  
4. **>>> EXIT** - Terminate session

### Running with Root Privileges (Recommended)

```bash
sudo debian-doctor
```

Root access enables:
- System service status checks
- System logs analysis
- Package system integrity verification
- Advanced network diagnostics

## Architecture

### Project Structure
```
debian-doctor/
├── cmd/                    # Command line interface
├── internal/
│   ├── checks/            # System check implementations
│   ├── diagnose/          # Problem diagnosis logic
│   ├── tui/              # Terminal user interface
│   └── utils/            # Utility functions
├── pkg/
│   ├── config/           # Configuration management
│   └── logger/           # Logging functionality
├── scripts/               # Development and build scripts
├── main.go               # Application entry point
└── CLAUDE.md             # Project documentation for AI assistants
```

### Key Components

1. **Checks Package**: Implements the `Check` interface for system diagnostics
2. **Diagnose Package**: Provides targeted problem analysis and fix suggestions
3. **TUI Package**: Simple text-based terminal interface for universal compatibility
4. **Config Package**: Application configuration and user preferences
5. **Logger Package**: Structured logging with file and console output

## Development

### Running Tests
```bash
go test ./...                             # Run all tests
go test ./internal/checks                 # Run specific package tests
go test -v ./...                         # Verbose test output
```

### Code Quality
```bash
go fmt ./...                              # Format code
go vet ./...                              # Static analysis
golint ./...                              # Linting (if installed)
```

### Adding New Checks

1. Implement the `Check` interface in `internal/checks/`:
```go
type Check interface {
    Name() string
    Run() CheckResult
    RequiresRoot() bool
}
```

2. Add the check to `GetAllChecks()` in `internal/checks/checks.go`

3. Create corresponding tests in `*_test.go` files

### Adding New Diagnosis Types

1. Create diagnosis function in `internal/diagnose/`
2. Add case to `RunDiagnosis()` in `internal/tui/simple.go`
3. Add menu item to diagnosis menu

## Features Implemented

### Core Functionality
- ✅ All system checks (disk, memory, network, services, system info)
- ✅ Interactive diagnosis menu system
- ✅ Custom issue diagnosis via CLI
- ✅ Fix suggestions with risk levels
- ✅ Permission-aware operations
- ✅ Comprehensive logging
- ✅ Service-specific diagnosis
- ✅ System log analysis
- ✅ Advanced package management
- ✅ Filesystem integrity checks

### User Interface
- ✅ Simple text-based interface for universal compatibility
- ✅ Real-time progress bars
- ✅ Interactive prompts
- ✅ Clear error messages
- ✅ Status indicators

### Safety Features
- ✅ Fix confirmation prompts
- ✅ Risk level indicators
- ✅ Root permission checks
- ✅ Reversible fixes where applicable
- ✅ Safe command execution

## Troubleshooting

### Build Issues
```bash
# Clean module cache
go clean -modcache
go mod download

# Verify Go version
go version

# Check dependencies
go mod verify
```

### Runtime Issues
```bash
# Check permissions
ls -la debian-doctor

# Verify terminal compatibility
echo $TERM

# Check system requirements
uname -a
```

### Common Problems

1. **"Permission denied" errors**: Run with `sudo` for full functionality
2. **Terminal display issues**: Ensure terminal supports colors and UTF-8
3. **Missing system tools**: Install required system utilities (systemctl, etc.)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project maintains the same license as the original Debian Doctor bash script.

## Support

For issues and feature requests, please use the project's issue tracker.