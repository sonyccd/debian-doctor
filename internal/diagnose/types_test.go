package diagnose

import (
	"testing"
	
	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

func TestDiagnosisTypes(t *testing.T) {
	// Test new Fix struct from fixes package
	fix := &fixes.Fix{
		ID:           "test_fix",
		Title:        "Test Fix",
		Description:  "Test fix",
		Commands:     []string{"echo 'test'"},
		RequiresRoot: false,
		RiskLevel:    fixes.RiskLow,
	}
	
	if fix.Description != "Test fix" {
		t.Errorf("Expected description 'Test fix', got '%s'", fix.Description)
	}
	
	if len(fix.Commands) == 0 || fix.Commands[0] != "echo 'test'" {
		t.Errorf("Expected command 'echo 'test'', got '%v'", fix.Commands)
	}
	
	if fix.RequiresRoot {
		t.Error("Expected RequiresRoot to be false")
	}
	
	// Test Diagnosis struct
	diagnosis := Diagnosis{
		Issue:    "Test Issue",
		Findings: []string{"Finding 1", "Finding 2"},
		Fixes:    []*fixes.Fix{fix},
	}
	
	if diagnosis.Issue != "Test Issue" {
		t.Errorf("Expected issue 'Test Issue', got '%s'", diagnosis.Issue)
	}
	
	if len(diagnosis.Findings) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(diagnosis.Findings))
	}
	
	if len(diagnosis.Fixes) != 1 {
		t.Errorf("Expected 1 fix, got %d", len(diagnosis.Fixes))
	}
	
	if diagnosis.Findings[0] != "Finding 1" {
		t.Errorf("Expected first finding 'Finding 1', got '%s'", diagnosis.Findings[0])
	}
}