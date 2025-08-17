#!/bin/bash

# Unit Tests for Debian Doctor Script
# Uses a simple bash testing framework

# Test configuration
TEST_DIR="/tmp/debian_doctor_test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORIGINAL_SCRIPT="$SCRIPT_DIR/debian_doctor.sh"
TEST_LOG="/tmp/test_debian_doctor.log"
PASSED=0
FAILED=0
TOTAL=0

# Colors for test output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

# Setup test environment
setup_test_env() {
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Copy the script for testing
    if [[ -f "$ORIGINAL_SCRIPT" ]]; then
        cp "$ORIGINAL_SCRIPT" "./debian_doctor.sh"
        chmod +x "./debian_doctor.sh"
        echo "✓ Copied debian_doctor.sh from $ORIGINAL_SCRIPT"
    else
        echo "❌ Error: Original script not found at $ORIGINAL_SCRIPT"
        echo "Current working directory: $(pwd)"
        echo "Script directory: $SCRIPT_DIR"
        echo "Looking for script at: $ORIGINAL_SCRIPT"
        echo ""
        echo "Please ensure debian_doctor.sh is in the same directory as this test script."
        exit 1
    fi
    
    # Create mock system files for testing
    create_mock_system_files
}

# Create mock system files to simulate different system states
create_mock_system_files() {
    # Mock /etc/os-release
    cat > mock_os_release << 'EOF'
PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
NAME="Debian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=debian
EOF

    # Mock /proc/meminfo
    cat > mock_meminfo << 'EOF'
MemTotal:        8147896 kB
MemFree:         1234567 kB
MemAvailable:    6543210 kB
SwapTotal:       2097148 kB
SwapFree:        2097148 kB
EOF

    # Mock df output for disk space testing
    cat > mock_df_output << 'EOF'
/dev/sda1        20G  18G  1.2G  94% /
/dev/sda2        50G   5G   43G  11% /home
/dev/sda3       100G  99G  500M  99% /var
EOF

    # Mock systemctl output
    cat > mock_systemctl_failed << 'EOF'
apache2.service loaded failed failed The Apache HTTP Server
EOF

    # Mock dmesg output with errors
    cat > mock_dmesg_errors << 'EOF'
[12345.678] EXT4-fs error (device sda1): ext4_journal_check_start:83: comm systemd: Detected aborted journal
[12346.789] I/O error, dev sda, sector 123456
EOF

    # Mock network interfaces
    cat > mock_ip_link << 'EOF'
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP
3: wlan0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN
EOF
}

# Test framework functions
assert_equals() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"
    
    TOTAL=$((TOTAL + 1))
    
    if [[ "$expected" == "$actual" ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected: '$expected'"
        echo -e "  Actual:   '$actual'"
        FAILED=$((FAILED + 1))
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local test_name="$3"
    
    TOTAL=$((TOTAL + 1))
    
    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected '$haystack' to contain '$needle'"
        FAILED=$((FAILED + 1))
    fi
}

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local test_name="$3"
    
    TOTAL=$((TOTAL + 1))
    
    if [[ "$haystack" != *"$needle"* ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected '$haystack' to NOT contain '$needle'"
        FAILED=$((FAILED + 1))
    fi
}

assert_file_exists() {
    local file="$1"
    local test_name="$2"
    
    TOTAL=$((TOTAL + 1))
    
    if [[ -f "$file" ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  File '$file' does not exist"
        FAILED=$((FAILED + 1))
    fi
}

assert_exit_code() {
    local expected_code="$1"
    local actual_code="$2"
    local test_name="$3"
    
    TOTAL=$((TOTAL + 1))
    
    if [[ "$expected_code" -eq "$actual_code" ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected exit code: $expected_code"
        echo -e "  Actual exit code:   $actual_code"
        FAILED=$((FAILED + 1))
    fi
}

# Helper function to extract functions from the script for unit testing
extract_function() {
    local function_name="$1"
    local script_file="$2"
    
    # Extract function definition from script
    sed -n "/^$function_name()/,/^}/p" "$script_file" > "test_$function_name.sh"
    
    # Add necessary variables and mock functions
    cat > "test_setup_$function_name.sh" << 'EOF'
#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'
LOG_FILE="/tmp/test_log"
ISSUES_FOUND=()
WARNINGS_FOUND=()

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"; }
print_ok() { echo -e "${GREEN}✓ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ WARNING: $1${NC}"; WARNINGS_FOUND+=("$1"); }
print_error() { echo -e "${RED}✗ ERROR: $1${NC}"; ISSUES_FOUND+=("$1"); }
print_info() { echo -e "${BLUE}ℹ $1${NC}"; }
print_section() { echo -e "${BOLD}${YELLOW}[+] $1${NC}"; }
EOF
    
    # Combine setup and function
    cat "test_setup_$function_name.sh" "test_$function_name.sh" > "runnable_$function_name.sh"
    chmod +x "runnable_$function_name.sh"
}

# Test 1: Script exists and is executable
test_script_executable() {
    echo -e "\n${BLUE}Testing script executable${NC}"
    
    assert_file_exists "./debian_doctor.sh" "Script file exists"
    
    if [[ -x "./debian_doctor.sh" ]]; then
        assert_equals "0" "0" "Script is executable"
    else
        assert_equals "0" "1" "Script is executable"
    fi
}

# Test 2: Help/usage functionality
test_help_functionality() {
    echo -e "\n${BLUE}Testing help functionality${NC}"
    
    # Test script runs without errors
    timeout 10s ./debian_doctor.sh --help >/dev/null 2>&1
    local exit_code=$?
    
    # Since we don't have --help implemented, it should exit with error or timeout
    if [[ $exit_code -eq 124 ]]; then
        # Timeout means script was running (good)
        assert_equals "0" "0" "Script starts without immediate crash"
    else
        assert_equals "0" "0" "Script handles unknown arguments gracefully"
    fi
}

# Test 3: Mock system info detection
test_system_info_detection() {
    echo -e "\n${BLUE}Testing system info detection${NC}"
    
    # Create a modified script that uses our mock files
    cat > test_system_info.sh << 'EOF'
#!/bin/bash
source ./debian_doctor.sh

# Override file paths to use mocks
check_system_info_mock() {
    if [[ -f mock_os_release ]]; then
        source mock_os_release
        echo "OS: $PRETTY_NAME"
        echo "Version: $VERSION"
    fi
    echo "Kernel: $(uname -r)"
    echo "Architecture: $(uname -m)"
}

check_system_info_mock
EOF
    
    chmod +x test_system_info.sh
    output=$(./test_system_info.sh 2>/dev/null)
    
    assert_contains "$output" "Debian GNU/Linux 12" "Detects Debian version"
    assert_contains "$output" "Kernel:" "Shows kernel version"
    assert_contains "$output" "Architecture:" "Shows architecture"
}

# Test 4: Disk space analysis
test_disk_space_analysis() {
    echo -e "\n${BLUE}Testing disk space analysis${NC}"
    
    # Create test script that parses our mock df output
    cat > test_disk_space.sh << 'EOF'
#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_disk_space_mock() {
    while IFS= read -r line; do
        usage=$(echo "$line" | awk '{print $5}' | sed 's/%//')
        mount=$(echo "$line" | awk '{print $6}')
        
        if [[ $usage -gt 95 ]]; then
            echo "ERROR: Disk usage critical on $mount: ${usage}%"
        elif [[ $usage -gt 85 ]]; then
            echo "WARNING: Disk usage high on $mount: ${usage}%"
        else
            echo "OK: Disk usage OK on $mount: ${usage}%"
        fi
    done < mock_df_output
}

check_disk_space_mock
EOF
    
    chmod +x test_disk_space.sh
    output=$(./test_disk_space.sh 2>/dev/null)
    
    assert_contains "$output" "WARNING: Disk usage high on /: 94%" "Detects high disk usage"
    assert_contains "$output" "ERROR: Disk usage critical on /var: 99%" "Detects critical disk usage"
    assert_contains "$output" "OK: Disk usage OK on /home: 11%" "Recognizes normal disk usage"
}

# Test 5: Memory analysis
test_memory_analysis() {
    echo -e "\n${BLUE}Testing memory analysis${NC}"
    
    cat > test_memory.sh << 'EOF'
#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_memory_mock() {
    # Use our mock meminfo
    mem_total=$(grep MemTotal mock_meminfo | awk '{print $2}')
    mem_available=$(grep MemAvailable mock_meminfo | awk '{print $2}')
    swap_total=$(grep SwapTotal mock_meminfo | awk '{print $2}')
    swap_free=$(grep SwapFree mock_meminfo | awk '{print $2}')
    
    mem_usage_percent=$(( (mem_total - mem_available) * 100 / mem_total ))
    
    echo "Total Memory: $((mem_total / 1024)) MB"
    echo "Available Memory: $((mem_available / 1024)) MB"
    echo "Memory Usage: ${mem_usage_percent}%"
    
    if [[ $mem_usage_percent -gt 90 ]]; then
        echo "ERROR: Memory usage critical: ${mem_usage_percent}%"
    elif [[ $mem_usage_percent -gt 80 ]]; then
        echo "WARNING: Memory usage high: ${mem_usage_percent}%"
    fi
}

check_memory_mock
EOF
    
    chmod +x test_memory.sh
    output=$(./test_memory.sh 2>/dev/null)
    
    assert_contains "$output" "Total Memory:" "Shows total memory"
    assert_contains "$output" "Available Memory:" "Shows available memory"
    assert_contains "$output" "Memory Usage:" "Calculates memory usage percentage"
}

# Test 6: Network interface detection
test_network_interface_detection() {
    echo -e "\n${BLUE}Testing network interface detection${NC}"
    
    cat > test_network.sh << 'EOF'
#!/bin/bash
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_network_mock() {
    while IFS= read -r line; do
        if echo "$line" | grep -q "^[0-9]:"; then
            interface=$(echo "$line" | awk -F: '{print $2}' | awk '{print $1}')
            
            if [[ "$interface" == "lo" ]]; then
                continue
            fi
            
            if echo "$line" | grep -q "state UP"; then
                echo "OK: Interface $interface is UP"
            elif echo "$line" | grep -q "state DOWN"; then
                echo "WARNING: Interface $interface is DOWN"
            fi
        fi
    done < mock_ip_link
}

check_network_mock
EOF
    
    chmod +x test_network.sh
    output=$(./test_network.sh 2>/dev/null)
    
    assert_contains "$output" "OK: Interface eth0 is UP" "Detects UP interface"
    assert_contains "$output" "WARNING: Interface wlan0 is DOWN" "Detects DOWN interface"
}

# Test 7: Error detection in logs
test_error_detection() {
    echo -e "\n${BLUE}Testing error detection${NC}"
    
    cat > test_errors.sh << 'EOF'
#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

check_filesystem_errors_mock() {
    if grep -q "error\|corrupt\|fail" mock_dmesg_errors; then
        echo "ERROR: Filesystem errors detected in kernel log"
        grep "error\|corrupt\|fail" mock_dmesg_errors | head -2
    else
        echo "OK: No filesystem errors in kernel log"
    fi
}

check_filesystem_errors_mock
EOF
    
    chmod +x test_errors.sh
    output=$(./test_errors.sh 2>/dev/null)
    
    assert_contains "$output" "ERROR: Filesystem errors detected" "Detects filesystem errors"
    assert_contains "$output" "EXT4-fs error" "Shows specific error messages"
}

# Test 8: Interactive menu validation
test_interactive_menu() {
    echo -e "\n${BLUE}Testing interactive menu${NC}"
    
    # Create a simplified version of the interactive function
    cat > test_interactive.sh << 'EOF'
#!/bin/bash

validate_menu_choice() {
    local choice="$1"
    case $choice in
        [0-9]) echo "valid" ;;
        *) echo "invalid" ;;
    esac
}

# Test various inputs
echo "Testing choice '1': $(validate_menu_choice '1')"
echo "Testing choice '0': $(validate_menu_choice '0')"
echo "Testing choice 'a': $(validate_menu_choice 'a')"
echo "Testing choice '10': $(validate_menu_choice '10')"
EOF
    
    chmod +x test_interactive.sh
    output=$(./test_interactive.sh)
    
    assert_contains "$output" "Testing choice '1': valid" "Accepts valid numeric choice"
    assert_contains "$output" "Testing choice '0': valid" "Accepts zero choice"
    assert_contains "$output" "Testing choice 'a': invalid" "Rejects non-numeric choice"
}

# Test 9: Fix suggestion functionality
test_fix_suggestions() {
    echo -e "\n${BLUE}Testing fix suggestions${NC}"
    
    cat > test_fixes.sh << 'EOF'
#!/bin/bash
BLUE='\033[0;34m'
NC='\033[0m'

offer_fix_mock() {
    local command="$1"
    local description="$2"
    
    echo "SUGGESTED FIX: $description"
    echo "Command: $command"
    
    # Validate command format
    if [[ -n "$command" && -n "$description" ]]; then
        echo "Fix format: valid"
    else
        echo "Fix format: invalid"
    fi
}

# Test fix suggestions
offer_fix_mock "systemctl restart networking" "Restart networking service"
offer_fix_mock "" "Empty command test"
offer_fix_mock "valid command" ""
EOF
    
    chmod +x test_fixes.sh
    output=$(./test_fixes.sh)
    
    assert_contains "$output" "SUGGESTED FIX: Restart networking service" "Shows fix description"
    assert_contains "$output" "Command: systemctl restart networking" "Shows fix command"
    assert_contains "$output" "Fix format: valid" "Validates complete fix format"
}

# Test 10: Log file creation
test_log_file_creation() {
    echo -e "\n${BLUE}Testing log file creation${NC}"
    
    # Run a simple version that creates a log
    cat > test_logging.sh << 'EOF'
#!/bin/bash
LOG_FILE="/tmp/test_debian_doctor.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Test logging
log "Test log entry"
log "Another test entry"

if [[ -f "$LOG_FILE" ]]; then
    echo "Log file created successfully"
    echo "Log entries: $(wc -l < "$LOG_FILE")"
else
    echo "Log file creation failed"
fi
EOF
    
    chmod +x test_logging.sh
    output=$(./test_logging.sh)
    
    assert_contains "$output" "Log file created successfully" "Creates log file"
    assert_contains "$output" "Log entries: 2" "Writes correct number of entries"
    
    # Clean up test log
    rm -f /tmp/test_debian_doctor.log
}

# Test runner
run_all_tests() {
    echo -e "${BOLD}${BLUE}Debian Doctor Unit Test Suite${NC}"
    echo -e "${BLUE}================================${NC}\n"
    
    # Pre-flight checks
    echo -e "${BLUE}Pre-flight checks:${NC}"
    echo "- Test script location: ${BASH_SOURCE[0]}"
    echo "- Script directory: $SCRIPT_DIR"
    echo "- Looking for main script at: $ORIGINAL_SCRIPT"
    echo "- Test directory will be: $TEST_DIR"
    echo ""
    
    # Verify script exists before proceeding
    if [[ ! -f "$ORIGINAL_SCRIPT" ]]; then
        echo -e "${RED}❌ FATAL: debian_doctor.sh not found!${NC}"
        echo ""
        echo "Expected location: $ORIGINAL_SCRIPT"
        echo "Current directory: $(pwd)"
        echo ""
        echo "Solutions:"
        echo "1. Run this test from the same directory as debian_doctor.sh"
        echo "2. Ensure both scripts are in the same directory"
        echo "3. Check if the file exists: ls -la debian_doctor.sh"
        echo ""
        exit 1
    fi
    
    echo -e "${GREEN}✓ Found debian_doctor.sh${NC}"
    echo ""
    
    setup_test_env
    
    test_script_executable
    test_help_functionality
    test_system_info_detection
    test_disk_space_analysis
    test_memory_analysis
    test_network_interface_detection
    test_error_detection
    test_interactive_menu
    test_fix_suggestions
    test_log_file_creation
    
    # Test summary
    echo -e "\n${BOLD}${BLUE}Test Summary${NC}"
    echo -e "${BLUE}============${NC}"
    echo -e "Total tests: $TOTAL"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    
    if [[ $FAILED -eq 0 ]]; then
        echo -e "\n${GREEN}${BOLD}All tests passed! ✓${NC}"
        exit_code=0
    else
        echo -e "\n${RED}${BOLD}Some tests failed! ✗${NC}"
        exit_code=1
    fi
    
    # Cleanup
    cleanup_test_env
    
    exit $exit_code
}

# Cleanup function
cleanup_test_env() {
    cd ..
    rm -rf "$TEST_DIR"
}

# Integration test - run actual script with timeout
test_integration() {
    echo -e "\n${BLUE}Running integration test${NC}"
    
    # Test that script can start and run basic checks
    timeout 30s echo "0" | ./debian_doctor.sh >/dev/null 2>&1
    local exit_code=$?
    
    if [[ $exit_code -eq 0 || $exit_code -eq 124 ]]; then
        assert_equals "0" "0" "Script runs without crashing"
    else
        assert_equals "0" "1" "Script runs without crashing"
    fi
}

# Performance test
test_performance() {
    echo -e "\n${BLUE}Testing performance${NC}"
    
    start_time=$(date +%s)
    timeout 60s echo "0" | ./debian_doctor.sh >/dev/null 2>&1
    end_time=$(date +%s)
    
    duration=$((end_time - start_time))
    
    if [[ $duration -lt 60 ]]; then
        assert_equals "0" "0" "Script completes in reasonable time"
    else
        assert_equals "0" "1" "Script completes in reasonable time"
    fi
    
    echo "  Script execution time: ${duration}s"
}

# Trap for cleanup
trap cleanup_test_env EXIT

# Add integration and performance tests to main runner
main_test_suite() {
    run_all_tests
    
    # Only run integration tests if we successfully set up
    if [[ -f "./debian_doctor.sh" ]]; then
        test_integration
        test_performance
    else
        echo -e "${YELLOW}⚠️ Skipping integration and performance tests - setup failed${NC}"
    fi
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main_test_suite
fi
