package diagnose

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseServiceIssues diagnoses service-related problems and provides fixes
func DiagnoseServiceIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Service Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check for failed services
	failedServices := checkFailedSystemdServices()
	if len(failedServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Failed services detected:")
		for _, service := range failedServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "restart_failed_services",
			Title:       "Restart Failed Services",
			Description: "Attempt to restart all failed services",
			Commands:    []string{"systemctl restart " + strings.Join(failedServices, " ")},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"systemctl stop " + strings.Join(failedServices, " ")},
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_service_logs",
			Title:       "Check Service Logs",
			Description: "Examine logs for failed services to understand issues",
			Commands:    generateServiceLogCommands(failedServices),
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for services in error state
	errorServices := checkServicesInErrorState()
	if len(errorServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Services in error state:")
		for _, service := range errorServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "reset_error_services",
			Title:       "Reset Services in Error State",
			Description: "Reset failed service states and attempt restart",
			Commands: []string{
				"systemctl reset-failed",
				"systemctl restart " + strings.Join(errorServices, " "),
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Check for disabled critical services
	criticalServices := checkCriticalServices()
	if len(criticalServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Critical services that are disabled:")
		for _, service := range criticalServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "enable_critical_services",
			Title:       "Enable Critical Services",
			Description: "Enable and start essential system services",
			Commands:    generateEnableServiceCommands(criticalServices),
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: generateDisableServiceCommands(criticalServices),
			RiskLevel:   fixes.RiskHigh,
		})
	}

	// Check for services with high restart rates
	flappingServices := checkFlappingServices()
	if len(flappingServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Services with high restart rates (potentially flapping):")
		for _, service := range flappingServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "analyze_flapping_services",
			Title:       "Analyze Flapping Services",
			Description: "Examine services that are restarting frequently",
			Commands:    generateFlappingAnalysisCommands(flappingServices),
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "stop_flapping_services",
			Title:       "Stop Flapping Services",
			Description: "Temporarily stop services that are restarting frequently",
			Commands:    []string{"systemctl stop " + strings.Join(flappingServices, " ")},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"systemctl start " + strings.Join(flappingServices, " ")},
			RiskLevel:   fixes.RiskHigh,
		})
	}

	// Check for masked services
	maskedServices := checkMaskedServices()
	if len(maskedServices) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Masked services that may need attention:")
		for _, service := range maskedServices {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", service))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "unmask_services",
			Title:       "Unmask Services",
			Description: "Unmask services that may have been accidentally masked",
			Commands:    []string{"systemctl unmask " + strings.Join(maskedServices, " ")},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"systemctl mask " + strings.Join(maskedServices, " ")},
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Check for dependency issues
	dependencyIssues := checkServiceDependencies()
	if len(dependencyIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Service dependency issues:")
		for _, issue := range dependencyIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "reload_systemd_daemon",
			Title:       "Reload Systemd Configuration",
			Description: "Reload systemd daemon to refresh service dependencies",
			Commands: []string{
				"systemctl daemon-reload",
				"systemctl restart systemd-logind",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Always add general service management fixes
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "service_overview",
		Title:       "Service Overview",
		Description: "Display comprehensive service status information",
		Commands: []string{
			"systemctl list-units --type=service --state=failed",
			"systemctl list-units --type=service --state=active",
			"systemctl list-unit-files --type=service | grep disabled",
			"systemctl status",
		},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "check_service_conflicts",
		Title:       "Check Service Conflicts",
		Description: "Identify conflicting services and dependency loops",
		Commands: []string{
			"systemctl list-dependencies --reverse --all",
			"systemctl list-dependencies --before --all",
			"systemd-analyze verify",
		},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No significant service issues detected")
	}

	return diagnosis
}

// checkFailedSystemdServices finds services in failed state
func checkFailedSystemdServices() []string {
	failed := []string{}

	cmd := exec.Command("systemctl", "list-units", "--failed", "--type=service", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return failed
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 && strings.HasSuffix(fields[0], ".service") {
			serviceName := strings.TrimSuffix(fields[0], ".service")
			failed = append(failed, serviceName)
		}
	}

	return failed
}

// checkServicesInErrorState finds services in error/activating state
func checkServicesInErrorState() []string {
	errorServices := []string{}

	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=activating,deactivating", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return errorServices
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 && strings.HasSuffix(fields[0], ".service") {
			serviceName := strings.TrimSuffix(fields[0], ".service")
			errorServices = append(errorServices, serviceName)
		}
	}

	return errorServices
}

// checkCriticalServices finds disabled critical services
func checkCriticalServices() []string {
	disabled := []string{}

	criticalServicesList := []string{
		"networking", "systemd-networkd", "NetworkManager",
		"ssh", "sshd", "systemd-logind", "dbus",
		"systemd-resolved", "systemd-timesyncd",
	}

	for _, service := range criticalServicesList {
		cmd := exec.Command("systemctl", "is-enabled", service)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		status := strings.TrimSpace(string(output))
		if status == "disabled" || status == "masked" {
			// Double-check if service exists
			checkCmd := exec.Command("systemctl", "status", service)
			if checkCmd.Run() == nil {
				disabled = append(disabled, service)
			}
		}
	}

	return disabled
}

// checkFlappingServices finds services restarting frequently
func checkFlappingServices() []string {
	flapping := []string{}

	cmd := exec.Command("journalctl", "--since", "1 hour ago", "--grep", "Started\\|Stopped", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return flapping
	}

	// Count service start/stop events
	serviceEvents := make(map[string]int)
	lines := strings.Split(string(output), "\n")

	serviceRegex := regexp.MustCompile(`(Started|Stopped) (.+)\.service`)
	for _, line := range lines {
		matches := serviceRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			serviceName := matches[2]
			serviceEvents[serviceName]++
		}
	}

	// Services with more than 6 events in the last hour are considered flapping
	for service, count := range serviceEvents {
		if count > 6 {
			flapping = append(flapping, service)
		}
	}

	return flapping
}

// checkMaskedServices finds masked services
func checkMaskedServices() []string {
	masked := []string{}

	cmd := exec.Command("systemctl", "list-unit-files", "--type=service", "--state=masked", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return masked
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 && strings.HasSuffix(fields[0], ".service") {
			serviceName := strings.TrimSuffix(fields[0], ".service")
			masked = append(masked, serviceName)
		}
	}

	return masked
}

// checkServiceDependencies finds dependency issues
func checkServiceDependencies() []string {
	issues := []string{}

	// Check for circular dependencies
	cmd := exec.Command("systemd-analyze", "verify")
	output, err := cmd.CombinedOutput()
	if err != nil {
		content := string(output)
		if strings.Contains(content, "circular") || strings.Contains(content, "dependency") {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && (strings.Contains(line, "circular") || strings.Contains(line, "dependency")) {
					issues = append(issues, line)
				}
			}
		}
	}

	return issues
}

// Helper functions for generating commands

func generateServiceLogCommands(services []string) []string {
	commands := []string{}
	for _, service := range services {
		commands = append(commands, fmt.Sprintf("journalctl -u %s --since '1 hour ago' --no-pager", service))
	}
	return commands
}

func generateEnableServiceCommands(services []string) []string {
	commands := []string{}
	for _, service := range services {
		commands = append(commands, fmt.Sprintf("systemctl enable %s", service))
		commands = append(commands, fmt.Sprintf("systemctl start %s", service))
	}
	return commands
}

func generateDisableServiceCommands(services []string) []string {
	commands := []string{}
	for _, service := range services {
		commands = append(commands, fmt.Sprintf("systemctl stop %s", service))
		commands = append(commands, fmt.Sprintf("systemctl disable %s", service))
	}
	return commands
}

func generateFlappingAnalysisCommands(services []string) []string {
	commands := []string{}
	for _, service := range services {
		commands = append(commands, fmt.Sprintf("systemctl status %s", service))
		commands = append(commands, fmt.Sprintf("journalctl -u %s --since '2 hours ago' | grep -E '(Started|Stopped|Failed)' | tail -10", service))
	}
	return commands
}

func removeDuplicateServiceStrings(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}