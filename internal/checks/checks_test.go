package checks

import (
	"os"
	"testing"
)

func TestGetAllChecks(t *testing.T) {
	// Test with current user privileges
	checks := GetAllChecks()
	
	if len(checks) == 0 {
		t.Error("Expected at least some checks to be available")
	}
	
	// Test that all basic checks are included
	checkNames := make(map[string]bool)
	for _, check := range checks {
		checkNames[check.Name()] = true
	}
	
	expectedChecks := []string{
		"System Information",
		"Disk Space",
		"Memory Usage",
		"Network Configuration",
	}
	
	for _, expectedCheck := range expectedChecks {
		if !checkNames[expectedCheck] {
			t.Errorf("Expected check '%s' to be included", expectedCheck)
		}
	}
	
	// Test that services check is included only if running as root
	isRoot := os.Geteuid() == 0
	hasServicesCheck := checkNames["System Services"]
	
	if isRoot && !hasServicesCheck {
		t.Error("Expected Services check when running as root")
	}
	
	if !isRoot && hasServicesCheck {
		t.Error("Did not expect Services check when not running as root")
	}
}

func TestCheckInterface(t *testing.T) {
	checks := GetAllChecks()
	
	for _, check := range checks {
		// Test that Name() returns non-empty string
		name := check.Name()
		if name == "" {
			t.Error("Check name should not be empty")
		}
		
		// Test that RequiresRoot() returns a boolean (this is implicit but good to verify)
		requiresRoot := check.RequiresRoot()
		_ = requiresRoot // Just ensure it doesn't crash
		
		// Test that Run() returns a valid result
		result := check.Run()
		
		if result.Name == "" {
			t.Errorf("Check '%s' returned empty result name", name)
		}
		
		if result.Name != name {
			t.Errorf("Check '%s' returned mismatched result name '%s'", name, result.Name)
		}
		
		if result.Timestamp.IsZero() {
			t.Errorf("Check '%s' returned zero timestamp", name)
		}
		
		// Test that severity is valid
		validSeverities := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
		validSeverity := false
		for _, valid := range validSeverities {
			if result.Severity == valid {
				validSeverity = true
				break
			}
		}
		
		if !validSeverity {
			t.Errorf("Check '%s' returned invalid severity %v", name, result.Severity)
		}
	}
}

func TestCheckConsistency(t *testing.T) {
	checks := GetAllChecks()
	
	// Run each check twice and ensure they're consistent in basic properties
	for _, check := range checks {
		result1 := check.Run()
		result2 := check.Run()
		
		if result1.Name != result2.Name {
			t.Errorf("Check '%s' returned inconsistent names", check.Name())
		}
		
		// Note: We don't test message/severity consistency as they may change
		// based on actual system state, which is expected behavior
	}
}

// Test that we can create a mock check and it works with the interface
func TestMockCheck(t *testing.T) {
	mock := &mockCheckImpl{
		name:         "Mock Check",
		requiresRoot: false,
	}
	
	// Test interface compliance
	var _ Check = mock
	
	// Test methods
	if mock.Name() != "Mock Check" {
		t.Errorf("Expected name 'Mock Check', got '%s'", mock.Name())
	}
	
	if mock.RequiresRoot() {
		t.Error("Expected RequiresRoot to be false")
	}
	
	result := mock.Run()
	if result.Name != "Mock Check" {
		t.Errorf("Expected result name 'Mock Check', got '%s'", result.Name)
	}
	
	if result.Severity != SeverityInfo {
		t.Errorf("Expected severity %v, got %v", SeverityInfo, result.Severity)
	}
}

// Mock check implementation for testing
type mockCheckImpl struct {
	name         string
	requiresRoot bool
}

func (m *mockCheckImpl) Name() string {
	return m.name
}

func (m *mockCheckImpl) RequiresRoot() bool {
	return m.requiresRoot
}

func (m *mockCheckImpl) Run() CheckResult {
	return CheckResult{
		Name:     m.name,
		Severity: SeverityInfo,
		Message:  "Mock check completed successfully",
		Details:  []string{"Mock detail 1", "Mock detail 2"},
	}
}