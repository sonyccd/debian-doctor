package checks

import (
	"testing"
)

func TestMemoryCheck(t *testing.T) {
	check := MemoryCheck{}
	
	// Test check properties
	if check.Name() != "Memory Usage" {
		t.Errorf("Expected name 'Memory Usage', got '%s'", check.Name())
	}
	
	if check.RequiresRoot() {
		t.Error("MemoryCheck should not require root")
	}
	
	// Test running the check
	result := check.Run()
	
	if result.Name != "Memory Usage" {
		t.Errorf("Expected result name 'Memory Usage', got '%s'", result.Name)
	}
	
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	// Should have details about memory usage
	if len(result.Details) == 0 {
		t.Error("Expected memory usage details")
	}
}