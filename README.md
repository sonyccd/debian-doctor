# Debian Doctor 🩺

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Shell](https://img.shields.io/badge/Shell-Bash-blue.svg)](https://www.gnu.org/software/bash/)
[![Platform](https://img.shields.io/badge/Platform-Debian%20%7C%20Ubuntu-orange.svg)](https://www.debian.org/)

A comprehensive offline system diagnostic and troubleshooting tool for Debian-based systems. Debian Doctor helps you identify and resolve common system issues without requiring internet connectivity.

## 🚀 Features

### Automatic System Analysis
- **System Information** - OS version, kernel, architecture, uptime
- **Disk Space Analysis** - Usage monitoring with critical/warning thresholds
- **Memory Analysis** - RAM and swap usage with pressure detection
- **Service Status** - Critical system services health check
- **Network Configuration** - Interface status, IP addresses, routing, DNS
- **Filesystem Health** - Error detection and mount status
- **System Logs** - Recent error analysis from systemd journal
- **Package System** - Broken packages and lock detection

### Interactive Problem Diagnosis
When automatic checks don't identify issues, Debian Doctor provides guided troubleshooting for:

1. **Boot Issues** - System startup problems
2. **Performance Problems** - Slow system response
3. **Network Connectivity** - Connection and DNS issues
4. **Disk/Storage Issues** - Space and I/O problems
5. **Service Problems** - Application startup failures
6. **Display/Graphics Issues** - GUI and driver problems
7. **Package Management** - APT and dpkg issues
8. **Permission/Access Issues** - File and directory access
9. **Custom Diagnosis** - General troubleshooting guidance

### Smart Fix Suggestions
- Identifies root causes automatically
- Provides specific fix commands with explanations
- Asks for user confirmation before applying changes
- Offers manual alternatives if automated fixes fail
- Includes safety checks and validation

## 📋 Requirements

- **Operating System**: Debian 9+ or Ubuntu 18.04+
- **Shell**: Bash 4.0+
- **Permissions**: Regular user (limited functionality) or root (full features)
- **Dependencies**: Standard system utilities (automatically available)

### Required Tools (usually pre-installed)
```bash
# Core utilities
df, free, ps, systemctl, journalctl, ip, mount, dpkg

# Optional but recommended
bc, nslookup, lspci
```

## 🔧 Installation

### Quick Install
```bash
# Download the script
wget https://raw.githubusercontent.com/sonyccd/debian-doctor/main/debian_doctor.sh

# Make it executable
chmod +x debian_doctor.sh

# Run it
./debian_doctor.sh
```

### System-wide Installation
```bash
# Install system-wide
sudo cp debian_doctor.sh /usr/local/bin/debian-doctor
sudo chmod +x /usr/local/bin/debian-doctor

# Run from anywhere
debian-doctor
```

### Git Clone
```bash
git clone https://github.com/sonyccd/debian-doctor.git
cd debian-doctor
chmod +x debian_doctor.sh
./debian_doctor.sh
```

## 🎯 Usage

### Basic Usage
```bash
# Run with current user privileges
./debian_doctor.sh

# Run with full system access (recommended)
sudo ./debian_doctor.sh
```

### Example Output
```
================================
    DEBIAN DOCTOR v1.0
    System Diagnostic Tool
================================

ℹ Running as root - full system access available

[+] System Information
----------------------------------------
✓ OS: Debian GNU/Linux 12 (bookworm)
✓ Version: 12 (bookworm)
✓ Kernel: 6.1.0-13-amd64
✓ Architecture: x86_64
✓ Hostname: myserver
✓ Uptime: up 2 days, 14 hours, 23 minutes

[+] Disk Space Analysis
----------------------------------------
✓ Disk usage OK on /: 45%
⚠ WARNING: Disk usage high on /var: 87%
✓ Disk usage OK on /home: 23%

[+] Memory Analysis
----------------------------------------
✓ Total Memory: 7956 MB
✓ Available Memory: 5234 MB
✓ Memory Usage: 34%
✓ Swap Usage: 2%
```

### Interactive Mode
After the automatic scan, you can enter interactive mode to diagnose specific issues:

```
Please select the issue you're experiencing:
1) System won't boot properly
2) System is running very slowly
3) Network connectivity issues
4) Disk/storage problems
5) Service/application won't start
6) Display/graphics issues
7) Package management problems
8) Permission/access issues
9) Other/Custom diagnosis
0) Exit

Enter your choice (0-9): 2
```

### Fix Application Example
```
SUGGESTED FIX: Restart networking service
Command: systemctl restart networking

Do you want to run this fix? (y/N): y
Executing: systemctl restart networking
✓ Fix applied successfully
```

## 📊 Understanding Output

### Status Indicators
- **✓ Green**: System is healthy
- **⚠ Yellow**: Warning - attention recommended
- **✗ Red**: Error - immediate action required
- **ℹ Blue**: Information - no action needed

### Severity Levels

#### Critical Issues (Red)
- Disk usage > 95%
- Memory usage > 90%
- Failed critical services
- Filesystem errors
- Root filesystem read-only

#### Warnings (Yellow)
- Disk usage > 85%
- Memory usage > 80%
- High swap usage
- Interface down
- No DNS servers

#### Information (Blue)
- System specifications
- Normal status messages
- Helpful tips and suggestions

## 🛠️ Advanced Usage

### Running Specific Checks
While the script doesn't support individual check selection, you can extract functions for custom use:

```bash
# Source the script to use individual functions
source debian_doctor.sh

# Run specific checks
check_disk_space
check_memory
check_services
```

### Customizing Thresholds
Edit the script to modify warning/error thresholds:

```bash
# Disk space thresholds
if [[ $usage -gt 95 ]]; then    # Critical threshold
if [[ $usage -gt 85 ]]; then    # Warning threshold

# Memory thresholds  
if [[ $mem_usage_percent -gt 90 ]]; then  # Critical
if [[ $mem_usage_percent -gt 80 ]]; then  # Warning
```

### Log Analysis
All activities are logged to `/tmp/debian_doctor.log`:

```bash
# View the log
cat /tmp/debian_doctor.log

# Monitor in real-time
tail -f /tmp/debian_doctor.log
```

## 🔍 Troubleshooting Guide

### Common Issues

#### "Permission denied" errors
**Solution**: Run with sudo for full functionality
```bash
sudo ./debian_doctor.sh
```

#### Script won't execute
**Solution**: Check permissions and shebang
```bash
chmod +x debian_doctor.sh
file debian_doctor.sh  # Should show "Bash script"
```

#### Missing dependencies
**Solution**: Install required packages
```bash
sudo apt update
sudo apt install bc net-tools iproute2
```

#### False positives
Some warnings may be expected in your environment:
- Container environments may show unusual filesystem layouts
- Virtual machines may report different hardware information
- Test systems may have intentionally high resource usage

## 🧪 Testing

### Running Tests
```bash
# Run the included test suite
./debian_doctor_tests.sh

# Check syntax only
bash -n debian_doctor.sh
```

### Test Coverage
The test suite covers:
- ✅ Core functionality
- ✅ Error conditions
- ✅ Edge cases
- ✅ Performance benchmarks
- ✅ Cross-platform compatibility

### CI/CD
This project uses GitHub Actions for automated testing:
- Runs on multiple Ubuntu versions
- Tests with different shell interpreters  
- Includes security scanning
- Validates performance benchmarks

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md).

### Development Setup
```bash
git clone https://github.com/sonyccd/debian-doctor.git
cd debian-doctor

# Run tests
./debian_doctor_tests.sh

# Check code quality
shellcheck debian_doctor.sh
```

### Reporting Issues
When reporting issues, please include:
- Operating system and version
- Error messages (if any)
- Output from `./debian_doctor.sh`
- Steps to reproduce

### Feature Requests
We're always looking to improve! Current roadmap includes:
- Configuration file support
- Plugin system for custom checks
- JSON output format
- Web dashboard interface
- Integration with monitoring systems

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


## 📚 Additional Resources

### Documentation
- [Debian System Administration](https://www.debian.org/doc/manuals/debian-reference/)
- [Ubuntu Server Guide](https://ubuntu.com/server/docs)
- [Systemd Documentation](https://www.freedesktop.org/wiki/Software/systemd/)

### Related Tools
- `htop` - Interactive process viewer
- `iotop` - I/O monitoring
- `netstat` - Network connections
- `lsof` - Open files and processes
- `dmesg` - Kernel messages

### Support
- [GitHub Issues](https://github.com/sonyccd/debian-doctor/issues)

---
