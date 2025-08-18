package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseNetworkIssues(t *testing.T) {
	diagnosis := DiagnoseNetworkIssues()
	
	// Test basic structure
	if diagnosis.Issue != "Network Issues" {
		t.Errorf("Expected issue 'Network Issues', got '%s'", diagnosis.Issue)
	}
	
	// Should have findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected at least one finding")
	}
	
	// Test that network components are checked
	hasServiceCheck := false
	hasInterfaceCheck := false
	hasDNSCheck := false
	
	for _, finding := range diagnosis.Findings {
		lower := strings.ToLower(finding)
		if strings.Contains(lower, "service") || strings.Contains(lower, "networking") {
			hasServiceCheck = true
		}
		if strings.Contains(lower, "interface") {
			hasInterfaceCheck = true
		}
		if strings.Contains(lower, "dns") || strings.Contains(lower, "resolution") {
			hasDNSCheck = true
		}
	}
	
	if !hasServiceCheck {
		t.Error("Expected networking service check in findings")
	}
	if !hasInterfaceCheck {
		t.Error("Expected interface check in findings")
	}
	if !hasDNSCheck {
		t.Error("Expected DNS check in findings")
	}
	
	// Test fixes structure and root requirements
	for _, fix := range diagnosis.Fixes {
		if fix.Description == "" {
			t.Error("Fix description should not be empty")
		}
		if len(fix.Commands) == 0 || fix.Commands[0] == "" {
			t.Error("Fix command should not be empty")
		}
		
		// Network fixes typically require root
		for _, cmd := range fix.Commands {
			if strings.Contains(cmd, "systemctl") || 
			   strings.Contains(cmd, "ip ") ||
			   strings.Contains(cmd, "/etc/") {
				if !fix.RequiresRoot {
					t.Errorf("Network command '%s' should require root", cmd)
				}
			}
		}
	}
}