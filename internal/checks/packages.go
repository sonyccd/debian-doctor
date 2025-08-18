package checks

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PackagesCheck checks the APT package system for issues
type PackagesCheck struct{}

func (c PackagesCheck) Name() string {
	return "Package System"
}

func (c PackagesCheck) RequiresRoot() bool {
	return false // Basic package checks don't require root
}

func (c PackagesCheck) Run() CheckResult {
	result := CheckResult{
		Name:      c.Name(),
		Severity:  SeverityInfo,
		Message:   "Package system analysis completed",
		Details:   []string{},
		Timestamp: time.Now(),
	}

	// Check for broken packages
	brokenPackages := c.checkBrokenPackages()
	if len(brokenPackages) > 0 {
		result.Severity = SeverityError
		result.Message = "Broken packages detected"
		result.Details = append(result.Details, fmt.Sprintf("Broken packages found: %d", len(brokenPackages)))
		for i, pkg := range brokenPackages {
			if i >= 5 { // Limit to first 5
				result.Details = append(result.Details, fmt.Sprintf("... and %d more", len(brokenPackages)-5))
				break
			}
			result.Details = append(result.Details, fmt.Sprintf("  - %s", pkg))
		}
	}

	// Check for held packages
	heldPackages := c.checkHeldPackages()
	if len(heldPackages) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Held packages detected"
		}
		result.Details = append(result.Details, fmt.Sprintf("Held packages: %d", len(heldPackages)))
		for i, pkg := range heldPackages {
			if i >= 3 {
				result.Details = append(result.Details, fmt.Sprintf("... and %d more", len(heldPackages)-3))
				break
			}
			result.Details = append(result.Details, fmt.Sprintf("  - %s", pkg))
		}
	}

	// Check for upgradeable packages
	upgradeableCount := c.checkUpgradeablePackages()
	if upgradeableCount > 0 {
		result.Details = append(result.Details, fmt.Sprintf("Packages available for upgrade: %d", upgradeableCount))
		if upgradeableCount > 50 {
			if result.Severity < SeverityWarning {
				result.Severity = SeverityWarning
				result.Message = "Many packages need upgrading"
			}
		}
	}

	// Check for autoremovable packages
	autoremovableCount := c.checkAutoremovablePackages()
	if autoremovableCount > 0 {
		result.Details = append(result.Details, fmt.Sprintf("Autoremovable packages: %d", autoremovableCount))
		if autoremovableCount > 20 {
			if result.Severity < SeverityWarning {
				result.Severity = SeverityWarning
				result.Message = "Many orphaned packages detected"
			}
		}
	}

	// Check APT sources validity
	invalidSources := c.checkAPTSources()
	if len(invalidSources) > 0 {
		result.Severity = SeverityError
		result.Message = "Invalid APT sources detected"
		result.Details = append(result.Details, "Invalid sources:")
		for _, source := range invalidSources {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", source))
		}
	}

	// Check for dpkg interruptions
	if c.checkDpkgInterrupted() {
		result.Severity = SeverityError
		result.Message = "Package installation was interrupted"
		result.Details = append(result.Details, "dpkg was interrupted - packages may be in inconsistent state")
	}

	// Check package cache size
	cacheSize := c.checkPackageCacheSize()
	if cacheSize > 1000 { // More than 1GB
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Large package cache detected"
		}
		result.Details = append(result.Details, fmt.Sprintf("Package cache size: %.1f MB", cacheSize))
	}

	// Check for unattended upgrades status
	unattendedStatus := c.checkUnattendedUpgrades()
	result.Details = append(result.Details, fmt.Sprintf("Unattended upgrades: %s", unattendedStatus))

	if result.Severity == SeverityInfo {
		result.Details = append(result.Details, "Package system appears healthy")
	}

	return result
}

// checkBrokenPackages finds packages in broken state
func (c PackagesCheck) checkBrokenPackages() []string {
	broken := []string{}

	// Check with dpkg
	cmd := exec.Command("dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return broken
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "iU") || strings.HasPrefix(line, "iF") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				broken = append(broken, fields[1])
			}
		}
	}

	// Also check with apt
	cmd = exec.Command("apt", "list", "--broken")
	output, err = cmd.Output()
	if err == nil {
		lines = strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "/") && !strings.HasPrefix(line, "WARNING") {
				parts := strings.Split(line, "/")
				if len(parts) > 0 {
					pkgName := strings.TrimSpace(parts[0])
					if pkgName != "" && pkgName != "Listing..." {
						broken = append(broken, pkgName)
					}
				}
			}
		}
	}

	return removeDuplicates(broken)
}

// checkHeldPackages finds packages on hold
func (c PackagesCheck) checkHeldPackages() []string {
	held := []string{}

	cmd := exec.Command("apt-mark", "showhold")
	output, err := cmd.Output()
	if err != nil {
		return held
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			held = append(held, line)
		}
	}

	return held
}

// checkUpgradeablePackages counts packages that can be upgraded
func (c PackagesCheck) checkUpgradeablePackages() int {
	cmd := exec.Command("apt", "list", "--upgradable")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "[upgradable from:") {
			count++
		}
	}

	return count
}

// checkAutoremovablePackages counts packages that can be autoremoved
func (c PackagesCheck) checkAutoremovablePackages() int {
	cmd := exec.Command("apt", "autoremove", "--dry-run")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	content := string(output)
	// Look for the line that says "The following packages will be REMOVED:"
	re := regexp.MustCompile(`The following packages will be REMOVED:\s*\n(.*?)(\n\n|\nNeed|\nAfter|$)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		packages := strings.Fields(matches[1])
		return len(packages)
	}

	return 0
}

// checkAPTSources validates APT source lists
func (c PackagesCheck) checkAPTSources() []string {
	invalid := []string{}

	cmd := exec.Command("apt", "update", "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		content := string(output)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Failed to fetch") ||
				strings.Contains(line, "404  Not Found") ||
				strings.Contains(line, "Connection failed") {
				invalid = append(invalid, strings.TrimSpace(line))
			}
		}
	}

	return invalid
}

// checkDpkgInterrupted checks if dpkg was interrupted
func (c PackagesCheck) checkDpkgInterrupted() bool {
	cmd := exec.Command("dpkg", "--audit")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}

// checkPackageCacheSize returns cache size in MB
func (c PackagesCheck) checkPackageCacheSize() float64 {
	cmd := exec.Command("du", "-sm", "/var/cache/apt/archives")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(output))
	if len(fields) >= 1 {
		size, err := strconv.ParseFloat(fields[0], 64)
		if err == nil {
			return size
		}
	}

	return 0
}

// checkUnattendedUpgrades checks unattended-upgrades status
func (c PackagesCheck) checkUnattendedUpgrades() string {
	// Check if unattended-upgrades is installed
	cmd := exec.Command("dpkg", "-l", "unattended-upgrades")
	output, err := cmd.Output()
	if err != nil {
		return "not installed"
	}

	if !strings.Contains(string(output), "ii  unattended-upgrades") {
		return "not installed"
	}

	// Check if it's enabled
	cmd = exec.Command("systemctl", "is-enabled", "unattended-upgrades")
	output, err = cmd.Output()
	if err != nil {
		return "installed but status unknown"
	}

	status := strings.TrimSpace(string(output))
	if status == "enabled" {
		return "enabled"
	} else if status == "disabled" {
		return "disabled"
	}

	return fmt.Sprintf("installed (%s)", status)
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
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