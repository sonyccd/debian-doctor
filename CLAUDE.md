# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

Debian Doctor is a comprehensive system diagnostic and troubleshooting tool for Debian-based systems. It performs automatic system health checks and provides interactive problem diagnosis with fix suggestions.

## Key Commands

### Development and Testing
```bash
# Run the main diagnostic script
./debian_doctor.sh              # Run with current user privileges
sudo ./debian_doctor.sh          # Run with full system access (recommended)

# Run test suite
./debian_doctor_tests.sh        # Execute all unit tests

# Validate syntax
bash -n debian_doctor.sh        # Check script syntax without execution
bash -n debian_doctor_tests.sh  # Check test script syntax

# Code quality checks
shellcheck debian_doctor.sh     # Static analysis for shell scripts
```

### Git Workflow
```bash
# Run tests before committing
./debian_doctor_tests.sh

# Ensure scripts are executable
chmod +x debian_doctor.sh
chmod +x debian_doctor_tests.sh
```

## Architecture and Code Structure

### Main Script (`debian_doctor.sh`)

The script follows a modular architecture with clearly separated concerns:

1. **Configuration Section (lines 1-26)**
   - Color definitions for terminal output
   - Global variables for log files and issue tracking
   - User-specific log file handling

2. **Utility Functions (lines 28-80)**
   - `log()` - Logging with error handling and fallback
   - `print_*()` - Formatted output functions (header, section, ok, warning, error, info)
   - `check_root()` - Permission checking

3. **System Check Functions (lines 82-369)**
   - `check_system_info()` - OS and hardware information
   - `check_disk_space()` - Disk usage analysis with thresholds
   - `check_memory()` - RAM and swap monitoring
   - `check_services()` - Critical service health
   - `check_network()` - Network configuration and connectivity
   - `check_filesystem()` - Filesystem health and mount status
   - `check_logs()` - System log error analysis
   - `check_packages()` - Package system integrity

4. **Interactive Diagnosis Functions (lines 370-674)**
   - `interactive_diagnosis()` - Main menu system
   - `diagnose_*()` - Specific diagnosis functions for each problem category
   - Each diagnosis function provides targeted checks and fix suggestions

5. **Fix Management (lines 675-705)**
   - `offer_fix()` - Prompts user and executes fixes with confirmation
   - Includes safety checks and validation

6. **Summary and Main (lines 706-end)**
   - `generate_summary()` - Produces final report
   - `main()` - Orchestrates execution flow

### Test Script (`debian_doctor_tests.sh`)

The test framework includes:
- Mock system file creation for isolated testing
- Unit test assertions (`assert_equals`, `assert_contains`, etc.)
- Integration and performance testing
- Test coverage for error conditions and edge cases

### CI/CD Pipeline (`.github/workflows/test.yml`)

GitHub Actions workflow provides:
- Multi-version Ubuntu testing (20.04, 22.04, latest)
- Syntax validation and shellcheck analysis
- Unit, integration, and performance tests
- Security scanning for dangerous patterns
- Shell compatibility testing (bash, dash)
- Automated PR commenting with test results

## Critical Design Patterns

1. **Error Handling**: All functions use defensive programming with fallback options
2. **Logging**: Dual logging to screen and file with permission-aware log rotation
3. **Color Coding**: Consistent visual feedback (✓ green, ⚠ yellow, ✗ red, ℹ blue)
4. **Threshold-Based Alerts**: Configurable warning (85%) and critical (95%) levels
5. **User Confirmation**: All fixes require explicit user approval
6. **Modular Testing**: Each system aspect has isolated check functions

## Important Implementation Details

- **Permission Handling**: Script adapts functionality based on root/user privileges
- **Log Files**: User-specific logs to avoid permission conflicts (`/tmp/debian_doctor_$(id -u).log`)
- **Exit on User Request**: Interactive mode respects user's choice to exit (option 0)
- **Mock Testing**: Tests use mock system files to simulate various system states
- **Performance Constraints**: Script must complete within 60 seconds (CI/CD requirement)

## Common Modifications

When modifying diagnostic checks:
1. Add new check function following the `check_*()` naming pattern
2. Use appropriate print functions for output consistency
3. Update `ISSUES_FOUND` or `WARNINGS_FOUND` arrays
4. Call new function from `main()` in the appropriate sequence

When adding interactive diagnosis options:
1. Add menu item in `interactive_diagnosis()`
2. Create corresponding `diagnose_*()` function
3. Include specific checks and offer fixes using `offer_fix()`
4. Ensure all fixes are safe and reversible when possible

## Testing Requirements

Before committing changes:
1. Run `./debian_doctor_tests.sh` to ensure all tests pass
2. Verify syntax with `bash -n debian_doctor.sh`
3. Test both with and without root privileges
4. Ensure new functions have corresponding test coverage
5. Check that execution completes within performance thresholds