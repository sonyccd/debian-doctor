package checks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FilesystemCheck checks filesystem health and integrity
type FilesystemCheck struct{}

func (c FilesystemCheck) Name() string {
	return "Filesystem Health"
}

func (c FilesystemCheck) RequiresRoot() bool {
	return false // Basic filesystem checks don't require root
}

func (c FilesystemCheck) Run() CheckResult {
	result := CheckResult{
		Name:      c.Name(),
		Severity:  SeverityInfo,
		Message:   "Filesystem health check completed",
		Details:   []string{},
		Timestamp: time.Now(),
	}

	// Check filesystem mount status
	mountIssues := c.checkMountStatus()
	if len(mountIssues) > 0 {
		result.Severity = SeverityError
		result.Message = "Filesystem mount issues detected"
		result.Details = append(result.Details, "Mount issues:")
		for _, issue := range mountIssues {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", issue))
		}
	}

	// Check for read-only filesystems
	readOnlyFS := c.checkReadOnlyFilesystems()
	if len(readOnlyFS) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Read-only filesystems detected"
		}
		result.Details = append(result.Details, "Read-only filesystems:")
		for _, fs := range readOnlyFS {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", fs))
		}
	}

	// Check filesystem errors in dmesg
	fsErrors := c.checkFilesystemErrors()
	if len(fsErrors) > 0 {
		result.Severity = SeverityCritical
		result.Message = "Filesystem errors detected"
		result.Details = append(result.Details, "Filesystem errors in kernel log:")
		for i, err := range fsErrors {
			if i >= 5 { // Limit to first 5
				result.Details = append(result.Details, fmt.Sprintf("... and %d more", len(fsErrors)-5))
				break
			}
			result.Details = append(result.Details, fmt.Sprintf("  - %s", err))
		}
	}

	// Check for full inodes
	inodeIssues := c.checkInodeUsage()
	if len(inodeIssues) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "High inode usage detected"
		}
		result.Details = append(result.Details, "Inode usage issues:")
		for _, issue := range inodeIssues {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", issue))
		}
	}

	// Check for filesystem corruption indicators
	corruptionSigns := c.checkCorruptionSigns()
	if len(corruptionSigns) > 0 {
		result.Severity = SeverityCritical
		result.Message = "Filesystem corruption signs detected"
		result.Details = append(result.Details, "Corruption indicators:")
		for _, sign := range corruptionSigns {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", sign))
		}
	}

	// Check disk usage patterns
	diskUsageIssues := c.checkDiskUsagePatterns()
	if len(diskUsageIssues) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Disk usage issues detected"
		}
		result.Details = append(result.Details, "Disk usage concerns:")
		for _, issue := range diskUsageIssues {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", issue))
		}
	}

	// Check for orphaned files and directories
	orphanedCount := c.checkOrphanedFiles()
	if orphanedCount > 0 {
		result.Details = append(result.Details, fmt.Sprintf("Potential orphaned files in /tmp: %d", orphanedCount))
		if orphanedCount > 1000 {
			if result.Severity < SeverityWarning {
				result.Severity = SeverityWarning
				result.Message = "Many orphaned files detected"
			}
		}
	}

	// Check for symbolic link issues
	symlinkIssues := c.checkSymbolicLinks()
	if len(symlinkIssues) > 0 {
		if result.Severity < SeverityWarning {
			result.Severity = SeverityWarning
			result.Message = "Symbolic link issues detected"
		}
		result.Details = append(result.Details, "Symbolic link issues:")
		for _, issue := range symlinkIssues {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", issue))
		}
	}

	// Check filesystem fragmentation (for ext filesystems)
	fragmentation := c.checkFragmentation()
	if len(fragmentation) > 0 {
		result.Details = append(result.Details, "Fragmentation status:")
		for _, frag := range fragmentation {
			result.Details = append(result.Details, fmt.Sprintf("  - %s", frag))
		}
	}

	if result.Severity == SeverityInfo {
		result.Details = append(result.Details, "Filesystem appears healthy")
	}

	return result
}

// checkMountStatus checks for mount-related issues
func (c FilesystemCheck) checkMountStatus() []string {
	issues := []string{}

	// Check /proc/mounts for any mount errors
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		issues = append(issues, "Failed to read mount information")
		return issues
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ro,") && !strings.Contains(line, "tmpfs") {
			// Extract filesystem name
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				issues = append(issues, fmt.Sprintf("%s mounted read-only", fields[2]))
			}
		}
	}

	// Check for failed mounts in systemd
	cmd = exec.Command("systemctl", "list-units", "--failed", "--type=mount")
	output, err = cmd.Output()
	if err == nil {
		content := string(output)
		if strings.Contains(content, "failed") && !strings.Contains(content, "0 loaded units") {
			issues = append(issues, "Failed mount units detected in systemd")
		}
	}

	return issues
}

// checkReadOnlyFilesystems finds filesystems mounted read-only
func (c FilesystemCheck) checkReadOnlyFilesystems() []string {
	readOnly := []string{}

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return readOnly
	}
	defer file.Close()

	cmd := exec.Command("cat", "/proc/mounts")
	output, err := cmd.Output()
	if err != nil {
		return readOnly
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			mountPoint := fields[1]
			options := fields[3]
			
			// Skip virtual filesystems
			if strings.HasPrefix(mountPoint, "/proc") ||
				strings.HasPrefix(mountPoint, "/sys") ||
				strings.HasPrefix(mountPoint, "/dev") ||
				strings.Contains(fields[2], "tmpfs") {
				continue
			}

			if strings.Contains(options, "ro") {
				readOnly = append(readOnly, mountPoint)
			}
		}
	}

	return readOnly
}

// checkFilesystemErrors looks for filesystem errors in kernel logs
func (c FilesystemCheck) checkFilesystemErrors() []string {
	errors := []string{}

	cmd := exec.Command("dmesg")
	output, err := cmd.Output()
	if err != nil {
		return errors
	}

	errorPatterns := []string{
		"ext4-fs error",
		"ext3-fs error",
		"ext2-fs error",
		"xfs: filesystem error",
		"btrfs: error",
		"filesystem error",
		"bad magic number",
		"corrupt",
		"journal commit i/o error",
		"remounting filesystem read-only",
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, pattern := range errorPatterns {
			if strings.Contains(lineLower, pattern) {
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				errors = append(errors, strings.TrimSpace(line))
				break
			}
		}
	}

	return removeDuplicateStrings(errors)
}

// checkInodeUsage checks for high inode usage
func (c FilesystemCheck) checkInodeUsage() []string {
	issues := []string{}

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
				strings.HasPrefix(filesystem, "devtmpfs") ||
				strings.HasPrefix(filesystem, "udev") {
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

// checkCorruptionSigns looks for signs of filesystem corruption
func (c FilesystemCheck) checkCorruptionSigns() []string {
	signs := []string{}

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

	// Check for bad blocks in ext filesystems
	cmd := exec.Command("dumpe2fs", "-h", "/dev/sda1")
	output, err := cmd.Output()
	if err == nil {
		content := string(output)
		if strings.Contains(content, "Bad block count:") {
			re := regexp.MustCompile(`Bad block count:\s+(\d+)`)
			matches := re.FindStringSubmatch(content)
			if len(matches) >= 2 {
				if count, err := strconv.Atoi(matches[1]); err == nil && count > 0 {
					signs = append(signs, fmt.Sprintf("Bad blocks detected: %d", count))
				}
			}
		}
	}

	return signs
}

// checkDiskUsagePatterns analyzes disk usage for concerning patterns
func (c FilesystemCheck) checkDiskUsagePatterns() []string {
	issues := []string{}

	// Check for rapid disk usage changes (simplified check)
	cmd := exec.Command("df", "-h")
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
			usageStr := fields[4]
			if strings.HasSuffix(usageStr, "%") {
				usageStr = strings.TrimSuffix(usageStr, "%")
				if usage, err := strconv.Atoi(usageStr); err == nil {
					mountPoint := fields[5]
					if usage > 95 {
						issues = append(issues, fmt.Sprintf("%s is %d%% full (critical)", mountPoint, usage))
					} else if usage > 85 {
						issues = append(issues, fmt.Sprintf("%s is %d%% full (warning)", mountPoint, usage))
					}
				}
			}
		}
	}

	return issues
}

// checkOrphanedFiles counts potentially orphaned files in /tmp
func (c FilesystemCheck) checkOrphanedFiles() int {
	tmpDir := "/tmp"
	count := 0

	// Count files older than 7 days in /tmp
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if path == tmpDir {
			return nil // Skip the root directory
		}

		// Check if file is older than 7 days
		if time.Since(info.ModTime()) > 7*24*time.Hour {
			count++
		}

		return nil
	})

	if err != nil {
		return 0
	}

	return count
}

// checkSymbolicLinks checks for broken symbolic links
func (c FilesystemCheck) checkSymbolicLinks() []string {
	issues := []string{}

	// Check common directories for broken symlinks
	checkDirs := []string{"/usr/bin", "/usr/local/bin", "/bin", "/sbin"}
	
	for _, dir := range checkDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			if info.Mode()&os.ModeSymlink != 0 {
				// Check if symlink target exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					relPath := strings.TrimPrefix(path, dir)
					issues = append(issues, fmt.Sprintf("Broken symlink: %s%s", dir, relPath))
				}
			}

			return nil
		})

		if err != nil {
			continue // Skip directories we can't access
		}

		// Limit the number of issues reported
		if len(issues) > 10 {
			issues = issues[:10]
			issues = append(issues, "... (truncated, more broken symlinks exist)")
			break
		}
	}

	return issues
}

// checkFragmentation checks filesystem fragmentation (basic implementation)
func (c FilesystemCheck) checkFragmentation() []string {
	fragmentation := []string{}

	// Check if e2freefrag is available and run it on ext filesystems
	cmd := exec.Command("which", "e2freefrag")
	if cmd.Run() == nil {
		// Try to run e2freefrag on the root filesystem
		cmd = exec.Command("e2freefrag", "/dev/sda1")
		output, err := cmd.Output()
		if err == nil {
			content := string(output)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "free fragments") || strings.Contains(line, "average free size") {
					fragmentation = append(fragmentation, strings.TrimSpace(line))
				}
			}
		}
	}

	// If no specific fragmentation info, provide general guidance
	if len(fragmentation) == 0 {
		fragmentation = append(fragmentation, "Fragmentation analysis requires e2freefrag tool")
	}

	return fragmentation
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