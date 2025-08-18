package checks

import (
	"strings"
	"testing"
)

func TestPackagesCheck_Name(t *testing.T) {
	check := PackagesCheck{}
	expected := "Package System"
	if got := check.Name(); got != expected {
		t.Errorf("PackagesCheck.Name() = %v, want %v", got, expected)
	}
}

func TestPackagesCheck_RequiresRoot(t *testing.T) {
	check := PackagesCheck{}
	if check.RequiresRoot() {
		t.Error("PackagesCheck.RequiresRoot() = true, want false")
	}
}

func TestPackagesCheck_Run(t *testing.T) {
	check := PackagesCheck{}
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

func TestPackagesCheck_checkBrokenPackages(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	broken := check.checkBrokenPackages()
	
	// Should return a slice (might be empty)
	if broken == nil {
		t.Error("checkBrokenPackages returned nil, expected slice")
	}
	
	// If packages exist, they should be non-empty strings
	for i, pkg := range broken {
		if strings.TrimSpace(pkg) == "" {
			t.Errorf("Broken package %d is empty or whitespace only", i)
		}
		
		// Package names shouldn't contain spaces or special chars typically
		if strings.Contains(pkg, " ") {
			t.Errorf("Broken package %d contains spaces, might be malformed: %s", i, pkg)
		}
	}
}

func TestPackagesCheck_checkHeldPackages(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	held := check.checkHeldPackages()
	
	// Should return a slice (might be empty)
	if held == nil {
		t.Error("checkHeldPackages returned nil, expected slice")
	}
	
	// If packages exist, they should be non-empty strings
	for i, pkg := range held {
		if strings.TrimSpace(pkg) == "" {
			t.Errorf("Held package %d is empty or whitespace only", i)
		}
	}
}

func TestPackagesCheck_checkUpgradeablePackages(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	count := check.checkUpgradeablePackages()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkUpgradeablePackages returned negative value: %d", count)
	}
	
	// Log the count for information
	t.Logf("Upgradeable packages: %d", count)
}

func TestPackagesCheck_checkAutoremovablePackages(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	count := check.checkAutoremovablePackages()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkAutoremovablePackages returned negative value: %d", count)
	}
	
	// Log the count for information
	t.Logf("Autoremovable packages: %d", count)
}

func TestPackagesCheck_checkAPTSources(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state and network
	invalid := check.checkAPTSources()
	
	// Should return a slice (might be empty)
	if invalid == nil {
		t.Error("checkAPTSources returned nil, expected slice")
	}
	
	// If invalid sources exist, they should be non-empty strings
	for i, source := range invalid {
		if strings.TrimSpace(source) == "" {
			t.Errorf("Invalid source %d is empty or whitespace only", i)
		}
	}
	
	// Log any issues for information
	if len(invalid) > 0 {
		t.Logf("Invalid sources found: %d", len(invalid))
	}
}

func TestPackagesCheck_checkDpkgInterrupted(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	interrupted := check.checkDpkgInterrupted()
	
	// Should return a boolean
	if interrupted {
		t.Log("dpkg interruption detected")
	} else {
		t.Log("No dpkg interruption detected")
	}
}

func TestPackagesCheck_checkPackageCacheSize(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system state
	size := check.checkPackageCacheSize()
	
	// Should return a non-negative number
	if size < 0 {
		t.Errorf("checkPackageCacheSize returned negative value: %f", size)
	}
	
	// Log the size for information
	t.Logf("Package cache size: %.1f MB", size)
}

func TestPackagesCheck_checkUnattendedUpgrades(t *testing.T) {
	check := PackagesCheck{}
	
	// This test will vary based on system configuration
	status := check.checkUnattendedUpgrades()
	
	// Should return a non-empty string
	if strings.TrimSpace(status) == "" {
		t.Error("checkUnattendedUpgrades returned empty status")
	}
	
	// Should be one of the expected values
	validStatuses := []string{
		"not installed",
		"enabled",
		"disabled",
	}
	
	validStatus := false
	for _, validStat := range validStatuses {
		if status == validStat || strings.Contains(status, "installed (") {
			validStatus = true
			break
		}
	}
	
	if !validStatus {
		t.Logf("Unexpected unattended upgrades status: %s", status)
	}
	
	t.Logf("Unattended upgrades status: %s", status)
}

func TestRemoveDuplicates(t *testing.T) {
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
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("removeDuplicates() length = %d, want %d", len(result), len(tt.expected))
			}
			
			// Check contents (order matters in this implementation)
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("removeDuplicates() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestPackagesCheck_Integration(t *testing.T) {
	// This is an integration test that checks the overall functionality
	check := PackagesCheck{}
	result := check.Run()

	// Verify the result structure is complete
	if result.Name == "" {
		t.Error("Result name is empty")
	}

	if result.Message == "" {
		t.Error("Result message is empty")
	}

	// Details should always be present
	if len(result.Details) == 0 {
		t.Error("No details provided in result")
	}

	// Check that the result provides useful information
	detailsText := strings.Join(result.Details, " ")
	
	// Should mention key package system aspects
	expectedKeywords := []string{"packages", "upgrade", "cache"}
	foundKeywords := 0
	for _, keyword := range expectedKeywords {
		if strings.Contains(strings.ToLower(detailsText), keyword) {
			foundKeywords++
		}
	}
	
	if foundKeywords == 0 {
		t.Error("Result details don't contain expected package-related information")
	}
}

func TestPackagesCheck_SeverityLogic(t *testing.T) {
	// Test that severity escalation works correctly
	check := PackagesCheck{}
	result := check.Run()

	// If there are any error conditions, severity should reflect that
	detailsText := strings.Join(result.Details, " ")
	
	if strings.Contains(detailsText, "Broken packages") && result.Severity < SeverityError {
		t.Error("Broken packages detected but severity is not Error or Critical")
	}
	
	if strings.Contains(detailsText, "Many packages need upgrading") && result.Severity < SeverityWarning {
		t.Error("Many upgradeable packages detected but severity is not Warning or higher")
	}
}