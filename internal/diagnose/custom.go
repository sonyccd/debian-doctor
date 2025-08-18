package diagnose

import (
	"fmt"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseCustomIssue provides general troubleshooting steps for user-described issues
func DiagnoseCustomIssue(userDescription string) Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Custom Issue Diagnosis",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Clean and analyze the user description
	description := strings.TrimSpace(strings.ToLower(userDescription))
	
	if description == "" {
		diagnosis.Findings = append(diagnosis.Findings, "No issue description provided")
		diagnosis.Fixes = append(diagnosis.Fixes, getGeneralTroubleshootingFixes()...)
		return diagnosis
	}

	diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("Analyzing issue: %s", userDescription))

	// Analyze keywords in the description and provide relevant fixes
	keywords := extractKeywords(description)
	diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("Detected keywords: %s", strings.Join(keywords, ", ")))

	// Add specific fixes based on detected keywords
	specificFixes := getKeywordBasedFixes(keywords)
	diagnosis.Fixes = append(diagnosis.Fixes, specificFixes...)

	// Always add general troubleshooting steps
	generalFixes := getGeneralTroubleshootingFixes()
	diagnosis.Fixes = append(diagnosis.Fixes, generalFixes...)

	// Add system information gathering fixes
	infoFixes := getInformationGatheringFixes()
	diagnosis.Fixes = append(diagnosis.Fixes, infoFixes...)

	if len(keywords) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No specific keywords detected - providing general troubleshooting steps")
	} else {
		diagnosis.Findings = append(diagnosis.Findings, "Providing targeted troubleshooting based on detected keywords")
	}

	return diagnosis
}

// extractKeywords identifies relevant keywords from user description
func extractKeywords(description string) []string {
	var keywords []string
	
	// Define keyword categories and their associated terms
	keywordCategories := map[string][]string{
		"boot": {"boot", "startup", "grub", "start", "starting", "boots", "booting"},
		"network": {"network", "internet", "wifi", "ethernet", "connection", "dns", "ip", "ping", "connect"},
		"performance": {"slow", "fast", "performance", "lag", "freeze", "hang", "cpu", "memory", "ram"},
		"disk": {"disk", "storage", "space", "full", "hdd", "ssd", "filesystem", "mount"},
		"services": {"service", "daemon", "systemd", "process", "running", "stopped"},
		"graphics": {"graphics", "display", "screen", "resolution", "x11", "wayland", "nvidia", "amd"},
		"audio": {"audio", "sound", "speaker", "microphone", "alsa", "pulseaudio"},
		"packages": {"package", "apt", "install", "software", "application", "program"},
		"permissions": {"permission", "access", "denied", "sudo", "root", "user", "group"},
		"logs": {"log", "error", "warning", "journal", "syslog", "dmesg"},
		"hardware": {"hardware", "device", "driver", "usb", "bluetooth", "keyboard", "mouse"},
		"security": {"security", "firewall", "ssh", "login", "password", "authentication"},
	}

	words := strings.Fields(description)
	foundCategories := make(map[string]bool)

	for category, terms := range keywordCategories {
		for _, word := range words {
			for _, term := range terms {
				if strings.Contains(word, term) {
					if !foundCategories[category] {
						keywords = append(keywords, category)
						foundCategories[category] = true
					}
				}
			}
		}
	}

	return keywords
}

// getKeywordBasedFixes returns fixes based on detected keywords
func getKeywordBasedFixes(keywords []string) []*fixes.Fix {
	var specificFixes []*fixes.Fix

	for _, keyword := range keywords {
		switch keyword {
		case "boot":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_boot_issues",
				Title:       "Check Boot Issues",
				Description: "Examine boot process and grub configuration",
				Commands: []string{
					"systemctl status",
					"journalctl -b -p err",
					"lsblk",
					"mount | grep boot",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "network":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "diagnose_network",
				Title:       "Diagnose Network Issues",
				Description: "Check network configuration and connectivity",
				Commands: []string{
					"ip addr show",
					"ip route show",
					"ping -c 3 8.8.8.8",
					"systemctl status networking",
					"cat /etc/resolv.conf",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "performance":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_performance",
				Title:       "Check System Performance",
				Description: "Analyze CPU, memory, and system load",
				Commands: []string{
					"top -b -n 1 | head -20",
					"free -h",
					"uptime",
					"iostat -x 1 3",
					"ps aux --sort=-%cpu | head -10",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "disk":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_disk_space",
				Title:       "Check Disk Usage",
				Description: "Analyze disk space and filesystem health",
				Commands: []string{
					"df -h",
					"df -i",
					"lsblk",
					"mount",
					"du -h /var /tmp /home | sort -rh | head -10",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "services":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_services",
				Title:       "Check System Services",
				Description: "Examine systemd services and processes",
				Commands: []string{
					"systemctl --failed",
					"systemctl list-units --state=failed",
					"ps aux | grep -v grep",
					"systemctl status",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "graphics":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_graphics",
				Title:       "Check Graphics Configuration",
				Description: "Examine display and graphics driver status",
				Commands: []string{
					"lspci | grep -i vga",
					"lsmod | grep -E '(nvidia|amd|intel)'",
					"xrandr",
					"echo $DISPLAY",
					"ps aux | grep -E '(X|wayland)'",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "audio":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_audio",
				Title:       "Check Audio Configuration",
				Description: "Examine audio devices and sound system",
				Commands: []string{
					"aplay -l",
					"amixer",
					"pulseaudio --check",
					"systemctl --user status pulseaudio",
					"lsmod | grep snd",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "packages":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_packages",
				Title:       "Check Package System",
				Description: "Examine APT package manager and installations",
				Commands: []string{
					"apt list --upgradable",
					"apt-get check",
					"dpkg --audit",
					"apt-cache policy",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "permissions":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_permissions",
				Title:       "Check File Permissions",
				Description: "Examine user permissions and access rights",
				Commands: []string{
					"id",
					"groups",
					"ls -la /home/$USER",
					"sudo -l",
					"getfacl /home/$USER 2>/dev/null || echo 'No ACL support'",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})

		case "hardware":
			specificFixes = append(specificFixes, &fixes.Fix{
				ID:          "check_hardware",
				Title:       "Check Hardware Status",
				Description: "Examine hardware devices and drivers",
				Commands: []string{
					"lspci",
					"lsusb",
					"lsmod",
					"dmesg | grep -i error | tail -10",
					"lshw -short",
				},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})
		}
	}

	return specificFixes
}

// getGeneralTroubleshootingFixes returns general troubleshooting steps
func getGeneralTroubleshootingFixes() []*fixes.Fix {
	return []*fixes.Fix{
		{
			ID:          "system_overview",
			Title:       "System Overview",
			Description: "Get a comprehensive overview of system status",
			Commands: []string{
				"uname -a",
				"lsb_release -a",
				"uptime",
				"whoami",
				"pwd",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
		{
			ID:          "check_recent_changes",
			Title:       "Check Recent Changes",
			Description: "Look for recent system changes that might have caused issues",
			Commands: []string{
				"last | head -10",
				"grep $(date '+%b %d') /var/log/syslog | tail -20",
				"journalctl --since '1 hour ago' -p warning",
				"apt list --installed | grep $(date '+%Y-%m-%d')",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
		{
			ID:          "basic_connectivity_test",
			Title:       "Basic Connectivity Test",
			Description: "Test basic network and system connectivity",
			Commands: []string{
				"ping -c 3 127.0.0.1",
				"ping -c 3 8.8.8.8",
				"curl -I http://example.com",
				"nslookup google.com",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
		{
			ID:          "restart_common_services",
			Title:       "Restart Common Services",
			Description: "Restart commonly problematic services",
			Commands: []string{
				"systemctl restart networking",
				"systemctl restart systemd-resolved",
				"systemctl restart dbus",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		},
	}
}

// getInformationGatheringFixes returns fixes for gathering system information
func getInformationGatheringFixes() []*fixes.Fix {
	return []*fixes.Fix{
		{
			ID:          "gather_system_info",
			Title:       "Gather Detailed System Information",
			Description: "Collect comprehensive system information for troubleshooting",
			Commands: []string{
				"cat /proc/version",
				"cat /proc/cpuinfo | grep 'model name' | head -1",
				"cat /proc/meminfo | grep -E '^(MemTotal|MemFree|MemAvailable)'",
				"lscpu",
				"env | sort",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
		{
			ID:          "check_system_logs",
			Title:       "Check System Logs",
			Description: "Examine system logs for error messages and warnings",
			Commands: []string{
				"journalctl -p err --since '24 hours ago' --no-pager",
				"dmesg | grep -i error | tail -10",
				"tail -50 /var/log/syslog | grep -i error",
				"systemctl status --failed",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
		{
			ID:          "create_diagnostic_report",
			Title:       "Create Diagnostic Report",
			Description: "Generate a comprehensive diagnostic report",
			Commands: []string{
				"echo '=== SYSTEM INFO ===' > /tmp/diagnostic_report.txt",
				"uname -a >> /tmp/diagnostic_report.txt",
				"echo '=== DISK USAGE ===' >> /tmp/diagnostic_report.txt",
				"df -h >> /tmp/diagnostic_report.txt",
				"echo '=== MEMORY USAGE ===' >> /tmp/diagnostic_report.txt",
				"free -h >> /tmp/diagnostic_report.txt",
				"echo '=== RECENT ERRORS ===' >> /tmp/diagnostic_report.txt",
				"journalctl -p err --since '24 hours ago' --no-pager | tail -20 >> /tmp/diagnostic_report.txt",
				"echo 'Report saved to /tmp/diagnostic_report.txt'",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		},
	}
}

// GetTroubleshootingSuggestions provides general troubleshooting advice
func GetTroubleshootingSuggestions() []string {
	return []string{
		"Try restarting the specific service or application that's causing issues",
		"Check system logs for error messages around the time the issue started",
		"Verify that you have sufficient disk space and memory",
		"Test in a different user account to rule out user-specific configuration issues",
		"Check if the issue persists after a system reboot",
		"Verify network connectivity if the issue involves internet access",
		"Look for recent system updates or changes that might have caused the issue",
		"Check for hardware issues by examining dmesg output",
		"Try running the problematic command with elevated privileges (sudo)",
		"Search online for error messages you encounter",
	}
}