package checks

import (
	"testing"
	"time"
)

func TestSeverityLevels(t *testing.T) {
	// Test severity constants
	if SeverityInfo != 0 {
		t.Errorf("Expected SeverityInfo to be 0, got %d", SeverityInfo)
	}
	
	if SeverityWarning != 1 {
		t.Errorf("Expected SeverityWarning to be 1, got %d", SeverityWarning)
	}
	
	if SeverityError != 2 {
		t.Errorf("Expected SeverityError to be 2, got %d", SeverityError)
	}
	
	if SeverityCritical != 3 {
		t.Errorf("Expected SeverityCritical to be 3, got %d", SeverityCritical)
	}
}

func TestCheckResult(t *testing.T) {
	timestamp := time.Now()
	result := CheckResult{
		Name:      "Test Check",
		Severity:  SeverityWarning,
		Message:   "Test message",
		Details:   []string{"detail1", "detail2"},
		Timestamp: timestamp,
	}
	
	if result.Name != "Test Check" {
		t.Errorf("Expected name 'Test Check', got '%s'", result.Name)
	}
	
	if result.Severity != SeverityWarning {
		t.Errorf("Expected severity %v, got %v", SeverityWarning, result.Severity)
	}
	
	if result.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", result.Message)
	}
	
	if len(result.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(result.Details))
	}
	
	if result.Details[0] != "detail1" {
		t.Errorf("Expected first detail 'detail1', got '%s'", result.Details[0])
	}
	
	if !result.Timestamp.Equal(timestamp) {
		t.Error("Expected timestamp to match")
	}
}

func TestNewResults(t *testing.T) {
	results := NewResults()
	
	if len(results.GetErrors()) != 0 {
		t.Error("Expected no errors initially")
	}
	
	if len(results.GetWarnings()) != 0 {
		t.Error("Expected no warnings initially")
	}
	
	if len(results.GetInfo()) != 0 {
		t.Error("Expected no info initially")
	}
	
	if len(results.GetAllChecks()) != 0 {
		t.Error("Expected no checks initially")
	}
}

func TestResultsAddResult(t *testing.T) {
	results := NewResults()
	
	// Add error result
	errorResult := CheckResult{
		Name:     "Error Check",
		Severity: SeverityError,
		Message:  "Error message",
	}
	results.AddResult(errorResult)
	
	errors := results.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
	if errors[0] != "Error message" {
		t.Errorf("Expected error message 'Error message', got '%s'", errors[0])
	}
	
	// Add critical result (should also be in errors)
	criticalResult := CheckResult{
		Name:     "Critical Check",
		Severity: SeverityCritical,
		Message:  "Critical message",
	}
	results.AddResult(criticalResult)
	
	errors = results.GetErrors()
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
	
	// Add warning result
	warningResult := CheckResult{
		Name:     "Warning Check",
		Severity: SeverityWarning,
		Message:  "Warning message",
	}
	results.AddResult(warningResult)
	
	warnings := results.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "Warning message" {
		t.Errorf("Expected warning message 'Warning message', got '%s'", warnings[0])
	}
	
	// Add info result
	infoResult := CheckResult{
		Name:     "Info Check",
		Severity: SeverityInfo,
		Message:  "Info message",
	}
	results.AddResult(infoResult)
	
	info := results.GetInfo()
	if len(info) != 1 {
		t.Errorf("Expected 1 info, got %d", len(info))
	}
	if info[0] != "Info message" {
		t.Errorf("Expected info message 'Info message', got '%s'", info[0])
	}
	
	// Check total count
	allChecks := results.GetAllChecks()
	if len(allChecks) != 4 {
		t.Errorf("Expected 4 total checks, got %d", len(allChecks))
	}
}

func TestResultsGetMethods(t *testing.T) {
	results := NewResults()
	
	// Test empty results
	if len(results.GetErrors()) != 0 {
		t.Error("Expected empty errors list")
	}
	
	if len(results.GetWarnings()) != 0 {
		t.Error("Expected empty warnings list")
	}
	
	if len(results.GetInfo()) != 0 {
		t.Error("Expected empty info list")
	}
	
	if len(results.GetAllChecks()) != 0 {
		t.Error("Expected empty checks list")
	}
	
	// Add mixed results
	results.AddResult(CheckResult{Severity: SeverityError, Message: "Error 1"})
	results.AddResult(CheckResult{Severity: SeverityError, Message: "Error 2"})
	results.AddResult(CheckResult{Severity: SeverityWarning, Message: "Warning 1"})
	results.AddResult(CheckResult{Severity: SeverityInfo, Message: "Info 1"})
	
	// Test counts
	if len(results.GetErrors()) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(results.GetErrors()))
	}
	
	if len(results.GetWarnings()) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(results.GetWarnings()))
	}
	
	if len(results.GetInfo()) != 1 {
		t.Errorf("Expected 1 info, got %d", len(results.GetInfo()))
	}
	
	if len(results.GetAllChecks()) != 4 {
		t.Errorf("Expected 4 total checks, got %d", len(results.GetAllChecks()))
	}
}