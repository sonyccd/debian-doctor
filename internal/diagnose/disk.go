package diagnose


import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseDiskIssues diagnoses disk-related problems
func DiagnoseDiskIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Disk Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check disk usage
	var stat syscall.Statfs_t
	filesystems := map[string]string{
		"/":     "Root",
		"/home": "Home",
		"/var":  "Var",
		"/tmp":  "Tmp",
	}

	fullFilesystems := []string{}
	for path, name := range filesystems {
		if err := syscall.Statfs(path, &stat); err == nil {
			total := stat.Blocks * uint64(stat.Bsize)
			free := stat.Bavail * uint64(stat.Bsize)
			used := total - free
			usagePercent := int((used * 100) / total)
			
			if usagePercent > 95 {
				fullFilesystems = append(fullFilesystems, fmt.Sprintf("%s (%d%%)", name, usagePercent))
				diagnosis.Findings = append(diagnosis.Findings, 
					fmt.Sprintf("%s filesystem critical: %d%% full", name, usagePercent))
			} else if usagePercent > 85 {
				diagnosis.Findings = append(diagnosis.Findings, 
					fmt.Sprintf("%s filesystem warning: %d%% full", name, usagePercent))
			}
		}
	}

	// Always provide cleanup fixes for disk maintenance
	commonFixes := fixes.GetCommonFixes()
	
	if cleanFix, exists := commonFixes["clean_package_cache"]; exists {
		diagnosis.Fixes = append(diagnosis.Fixes, cleanFix)
	}
	
	if removeFix, exists := commonFixes["remove_orphaned_packages"]; exists {
		diagnosis.Fixes = append(diagnosis.Fixes, removeFix)
	}
	
	// Add custom fixes for disk analysis and cleanup
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "find_large_files",
		Title:       "Find Large Files",
		Description: "Find files larger than 100MB to identify disk space usage",
		Commands:    []string{"find / -type f -size +100M 2>/dev/null | head -20"},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})
	
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "clear_old_logs",
		Title:       "Clear Old System Logs",
		Description: "Remove system logs older than 7 days to free space",
		Commands:    []string{"journalctl --vacuum-time=7d"},
		RequiresRoot: true,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	// Check for I/O errors
	if output, err := exec.Command("dmesg").Output(); err == nil {
		outputStr := string(output)
		if strings.Contains(strings.ToLower(outputStr), "i/o error") || 
		   strings.Contains(strings.ToLower(outputStr), "disk error") {
			diagnosis.Findings = append(diagnosis.Findings, "Disk I/O errors detected in kernel log")
			
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:          "check_disk_health",
				Title:       "Check Disk Health",
				Description: "Use SMART tools to check disk health and identify potential failures",
				Commands:    []string{"smartctl -a /dev/sda"},
				RequiresRoot: true,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})
			
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:          "filesystem_check",
				Title:       "Filesystem Check",
				Description: "Run filesystem check to repair errors (WARNING: requires unmounting filesystem)",
				Commands:    []string{"umount /dev/sda1", "fsck -f /dev/sda1", "mount /dev/sda1"},
				RequiresRoot: true,
				Reversible:  true,
				ReverseCommands: []string{"mount /dev/sda1"},
				RiskLevel:   fixes.RiskHigh,
			})
		}
	}

	// Add disk speed test as an informational fix
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "test_disk_speed",
		Title:       "Test Disk Write Speed",
		Description: "Test disk write performance (creates and removes a 1GB test file)",
		Commands:    []string{
			"dd if=/dev/zero of=/tmp/test bs=1M count=1024 conv=fdatasync",
			"rm -f /tmp/test",
		},
		RequiresRoot: false,
		Reversible:  true,
		ReverseCommands: []string{"rm -f /tmp/test"},
		RiskLevel:   fixes.RiskLow,
	})

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No disk issues detected")
	}

	return diagnosis
}