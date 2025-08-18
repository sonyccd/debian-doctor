# Debian Doctor

A comprehensive system diagnostic and troubleshooting tool for Debian-based systems with a modern Terminal User Interface (TUI) built using Bubble Tea.

> **Note**: This is a complete rewrite of the original bash script in Go, providing enhanced performance, better error handling, and a modern interactive interface.

## Features

### System Checks
- **System Information**: OS version, kernel, hostname, uptime
- **Disk Space Analysis**: Usage monitoring with configurable thresholds
- **Memory Usage**: RAM and swap monitoring
- **Network Configuration**: Interface status, IP addresses, DNS
- **System Services**: Critical service health monitoring (requires root)

### Interactive Diagnosis
- **Boot Issues**: System startup problems
- **Performance Issues**: CPU, memory, and load analysis
- **Network Issues**: Connectivity troubleshooting
- **Disk Issues**: Storage problems and cleanup suggestions
- **Service Issues**: Service management problems
- **Display Issues**: Graphics and X11 problems
- **Package Issues**: APT package system problems
- **Permission Issues**: File access problems

### Modern TUI Features
- Beautiful terminal interface with colors and animations
- Progress bars for system checks
- Interactive menus with keyboard navigation
- Real-time status updates
- Comprehensive logging

## Installation

### Prerequisites
- Go 1.21 or later
- Debian-based Linux system
- Terminal with color support

### Build from Source

```bash
git clone <repository>
cd go-debian-doctor
go mod tidy
go build -o debian-doctor
```

### Install System-wide (Optional)

```bash
sudo cp debian-doctor /usr/local/bin/
sudo chmod +x /usr/local/bin/debian-doctor
```

## Usage

### Interactive Mode (Default)
```bash
./debian-doctor
```

### Command Line Options
```bash
./debian-doctor --help                    # Show help
./debian-doctor --non-interactive         # Run without TUI
./debian-doctor --verbose                 # Enable verbose output
```

### Running with Root Privileges (Recommended)
```bash
sudo ./debian-doctor
```

Running as root enables additional checks:
- System service status
- System logs analysis
- Package system integrity
- Advanced network diagnostics

## Navigation

### Main Menu
- `↑/↓` - Navigate menu options
- `Enter` - Select option
- `q` or `Ctrl+C` - Quit

### During System Checks
- System checks run automatically with progress indication
- Press `Ctrl+C` to cancel

### Results View
- `↑/↓` - Scroll through results
- `Esc` - Return to main menu

### Interactive Diagnosis
- `↑/↓` - Navigate diagnosis options
- `Enter` - Select diagnosis type
- `f` - Apply suggested fix (when available)
- `Esc` - Go back

### Fix Application
- `y` - Confirm fix application
- `n` - Cancel fix
- `Esc` - Go back

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
├── Makefile              # Build automation
└── CLAUDE.md             # Project documentation for AI assistants
```

### Key Components

1. **Checks Package**: Implements the `Check` interface for system diagnostics
2. **Diagnose Package**: Provides targeted problem analysis and fix suggestions
3. **TUI Package**: Bubble Tea based terminal interface with multiple views
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
2. Add case to `runDiagnosis()` in `internal/tui/model.go`
3. Add menu item to diagnosis menu

## Comparison with Bash Version

### Advantages of Go Version
- **Better Performance**: Compiled binary vs interpreted script
- **Modern Interface**: Rich TUI with colors, progress bars, and animations
- **Type Safety**: Go's type system prevents runtime errors
- **Better Testing**: Comprehensive test suite with mocks
- **Cross-compilation**: Can build for different architectures
- **Structured Logging**: Better log management and formatting
- **Memory Safety**: No shell injection vulnerabilities
- **Dependency Management**: Go modules for reliable builds

### Feature Parity
- ✅ All system checks from bash version
- ✅ Interactive diagnosis with fix suggestions
- ✅ Permission-aware operations
- ✅ Comprehensive logging
- ✅ Root/non-root operation modes
- ✅ Error handling and recovery

### Additional Features
- Progress indicators during checks
- Real-time UI updates
- Better error reporting
- Structured configuration
- Improved navigation
- Professional appearance

## Current Capabilities vs Original Bash Script

### ✅ **Fully Implemented**
- All system checks (disk, memory, network, services, system info)
- Interactive diagnosis menu system
- Modern TUI with animations and visual feedback
- Structured logging and configuration
- Comprehensive test suite
- Cross-platform builds

### 🚧 **Partially Implemented**
- **Fix Suggestions**: Currently shows fixes but doesn't execute them
- **Package Management**: Basic checks implemented, advanced APT diagnostics pending

### ❌ **Not Yet Implemented** (from original bash script)
- **Fix Execution**: `offer_fix()` functionality - actually running suggested commands
- **Custom Diagnosis**: Free-form issue description with general troubleshooting steps
- **Advanced Package Analysis**: Comprehensive APT package system checks
- **Service-Specific Diagnosis**: User input for specific service troubleshooting  
- **File Permission Analysis**: Detailed permission diagnosis for user-specified paths
- **Comprehensive Summary**: End-of-scan summary report generation
- **System Log Analysis**: Parsing systemd journal and log files for errors
- **Filesystem Integrity**: Read-only filesystem detection and remount suggestions

### 🎯 **Roadmap**
1. **Fix Execution System**: Implement safe command execution with confirmation
2. **Advanced Diagnostics**: Add custom diagnosis and detailed service analysis
3. **Log Analysis**: Implement comprehensive system log parsing
4. **Package Management**: Enhanced APT system checks and fixes
5. **Filesystem Checks**: Advanced filesystem integrity and permissions analysis

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