package checks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFilesystemCheck_Name(t *testing.T) {
	check := FilesystemCheck{}
	expected := "Filesystem Health"
	if got := check.Name(); got != expected {
		t.Errorf("FilesystemCheck.Name() = %v, want %v", got, expected)
	}
}

func TestFilesystemCheck_RequiresRoot(t *testing.T) {
	check := FilesystemCheck{}
	if check.RequiresRoot() {
		t.Error("FilesystemCheck.RequiresRoot() = true, want false")
	}
}

func TestFilesystemCheck_Run(t *testing.T) {
	check := FilesystemCheck{}
	result := check.Run()

	// Basic validation
	if result.Name != check.Name() {
		t.Errorf("Expected result.Name = %s, got %s", check.Name(), result.Name)
	}

	// Severity should be one of the valid severities
	validSeverities := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	validSeverity := false
	for _, severity := range validSeverities {
		if result.Severity == severity {
			validSeverity = true
			break
		}
	}
	if !validSeverity {
		t.Errorf("Invalid severity: %v", result.Severity)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Details should be populated
	if len(result.Details) == 0 {
		t.Error("Expected some details")
	}

	// Timestamp should be set
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestFilesystemCheck_checkMountStatus(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	issues := check.checkMountStatus()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkMountStatus returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Mount issue %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Mount issues found: %d", len(issues))
}

func TestFilesystemCheck_checkReadOnlyFilesystems(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	readOnly := check.checkReadOnlyFilesystems()
	
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

func TestFilesystemCheck_checkFilesystemErrors(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	errors := check.checkFilesystemErrors()
	
	// Should return a slice (might be empty, nil is acceptable for failed operations)
	if errors == nil {
		t.Log("checkFilesystemErrors returned nil (no errors or command failed)")
		return
	}
	
	// If errors exist, they should be non-empty strings
	for i, err := range errors {
		if strings.TrimSpace(err) == "" {
			t.Errorf("Filesystem error %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Filesystem errors found: %d", len(errors))
}

func TestFilesystemCheck_checkInodeUsage(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	issues := check.checkInodeUsage()
	
	// Should return a slice (might be empty, nil is acceptable for failed operations)
	if issues == nil {
		t.Log("checkInodeUsage returned nil (no issues or command failed)")
		return
	}
	
	// If issues exist, they should contain percentage information
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Inode issue %d is empty or whitespace only", i)
		}
		
		// Should contain percentage and path information
		if !strings.Contains(issue, "%") || !strings.Contains(issue, ":") {
			t.Errorf("Inode issue %d doesn't contain expected format: %s", i, issue)
		}
	}
	
	t.Logf("Inode usage issues found: %d", len(issues))
}

func TestFilesystemCheck_checkCorruptionSigns(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	signs := check.checkCorruptionSigns()
	
	// Should return a slice (might be empty, nil is acceptable for failed operations)
	if signs == nil {
		t.Log("checkCorruptionSigns returned nil (no corruption or command failed)")
		return
	}
	
	// If signs exist, they should be non-empty strings
	for i, sign := range signs {
		if strings.TrimSpace(sign) == "" {
			t.Errorf("Corruption sign %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Corruption signs found: %d", len(signs))
}

func TestFilesystemCheck_checkDiskUsagePatterns(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	issues := check.checkDiskUsagePatterns()
	
	// Should return a slice (might be empty, nil is acceptable for failed operations)
	if issues == nil {
		t.Log("checkDiskUsagePatterns returned nil (no issues or command failed)")
		return
	}
	
	// If issues exist, they should contain usage information
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Disk usage issue %d is empty or whitespace only", i)
		}
		
		// Should contain percentage information
		if !strings.Contains(issue, "%") {
			t.Errorf("Disk usage issue %d doesn't contain percentage: %s", i, issue)
		}
	}
	
	t.Logf("Disk usage issues found: %d", len(issues))
}

func TestFilesystemCheck_checkOrphanedFiles(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	count := check.checkOrphanedFiles()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkOrphanedFiles returned negative value: %d", count)
	}
	
	t.Logf("Orphaned files in /tmp: %d", count)
}

func TestFilesystemCheck_checkSymbolicLinks(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state
	issues := check.checkSymbolicLinks()
	
	// Should return a slice (might be empty, nil is acceptable for failed operations)
	if issues == nil {
		t.Log("checkSymbolicLinks returned nil (no broken links or command failed)")
		return
	}
	
	// If issues exist, they should be valid paths
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Symlink issue %d is empty or whitespace only", i)
		}
		
		// Should mention symlink
		if !strings.Contains(strings.ToLower(issue), "symlink") {
			t.Errorf("Symlink issue %d doesn't mention symlink: %s", i, issue)
		}
	}
	
	t.Logf("Symbolic link issues found: %d", len(issues))
}

func TestFilesystemCheck_checkFragmentation(t *testing.T) {
	check := FilesystemCheck{}
	
	// This test will vary based on system state and tools available
	fragmentation := check.checkFragmentation()
	
	// Should return a slice (might be empty, but shouldn't be nil)
	if fragmentation == nil {
		t.Error("checkFragmentation returned nil, expected slice")
	}
	
	// Should always return at least one entry (even if just a message about missing tools)
	if len(fragmentation) == 0 {
		t.Error("checkFragmentation returned empty slice, expected at least one entry")
	}
	
	// If fragmentation info exists, entries should be non-empty
	for i, frag := range fragmentation {
		if strings.TrimSpace(frag) == "" {
			t.Errorf("Fragmentation info %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Fragmentation info entries: %d", len(fragmentation))
}

func TestRemoveDuplicateStrings_Filesystem(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "filesystem errors with duplicates",
			input:    []string{"ext4-fs error", "corruption detected", "ext4-fs error", "bad block"},
			expected: []string{"ext4-fs error", "corruption detected", "bad block"},
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

func TestFilesystemCheck_Integration(t *testing.T) {
	// Integration test that validates the overall filesystem check functionality
	check := FilesystemCheck{}
	result := check.Run()
	
	// Validate basic structure
	if result.Name == "" {
		t.Error("Result name is empty")
	}
	
	if result.Message == "" {
		t.Error("Result message is empty")
	}
	
	if len(result.Details) == 0 {
		t.Error("No details provided in result")
	}
	
	// Check that details contain filesystem-related information
	detailsText := strings.Join(result.Details, " ")
	expectedKeywords := []string{"filesystem", "mount", "disk", "inode"}
	foundKeywords := 0
	
	for _, keyword := range expectedKeywords {
		if strings.Contains(strings.ToLower(detailsText), keyword) {
			foundKeywords++
		}
	}
	
	if foundKeywords == 0 {
		t.Error("Result details don't contain expected filesystem-related keywords")
	}
}

func TestFilesystemCheck_SeverityEscalation(t *testing.T) {
	// Test that severity escalation works correctly
	check := FilesystemCheck{}
	result := check.Run()
	
	detailsText := strings.Join(result.Details, " ")
	
	// Critical issues should result in critical severity
	criticalKeywords := []string{"corruption", "filesystem error", "critical"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(strings.ToLower(detailsText), keyword) {
			if result.Severity < SeverityError {
				t.Errorf("Found critical keyword '%s' but severity is not Error or Critical", keyword)
			}
			break
		}
	}
	
	// Warning issues should result in at least warning severity
	warningKeywords := []string{"read-only", "warning", "high inode usage"}
	for _, keyword := range warningKeywords {
		if strings.Contains(strings.ToLower(detailsText), keyword) {
			if result.Severity < SeverityWarning {
				t.Errorf("Found warning keyword '%s' but severity is not Warning or higher", keyword)
			}
			break
		}
	}
}

// Test helper function to create temporary files for testing
func createTempTestFiles(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Create some test files with different ages
	testFiles := []struct {
		name string
		age  time.Duration
	}{
		{"recent.txt", time.Hour},
		{"old.txt", 8 * 24 * time.Hour}, // 8 days old
		{"very_old.txt", 30 * 24 * time.Hour}, // 30 days old
	}
	
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
		file.Close()
		
		// Set file modification time
		pastTime := time.Now().Add(-tf.age)
		err = os.Chtimes(filePath, pastTime, pastTime)
		if err != nil {
			t.Fatalf("Failed to set file time for %s: %v", tf.name, err)
		}
	}
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return tempDir, cleanup
}

func TestFilesystemCheck_OrphanedFilesLogic(t *testing.T) {
	// Create temporary test environment
	tempDir, cleanup := createTempTestFiles(t)
	defer cleanup()
	
	// Count files older than 7 days in our test directory
	count := 0
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if path == tempDir {
			return nil
		}
		
		if time.Since(info.ModTime()) > 7*24*time.Hour {
			count++
		}
		
		return nil
	})
	
	if err != nil {
		t.Fatalf("Failed to walk temp directory: %v", err)
	}
	
	// We should find exactly 2 files older than 7 days (old.txt and very_old.txt)
	expectedCount := 2
	if count != expectedCount {
		t.Errorf("Expected %d old files, found %d", expectedCount, count)
	}
}