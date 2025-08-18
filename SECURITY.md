# Security Policy

## 🔒 Security Overview

Debian Doctor is a system diagnostic tool that analyzes sensitive system information and can execute privileged operations. We take security seriously and appreciate your help in keeping the project secure.

## 🛡️ Security Considerations

### **Privileged Operations**
- Debian Doctor can run with root privileges for comprehensive system analysis
- The tool can execute system commands as part of fix suggestions
- File permission analysis may access sensitive system files
- Network diagnostics may reveal system configuration details

### **Data Handling**
- System information is logged to temporary files
- Log files may contain sensitive system details
- No data is transmitted externally by default
- All operations are performed locally on the target system

### **Code Execution**
- Fix suggestions may include shell commands
- All fixes require explicit user confirmation
- Risk levels are assigned to all suggested fixes
- Reversible operations are preferred where possible

## 🚨 Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | ✅ Fully supported |
| < 1.0   | ❌ Not supported   |

**Supported Distribution Channels:**
- ✅ Snap Store (stable and edge channels)
- ✅ GitHub Releases (official releases)
- ✅ Debian packages (.deb files)

## 📢 Reporting Security Vulnerabilities

**Please DO NOT report security vulnerabilities through public GitHub issues.**

### **Preferred Reporting Method**

**Email**: bankersalgo@gmail.com  
**Subject**: `[SECURITY] Debian Doctor - [Brief Description]`

### **What to Include**

Please provide as much information as possible:

1. **Vulnerability Description**
   - Clear description of the security issue
   - Potential impact and severity assessment
   - Affected versions or configurations

2. **Reproduction Steps**
   - Step-by-step instructions to reproduce the issue
   - Sample commands or inputs that trigger the vulnerability
   - System environment details (OS, version, etc.)

3. **Proof of Concept**
   - Code snippets, scripts, or commands (if applicable)
   - Screenshots or logs demonstrating the issue
   - Any additional evidence

4. **Suggested Mitigation**
   - Potential fixes or workarounds (if known)
   - Recommendations for users to protect themselves

### **Security Report Template**

```
Subject: [SECURITY] Debian Doctor - [Brief Description]

**Summary:**
Brief description of the vulnerability

**Severity:** [Critical/High/Medium/Low]

**Affected Versions:**
- Version range or specific versions affected

**Description:**
Detailed description of the security issue

**Steps to Reproduce:**
1. Step one
2. Step two
3. Step three

**Impact:**
What could an attacker accomplish?

**Proof of Concept:**
Include code, commands, or screenshots

**Suggested Fix:**
Your recommendations (if any)

**Discovered By:**
Your name/handle (for attribution if you wish)
```

## ⏰ Response Timeline

We commit to the following response times:

- **Initial Response**: Within 48 hours
- **Vulnerability Assessment**: Within 1 week  
- **Fix Development**: 2-4 weeks (depending on complexity)
- **Security Release**: As soon as fix is ready and tested

## 🔧 Security Release Process

1. **Acknowledge**: We confirm receipt and begin investigation
2. **Assess**: We evaluate severity and impact
3. **Develop**: We create and test a fix
4. **Coordinate**: We work with you on disclosure timeline
5. **Release**: We publish the security fix
6. **Disclose**: We publicly acknowledge the issue (with credit if desired)

## 🏆 Recognition

Security researchers who responsibly disclose vulnerabilities will be:

- Credited in the security advisory (unless they prefer to remain anonymous)
- Listed in our security acknowledgments
- Given priority support for future security research

## ⚠️ Security Best Practices for Users

### **Running Debian Doctor Safely**

1. **Review Fix Suggestions**
   - Always read fix descriptions before applying
   - Understand the risk level of each suggested fix
   - Test fixes in non-production environments when possible

2. **Log File Security**
   - Debian Doctor creates log files in `/tmp/debian-doctor-*.log`
   - These files may contain sensitive system information
   - Review and clean up log files after diagnosis
   - Consider running `shred -u /tmp/debian-doctor-*.log` to securely delete

3. **Privilege Management**
   - Run with minimal necessary privileges
   - Use `sudo` only when full system analysis is required
   - Avoid running as root unless necessary

4. **Network Considerations**
   - Network diagnostic features test connectivity
   - Be aware of what network information might be exposed
   - Review network tests before running in sensitive environments

### **Installation Security**

- **Verify Checksums**: Always verify SHA256 checksums for downloaded binaries
- **Use Official Sources**: Download only from GitHub Releases or Snap Store
- **Keep Updated**: Install security updates promptly

## 📋 Known Security Considerations

### **By Design (Not Vulnerabilities)**

1. **Root Access**: Tool requires root for comprehensive system analysis
2. **Command Execution**: Fix suggestions may execute shell commands
3. **File System Access**: Permission analysis accesses system files
4. **Log Files**: Diagnostic information is logged to disk

### **Mitigation Controls**

- All fix executions require explicit user confirmation
- Risk levels are clearly displayed for all suggested fixes
- Commands are shown to users before execution
- Logging can be disabled or customized
- All operations are performed locally (no external communication)

## 📞 Contact Information

- **Security Issues**: bankersalgo@gmail.com
- **General Issues**: [GitHub Issues](https://github.com/sonyccd/debian-doctor/issues)
- **Project Maintainer**: Brad Bazemore

---

**Last Updated**: August 2025  
**Version**: 1.0

Thank you for helping keep Debian Doctor secure! 🛡️