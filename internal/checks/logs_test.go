package checks

import (
	"strings"
	"testing"
)

func TestLogsCheck_Name(t *testing.T) {
	check := LogsCheck{}
	expected := "System Logs"
	if got := check.Name(); got != expected {
		t.Errorf("LogsCheck.Name() = %v, want %v", got, expected)
	}
}

func TestLogsCheck_RequiresRoot(t *testing.T) {
	check := LogsCheck{}
	if check.RequiresRoot() {
		t.Error("LogsCheck.RequiresRoot() = true, want false")
	}
}

func TestLogsCheck_Run(t *testing.T) {
	check := LogsCheck{}
	result := check.Run()

	// Basic validation
	if result.Name != check.Name() {
		t.Errorf("Expected result.Name = %s, got %s", check.Name(), result.Name)
	}

	// Severity should be one of the valid severities
	validSeverities := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	validSeverity := false
	for _, severity := range validSeverities {
		if result.Severity == severity {
			validSeverity = true
			break
		}
	}
	if !validSeverity {
		t.Errorf("Invalid severity: %v", result.Severity)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Details should be populated
	if len(result.Details) == 0 {
		t.Error("Expected some details")
	}
}

func TestLogsCheck_isSignificantError(t *testing.T) {
	check := LogsCheck{}

	tests := []struct {
		message     string
		significant bool
	}{
		{"Critical system error occurred", true},
		{"Failed to start service", true},
		{"Connection reset by peer", false},
		{"Broken pipe", false},
		{"No route to host", false},
		{"Network is unreachable", false},
		{"Temporary failure in name resolution", false},
		{"Device busy", false},
		{"Resource temporarily unavailable", false},
		{"Kernel panic - not syncing", true},
		{"Out of memory: Kill process", true},
		{"I/O error on device", true},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := check.isSignificantError(tt.message)
			if result != tt.significant {
				t.Errorf("isSignificantError(%q) = %v, want %v", tt.message, result, tt.significant)
			}
		})
	}
}

func TestLogsCheck_checkJournalErrors(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state, so we mainly check that it doesn't crash
	errors := check.checkJournalErrors()
	
	// Should return a slice (might be empty)
	if errors == nil {
		t.Error("checkJournalErrors returned nil, expected slice")
	}
	
	// If errors exist, they should be non-empty strings
	for i, err := range errors {
		if strings.TrimSpace(err) == "" {
			t.Errorf("Error %d is empty or whitespace only", i)
		}
	}
}

func TestLogsCheck_checkAuthFailures(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state
	failures := check.checkAuthFailures()
	
	// Should return a non-negative number
	if failures < 0 {
		t.Errorf("checkAuthFailures returned negative value: %d", failures)
	}
}

func TestLogsCheck_checkDiskErrors(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state
	errors := check.checkDiskErrors()
	
	// Should return a slice (might be empty)
	if errors == nil {
		t.Error("checkDiskErrors returned nil, expected slice")
	}
	
	// If errors exist, they should be non-empty strings
	for i, err := range errors {
		if strings.TrimSpace(err) == "" {
			t.Errorf("Disk error %d is empty or whitespace only", i)
		}
	}
}

func TestLogsCheck_checkMemoryIssues(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state
	issues := check.checkMemoryIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkMemoryIssues returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Memory issue %d is empty or whitespace only", i)
		}
	}
}

func TestLogsCheck_checkServiceFailures(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state
	failures := check.checkServiceFailures()
	
	// Should return a slice (might be empty)
	if failures == nil {
		t.Error("checkServiceFailures returned nil, expected slice")
	}
	
	// If failures exist, they should be non-empty strings
	for i, failure := range failures {
		if strings.TrimSpace(failure) == "" {
			t.Errorf("Service failure %d is empty or whitespace only", i)
		}
	}
}

func TestLogsCheck_checkLogSizes(t *testing.T) {
	check := LogsCheck{}
	
	// This test will vary based on system state
	largeLogs := check.checkLogSizes()
	
	// Should return a slice (might be empty)
	if largeLogs == nil {
		t.Error("checkLogSizes returned nil, expected slice")
	}
	
	// If large logs exist, they should be non-empty strings and contain size info
	for i, log := range largeLogs {
		if strings.TrimSpace(log) == "" {
			t.Errorf("Large log %d is empty or whitespace only", i)
		}
		
		// Should contain size information (MB or GB)
		if !strings.Contains(log, "MB") && !strings.Contains(log, "GB") && !strings.Contains(log, "B") {
			t.Errorf("Large log %d doesn't contain size information: %s", i, log)
		}
	}
}

func TestLogsCheck_Integration(t *testing.T) {
	// This is an integration test that checks the overall functionality
	check := LogsCheck{}
	result := check.Run()

	// Verify the result structure is complete
	if result.Name == "" {
		t.Error("Result name is empty")
	}

	if result.Message == "" {
		t.Error("Result message is empty")
	}

	// The result should either be Info with no critical issues, or have appropriate severity
	if result.Severity == SeverityInfo {
		t.Log("System logs appear healthy")
	} else {
		t.Logf("Log issues detected with severity: %v", result.Severity)
	}

	// Details should always be present
	if len(result.Details) == 0 {
		t.Error("No details provided in result")
	}
}