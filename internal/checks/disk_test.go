package checks

import (
	"testing"
)

func TestDiskSpaceCheck(t *testing.T) {
	check := DiskSpaceCheck{}
	
	// Test check properties
	if check.Name() != "Disk Space" {
		t.Errorf("Expected name 'Disk Space', got '%s'", check.Name())
	}
	
	if check.RequiresRoot() {
		t.Error("DiskSpaceCheck should not require root")
	}
	
	// Test running the check
	result := check.Run()
	
	if result.Name != "Disk Space" {
		t.Errorf("Expected result name 'Disk Space', got '%s'", result.Name)
	}
	
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	// Should have details about disk usage
	if len(result.Details) == 0 {
		t.Error("Expected disk usage details")
	}
}