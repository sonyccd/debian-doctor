package checks

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogsCheck checks system logs for errors and issues
type LogsCheck struct{}

func (c LogsCheck) Name() string {
	return "System Logs"
}

func (c LogsCheck) RequiresRoot() bool {
	return false // Most log viewing doesn't require root
}

func (c LogsCheck) Run() CheckResult {
	result := CheckResult{
		Name:      c.Name(),
		Severity:  SeverityInfo,
		Message:   "System logs analysis completed",
		Details:   []string{},
		Timestamp: time.Now(),
	}

	// Check systemd journal for recent errors
	journalErrors := c.checkJournalErrors()
	if len(journalErrors) > 0 {
		result.Severity = SeverityWarning
		result.Message = "Errors found in system journal"
		result.Details = append(result.Details, fmt.Sprintf("Recent journal errors: %d", len(journalErrors)))
		
		// Show first few errors as examples
		for i, err := range journalErrors {
			if i >= 3 { // Limit to first 3 errors
				result.Details = append(result.Details, fmt.Sprintf("... and %d more errors", len(journalErrors)-3))
				break
			}
			result.Details = append(result.Details, fmt.Sprintf("  - %s", err))
		}
	}

	// Check for authentication failures
	authFailures := c.checkAuthFailures()
	if authFailures > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Authentication failures detected"
		}
		result.Details = append(result.Details, fmt.Sprintf("Recent auth failures: %d", authFailures))
	}

	// Check for disk errors
	diskErrors := c.checkDiskErrors()
	if len(diskErrors) > 0 {
		result.Severity = SeverityCritical
		result.Message = "Disk errors detected in logs"
		result.Details = append(result.Details, "Disk errors found:")
		for _, err := range diskErrors {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", err))
		}
	}

	// Check for memory issues
	memoryIssues := c.checkMemoryIssues()
	if len(memoryIssues) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Memory issues detected"
		}
		result.Details = append(result.Details, "Memory issues:")
		for _, issue := range memoryIssues {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", issue))
		}
	}

	// Check for service failures
	serviceFailures := c.checkServiceFailures()
	if len(serviceFailures) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Service failures detected"
		}
		result.Details = append(result.Details, "Failed services:")
		for _, service := range serviceFailures {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", service))
		}
	}

	// Check log file sizes
	logSizes := c.checkLogSizes()
	if len(logSizes) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Large log files detected"
		}
		result.Details = append(result.Details, "Large log files:")
		for _, logInfo := range logSizes {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", logInfo))
		}
	}

	if result.Severity == SeverityInfo {
		result.Details = append(result.Details, "No significant issues found in system logs")
	}

	return result
}

// checkJournalErrors checks systemd journal for recent errors
func (c LogsCheck) checkJournalErrors() []string {
	errors := []string{}

	// Get errors from the last 24 hours
	cmd := exec.Command("journalctl", "--since", "24 hours ago", "-p", "err", "--no-pager", "-n", "20")
	output, err := cmd.Output()
	if err != nil {
		return errors
	}

	lines := strings.Split(string(output), "\n")
	errorPattern := regexp.MustCompile(`(\w+\s+\d+\s+\d+:\d+:\d+)\s+\S+\s+(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := errorPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			// Extract timestamp and message
			timestamp := matches[1]
			message := matches[2]
			
			// Filter out common non-critical errors
			if c.isSignificantError(message) {
				errors = append(errors, fmt.Sprintf("%s: %s", timestamp, message))
			}
		}
	}

	return errors
}

// checkAuthFailures counts recent authentication failures
func (c LogsCheck) checkAuthFailures() int {
	cmd := exec.Command("journalctl", "--since", "24 hours ago", "-u", "ssh", "-u", "systemd-logind", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	failurePatterns := []string{
		"Failed password",
		"authentication failure",
		"Invalid user",
		"Connection closed by authenticating user",
		"PAM authentication failed",
	}

	failures := 0
	content := string(output)
	for _, pattern := range failurePatterns {
		failures += strings.Count(strings.ToLower(content), strings.ToLower(pattern))
	}

	return failures
}

// checkDiskErrors looks for disk-related errors in logs
func (c LogsCheck) checkDiskErrors() []string {
	errors := []string{}

	cmd := exec.Command("journalctl", "--since", "7 days ago", "-p", "err", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return errors
	}

	diskErrorPatterns := []string{
		"i/o error",
		"disk error",
		"ata error",
		"scsi error",
		"read error",
		"write error",
		"bad sector",
		"medium error",
		"critical medium error",
		"sense key: medium error",
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, pattern := range diskErrorPatterns {
			if strings.Contains(lineLower, pattern) {
				// Extract relevant part of the error message
				if len(line) > 200 {
					line = line[:200] + "..."
				}
				errors = append(errors, strings.TrimSpace(line))
				break
			}
		}
	}

	return errors
}

// checkMemoryIssues looks for memory-related problems
func (c LogsCheck) checkMemoryIssues() []string {
	issues := []string{}

	cmd := exec.Command("journalctl", "--since", "24 hours ago", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return issues
	}

	memoryPatterns := []string{
		"out of memory",
		"oom killer",
		"memory allocation failed",
		"cannot allocate memory",
		"killed process",
		"memory pressure",
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, pattern := range memoryPatterns {
			if strings.Contains(lineLower, pattern) {
				if len(line) > 150 {
					line = line[:150] + "..."
				}
				issues = append(issues, strings.TrimSpace(line))
				break
			}
		}
	}

	return issues
}

// checkServiceFailures looks for failed systemd services
func (c LogsCheck) checkServiceFailures() []string {
	failures := []string{}

	cmd := exec.Command("systemctl", "--failed", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return failures
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			serviceName := fields[0]
			if serviceName != "" {
				failures = append(failures, serviceName)
			}
		}
	}

	return failures
}

// checkLogSizes checks for excessively large log files
func (c LogsCheck) checkLogSizes() []string {
	largeLogs := []string{}

	// Check journal size
	cmd := exec.Command("journalctl", "--disk-usage")
	output, err := cmd.Output()
	if err == nil {
		content := string(output)
		if strings.Contains(content, "Archived and active journals take up") {
			// Extract size information
			re := regexp.MustCompile(`take up ([0-9.]+)([KMGT]?)B`)
			matches := re.FindStringSubmatch(content)
			if len(matches) >= 3 {
				size, _ := strconv.ParseFloat(matches[1], 64)
				unit := matches[2]
				
				// Convert to MB for comparison
				sizeMB := size
				switch unit {
				case "G":
					sizeMB *= 1024
				case "T":
					sizeMB *= 1024 * 1024
				case "K":
					sizeMB /= 1024
				}

				if sizeMB > 1000 { // More than 1GB
					largeLogs = append(largeLogs, fmt.Sprintf("systemd journal: %.1f%sB", size, unit))
				}
			}
		}
	}

	// Check common log files
	logFiles := []string{
		"/var/log/syslog",
		"/var/log/messages",
		"/var/log/kern.log",
		"/var/log/auth.log",
		"/var/log/apache2/error.log",
		"/var/log/apache2/access.log",
		"/var/log/nginx/error.log",
		"/var/log/nginx/access.log",
	}

	for _, logFile := range logFiles {
		cmd := exec.Command("stat", "-c", "%s", logFile)
		output, err := cmd.Output()
		if err != nil {
			continue // File doesn't exist or can't be accessed
		}

		sizeStr := strings.TrimSpace(string(output))
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			continue
		}

		// Flag files larger than 100MB
		if size > 100*1024*1024 {
			sizeMB := float64(size) / (1024 * 1024)
			largeLogs = append(largeLogs, fmt.Sprintf("%s: %.1f MB", logFile, sizeMB))
		}
	}

	return largeLogs
}

// isSignificantError filters out common non-critical error messages
func (c LogsCheck) isSignificantError(message string) bool {
	message = strings.ToLower(message)
	
	// Filter out common, usually non-critical errors
	ignoredPatterns := []string{
		"connection reset by peer",
		"broken pipe",
		"no route to host",
		"network is unreachable",
		"temporary failure in name resolution",
		"device busy",
		"resource temporarily unavailable",
	}

	for _, pattern := range ignoredPatterns {
		if strings.Contains(message, pattern) {
			return false
		}
	}

	return true
}