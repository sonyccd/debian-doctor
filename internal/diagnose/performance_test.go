package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnosePerformanceIssues(t *testing.T) {
	diagnosis := DiagnosePerformanceIssues()
	
	// Test basic structure
	if diagnosis.Issue != "Performance Issues" {
		t.Errorf("Expected issue 'Performance Issues', got '%s'", diagnosis.Issue)
	}
	
	// Should have findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected at least one finding")
	}
	
	// Test that CPU and memory checks are performed
	hasCPUCheck := false
	hasMemoryCheck := false
	hasLoadCheck := false
	
	for _, finding := range diagnosis.Findings {
		lower := strings.ToLower(finding)
		if strings.Contains(lower, "cpu") {
			hasCPUCheck = true
		}
		if strings.Contains(lower, "memory") {
			hasMemoryCheck = true
		}
		if strings.Contains(lower, "load") {
			hasLoadCheck = true
		}
	}
	
	if !hasCPUCheck {
		t.Error("Expected CPU usage check in findings")
	}
	if !hasMemoryCheck {
		t.Error("Expected memory usage check in findings")
	}
	if !hasLoadCheck {
		t.Error("Expected load average check in findings")
	}
	
	// Test fixes structure
	for _, fix := range diagnosis.Fixes {
		if fix.Description == "" {
			t.Error("Fix description should not be empty")
		}
		if len(fix.Commands) == 0 || fix.Commands[0] == "" {
			t.Error("Fix command should not be empty")
		}
		// Performance fixes often require root
		for _, cmd := range fix.Commands {
			if strings.Contains(cmd, "echo") && strings.Contains(cmd, "/proc/sys") {
				if !fix.RequiresRoot {
					t.Error("System cache clearing should require root")
				}
			}
		}
	}
}