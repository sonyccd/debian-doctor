package checks

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ServicesCheck checks critical system services
type ServicesCheck struct{}

func (s ServicesCheck) Name() string {
	return "System Services"
}

func (s ServicesCheck) RequiresRoot() bool {
	return true
}

func (s ServicesCheck) Run() CheckResult {
	result := CheckResult{
		Name:      s.Name(),
		Severity:  SeverityInfo,
		Timestamp: time.Now(),
		Details:   []string{},
	}

	// Check if systemctl is available
	if _, err := exec.LookPath("systemctl"); err != nil {
		result.Severity = SeverityWarning
		result.Message = "systemctl not found - cannot check services"
		return result
	}

	// List of critical services to check
	criticalServices := []string{
		"systemd-logind",
		"dbus",
		"networking",
		"ssh",
		"cron",
	}

	failedServices := []string{}
	for _, service := range criticalServices {
		status := checkServiceStatus(service)
		result.Details = append(result.Details, status)
		
		if strings.Contains(status, "not running") || strings.Contains(status, "failed") {
			failedServices = append(failedServices, service)
		}
	}

	// Check for any failed services
	cmd := exec.Command("systemctl", "--failed", "--no-legend", "--no-pager")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line != "" {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					failedServices = append(failedServices, parts[0])
				}
			}
		}
	}

	// Set result based on findings
	if len(failedServices) > 0 {
		result.Severity = SeverityError
		result.Message = fmt.Sprintf("%d failed services detected", len(failedServices))
		result.Details = append(result.Details, fmt.Sprintf("Failed: %s", strings.Join(failedServices, ", ")))
	} else {
		result.Message = "All critical services are running"
	}

	return result
}

func checkServiceStatus(service string) string {
	cmd := exec.Command("systemctl", "is-active", service)
	var out bytes.Buffer
	cmd.Stdout = &out
	
	err := cmd.Run()
	status := strings.TrimSpace(out.String())
	
	if err != nil || status != "active" {
		// Try to get more info
		cmd = exec.Command("systemctl", "status", service, "--no-pager", "-n", "0")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				for _, line := range lines {
					if strings.Contains(line, "Active:") {
						return fmt.Sprintf("%s: %s", service, strings.TrimSpace(line))
					}
				}
			}
		}
		return fmt.Sprintf("%s is not running", service)
	}
	
	return fmt.Sprintf("%s is running", service)
}