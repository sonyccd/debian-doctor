package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseFilesystemIssues(t *testing.T) {
	diagnosis := DiagnoseFilesystemIssues()

	// Basic validation
	if diagnosis.Issue != "Filesystem Issues" {
		t.Errorf("Expected issue 'Filesystem Issues', got '%s'", diagnosis.Issue)
	}

	// Should have findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected some findings")
	}

	// Should have fixes
	if len(diagnosis.Fixes) == 0 {
		t.Error("Expected some fixes")
	}

	// All fixes should have required fields
	for i, fix := range diagnosis.Fixes {
		if fix.ID == "" {
			t.Errorf("Fix %d has empty ID", i)
		}
		if fix.Title == "" {
			t.Errorf("Fix %d has empty Title", i)
		}
		if fix.Description == "" {
			t.Errorf("Fix %d has empty Description", i)
		}
		if len(fix.Commands) == 0 {
			t.Errorf("Fix %d has no commands", i)
		}
	}
}

func TestCheckReadOnlyFilesystems(t *testing.T) {
	readOnly := checkReadOnlyFilesystems()
	
	// Should return a slice (might be empty)
	if readOnly == nil {
		t.Error("checkReadOnlyFilesystems returned nil, expected slice")
	}
	
	// If filesystems exist, they should be valid paths
	for i, fs := range readOnly {
		if strings.TrimSpace(fs) == "" {
			t.Errorf("Read-only filesystem %d is empty or whitespace only", i)
		}
		
		// Should be absolute paths
		if !strings.HasPrefix(fs, "/") {
			t.Errorf("Read-only filesystem %d is not an absolute path: %s", i, fs)
		}
	}
	
	t.Logf("Read-only filesystems found: %d", len(readOnly))
}

func TestCheckDiskSpaceIssues(t *testing.T) {
	issues := checkDiskSpaceIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkDiskSpaceIssues returned nil, expected slice")
	}
	
	// If issues exist, they should contain percentage information
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Disk space issue %d is empty or whitespace only", i)
		}
		
		// Should contain percentage and severity information
		if !strings.Contains(issue, "%") {
			t.Errorf("Disk space issue %d doesn't contain percentage: %s", i, issue)
		}
		
		// Should indicate severity
		if !strings.Contains(issue, "critical") && !strings.Contains(issue, "warning") {
			t.Errorf("Disk space issue %d doesn't indicate severity: %s", i, issue)
		}
	}
	
	t.Logf("Disk space issues found: %d", len(issues))
}

func TestCheckInodeIssues(t *testing.T) {
	issues := checkInodeIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkInodeIssues returned nil, expected slice")
	}
	
	// If issues exist, they should contain inode percentage information
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Inode issue %d is empty or whitespace only", i)
		}
		
		// Should contain percentage and colon
		if !strings.Contains(issue, "%") || !strings.Contains(issue, ":") {
			t.Errorf("Inode issue %d doesn't contain expected format: %s", i, issue)
		}
		
		// Should mention inode
		if !strings.Contains(strings.ToLower(issue), "inode") {
			t.Errorf("Inode issue %d doesn't mention inode: %s", i, issue)
		}
	}
	
	t.Logf("Inode issues found: %d", len(issues))
}

func TestCheckFilesystemCorruption(t *testing.T) {
	signs := checkFilesystemCorruption()
	
	// Should return a slice (might be empty)
	if signs == nil {
		t.Error("checkFilesystemCorruption returned nil, expected slice")
	}
	
	// If signs exist, they should be meaningful
	for i, sign := range signs {
		if strings.TrimSpace(sign) == "" {
			t.Errorf("Corruption sign %d is empty or whitespace only", i)
		}
		
		// Should contain relevant keywords
		lowerSign := strings.ToLower(sign)
		relevantKeywords := []string{"lost+found", "error", "corruption", "magic"}
		hasKeyword := false
		for _, keyword := range relevantKeywords {
			if strings.Contains(lowerSign, keyword) {
				hasKeyword = true
				break
			}
		}
		
		if !hasKeyword {
			t.Errorf("Corruption sign %d doesn't contain relevant keywords: %s", i, sign)
		}
	}
	
	t.Logf("Corruption signs found: %d", len(signs))
}

func TestCheckMountIssues(t *testing.T) {
	issues := checkMountIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkMountIssues returned nil, expected slice")
	}
	
	// If issues exist, they should be meaningful
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Mount issue %d is empty or whitespace only", i)
		}
		
		// Should mention mount-related terms
		lowerIssue := strings.ToLower(issue)
		mountKeywords := []string{"mount", "fstab", "systemd", "failed"}
		hasKeyword := false
		for _, keyword := range mountKeywords {
			if strings.Contains(lowerIssue, keyword) {
				hasKeyword = true
				break
			}
		}
		
		if !hasKeyword {
			t.Errorf("Mount issue %d doesn't contain mount-related keywords: %s", i, issue)
		}
	}
	
	t.Logf("Mount issues found: %d", len(issues))
}

func TestCheckBrokenSymlinks(t *testing.T) {
	broken := checkBrokenSymlinks()
	
	// Should return a slice (might be empty)
	if broken == nil {
		t.Error("checkBrokenSymlinks returned nil, expected slice")
	}
	
	// If broken symlinks exist, they should be valid paths
	for i, link := range broken {
		if strings.TrimSpace(link) == "" {
			t.Errorf("Broken symlink %d is empty or whitespace only", i)
		}
		
		// Should be absolute paths
		if !strings.HasPrefix(link, "/") {
			t.Errorf("Broken symlink %d is not an absolute path: %s", i, link)
		}
		
		// Should be in expected directories
		expectedDirs := []string{"/usr/bin", "/usr/local/bin", "/bin", "/sbin"}
		inExpectedDir := false
		for _, dir := range expectedDirs {
			if strings.HasPrefix(link, dir) {
				inExpectedDir = true
				break
			}
		}
		
		if !inExpectedDir {
			t.Errorf("Broken symlink %d is not in expected directory: %s", i, link)
		}
	}
	
	t.Logf("Broken symlinks found: %d", len(broken))
}

func TestCheckFilesystemPerformance(t *testing.T) {
	issues := checkFilesystemPerformance()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkFilesystemPerformance returned nil, expected slice")
	}
	
	// If issues exist, they should contain performance metrics
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Performance issue %d is empty or whitespace only", i)
		}
		
		// Should mention performance-related terms
		lowerIssue := strings.ToLower(issue)
		perfKeywords := []string{"load", "i/o", "wait", "high"}
		hasKeyword := false
		for _, keyword := range perfKeywords {
			if strings.Contains(lowerIssue, keyword) {
				hasKeyword = true
				break
			}
		}
		
		if !hasKeyword {
			t.Errorf("Performance issue %d doesn't contain performance keywords: %s", i, issue)
		}
	}
	
	t.Logf("Performance issues found: %d", len(issues))
}

func TestRemoveDuplicateStrings_Filesystem(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "corruption signs with duplicates",
			input:    []string{"lost+found files", "ext4-fs error", "lost+found files", "corruption"},
			expected: []string{"lost+found files", "ext4-fs error", "corruption"},
		},
		{
			name:     "mount points with duplicates",
			input:    []string{"/home", "/var", "/home", "/tmp"},
			expected: []string{"/home", "/var", "/tmp"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single item",
			input:    []string{"filesystem error"},
			expected: []string{"filesystem error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicateStrings(tt.input)
			
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("removeDuplicateStrings() length = %d, want %d", len(result), len(tt.expected))
			}
			
			// Check contents
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("removeDuplicateStrings() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestFilesystemDiagnosisFixValidation(t *testing.T) {
	diagnosis := DiagnoseFilesystemIssues()
	
	// Verify that all fix IDs are unique
	fixIDs := make(map[string]bool)
	for i, fix := range diagnosis.Fixes {
		if fixIDs[fix.ID] {
			t.Errorf("Duplicate fix ID found: %s at index %d", fix.ID, i)
		}
		fixIDs[fix.ID] = true
	}
	
	// Common fix IDs that should always be present
	expectedFixes := []string{
		"filesystem_overview",
	}
	
	for _, expectedID := range expectedFixes {
		if !fixIDs[expectedID] {
			t.Errorf("Expected fix ID '%s' not found", expectedID)
		}
	}
}

func TestFilesystemDiagnosisIntegration(t *testing.T) {
	// Integration test that validates the overall filesystem diagnosis functionality
	diagnosis := DiagnoseFilesystemIssues()
	
	// Validate basic structure
	if diagnosis.Issue == "" {
		t.Error("Diagnosis issue is empty")
	}
	
	if len(diagnosis.Findings) == 0 {
		t.Error("No findings in diagnosis")
	}
	
	if len(diagnosis.Fixes) == 0 {
		t.Error("No fixes in diagnosis")
	}
	
	// Check that findings contain filesystem-related information
	findingsText := strings.Join(diagnosis.Findings, " ")
	expectedKeywords := []string{"filesystem", "disk", "mount", "inode"}
	foundKeywords := 0
	
	for _, keyword := range expectedKeywords {
		if strings.Contains(strings.ToLower(findingsText), keyword) {
			foundKeywords++
		}
	}
	
	if foundKeywords == 0 {
		t.Error("Findings don't contain expected filesystem-related keywords")
	}
	
	// Validate fix commands contain filesystem management commands
	allCommands := make([]string, 0)
	for _, fix := range diagnosis.Fixes {
		allCommands = append(allCommands, fix.Commands...)
	}
	
	commandsText := strings.Join(allCommands, " ")
	fsCommands := []string{"mount", "fsck", "df", "find"}
	foundCommands := 0
	
	for _, cmd := range fsCommands {
		if strings.Contains(commandsText, cmd) {
			foundCommands++
		}
	}
	
	if foundCommands == 0 {
		t.Error("Fix commands don't contain expected filesystem management tools")
	}
}

func TestFilesystemDiagnosisRiskLevels(t *testing.T) {
	diagnosis := DiagnoseFilesystemIssues()
	
	// Check that dangerous operations have appropriate risk levels
	for _, fix := range diagnosis.Fixes {
		// Filesystem check should be high risk
		if strings.Contains(fix.ID, "check_filesystem") && fix.RiskLevel.String() != "High" {
			t.Errorf("Filesystem check should be high risk, got %s", fix.RiskLevel.String())
		}
		
		// File deletion should be at least medium risk
		if strings.Contains(fix.Description, "delete") || strings.Contains(fix.Description, "remove") {
			if fix.RiskLevel.String() == "Low" {
				t.Errorf("File deletion operation marked as low risk: %s", fix.Title)
			}
		}
		
		// Information gathering should be low risk
		if strings.Contains(fix.Description, "Display") || strings.Contains(fix.Description, "Check") ||
		   strings.Contains(fix.Description, "Find") || strings.Contains(fix.Description, "List") {
			if fix.RiskLevel.String() != "Low" {
				t.Errorf("Information gathering operation should be low risk: %s", fix.Title)
			}
		}
		
		// Mount operations should be medium or high risk
		if strings.Contains(fix.Description, "mount") || strings.Contains(fix.Description, "Mount") {
			if fix.RiskLevel.String() == "Low" {
				t.Errorf("Mount operation marked as low risk: %s", fix.Title)
			}
		}
	}
}

func TestFilesystemDiagnosisComprehensiveness(t *testing.T) {
	diagnosis := DiagnoseFilesystemIssues()
	
	// Check that we cover major filesystem issue categories
	findingsText := strings.Join(diagnosis.Findings, " ")
	
	// Should mention various filesystem aspects
	aspectsCovered := make(map[string]bool)
	aspects := map[string][]string{
		"space": {"disk", "space", "full"},
		"corruption": {"corruption", "lost+found", "error"},
		"mount": {"mount", "read-only"},
		"inodes": {"inode", "usage"},
		"symlinks": {"symlink", "broken"},
		"performance": {"performance", "load", "i/o"},
	}
	
	lowerFindings := strings.ToLower(findingsText)
	for aspect, keywords := range aspects {
		for _, keyword := range keywords {
			if strings.Contains(lowerFindings, keyword) {
				aspectsCovered[aspect] = true
				break
			}
		}
	}
	
	// We should cover at least some major aspects (not all systems will have all issues)
	if len(aspectsCovered) == 0 {
		t.Error("Diagnosis doesn't cover any major filesystem aspects")
	}
	
	t.Logf("Filesystem aspects covered: %v", aspectsCovered)
}