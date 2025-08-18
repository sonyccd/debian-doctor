package checks

import (
	"testing"
)

func TestSystemInfoCheck(t *testing.T) {
	check := SystemInfoCheck{}
	
	// Test check properties
	if check.Name() != "System Information" {
		t.Errorf("Expected name 'System Information', got '%s'", check.Name())
	}
	
	if check.RequiresRoot() {
		t.Error("SystemInfoCheck should not require root")
	}
	
	// Test running the check
	result := check.Run()
	
	if result.Name != "System Information" {
		t.Errorf("Expected result name 'System Information', got '%s'", result.Name)
	}
	
	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	// Should have system information details
	if len(result.Details) == 0 {
		t.Error("Expected system information details")
	}
}

func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}
	
	if info.Hostname == "" {
		t.Error("Expected hostname to be set")
	}
	
	if info.Architecture == "" {
		t.Error("Expected architecture to be set")
	}
}

func TestGetOSRelease(t *testing.T) {
	osInfo, err := getOSRelease()
	if err != nil {
		t.Fatalf("getOSRelease failed: %v", err)
	}
	
	// Should have at least some basic fields
	if len(osInfo) == 0 {
		t.Error("Expected at least some OS release information")
	}
	
	// Test that common fields exist (at least one should be present)
	hasCommonField := false
	commonFields := []string{"ID", "NAME", "VERSION", "PRETTY_NAME"}
	for _, field := range commonFields {
		if _, exists := osInfo[field]; exists {
			hasCommonField = true
			break
		}
	}
	
	if !hasCommonField {
		t.Error("Expected at least one common OS release field")
	}
}

func TestGetDistributionInfo(t *testing.T) {
	name, version, err := GetDistributionInfo()
	if err != nil {
		t.Fatalf("GetDistributionInfo failed: %v", err)
	}
	
	// Name should not be empty
	if name == "" {
		t.Error("Expected distribution name to be set")
	}
	
	// Version may be empty on some systems, but that's okay
	// Just test that the function doesn't crash
	_ = version
}

func TestIsSystemdSystem(t *testing.T) {
	// This test just ensures the function doesn't crash
	// The result depends on the system
	result := IsSystemdSystem()
	
	// Result should be boolean (this is just a type check)
	_ = result
}