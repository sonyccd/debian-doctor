package diagnose


import (
	"os/exec"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseBootIssues diagnoses boot-related problems
func DiagnoseBootIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Boot Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check systemd state
	if output, err := exec.Command("systemctl", "is-system-running").Output(); err == nil {
		state := strings.TrimSpace(string(output))
		if state == "degraded" {
			diagnosis.Findings = append(diagnosis.Findings, "System is in degraded state")
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:          "show_failed_services",
				Title:       "Show Failed Services",
				Description: "Display services that failed to start during boot",
				Commands:    []string{"systemctl --failed"},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})
		} else if state == "running" {
			diagnosis.Findings = append(diagnosis.Findings, "System is running normally")
		} else {
			diagnosis.Findings = append(diagnosis.Findings, "System state: "+state)
		}
	}

	// Check for boot errors in journal
	cmd := exec.Command("journalctl", "-b", "--no-pager", "-p", "err", "-n", "10")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		errorCount := 0
		for _, line := range lines {
			if line != "" {
				errorCount++
			}
		}
		if errorCount > 0 {
			diagnosis.Findings = append(diagnosis.Findings, "Boot errors detected in system journal")
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:          "view_boot_errors",
				Title:       "View Boot Errors",
				Description: "Display boot-time errors from the system journal",
				Commands:    []string{"journalctl -b -p err"},
				RequiresRoot: false,
				Reversible:  false,
				RiskLevel:   fixes.RiskLow,
			})
		}
	}

	// Check filesystem mount status
	if output, err := exec.Command("mount").Output(); err == nil {
		if strings.Contains(string(output), "ro,") {
			diagnosis.Findings = append(diagnosis.Findings, "Read-only filesystem detected")
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:          "remount_rw",
				Title:       "Remount Filesystem Read-Write",
				Description: "Remount the root filesystem as read-write to allow modifications",
				Commands:    []string{"mount -o remount,rw /"},
				RequiresRoot: true,
				Reversible:  true,
				ReverseCommands: []string{"mount -o remount,ro /"},
				RiskLevel:   fixes.RiskMedium,
			})
		}
	}

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No boot issues detected")
	}

	return diagnosis
}