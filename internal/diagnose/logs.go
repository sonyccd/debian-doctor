package diagnose

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseLogIssues diagnoses system log-related problems and provides fixes
func DiagnoseLogIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "System Log Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check journal disk usage
	journalSize := checkJournalSize()
	if journalSize > 1000 { // More than 1GB
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("systemd journal is using %.1f MB of disk space", journalSize))
		
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "vacuum_journal_time",
			Title:       "Clean Old Journal Entries (30 days)",
			Description: "Remove journal entries older than 30 days to free disk space",
			Commands:    []string{"journalctl --vacuum-time=30d"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "vacuum_journal_size",
			Title:       "Limit Journal Size (500MB)",
			Description: "Limit systemd journal to 500MB total size",
			Commands:    []string{"journalctl --vacuum-size=500M"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for persistent errors
	persistentErrors := checkPersistentErrors()
	if len(persistentErrors) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Found %d persistent error patterns in logs", len(persistentErrors)))
		
		for i, errPattern := range persistentErrors {
			if i < 3 { // Show first 3 as examples
				diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", errPattern))
			}
		}
		if len(persistentErrors) > 3 {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("  ... and %d more error patterns", len(persistentErrors)-3))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "analyze_errors",
			Title:       "Analyze Recent Errors",
			Description: "Display recent error messages for detailed analysis",
			Commands:    []string{"journalctl -p err --since '24 hours ago' --no-pager"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for log rotation issues
	logRotationIssues := checkLogRotation()
	if len(logRotationIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Log rotation issues detected:")
		for _, issue := range logRotationIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "force_logrotate",
			Title:       "Force Log Rotation",
			Description: "Force immediate log rotation for all configured logs",
			Commands:    []string{"logrotate -f /etc/logrotate.conf"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_logrotate_config",
			Title:       "Check Logrotate Configuration",
			Description: "Test logrotate configuration for syntax errors",
			Commands:    []string{"logrotate -d /etc/logrotate.conf"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for failed services based on logs
	failedServices := checkFailedServices()
	if len(failedServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Services with errors detected:")
		for _, service := range failedServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "restart_failed_services",
			Title:       "Restart Failed Services",
			Description: "Attempt to restart all currently failed services",
			Commands:    []string{"systemctl --failed --no-legend | awk '{print $1}' | xargs -r systemctl restart"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "show_service_status",
			Title:       "Show Failed Service Details",
			Description: "Display detailed status of all failed services",
			Commands:    []string{"systemctl --failed", "systemctl status --failed"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for core dumps
	coreDumps := checkCoreDumps()
	if coreDumps > 0 {
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Found %d core dumps on system", coreDumps))
		
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "list_core_dumps",
			Title:       "List Core Dumps",
			Description: "Show all core dumps with details",
			Commands:    []string{"coredumpctl list"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "clean_core_dumps",
			Title:       "Clean Old Core Dumps",
			Description: "Remove core dumps older than 3 days",
			Commands:    []string{"coredumpctl --vacuum-time=3d"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for kernel messages
	kernelIssues := checkKernelIssues()
	if len(kernelIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Kernel issues detected:")
		for _, issue := range kernelIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "show_kernel_messages",
			Title:       "Show Recent Kernel Messages",
			Description: "Display recent kernel messages and errors",
			Commands:    []string{"dmesg | tail -50", "journalctl -k --since '24 hours ago' --no-pager"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Always add general log analysis fixes
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "show_system_overview",
		Title:       "System Overview",
		Description: "Show comprehensive system status and recent important events",
		Commands:    []string{
			"systemctl status",
			"journalctl --since '1 hour ago' -p warning --no-pager",
		},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No significant log issues detected")
	}

	return diagnosis
}

// checkJournalSize returns journal size in MB
func checkJournalSize() float64 {
	cmd := exec.Command("journalctl", "--disk-usage")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	content := string(output)
	re := regexp.MustCompile(`take up ([0-9.]+)([KMGT]?)B`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 3 {
		size, _ := strconv.ParseFloat(matches[1], 64)
		unit := matches[2]

		// Convert to MB
		switch unit {
		case "G":
			return size * 1024
		case "T":
			return size * 1024 * 1024
		case "K":
			return size / 1024
		default: // B or MB
			return size
		}
	}

	return 0
}

// checkPersistentErrors looks for repeated error patterns
func checkPersistentErrors() []string {
	errors := []string{}

	cmd := exec.Command("journalctl", "-p", "err", "--since", "24 hours ago", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return errors
	}

	lines := strings.Split(string(output), "\n")
	errorCounts := make(map[string]int)

	// Count error patterns
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract the error message part (remove timestamp and hostname)
		parts := strings.SplitN(line, " ", 6)
		if len(parts) >= 6 {
			errorMsg := parts[5]
			// Normalize similar errors
			normalized := normalizeErrorMessage(errorMsg)
			errorCounts[normalized]++
		}
	}

	// Find patterns that appear more than 3 times
	for errorMsg, count := range errorCounts {
		if count > 3 {
			errors = append(errors, fmt.Sprintf("%s (occurred %d times)", errorMsg, count))
		}
	}

	return errors
}

// checkLogRotation checks for log rotation issues
func checkLogRotation() []string {
	issues := []string{}

	// Check logrotate status
	cmd := exec.Command("logrotate", "-d", "/etc/logrotate.conf")
	output, err := cmd.Output()
	if err != nil {
		issues = append(issues, "Logrotate configuration test failed")
	} else {
		content := strings.ToLower(string(output))
		if strings.Contains(content, "error") {
			issues = append(issues, "Logrotate configuration contains errors")
		}
	}

	// Check for large unrotated logs
	logFiles := []string{
		"/var/log/syslog",
		"/var/log/auth.log",
		"/var/log/kern.log",
		"/var/log/daemon.log",
	}

	for _, logFile := range logFiles {
		cmd := exec.Command("stat", "-c", "%s", logFile)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		sizeStr := strings.TrimSpace(string(output))
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			continue
		}

		// Flag files larger than 50MB as potentially unrotated
		if size > 50*1024*1024 {
			sizeMB := float64(size) / (1024 * 1024)
			issues = append(issues, fmt.Sprintf("%s is %.1f MB (may need rotation)", logFile, sizeMB))
		}
	}

	return issues
}

// checkFailedServices returns services that have recently failed
func checkFailedServices() []string {
	services := []string{}

	cmd := exec.Command("systemctl", "--failed", "--no-legend", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return services
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			services = append(services, fields[0])
		}
	}

	return services
}

// checkCoreDumps counts core dumps
func checkCoreDumps() int {
	cmd := exec.Command("coredumpctl", "list", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

// checkKernelIssues looks for kernel-related problems
func checkKernelIssues() []string {
	issues := []string{}

	cmd := exec.Command("dmesg")
	output, err := cmd.Output()
	if err != nil {
		return issues
	}

	content := strings.ToLower(string(output))
	kernelPatterns := []string{
		"kernel panic",
		"oops:",
		"call trace:",
		"segfault",
		"general protection fault",
		"hardware error",
		"mce: machine check events",
	}

	for _, pattern := range kernelPatterns {
		if strings.Contains(content, pattern) {
			issues = append(issues, fmt.Sprintf("Detected: %s", pattern))
		}
	}

	return issues
}

// normalizeErrorMessage normalizes error messages for pattern matching
func normalizeErrorMessage(msg string) string {
	// Remove timestamps, PIDs, and other variable parts
	msg = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`).ReplaceAllString(msg, "[TIME]")
	msg = regexp.MustCompile(`\[\d+\]`).ReplaceAllString(msg, "[PID]")
	msg = regexp.MustCompile(`pid \d+`).ReplaceAllString(msg, "pid [NUM]")
	msg = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`).ReplaceAllString(msg, "[IP]")
	msg = regexp.MustCompile(`/dev/\w+\d+`).ReplaceAllString(msg, "[DEVICE]")
	
	// Limit length for grouping
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}

	return msg
}