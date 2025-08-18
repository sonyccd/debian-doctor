package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseBootIssues(t *testing.T) {
	diagnosis := DiagnoseBootIssues()
	
	// Test basic structure
	if diagnosis.Issue != "Boot Issues" {
		t.Errorf("Expected issue 'Boot Issues', got '%s'", diagnosis.Issue)
	}
	
	// Should have findings (even if no issues detected)
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected at least one finding")
	}
	
	// Test that findings contain meaningful information
	hasSystemStatusCheck := false
	for _, finding := range diagnosis.Findings {
		if strings.Contains(strings.ToLower(finding), "system") {
			hasSystemStatusCheck = true
			break
		}
	}
	if !hasSystemStatusCheck {
		t.Error("Expected system status check in findings")
	}
	
	// Test fixes structure
	for _, fix := range diagnosis.Fixes {
		if fix.Description == "" {
			t.Error("Fix description should not be empty")
		}
		if len(fix.Commands) == 0 || fix.Commands[0] == "" {
			t.Error("Fix command should not be empty")
		}
	}
}