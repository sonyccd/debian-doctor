package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseDiskIssues(t *testing.T) {
	diagnosis := DiagnoseDiskIssues()
	
	// Test basic structure
	if diagnosis.Issue != "Disk Issues" {
		t.Errorf("Expected issue 'Disk Issues', got '%s'", diagnosis.Issue)
	}
	
	// Should have findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected at least one finding")
	}
	
	// Test that disk components are checked
	hasDiskUsageCheck := false
	
	for _, finding := range diagnosis.Findings {
		lower := strings.ToLower(finding)
		if strings.Contains(lower, "disk") || 
		   strings.Contains(lower, "filesystem") ||
		   strings.Contains(lower, "usage") ||
		   strings.Contains(lower, "no issues detected") {
			hasDiskUsageCheck = true
			break
		}
	}
	
	if !hasDiskUsageCheck {
		t.Error("Expected disk usage or status check in findings")
	}
	
	// Test that fixes include common cleanup operations
	hasCleanupFix := false
	for _, fix := range diagnosis.Fixes {
		if fix.Description == "" {
			t.Error("Fix description should not be empty")
		}
		if len(fix.Commands) == 0 || fix.Commands[0] == "" {
			t.Error("Fix command should not be empty")
		}
		
		for _, cmd := range fix.Commands {
			if strings.Contains(cmd, "apt-get clean") ||
			   strings.Contains(cmd, "apt-get autoremove") ||
			   strings.Contains(cmd, "journalctl --vacuum") {
				hasCleanupFix = true
			}
			
			// Most disk fixes require root
			if strings.Contains(cmd, "apt-get") || 
			   strings.Contains(cmd, "journalctl") {
				if !fix.RequiresRoot {
					t.Errorf("Disk cleanup command '%s' should require root", cmd)
				}
			}
		}
	}
	
	// Should have at least one cleanup fix suggested
	if len(diagnosis.Fixes) > 0 && !hasCleanupFix {
		t.Error("Expected at least one cleanup fix to be suggested")
	}
}