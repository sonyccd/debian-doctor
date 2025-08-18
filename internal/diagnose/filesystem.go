package diagnose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseFilesystemIssues diagnoses filesystem-related problems and provides fixes
func DiagnoseFilesystemIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Filesystem Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check for read-only filesystems
	readOnlyFS := checkReadOnlyFilesystems()
	if len(readOnlyFS) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Read-only filesystems detected:")
		for _, fs := range readOnlyFS {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", fs))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "remount_rw",
			Title:       "Remount Filesystems Read-Write",
			Description: "Attempt to remount read-only filesystems as read-write",
			Commands:    []string{"mount -o remount,rw /"},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"mount -o remount,ro /"},
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_filesystem_errors",
			Title:       "Check for Filesystem Errors",
			Description: "Check filesystem for errors that might cause read-only state",
			Commands:    []string{"dmesg | grep -i 'filesystem\\|ext4\\|ext3'"},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check disk space issues
	spaceIssues := checkDiskSpaceIssues()
	if len(spaceIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Disk space issues:")
		for _, issue := range spaceIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "clean_temp_files",
			Title:       "Clean Temporary Files",
			Description: "Remove old temporary files to free disk space",
			Commands:    []string{
				"find /tmp -type f -atime +7 -delete",
				"find /var/tmp -type f -atime +7 -delete",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "clean_log_files",
			Title:       "Clean Old Log Files",
			Description: "Remove or compress old log files to free space",
			Commands:    []string{
				"journalctl --vacuum-time=30d",
				"find /var/log -name '*.log' -type f -mtime +30 -delete",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "find_large_files",
			Title:       "Find Large Files",
			Description: "Locate large files that may be consuming excessive disk space",
			Commands:    []string{
				"find / -type f -size +100M 2>/dev/null | head -20",
				"du -h /var /tmp /home | sort -rh | head -10",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check inode issues
	inodeIssues := checkInodeIssues()
	if len(inodeIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Inode usage issues:")
		for _, issue := range inodeIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "clean_small_files",
			Title:       "Clean Small/Empty Files",
			Description: "Remove small and empty files that consume inodes",
			Commands:    []string{
				"find /tmp -type f -size 0 -delete",
				"find /var/tmp -type f -size 0 -delete",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "find_inode_consumers",
			Title:       "Find Directories with Many Files",
			Description: "Locate directories consuming large numbers of inodes",
			Commands:    []string{
				"for dir in /tmp /var /home; do echo \"$dir:\"; find \"$dir\" -type d -exec sh -c 'echo \"$(find \"$1\" -maxdepth 1 | wc -l) $1\"' _ {} \\; 2>/dev/null | sort -rn | head -5; done",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for filesystem corruption
	corruptionSigns := checkFilesystemCorruption()
	if len(corruptionSigns) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Filesystem corruption detected:")
		for _, sign := range corruptionSigns {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", sign))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_filesystem",
			Title:       "Check Filesystem Integrity",
			Description: "Run filesystem check on unmounted filesystem (REQUIRES REBOOT)",
			Commands:    []string{
				"fsck -f /dev/sda1",
				"touch /forcefsck",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskHigh,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "backup_lost_found",
			Title:       "Backup Lost+Found Files",
			Description: "Create backup of files in lost+found directories",
			Commands:    []string{
				"tar -czf /root/lost_found_backup_$(date +%Y%m%d).tar.gz /lost+found /home/lost+found /var/lost+found 2>/dev/null || true",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for mount issues
	mountIssues := checkMountIssues()
	if len(mountIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Mount issues detected:")
		for _, issue := range mountIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "reload_systemd_mounts",
			Title:       "Reload Systemd Mount Units",
			Description: "Reload and restart failed mount units",
			Commands:    []string{
				"systemctl daemon-reload",
				"systemctl restart local-fs.target",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_fstab",
			Title:       "Validate fstab Configuration",
			Description: "Check /etc/fstab for syntax errors and missing devices",
			Commands:    []string{
				"mount -a --test",
				"findmnt --verify",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check for broken symbolic links
	brokenSymlinks := checkBrokenSymlinks()
	if len(brokenSymlinks) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("Broken symbolic links found: %d", len(brokenSymlinks)))
		
		for i, link := range brokenSymlinks {
			if i < 5 { // Show first 5
				diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", link))
			}
		}
		if len(brokenSymlinks) > 5 {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  ... and %d more", len(brokenSymlinks)-5))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "remove_broken_symlinks",
			Title:       "Remove Broken Symbolic Links",
			Description: "Remove broken symbolic links from common directories",
			Commands:    []string{
				"find /usr/bin /usr/local/bin /bin /sbin -type l ! -exec test -e {} \\; -delete 2>/dev/null || true",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskMedium,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "list_broken_symlinks",
			Title:       "List All Broken Symbolic Links",
			Description: "Find and list all broken symbolic links for manual review",
			Commands:    []string{
				"find /usr /etc /var -type l ! -exec test -e {} \\; -print 2>/dev/null | head -20",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Check filesystem performance
	performanceIssues := checkFilesystemPerformance()
	if len(performanceIssues) > 0 {
		diagnosis.Findings = append(diagnosis.Findings, "Filesystem performance issues:")
		for _, issue := range performanceIssues {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("  - %s", issue))
		}

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "optimize_filesystem",
			Title:       "Optimize Filesystem Performance",
			Description: "Run filesystem optimization commands",
			Commands:    []string{
				"sync",
				"echo 3 > /proc/sys/vm/drop_caches",
			},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})

		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:          "check_io_stats",
			Title:       "Check I/O Statistics",
			Description: "Display current I/O statistics and performance metrics",
			Commands:    []string{
				"iostat -x 1 5",
				"iotop -o -d 1 -n 5",
			},
			RequiresRoot: false,
			Reversible:  false,
			RiskLevel:   fixes.RiskLow,
		})
	}

	// Always add general filesystem maintenance fixes
	diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
		ID:          "filesystem_overview",
		Title:       "Filesystem Overview",
		Description: "Display comprehensive filesystem information",
		Commands:    []string{
			"df -h",
			"df -i",
			"mount | grep -E '^/dev'",
			"findmnt",
		},
		RequiresRoot: false,
		Reversible:  false,
		RiskLevel:   fixes.RiskLow,
	})

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No significant filesystem issues detected")
	}

	return diagnosis
}

// checkReadOnlyFilesystems finds filesystems mounted read-only
func checkReadOnlyFilesystems() []string {
	var readOnly []string

	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return readOnly
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, " ro,") && !strings.Contains(line, "tmpfs") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				readOnly = append(readOnly, fields[2])
			}
		}
	}

	return readOnly
}

// checkDiskSpaceIssues checks for disk space problems
func checkDiskSpaceIssues() []string {
	var issues []string

	var stat syscall.Statfs_t
	filesystems := map[string]string{
		"/":     "Root",
		"/home": "Home",
		"/var":  "Var",
		"/tmp":  "Tmp",
	}

	for path, name := range filesystems {
		if err := syscall.Statfs(path, &stat); err == nil {
			total := stat.Blocks * uint64(stat.Bsize)
			free := stat.Bavail * uint64(stat.Bsize)
			used := total - free
			usagePercent := int((used * 100) / total)
			
			if usagePercent > 95 {
				issues = append(issues, fmt.Sprintf("%s filesystem critical: %d%% full", name, usagePercent))
			} else if usagePercent > 85 {
				issues = append(issues, fmt.Sprintf("%s filesystem warning: %d%% full", name, usagePercent))
			}
		}
	}

	return issues
}

// checkInodeIssues checks for inode usage problems
func checkInodeIssues() []string {
	var issues []string

	cmd := exec.Command("df", "-i")
	output, err := cmd.Output()
	if err != nil {
		return issues
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			filesystem := fields[0]
			usageStr := fields[4]
			
			// Skip virtual filesystems
			if strings.HasPrefix(filesystem, "tmpfs") ||
				strings.HasPrefix(filesystem, "devtmpfs") {
				continue
			}

			if strings.HasSuffix(usageStr, "%") {
				usageStr = strings.TrimSuffix(usageStr, "%")
				if usage, err := strconv.Atoi(usageStr); err == nil {
					if usage > 90 {
						mountPoint := fields[5]
						issues = append(issues, fmt.Sprintf("%s: %d%% inode usage", mountPoint, usage))
					}
				}
			}
		}
	}

	return issues
}

// checkFilesystemCorruption looks for signs of filesystem corruption
func checkFilesystemCorruption() []string {
	var signs []string

	// Check for lost+found directories with content
	lostFoundDirs := []string{"/lost+found", "/home/lost+found", "/var/lost+found"}
	for _, dir := range lostFoundDirs {
		if _, err := os.Stat(dir); err == nil {
			entries, err := os.ReadDir(dir)
			if err == nil && len(entries) > 0 {
				signs = append(signs, fmt.Sprintf("Files found in %s (%d items)", dir, len(entries)))
			}
		}
	}

	// Check for filesystem errors in dmesg
	cmd := exec.Command("dmesg")
	output, err := cmd.Output()
	if err == nil {
		content := strings.ToLower(string(output))
		errorPatterns := []string{
			"ext4-fs error",
			"filesystem error",
			"corruption",
			"bad magic number",
		}

		for _, pattern := range errorPatterns {
			if strings.Contains(content, pattern) {
				signs = append(signs, fmt.Sprintf("Kernel log contains: %s", pattern))
			}
		}
	}

	return removeDuplicateStrings(signs)
}

// checkMountIssues checks for mount-related problems
func checkMountIssues() []string {
	var issues []string

	// Check for failed mount units
	cmd := exec.Command("systemctl", "list-units", "--failed", "--type=mount")
	output, err := cmd.Output()
	if err == nil {
		content := string(output)
		if strings.Contains(content, "failed") && !strings.Contains(content, "0 loaded units") {
			issues = append(issues, "Failed mount units in systemd")
		}
	}

	// Check fstab validity
	cmd = exec.Command("findmnt", "--verify")
	output, err = cmd.CombinedOutput()
	if err != nil {
		content := string(output)
		if content != "" {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, "Success") {
					issues = append(issues, fmt.Sprintf("fstab issue: %s", line))
				}
			}
		}
	}

	return issues
}

// checkBrokenSymlinks finds broken symbolic links
func checkBrokenSymlinks() []string {
	var broken []string

	checkDirs := []string{"/usr/bin", "/usr/local/bin", "/bin", "/sbin"}
	
	for _, dir := range checkDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					broken = append(broken, path)
				}
			}

			return nil
		})

		if err != nil {
			continue
		}

		// Limit the number of broken symlinks reported
		if len(broken) > 20 {
			break
		}
	}

	return broken
}

// checkFilesystemPerformance checks for performance issues
func checkFilesystemPerformance() []string {
	var issues []string

	// Check for high load average
	loadavg, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		fields := strings.Fields(string(loadavg))
		if len(fields) >= 1 {
			var load float64
			if _, err := fmt.Sscanf(fields[0], "%f", &load); err == nil {
				if load > 5.0 {
					issues = append(issues, fmt.Sprintf("High system load: %.2f", load))
				}
			}
		}
	}

	// Check for high I/O wait
	stat, err := os.ReadFile("/proc/stat")
	if err == nil {
		lines := strings.Split(string(stat), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu ") {
				fields := strings.Fields(line)
				if len(fields) >= 6 {
					var iowait, total int64
					fmt.Sscanf(fields[5], "%d", &iowait)
					for i := 1; i < len(fields); i++ {
						var val int64
						fmt.Sscanf(fields[i], "%d", &val)
						total += val
					}
					
					if total > 0 {
						iowaitPercent := (iowait * 100) / total
						if iowaitPercent > 10 {
							issues = append(issues, fmt.Sprintf("High I/O wait: %d%%", iowaitPercent))
						}
					}
				}
				break
			}
		}
	}

	return issues
}

