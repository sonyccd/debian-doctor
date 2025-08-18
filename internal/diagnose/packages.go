package diagnose

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnosePackageIssues diagnoses APT package system problems and provides fixes
func DiagnosePackageIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Package System Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check for broken packages
	brokenPackages := checkBrokenPackages()
	if len(brokenPackages) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Broken packages detected: %d", len(brokenPackages)))
		
		for i, pkg := range brokenPackages {
			if i < 5 { // Show first 5
				diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", pkg))
			}
		}
		if len(brokenPackages) > 5 {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("  ... and %d more", len(brokenPackages)-5))
		}

		// Get common fixes for broken packages
		commonFixes := fixes.GetCommonFixes()
		if fixBroken, exists := commonFixes["fix_broken_packages"]; exists {
			diagnosis.Fixes = append(diagnosis.Fixes, fixBroken)
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "dpkg_configure_all",
			Title:       "Configure All Packages",
			Description: "Configure all unpacked but unconfigured packages",
			Commands:    []string{"dpkg --configure -a"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Check for dependency issues
	dependencyIssues := checkDependencyIssues()
	if len(dependencyIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Dependency issues found:")
		for _, issue := range dependencyIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "fix_dependencies",
			Title:       "Fix Missing Dependencies",
			Description: "Install missing dependencies and fix broken dependencies",
			Commands:    []string{"apt-get -f install"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Check for lock file issues
	if checkAPTLocked() {
		diagnosis.Findings = append(diagnosis.Findings, "APT is currently locked (another package operation in progress)")
		
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "show_apt_processes",
			Title:       "Show Running APT Processes",
			Description: "Display processes that may be using APT/dpkg",
			Commands:    []string{"ps aux | grep -E '(apt|dpkg|unattended-upgrade)'"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "remove_apt_lock",
			Title:       "Remove APT Lock Files (DANGEROUS)",
			Description: "Force remove APT lock files - only use if no APT processes are running",
			Commands:    []string{
				"rm -f /var/lib/dpkg/lock-frontend",
				"rm -f /var/lib/dpkg/lock",
				"rm -f /var/cache/apt/archives/lock",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskHigh,
		})
	}

	// Check for repository issues
	repoIssues := checkRepositoryIssues()
	if len(repoIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Repository issues detected:")
		for _, issue := range repoIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		commonFixes := fixes.GetCommonFixes()
		if updateFix, exists := commonFixes["update_package_cache"]; exists {
			diagnosis.Fixes = append(diagnosis.Fixes, updateFix)
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "fix_repository_keys",
			Title:       "Fix Repository Keys",
			Description: "Refresh and fix APT repository keys",
			Commands:    []string{
				"apt-key adv --refresh-keys --keyserver keyserver.ubuntu.com",
				"apt-get update",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for package cache issues
	cacheSize := checkPackageCacheSize()
	if cacheSize > 1000 { // More than 1GB
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Large package cache detected: %.1f MB", cacheSize))
		
		commonFixes := fixes.GetCommonFixes()
		if cleanFix, exists := commonFixes["clean_package_cache"]; exists {
			diagnosis.Fixes = append(diagnosis.Fixes, cleanFix)
		}
	}

	// Check for many upgradeable packages
	upgradeableCount := checkUpgradeableCount()
	if upgradeableCount > 20 {
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Many packages available for upgrade: %d", upgradeableCount))
		
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "upgrade_packages",
			Title:       "Upgrade All Packages",
			Description: "Upgrade all packages to their latest versions",
			Commands:    []string{"apt-get update", "apt-get upgrade -y"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "list_upgradeable",
			Title:       "List Upgradeable Packages",
			Description: "Show which packages can be upgraded",
			Commands:    []string{"apt list --upgradable"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for orphaned packages
	orphanedCount := checkOrphanedPackages()
	if orphanedCount > 10 {
		diagnosis.Findings = append(diagnosis.Findings, 
			fmt.Sprintf("Many orphaned packages detected: %d", orphanedCount))
		
		commonFixes := fixes.GetCommonFixes()
		if removeFix, exists := commonFixes["remove_orphaned_packages"]; exists {
			diagnosis.Fixes = append(diagnosis.Fixes, removeFix)
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "list_orphaned",
			Title:       "List Orphaned Packages",
			Description: "Show packages that can be automatically removed",
			Commands:    []string{"apt autoremove --dry-run"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for package configuration issues
	configIssues := checkPackageConfiguration()
	if len(configIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Package configuration issues:")
		for _, issue := range configIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "reconfigure_packages",
			Title:       "Reconfigure Packages",
			Description: "Reconfigure packages that failed configuration",
			Commands:    []string{"dpkg-reconfigure -a"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Check for duplicate packages
	duplicates := checkDuplicatePackages()
	if len(duplicates) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Duplicate packages detected:")
		for _, dup := range duplicates {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", dup))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "remove_duplicates",
			Title:       "Remove Duplicate Packages",
			Description: "Remove older versions of duplicate packages",
			Commands:    []string{"aptitude purge '~o'"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})
	}

	// Always add general maintenance fixes
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "package_system_check",
		Title:       "Comprehensive Package Check",
		Description: "Run comprehensive package system diagnostics",
		Commands:    []string{
			"apt-get check",
			"dpkg --audit",
			"apt list --installed | wc -l",
		},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No significant package system issues detected")
	}

	return diagnosis
}

// checkBrokenPackages finds packages in broken state
func checkBrokenPackages() []string {
	broken := []string{}

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

	return removeDuplicateStrings(broken)
}

// checkDependencyIssues checks for unmet dependencies
func checkDependencyIssues() []string {
	issues := []string{}

	cmd := exec.Command("apt-get", "check")
	output, err := cmd.CombinedOutput()
	if err != nil {
		content := string(output)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "Reading package lists") {
				issues = append(issues, line)
			}
		}
	}

	return issues
}

// checkAPTLocked checks if APT is currently locked
func checkAPTLocked() bool {
	lockFiles := []string{
		"/var/lib/dpkg/lock-frontend",
		"/var/lib/dpkg/lock",
		"/var/cache/apt/archives/lock",
	}

	cmd := exec.Command("lsof")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	content := string(output)
	for _, lockFile := range lockFiles {
		if strings.Contains(content, lockFile) {
			return true
		}
	}

	return false
}

// checkRepositoryIssues checks for repository problems
func checkRepositoryIssues() []string {
	issues := []string{}

	cmd := exec.Command("apt-get", "update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		content := string(output)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Failed to fetch") ||
				strings.Contains(line, "404  Not Found") ||
				strings.Contains(line, "NO_PUBKEY") ||
				strings.Contains(line, "KEYEXPIRED") {
				issues = append(issues, strings.TrimSpace(line))
			}
		}
	}

	return issues
}

// checkPackageCacheSize returns cache size in MB
func checkPackageCacheSize() float64 {
	cmd := exec.Command("du", "-sm", "/var/cache/apt/archives")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(output))
	if len(fields) >= 1 {
		var size float64
		fmt.Sscanf(fields[0], "%f", &size)
		return size
	}

	return 0
}

// checkUpgradeableCount counts packages that can be upgraded
func checkUpgradeableCount() int {
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

// checkOrphanedPackages counts orphaned packages
func checkOrphanedPackages() int {
	cmd := exec.Command("apt", "autoremove", "--dry-run")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	content := string(output)
	re := regexp.MustCompile(`The following packages will be REMOVED:\s*\n(.*?)(\n\n|\nNeed|\nAfter|$)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		packages := strings.Fields(matches[1])
		return len(packages)
	}

	return 0
}

// checkPackageConfiguration checks for configuration issues
func checkPackageConfiguration() []string {
	issues := []string{}

	cmd := exec.Command("dpkg", "--audit")
	output, err := cmd.Output()
	if err != nil {
		return issues
	}

	content := strings.TrimSpace(string(output))
	if content != "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				issues = append(issues, line)
			}
		}
	}

	return issues
}

// checkDuplicatePackages finds packages with multiple versions
func checkDuplicatePackages() []string {
	duplicates := []string{}

	cmd := exec.Command("dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return duplicates
	}

	packageCounts := make(map[string]int)
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "ii") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// Extract base package name (remove architecture suffix)
				pkgName := fields[1]
				if colonIndex := strings.Index(pkgName, ":"); colonIndex != -1 {
					pkgName = pkgName[:colonIndex]
				}
				packageCounts[pkgName]++
			}
		}
	}

	for pkg, count := range packageCounts {
		if count > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s (%d versions)", pkg, count))
		}
	}

	return duplicates
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(slice []string) []string {
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