#!/bin/bash

# Debian Doctor - Offline System Diagnostic Script
# Version 1.0
# Diagnoses common Debian system issues without internet connectivity

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Global variables
# Create user-specific log file to avoid permission conflicts
if [[ $EUID -eq 0 ]]; then
    LOG_FILE="/tmp/debian_doctor_root.log"
else
    LOG_FILE="/tmp/debian_doctor_$(id -u).log"
fi
ISSUES_FOUND=()
WARNINGS_FOUND=()

# Logging function with error handling
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE" 2>/dev/null || {
        # If logging fails, try to create a new log file
        local new_log="/tmp/debian_doctor_backup_$(date +%s).log"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Log permission error, switching to: $new_log" > "$new_log"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$new_log"
        LOG_FILE="$new_log"
    }
}

# Print functions
print_header() {
    echo -e "${BOLD}${BLUE}================================${NC}"
    echo -e "${BOLD}${BLUE}    DEBIAN DOCTOR v1.0${NC}"
    echo -e "${BOLD}${BLUE}    System Diagnostic Tool${NC}"
    echo -e "${BOLD}${BLUE}================================${NC}"
    echo
}

print_section() {
    echo -e "${BOLD}${YELLOW}[+] $1${NC}"
    echo "----------------------------------------"
}

print_ok() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
    WARNINGS_FOUND+=("$1")
}

print_error() {
    echo -e "${RED}✗ ERROR: $1${NC}"
    ISSUES_FOUND+=("$1")
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_info "Running as root - full system access available"
        return 0
    else
        print_warning "Not running as root - some checks may be limited"
        return 1
    fi
}

# System Information
check_system_info() {
    print_section "System Information"
    
    # Basic system info
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        print_ok "OS: $PRETTY_NAME"
        print_ok "Version: $VERSION"
    fi
    
    print_ok "Kernel: $(uname -r)"
    print_ok "Architecture: $(uname -m)"
    print_ok "Hostname: $(hostname)"
    print_ok "Uptime: $(uptime -p)"
    
    # Check if it's actually Debian
    if ! grep -q "debian" /etc/os-release 2>/dev/null; then
        print_warning "This doesn't appear to be a Debian system"
    fi
    
    echo
}

# Check disk space
check_disk_space() {
    print_section "Disk Space Analysis"
    
    while IFS= read -r line; do
        usage=$(echo "$line" | awk '{print $5}' | sed 's/%//')
        mount=$(echo "$line" | awk '{print $6}')
        
        # Skip if usage is empty or not a number
        if [[ -n "$usage" && "$usage" =~ ^[0-9]+$ ]]; then
            if [[ $usage -gt 95 ]]; then
                print_error "Disk usage critical on $mount: ${usage}%"
            elif [[ $usage -gt 85 ]]; then
                print_warning "Disk usage high on $mount: ${usage}%"
            else
                print_ok "Disk usage OK on $mount: ${usage}%"
            fi
        fi
    done < <(df -h | grep -E '^/dev/')
    
    # Check inode usage
    print_info "Checking inode usage..."
    while IFS= read -r line; do
        usage=$(echo "$line" | awk '{print $5}' | sed 's/%//')
        mount=$(echo "$line" | awk '{print $6}')
        
        # Skip if usage is empty or not a number
        if [[ -n "$usage" && "$usage" =~ ^[0-9]+$ ]]; then
            if [[ $usage -gt 90 ]]; then
                print_error "Inode usage critical on $mount: ${usage}%"
            elif [[ $usage -gt 80 ]]; then
                print_warning "Inode usage high on $mount: ${usage}%"
            fi
        fi
    done < <(df -i | grep -E '^/dev/')
    
    echo
}

# Check memory usage
check_memory() {
    print_section "Memory Analysis"
    
    # Parse /proc/meminfo
    mem_total=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    mem_available=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
    swap_total=$(grep SwapTotal /proc/meminfo | awk '{print $2}')
    swap_free=$(grep SwapFree /proc/meminfo | awk '{print $2}')
    
    mem_usage_percent=$(( (mem_total - mem_available) * 100 / mem_total ))
    
    print_ok "Total Memory: $((mem_total / 1024)) MB"
    print_ok "Available Memory: $((mem_available / 1024)) MB"
    print_ok "Memory Usage: ${mem_usage_percent}%"
    
    if [[ $mem_usage_percent -gt 90 ]]; then
        print_error "Memory usage critical: ${mem_usage_percent}%"
    elif [[ $mem_usage_percent -gt 80 ]]; then
        print_warning "Memory usage high: ${mem_usage_percent}%"
    fi
    
    if [[ $swap_total -gt 0 ]]; then
        swap_usage_percent=$(( (swap_total - swap_free) * 100 / swap_total ))
        print_ok "Swap Usage: ${swap_usage_percent}%"
        
        if [[ $swap_usage_percent -gt 50 ]]; then
            print_warning "High swap usage may indicate memory pressure"
        fi
    else
        print_warning "No swap space configured"
    fi
    
    echo
}

# Check system services
check_services() {
    print_section "Critical Services Status"
    
    critical_services=("systemd-logind" "dbus" "networking" "ssh" "cron")
    
    for service in "${critical_services[@]}"; do
        if systemctl is-active --quiet "$service" 2>/dev/null; then
            print_ok "$service is running"
        else
            if systemctl list-unit-files | grep -q "^$service"; then
                print_error "$service is not running"
            else
                print_info "$service is not installed"
            fi
        fi
    done
    
    # Check for failed services with better parsing
    print_info "Checking for failed services..."
    failed_output=$(systemctl --failed --no-legend --no-pager 2>/dev/null)
    
    if [[ -n "$failed_output" ]]; then
        print_error "Failed services detected:"
        
        # Parse each line of failed services output
        echo "$failed_output" | while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Extract service name - it's the first field before any spaces
                service_name=$(echo "$line" | awk '{print $1}' | sed 's/^●\s*//' | sed 's/^\*\s*//')
                
                # Only show if we have a valid service name
                if [[ -n "$service_name" && "$service_name" != "●" && "$service_name" != "*" ]]; then
                    echo "  - $service_name"
                    
                    # Try to get failure reason
                    if systemctl status "$service_name" --no-pager -l 2>/dev/null | grep -q "failed"; then
                        failure_info=$(systemctl status "$service_name" --no-pager -l 2>/dev/null | grep -E "failed|error|exit" | head -1 | sed 's/^[[:space:]]*//')
                        if [[ -n "$failure_info" ]]; then
                            echo "    Reason: $failure_info"
                        fi
                    fi
                fi
            fi
        done
    else
        print_ok "No failed services found"
    fi
    
    echo
}

# Check network configuration
check_network() {
    print_section "Network Configuration"
    
    # Check interfaces
    interfaces=$(ip link show | grep -E '^[0-9]+:' | awk -F: '{print $2}' | tr -d ' ')
    
    for interface in $interfaces; do
        if [[ "$interface" == "lo" ]]; then
            continue
        fi
        
        state=$(ip link show "$interface" | grep -o 'state [A-Z]*' | awk '{print $2}')
        
        if [[ "$state" == "UP" ]]; then
            print_ok "Interface $interface is UP"
            
            # Check if it has an IP
            if ip addr show "$interface" | grep -q "inet "; then
                ip_addr=$(ip addr show "$interface" | grep "inet " | awk '{print $2}' | head -n1)
                print_ok "  IP: $ip_addr"
            else
                print_warning "  No IP address assigned"
            fi
        else
            print_warning "Interface $interface is $state"
        fi
    done
    
    # Check routing
    if ip route | grep -q default; then
        default_route=$(ip route | grep default | head -n1)
        print_ok "Default route: $default_route"
    else
        print_warning "No default route configured"
    fi
    
    # Check DNS
    if [[ -f /etc/resolv.conf ]]; then
        dns_servers=$(grep nameserver /etc/resolv.conf | awk '{print $2}')
        if [[ -n "$dns_servers" ]]; then
            print_ok "DNS servers configured: $(echo $dns_servers | tr '\n' ' ')"
        else
            print_warning "No DNS servers configured"
        fi
    fi
    
    echo
}

# Check filesystem integrity
check_filesystem() {
    print_section "Filesystem Health"
    
    # Check for filesystem errors in dmesg
    if dmesg | grep -i "filesystem\|ext[234]\|xfs\|btrfs" | grep -i "error\|corrupt\|fail" | head -5 | grep -q .; then
        print_error "Filesystem errors detected in kernel log:"
        dmesg | grep -i "filesystem\|ext[234]\|xfs\|btrfs" | grep -i "error\|corrupt\|fail" | tail -3
    else
        print_ok "No filesystem errors in kernel log"
    fi
    
    # Check mounted filesystems
    while IFS= read -r line; do
        filesystem=$(echo "$line" | awk '{print $1}')
        mount_point=$(echo "$line" | awk '{print $3}')
        fs_type=$(echo "$line" | awk '{print $5}')
        options=$(echo "$line" | awk '{print $6}')
        
        if echo "$options" | grep -q "ro,"; then
            print_error "Filesystem $mount_point is mounted read-only"
        else
            print_ok "Filesystem $mount_point ($fs_type) is writable"
        fi
    done < <(mount | grep -E '^/dev/')
    
    echo
}

# Check system logs for errors
check_logs() {
    print_section "Recent System Errors"
    
    # Check systemd journal for errors
    if command -v journalctl >/dev/null; then
        error_count=$(journalctl --since "1 hour ago" -p err --no-pager | wc -l)
        if [[ $error_count -gt 0 ]]; then
            print_warning "$error_count errors in the last hour"
            print_info "Recent errors:"
            journalctl --since "1 hour ago" -p err --no-pager | tail -5
        else
            print_ok "No errors in the last hour"
        fi
    fi
    
    # Check for kernel panics or oops
    if dmesg | grep -i "panic\|oops\|segfault" | head -3 | grep -q .; then
        print_error "Kernel issues detected:"
        dmesg | grep -i "panic\|oops\|segfault" | tail -3
    fi
    
    echo
}

# Check package system
check_packages() {
    print_section "Package System Status"
    
    # Check if dpkg is locked
    if [[ -f /var/lib/dpkg/lock-frontend ]]; then
        print_warning "dpkg frontend lock exists - package operations may be blocked"
    fi
    
    if [[ -f /var/lib/dpkg/lock ]]; then
        print_warning "dpkg lock exists - package operations may be blocked"
    fi
    
    # Check for broken packages
    if command -v dpkg >/dev/null; then
        broken_packages=$(dpkg -l | grep "^i[^i]" | wc -l)
        if [[ $broken_packages -gt 0 ]]; then
            print_error "$broken_packages broken packages detected"
            print_info "Run 'dpkg -l | grep \"^i[^i]\"' to see them"
        else
            print_ok "No broken packages detected"
        fi
    fi
    
    # Check for held packages
    held_packages=$(dpkg --get-selections | grep hold | wc -l)
    if [[ $held_packages -gt 0 ]]; then
        print_info "$held_packages packages are held"
    fi
    
    echo
}

# Interactive problem selection
interactive_diagnosis() {
    print_section "Interactive Problem Diagnosis"
    
    echo "Please select the issue you're experiencing:"
    echo "1) System won't boot properly"
    echo "2) System is running very slowly"
    echo "3) Network connectivity issues"
    echo "4) Disk/storage problems"
    echo "5) Service/application won't start"
    echo "6) Display/graphics issues"
    echo "7) Package management problems"
    echo "8) Permission/access issues"
    echo "9) Other/Custom diagnosis"
    echo "0) Exit"
    echo
    
    read -p "Enter your choice (0-9): " choice
    
    case $choice in
        1) diagnose_boot_issues ;;
        2) diagnose_performance_issues ;;
        3) diagnose_network_issues ;;
        4) diagnose_disk_issues ;;
        5) diagnose_service_issues ;;
        6) diagnose_display_issues ;;
        7) diagnose_package_issues ;;
        8) diagnose_permission_issues ;;
        9) custom_diagnosis ;;
        0) exit 0 ;;
        *) echo "Invalid choice"; interactive_diagnosis ;;
    esac
}

# Boot issues diagnosis
diagnose_boot_issues() {
    print_section "Boot Issues Diagnosis"
    
    # Check boot logs
    if command -v journalctl >/dev/null; then
        print_info "Checking boot logs..."
        boot_errors=$(journalctl -b --no-pager | grep -i "fail\|error\|fatal" | wc -l)
        if [[ $boot_errors -gt 0 ]]; then
            print_warning "$boot_errors potential boot issues found"
            echo "Recent boot errors:"
            journalctl -b --no-pager | grep -i "fail\|error\|fatal" | tail -5
        fi
    fi
    
    # Check systemd status
    if systemctl is-system-running | grep -q degraded; then
        print_error "System is in degraded state"
        offer_fix "systemctl --failed" "Show failed services"
    fi
    
    # Check filesystem integrity
    print_info "Checking filesystem status..."
    if mount | grep " / " | grep -q "ro,"; then
        print_error "Root filesystem is mounted read-only"
        offer_fix "mount -o remount,rw /" "Remount root filesystem as read-write"
    fi
}

# Performance issues diagnosis
diagnose_performance_issues() {
    print_section "Performance Issues Diagnosis"
    
    # Check CPU usage
    cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    if command -v bc >/dev/null && (( $(echo "$cpu_usage > 80" | bc -l) )); then
        print_warning "High CPU usage detected: ${cpu_usage}%"
        print_info "Top CPU processes:"
        ps aux --sort=-%cpu | head -5
    fi
    
    # Check memory pressure
    mem_total=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    mem_available=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
    mem_usage_percent=$(( (mem_total - mem_available) * 100 / mem_total ))
    
    if [[ $mem_usage_percent -gt 85 ]]; then
        print_warning "High memory usage detected"
        print_info "Top memory processes:"
        ps aux --sort=-%mem | head -5
        
        offer_fix "systemctl restart systemd-oomd" "Restart OOM daemon"
    fi
    
    # Check for swap thrashing
    if [[ -f /proc/vmstat ]]; then
        swap_in=$(grep pswpin /proc/vmstat | awk '{print $2}')
        swap_out=$(grep pswpout /proc/vmstat | awk '{print $2}')
        if [[ $swap_in -gt 1000 || $swap_out -gt 1000 ]]; then
            print_warning "High swap activity detected - possible thrashing"
        fi
    fi
    
    # Check load average
    load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    cpu_cores=$(nproc)
    if command -v bc >/dev/null && (( $(echo "$load_avg > $cpu_cores * 2" | bc -l) )); then
        print_warning "High system load: $load_avg (cores: $cpu_cores)"
    fi
}

# Network issues diagnosis
diagnose_network_issues() {
    print_section "Network Issues Diagnosis"
    
    # Check if networking service is running
    if ! systemctl is-active --quiet networking; then
        print_error "Networking service is not running"
        offer_fix "systemctl restart networking" "Restart networking service"
    fi
    
    # Check for interface issues
    down_interfaces=$(ip link show | grep -B1 "state DOWN" | grep -E '^[0-9]+:' | awk -F: '{print $2}' | tr -d ' ')
    if [[ -n "$down_interfaces" && "$down_interfaces" != "lo" ]]; then
        print_warning "Interfaces down: $down_interfaces"
        for iface in $down_interfaces; do
            if [[ "$iface" != "lo" ]]; then
                offer_fix "ip link set $iface up" "Bring up interface $iface"
            fi
        done
    fi
    
    # Check DNS resolution
    if [[ -f /etc/resolv.conf ]] && grep -q nameserver /etc/resolv.conf; then
        print_info "Testing DNS resolution..."
        if ! nslookup debian.org >/dev/null 2>&1; then
            print_warning "DNS resolution test failed"
            print_info "Consider checking /etc/resolv.conf"
        fi
    fi
}

# Disk issues diagnosis
diagnose_disk_issues() {
    print_section "Disk Issues Diagnosis"
    
    # Check for full filesystems
    full_filesystems=$(df -h | awk 'NR>1 {gsub(/%/, "", $5); if($5 > 95) print $6 " (" $5 "%)"}')
    if [[ -n "$full_filesystems" ]]; then
        print_error "Full filesystems detected:"
        echo "$full_filesystems"
        
        # Offer cleanup suggestions
        print_info "Cleanup suggestions:"
        echo "- Clear package cache: apt clean"
        echo "- Remove old kernels: apt autoremove"
        echo "- Check log files: find /var/log -name '*.log' -size +100M"
        echo "- Clear tmp files: find /tmp -type f -atime +7 -delete"
    fi
    
    # Check for I/O errors
    if dmesg | grep -i "i/o error\|disk.*error" | head -3 | grep -q .; then
        print_error "Disk I/O errors detected:"
        dmesg | grep -i "i/o error\|disk.*error" | tail -3
        print_info "Consider running fsck on affected filesystems"
    fi
}

# Service issues diagnosis
diagnose_service_issues() {
    print_section "Service Issues Diagnosis"
    
    read -p "Enter the service name you're having trouble with: " service_name
    
    if [[ -z "$service_name" ]]; then
        echo "No service name provided"
        return
    fi
    
    # Check if service exists
    if ! systemctl list-unit-files | grep -q "^$service_name"; then
        print_error "Service '$service_name' not found"
        return
    fi
    
    # Check service status
    if systemctl is-active --quiet "$service_name"; then
        print_ok "Service '$service_name' is running"
    else
        print_error "Service '$service_name' is not running"
        
        # Show status
        print_info "Service status:"
        systemctl status "$service_name" --no-pager -l
        
        offer_fix "systemctl start $service_name" "Start service $service_name"
        offer_fix "systemctl restart $service_name" "Restart service $service_name"
    fi
    
    # Check if enabled
    if systemctl is-enabled --quiet "$service_name"; then
        print_ok "Service '$service_name' is enabled"
    else
        print_warning "Service '$service_name' is not enabled"
        offer_fix "systemctl enable $service_name" "Enable service $service_name"
    fi
}

# Display issues diagnosis
diagnose_display_issues() {
    print_section "Display Issues Diagnosis"
    
    # Check if running in GUI
    if [[ -n "$DISPLAY" || -n "$WAYLAND_DISPLAY" ]]; then
        print_ok "Display environment detected"
        
        # Check graphics drivers
        if command -v lspci >/dev/null; then
            gpu_info=$(lspci | grep -i vga)
            print_info "Graphics hardware: $gpu_info"
        fi
        
        # Check X server logs
        if [[ -f /var/log/Xorg.0.log ]]; then
            x_errors=$(grep -i "error\|fatal" /var/log/Xorg.0.log | wc -l)
            if [[ $x_errors -gt 0 ]]; then
                print_warning "$x_errors X server errors found"
                print_info "Check /var/log/Xorg.0.log for details"
            fi
        fi
    else
        print_info "Running in console mode"
        print_info "To start GUI: systemctl start display-manager"
    fi
}

# Package issues diagnosis
diagnose_package_issues() {
    print_section "Package Issues Diagnosis"
    
    # Check for locks
    if [[ -f /var/lib/dpkg/lock-frontend || -f /var/lib/dpkg/lock ]]; then
        print_error "Package system is locked"
        offer_fix "rm -f /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock" "Remove package locks"
    fi
    
    # Check for interrupted installations
    if dpkg -C 2>/dev/null | grep -q .; then
        print_error "Interrupted package installations detected"
        offer_fix "dpkg --configure -a" "Configure interrupted packages"
    fi
    
    # Check for broken dependencies
    if apt-get check 2>&1 | grep -q "broken dependencies"; then
        print_error "Broken dependencies detected"
        offer_fix "apt-get install -f" "Fix broken dependencies"
    fi
}

# Permission issues diagnosis
diagnose_permission_issues() {
    print_section "Permission Issues Diagnosis"
    
    read -p "Enter the file or directory path with permission issues: " file_path
    
    if [[ -z "$file_path" ]]; then
        echo "No path provided"
        return
    fi
    
    if [[ ! -e "$file_path" ]]; then
        print_error "Path '$file_path' does not exist"
        return
    fi
    
    # Show current permissions
    perms=$(ls -ld "$file_path")
    print_info "Current permissions: $perms"
    
    # Check ownership
    owner=$(stat -c "%U:%G" "$file_path")
    print_info "Owner: $owner"
    
    # Suggest fixes
    echo "Common permission fixes:"
    echo "1) Make readable by all: chmod +r '$file_path'"
    echo "2) Make executable: chmod +x '$file_path'"
    echo "3) Change ownership: chown user:group '$file_path'"
    echo "4) Reset to standard: chmod 644 '$file_path' (files) or chmod 755 '$file_path' (dirs)"
}

# Custom diagnosis
custom_diagnosis() {
    print_section "Custom Diagnosis"
    
    echo "Please describe the issue you're experiencing:"
    read -p "> " issue_description
    
    print_info "Based on your description: '$issue_description'"
    print_info "Here are some general troubleshooting steps:"
    echo
    echo "1. Check system logs: journalctl -f"
    echo "2. Check running processes: ps aux"
    echo "3. Check network: ip addr show"
    echo "4. Check disk space: df -h"
    echo "5. Check memory: free -h"
    echo "6. Check services: systemctl --failed"
    echo
    echo "For more specific help, please provide more details about the issue."
}

# Offer fix function
offer_fix() {
    local command="$1"
    local description="$2"
    
    echo
    print_info "SUGGESTED FIX: $description"
    print_info "Command: $command"
    echo
    
    read -p "Do you want to run this fix? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Executing: $command"
        
        if eval "$command"; then
            print_ok "Fix applied successfully"
        else
            print_error "Fix failed. You may need to:"
            echo "  - Run as root/sudo"
            echo "  - Check the command syntax"
            echo "  - Investigate the specific error"
        fi
    else
        print_info "Fix skipped"
    fi
    
    echo
}

# Generate summary report
generate_summary() {
    print_section "Diagnosis Summary"
    
    echo "Scan completed at $(date)"
    echo "Log file: $LOG_FILE"
    echo
    
    if [[ ${#ISSUES_FOUND[@]} -gt 0 ]]; then
        print_error "CRITICAL ISSUES FOUND (${#ISSUES_FOUND[@]}):"
        for issue in "${ISSUES_FOUND[@]}"; do
            echo "  - $issue"
        done
        echo
    fi
    
    if [[ ${#WARNINGS_FOUND[@]} -gt 0 ]]; then
        print_warning "WARNINGS (${#WARNINGS_FOUND[@]}):"
        for warning in "${WARNINGS_FOUND[@]}"; do
            echo "  - $warning"
        done
        echo
    fi
    
    if [[ ${#ISSUES_FOUND[@]} -eq 0 && ${#WARNINGS_FOUND[@]} -eq 0 ]]; then
        print_ok "No critical issues detected!"
        print_info "System appears to be healthy"
    fi
    
    echo
    print_info "For additional help:"
    echo "  - Review the log file: $LOG_FILE"
    echo "  - Check Debian documentation"
    echo "  - Use interactive diagnosis for specific issues"
}

# Main function
main() {
    # Initialize log with proper permissions handling
    # Remove existing log if we can write to it, or create new one
    if [[ -w "$LOG_FILE" ]] || [[ ! -f "$LOG_FILE" ]]; then
        echo "Debian Doctor started at $(date)" > "$LOG_FILE" 2>/dev/null || {
            # If we can't write to the intended log file, use a backup location
            LOG_FILE="/tmp/debian_doctor_$(date +%s)_$$.log"
            echo "Debian Doctor started at $(date)" > "$LOG_FILE"
            print_warning "Using alternate log file due to permissions: $LOG_FILE"
        }
    else
        # Log file exists but not writable, create new one with timestamp
        LOG_FILE="/tmp/debian_doctor_$(date +%s)_$$.log"
        echo "Debian Doctor started at $(date)" > "$LOG_FILE"
        print_warning "Created new log file due to permission conflict: $LOG_FILE"
    fi
    
    print_header
    
    # Check if we're on a Debian system
    if [[ ! -f /etc/debian_version ]]; then
        print_error "This script is designed for Debian systems"
        exit 1
    fi
    
    IS_ROOT=false
    if check_root; then
        IS_ROOT=true
    fi
    
    echo "Starting comprehensive system check..."
    echo
    
    # Run all diagnostic checks
    check_system_info
    check_disk_space
    check_memory
    check_filesystem
    
    if [[ $IS_ROOT == true ]]; then
        check_services
        check_logs
        check_packages
    else
        print_info "Skipping root-only checks (services, logs, packages)"
        echo
    fi
    
    check_network
    
    # Generate summary
    generate_summary
    
    # Offer interactive diagnosis
    echo "Would you like to run interactive diagnosis for specific issues?"
    read -p "(y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        interactive_diagnosis
    fi
    
    print_info "Debian Doctor completed. Log saved to: $LOG_FILE"
}

# Trap to clean up on exit
trap 'echo "Scan interrupted"; exit 1' INT TERM

# Run main function
main "$@"
