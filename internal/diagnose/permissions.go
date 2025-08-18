package diagnose

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnosePermissionIssues performs comprehensive permission analysis
func DiagnosePermissionIssues() Diagnosis {
	findings := []string{}
	allFixes := []*fixes.Fix{}
	
	// Check common permission issues
	findings = append(findings, checkUserPermissions()...)
	findings = append(findings, checkHomeDirectoryPermissions()...)
	findings = append(findings, checkSystemDirectoryPermissions()...)
	findings = append(findings, checkExecutablePermissions()...)
	findings = append(findings, checkConfigFilePermissions()...)
	findings = append(findings, checkSSHPermissions()...)
	findings = append(findings, checkSudoPermissions()...)
	
	// Generate fixes
	allFixes = append(allFixes, generatePermissionFixes(findings)...)
	
	if len(findings) == 0 {
		findings = append(findings, "No permission issues detected")
	}
	
	return Diagnosis{
		Issue:    "Permission Issues",
		Findings: findings,
		Fixes:    allFixes,
	}
}

// DiagnoseFilePermissions analyzes permissions for a specific file or directory
func DiagnoseFilePermissions(path string) Diagnosis {
	findings := []string{}
	allFixes := []*fixes.Fix{}
	
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			findings = append(findings, fmt.Sprintf("Path does not exist: %s", path))
		} else if os.IsPermission(err) {
			findings = append(findings, fmt.Sprintf("Permission denied accessing: %s", path))
			allFixes = append(allFixes, &fixes.Fix{
				ID:           "fix_access_permission",
				Title:        "Fix Access Permission",
				Description:  fmt.Sprintf("Add read permission to access %s", path),
				Commands:     []string{fmt.Sprintf("sudo chmod +r '%s'", path)},
				RequiresRoot: true,
				RiskLevel:    fixes.RiskMedium,
			})
		} else {
			findings = append(findings, fmt.Sprintf("Error accessing path: %v", err))
		}
		return Diagnosis{
			Issue:    fmt.Sprintf("File Permission Analysis: %s", path),
			Findings: findings,
			Fixes:    allFixes,
		}
	}
	
	// Analyze permissions
	mode := info.Mode()
	findings = append(findings, fmt.Sprintf("Path: %s", path))
	findings = append(findings, fmt.Sprintf("Type: %s", getFileType(mode)))
	findings = append(findings, fmt.Sprintf("Permissions: %s (%04o)", mode.String(), mode.Perm()))
	
	// Get ownership information (Unix-specific)
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid := stat.Uid
		gid := stat.Gid
		
		// Get user and group names
		userName := getUsername(uid)
		groupName := getGroupname(gid)
		
		findings = append(findings, fmt.Sprintf("Owner: %s (UID: %d)", userName, uid))
		findings = append(findings, fmt.Sprintf("Group: %s (GID: %d)", groupName, gid))
		
		// Check if current user owns the file
		currentUser, err := user.Current()
		if err == nil {
			currentUID, _ := strconv.ParseUint(currentUser.Uid, 10, 32)
			if uint32(currentUID) != uid {
				findings = append(findings, "You do not own this file")
			}
		}
	}
	
	// Check specific permission issues
	if mode.IsDir() {
		findings = append(findings, checkDirectoryPermissions(path, mode)...)
		allFixes = append(allFixes, generateDirectoryFixes(path, mode)...)
	} else {
		findings = append(findings, checkFilePermissions(path, mode)...)
		allFixes = append(allFixes, generateFileFixes(path, mode)...)
	}
	
	// Check for security issues
	securityIssues := checkSecurityIssues(path, mode)
	if len(securityIssues) > 0 {
		findings = append(findings, "SECURITY ISSUES:")
		findings = append(findings, securityIssues...)
		allFixes = append(allFixes, generateSecurityFixes(path, mode)...)
	}
	
	return Diagnosis{
		Issue:    fmt.Sprintf("File Permission Analysis: %s", path),
		Findings: findings,
		Fixes:    allFixes,
	}
}

func checkUserPermissions() []string {
	findings := []string{}
	
	currentUser, err := user.Current()
	if err != nil {
		findings = append(findings, fmt.Sprintf("Cannot determine current user: %v", err))
		return findings
	}
	
	findings = append(findings, fmt.Sprintf("Current user: %s (UID: %s)", currentUser.Username, currentUser.Uid))
	
	// Check if user is in important groups
	groups, err := currentUser.GroupIds()
	if err == nil {
		importantGroups := []string{"sudo", "admin", "wheel", "docker", "vboxusers"}
		for _, group := range importantGroups {
			if hasGroup(groups, group) {
				findings = append(findings, fmt.Sprintf("User is in '%s' group", group))
			}
		}
	}
	
	return findings
}

func checkHomeDirectoryPermissions() []string {
	findings := []string{}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return findings
	}
	
	info, err := os.Stat(homeDir)
	if err != nil {
		return findings
	}
	
	mode := info.Mode()
	perm := mode.Perm()
	
	// Home directory should typically be 755 or 750
	if perm&0077 != 0 {
		findings = append(findings, fmt.Sprintf("Home directory has overly permissive permissions: %04o", perm))
	}
	
	// Check important subdirectories
	importantDirs := []string{".ssh", ".gnupg", ".config"}
	for _, dir := range importantDirs {
		dirPath := filepath.Join(homeDir, dir)
		if info, err := os.Stat(dirPath); err == nil {
			mode := info.Mode()
			if dir == ".ssh" && mode.Perm()&0077 != 0 {
				findings = append(findings, fmt.Sprintf("%s directory has insecure permissions: %04o", dir, mode.Perm()))
			}
		}
	}
	
	return findings
}

func checkSystemDirectoryPermissions() []string {
	findings := []string{}
	
	// Check critical system directories
	criticalDirs := map[string]os.FileMode{
		"/etc":     0755,
		"/bin":     0755,
		"/sbin":    0755,
		"/usr/bin": 0755,
		"/var/log": 0755,
	}
	
	for dir, expectedPerm := range criticalDirs {
		if info, err := os.Stat(dir); err == nil {
			perm := info.Mode().Perm()
			if perm != expectedPerm {
				findings = append(findings, fmt.Sprintf("%s has unexpected permissions: %04o (expected %04o)", 
					dir, perm, expectedPerm))
			}
		}
	}
	
	return findings
}

func checkExecutablePermissions() []string {
	findings := []string{}
	
	// Check if common executables are accessible
	executables := []string{
		"/bin/ls",
		"/bin/cat",
		"/usr/bin/sudo",
		"/usr/bin/apt",
		"/bin/systemctl",
	}
	
	for _, exe := range executables {
		if _, err := os.Stat(exe); err != nil {
			if os.IsPermission(err) {
				findings = append(findings, fmt.Sprintf("Cannot access executable: %s", exe))
			}
		} else {
			// Check if executable
			if info, err := os.Stat(exe); err == nil {
				if info.Mode()&0111 == 0 {
					findings = append(findings, fmt.Sprintf("File is not executable: %s", exe))
				}
			}
		}
	}
	
	return findings
}

func checkConfigFilePermissions() []string {
	findings := []string{}
	
	// Check sensitive configuration files
	sensitiveFiles := map[string]os.FileMode{
		"/etc/passwd":     0644,
		"/etc/shadow":     0640,
		"/etc/gshadow":    0640,
		"/etc/sudoers":    0440,
		"/etc/ssh/sshd_config": 0644,
	}
	
	for file, expectedPerm := range sensitiveFiles {
		if info, err := os.Stat(file); err == nil {
			perm := info.Mode().Perm()
			// Check if too permissive
			if perm&0007 != 0 {
				findings = append(findings, fmt.Sprintf("%s is world-readable/writable: %04o (expected %04o)", file, perm, expectedPerm))
			}
		}
	}
	
	return findings
}

func checkSSHPermissions() []string {
	findings := []string{}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return findings
	}
	
	sshDir := filepath.Join(homeDir, ".ssh")
	if info, err := os.Stat(sshDir); err == nil {
		perm := info.Mode().Perm()
		if perm != 0700 {
			findings = append(findings, fmt.Sprintf(".ssh directory has incorrect permissions: %04o (should be 0700)", perm))
		}
		
		// Check SSH key files
		sshFiles := map[string]os.FileMode{
			"id_rsa":           0600,
			"id_ed25519":       0600,
			"authorized_keys":  0600,
			"known_hosts":      0644,
			"config":           0644,
		}
		
		for file, expectedPerm := range sshFiles {
			filePath := filepath.Join(sshDir, file)
			if info, err := os.Stat(filePath); err == nil {
				perm := info.Mode().Perm()
				if strings.Contains(file, "id_") && perm != expectedPerm {
					findings = append(findings, fmt.Sprintf("SSH private key %s has insecure permissions: %04o (should be %04o)", 
						file, perm, expectedPerm))
				}
			}
		}
	}
	
	return findings
}

func checkSudoPermissions() []string {
	findings := []string{}
	
	// Check if user can use sudo
	currentUser, err := user.Current()
	if err == nil {
		// Check sudoers file (limited check without root)
		sudoersPath := "/etc/sudoers"
		if _, err := os.Stat(sudoersPath); err != nil {
			if os.IsPermission(err) {
				findings = append(findings, "Cannot check sudoers file (permission denied)")
			}
		}
		
		// Check if in sudo group
		groups, err := currentUser.GroupIds()
		if err == nil {
			if !hasGroup(groups, "sudo") && !hasGroup(groups, "admin") && !hasGroup(groups, "wheel") {
				findings = append(findings, "User is not in sudo/admin group")
			}
		}
	}
	
	return findings
}

func checkDirectoryPermissions(path string, mode os.FileMode) []string {
	findings := []string{}
	perm := mode.Perm()
	
	// Check execute permission (needed to access directory)
	if perm&0111 == 0 {
		findings = append(findings, "Directory is not accessible (no execute permission)")
	}
	
	// Check write permission
	if perm&0222 == 0 {
		findings = append(findings, "Directory is read-only")
	}
	
	// Check for sticky bit
	if mode&os.ModeSticky != 0 {
		findings = append(findings, "Directory has sticky bit set")
	}
	
	// Check for setuid/setgid
	if mode&os.ModeSetuid != 0 {
		findings = append(findings, "Directory has setuid bit set")
	}
	if mode&os.ModeSetgid != 0 {
		findings = append(findings, "Directory has setgid bit set")
	}
	
	return findings
}

func checkFilePermissions(path string, mode os.FileMode) []string {
	findings := []string{}
	perm := mode.Perm()
	
	// Check if executable
	if perm&0111 != 0 {
		findings = append(findings, "File is executable")
		
		// Check for setuid/setgid on executables
		if mode&os.ModeSetuid != 0 {
			findings = append(findings, "SECURITY: Executable has setuid bit set")
		}
		if mode&os.ModeSetgid != 0 {
			findings = append(findings, "SECURITY: Executable has setgid bit set")
		}
	}
	
	// Check world-writable
	if perm&0002 != 0 {
		findings = append(findings, "SECURITY: File is world-writable")
	}
	
	return findings
}

func checkSecurityIssues(path string, mode os.FileMode) []string {
	issues := []string{}
	perm := mode.Perm()
	
	// Check for overly permissive permissions
	if perm&0002 != 0 {
		issues = append(issues, "File/directory is world-writable")
	}
	
	// Check for setuid/setgid
	if mode&os.ModeSetuid != 0 {
		issues = append(issues, "Setuid bit is set (runs with owner privileges)")
	}
	if mode&os.ModeSetgid != 0 {
		issues = append(issues, "Setgid bit is set (runs with group privileges)")
	}
	
	// Check for sensitive file patterns
	sensitivePatterns := []string{
		"password", "passwd", "shadow", "private", "secret",
		"key", "token", ".pem", ".key", ".crt",
	}
	
	baseName := strings.ToLower(filepath.Base(path))
	for _, pattern := range sensitivePatterns {
		if strings.Contains(baseName, pattern) && perm&0077 != 0 {
			issues = append(issues, fmt.Sprintf("Potentially sensitive file has permissive permissions: %04o", perm))
			break
		}
	}
	
	return issues
}

func generatePermissionFixes(findings []string) []*fixes.Fix {
	allFixes := []*fixes.Fix{}
	
	// Fix for home directory permissions
	for _, finding := range findings {
		if strings.Contains(finding, "Home directory has overly permissive") {
			homeDir, _ := os.UserHomeDir()
			allFixes = append(allFixes, &fixes.Fix{
				ID:           "fix_home_permissions",
				Title:        "Fix Home Directory Permissions",
				Description:  "Set secure permissions on home directory",
				Commands:     []string{fmt.Sprintf("chmod 750 '%s'", homeDir)},
				RequiresRoot: false,
				RiskLevel:    fixes.RiskLow,
			})
		}
		
		if strings.Contains(finding, ".ssh directory has incorrect permissions") {
			homeDir, _ := os.UserHomeDir()
			sshDir := filepath.Join(homeDir, ".ssh")
			allFixes = append(allFixes, &fixes.Fix{
				ID:           "fix_ssh_dir_permissions",
				Title:        "Fix SSH Directory Permissions",
				Description:  "Set correct permissions on .ssh directory",
				Commands:     []string{
					fmt.Sprintf("chmod 700 '%s'", sshDir),
					fmt.Sprintf("chmod 600 '%s'/id_*", sshDir),
					fmt.Sprintf("chmod 600 '%s'/authorized_keys", sshDir),
					fmt.Sprintf("chmod 644 '%s'/known_hosts", sshDir),
				},
				RequiresRoot: false,
				RiskLevel:    fixes.RiskLow,
			})
		}
	}
	
	return allFixes
}

func generateDirectoryFixes(path string, mode os.FileMode) []*fixes.Fix {
	allFixes := []*fixes.Fix{}
	perm := mode.Perm()
	
	if perm&0111 == 0 {
		allFixes = append(allFixes, &fixes.Fix{
			ID:           "fix_dir_access",
			Title:        "Make Directory Accessible",
			Description:  "Add execute permission to access directory",
			Commands:     []string{fmt.Sprintf("chmod +x '%s'", path)},
			RequiresRoot: false,
			RiskLevel:    fixes.RiskLow,
		})
	}
	
	if perm&0222 == 0 {
		allFixes = append(allFixes, &fixes.Fix{
			ID:           "fix_dir_readonly",
			Title:        "Make Directory Writable",
			Description:  "Add write permission to directory",
			Commands:     []string{fmt.Sprintf("chmod u+w '%s'", path)},
			RequiresRoot: false,
			RiskLevel:    fixes.RiskLow,
		})
	}
	
	return allFixes
}

func generateFileFixes(path string, mode os.FileMode) []*fixes.Fix {
	allFixes := []*fixes.Fix{}
	perm := mode.Perm()
	
	if perm&0444 == 0 {
		allFixes = append(allFixes, &fixes.Fix{
			ID:           "fix_file_readable",
			Title:        "Make File Readable",
			Description:  "Add read permission to file",
			Commands:     []string{fmt.Sprintf("chmod +r '%s'", path)},
			RequiresRoot: false,
			RiskLevel:    fixes.RiskLow,
		})
	}
	
	return allFixes
}

func generateSecurityFixes(path string, mode os.FileMode) []*fixes.Fix {
	allFixes := []*fixes.Fix{}
	perm := mode.Perm()
	
	if perm&0002 != 0 {
		allFixes = append(allFixes, &fixes.Fix{
			ID:           "fix_world_writable",
			Title:        "Remove World-Writable Permission",
			Description:  "Remove world-writable permission for security",
			Commands:     []string{fmt.Sprintf("chmod o-w '%s'", path)},
			RequiresRoot: false,
			RiskLevel:    fixes.RiskLow,
			Reversible:   true,
			ReverseCommands: []string{fmt.Sprintf("chmod o+w '%s'", path)},
		})
	}
	
	if mode&os.ModeSetuid != 0 {
		allFixes = append(allFixes, &fixes.Fix{
			ID:           "remove_setuid",
			Title:        "Remove Setuid Bit",
			Description:  "Remove setuid bit for security",
			Commands:     []string{fmt.Sprintf("chmod u-s '%s'", path)},
			RequiresRoot: true,
			RiskLevel:    fixes.RiskHigh,
			Reversible:   true,
			ReverseCommands: []string{fmt.Sprintf("chmod u+s '%s'", path)},
		})
	}
	
	return allFixes
}

// Helper functions

func getFileType(mode os.FileMode) string {
	switch {
	case mode.IsRegular():
		return "Regular File"
	case mode.IsDir():
		return "Directory"
	case mode&os.ModeSymlink != 0:
		return "Symbolic Link"
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			return "Character Device"
		}
		return "Block Device"
	case mode&os.ModeNamedPipe != 0:
		return "Named Pipe"
	case mode&os.ModeSocket != 0:
		return "Socket"
	default:
		return "Unknown"
	}
}

func getUsername(uid uint32) string {
	u, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return fmt.Sprintf("UID:%d", uid)
	}
	return u.Username
}

func getGroupname(gid uint32) string {
	g, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return fmt.Sprintf("GID:%d", gid)
	}
	return g.Name
}

func hasGroup(groups []string, groupName string) bool {
	for _, gid := range groups {
		if g, err := user.LookupGroupId(gid); err == nil {
			if g.Name == groupName {
				return true
			}
		}
	}
	return false
}