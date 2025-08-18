package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnosePackageIssues(t *testing.T) {
	diagnosis := DiagnosePackageIssues()

	// Basic validation
	if diagnosis.Issue != "Package System Issues" {
		t.Errorf("Expected issue 'Package System Issues', got '%s'", diagnosis.Issue)
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

func TestCheckBrokenPackages(t *testing.T) {
	packages := checkBrokenPackages()
	
	// Should return a slice (might be empty)
	if packages == nil {
		t.Error("checkBrokenPackages returned nil, expected slice")
	}
	
	// If packages exist, they should be non-empty strings
	for i, pkg := range packages {
		if strings.TrimSpace(pkg) == "" {
			t.Errorf("Broken package %d is empty or whitespace only", i)
		}
		
		// Package names shouldn't contain spaces typically
		if strings.Contains(pkg, " ") {
			t.Errorf("Broken package %d contains spaces, might be malformed: %s", i, pkg)
		}
	}
	
	t.Logf("Broken packages found: %d", len(packages))
}

func TestCheckDependencyIssues(t *testing.T) {
	issues := checkDependencyIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkDependencyIssues returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Dependency issue %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Dependency issues found: %d", len(issues))
}

func TestCheckAPTLocked(t *testing.T) {
	locked := checkAPTLocked()
	
	// Should return a boolean
	if locked {
		t.Log("APT is currently locked")
	} else {
		t.Log("APT is not locked")
	}
}

func TestCheckRepositoryIssues(t *testing.T) {
	issues := checkRepositoryIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkRepositoryIssues returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Repository issue %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Repository issues found: %d", len(issues))
}

func TestCheckPackageCacheSize(t *testing.T) {
	size := checkPackageCacheSize()
	
	// Should return a non-negative number
	if size < 0 {
		t.Errorf("checkPackageCacheSize returned negative value: %f", size)
	}
	
	t.Logf("Package cache size: %.1f MB", size)
}

func TestCheckUpgradeableCount(t *testing.T) {
	count := checkUpgradeableCount()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkUpgradeableCount returned negative value: %d", count)
	}
	
	t.Logf("Upgradeable packages: %d", count)
}

func TestCheckOrphanedPackages(t *testing.T) {
	count := checkOrphanedPackages()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkOrphanedPackages returned negative value: %d", count)
	}
	
	t.Logf("Orphaned packages: %d", count)
}

func TestCheckPackageConfiguration(t *testing.T) {
	issues := checkPackageConfiguration()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkPackageConfiguration returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Configuration issue %d is empty or whitespace only", i)
		}
	}
	
	t.Logf("Configuration issues found: %d", len(issues))
}

func TestCheckDuplicatePackages(t *testing.T) {
	duplicates := checkDuplicatePackages()
	
	// Should return a slice (might be empty)
	if duplicates == nil {
		t.Error("checkDuplicatePackages returned nil, expected slice")
	}
	
	// If duplicates exist, they should be non-empty strings and contain version info
	for i, dup := range duplicates {
		if strings.TrimSpace(dup) == "" {
			t.Errorf("Duplicate package %d is empty or whitespace only", i)
		}
		
		// Should contain version count information
		if !strings.Contains(dup, "versions)") {
			t.Errorf("Duplicate package %d doesn't contain version info: %s", i, dup)
		}
	}
	
	t.Logf("Duplicate packages found: %d", len(duplicates))
}

func TestRemoveDuplicateStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"package1", "package2", "package3"},
			expected: []string{"package1", "package2", "package3"},
		},
		{
			name:     "with duplicates",
			input:    []string{"package1", "package2", "package1", "package3", "package2"},
			expected: []string{"package1", "package2", "package3"},
		},
		{
			name:     "all duplicates",
			input:    []string{"package1", "package1", "package1"},
			expected: []string{"package1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicateStrings(tt.input)
			
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("removeDuplicateStrings() length = %d, want %d", len(result), len(tt.expected))
			}
			
			// Check contents (order matters in this implementation)
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("removeDuplicateStrings() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestPackageDiagnosisFixValidation(t *testing.T) {
	diagnosis := DiagnosePackageIssues()
	
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
		"package_system_check",
	}
	
	for _, expectedID := range expectedFixes {
		if !fixIDs[expectedID] {
			t.Errorf("Expected fix ID '%s' not found", expectedID)
		}
	}
}

func TestPackageDiagnosisIntegration(t *testing.T) {
	// Integration test that validates the overall package diagnosis functionality
	diagnosis := DiagnosePackageIssues()
	
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
	
	// Check that findings contain package-related information
	findingsText := strings.Join(diagnosis.Findings, " ")
	expectedKeywords := []string{"package", "apt", "dpkg"}
	foundKeywords := 0
	
	for _, keyword := range expectedKeywords {
		if strings.Contains(strings.ToLower(findingsText), keyword) {
			foundKeywords++
		}
	}
	
	if foundKeywords == 0 {
		t.Error("Findings don't contain expected package-related keywords")
	}
	
	// Validate fix commands contain package management commands
	allCommands := make([]string, 0)
	for _, fix := range diagnosis.Fixes {
		allCommands = append(allCommands, fix.Commands...)
	}
	
	commandsText := strings.Join(allCommands, " ")
	packageCommands := []string{"apt", "dpkg", "aptitude"}
	foundCommands := 0
	
	for _, cmd := range packageCommands {
		if strings.Contains(commandsText, cmd) {
			foundCommands++
		}
	}
	
	if foundCommands == 0 {
		t.Error("Fix commands don't contain expected package management tools")
	}
}

func TestPackageDiagnosisRiskLevels(t *testing.T) {
	diagnosis := DiagnosePackageIssues()
	
	// Check that dangerous operations have appropriate risk levels
	for _, fix := range diagnosis.Fixes {
		// Lock file removal should be high risk
		if strings.Contains(fix.ID, "remove_apt_lock") && fix.RiskLevel.String() != "High" {
			t.Errorf("Lock file removal fix should be high risk, got %s", fix.RiskLevel.String())
		}
		
		// Package removal should be medium or high risk
		if strings.Contains(fix.Description, "remove") || strings.Contains(fix.Description, "Remove") {
			if fix.RiskLevel.String() == "Low" {
				t.Errorf("Package removal operation marked as low risk: %s", fix.Title)
			}
		}
		
		// Information gathering should be low risk
		if strings.Contains(fix.Description, "List") || strings.Contains(fix.Description, "Show") {
			if fix.RiskLevel.String() != "Low" {
				t.Errorf("Information gathering operation should be low risk: %s", fix.Title)
			}
		}
	}
}