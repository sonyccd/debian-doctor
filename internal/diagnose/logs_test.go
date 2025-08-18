package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseLogIssues(t *testing.T) {
	diagnosis := DiagnoseLogIssues()

	// Basic validation
	if diagnosis.Issue != "System Log Issues" {
		t.Errorf("Expected issue 'System Log Issues', got '%s'", diagnosis.Issue)
	}

	// Should have findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected some findings")
	}

	// Should have fixes
	if len(diagnosis.Fixes) == 0 {
		t.Error("Expected some fixes")
	}

	// All fixes should have required fields
	for i, fix := range diagnosis.Fixes {
		if fix.ID == "" {
			t.Errorf("Fix %d has empty ID", i)
		}
		if fix.Title == "" {
			t.Errorf("Fix %d has empty Title", i)
		}
		if fix.Description == "" {
			t.Errorf("Fix %d has empty Description", i)
		}
		if len(fix.Commands) == 0 {
			t.Errorf("Fix %d has no commands", i)
		}
	}
}

func TestCheckJournalSize(t *testing.T) {
	size := checkJournalSize()
	
	// Should return a non-negative number
	if size < 0 {
		t.Errorf("checkJournalSize returned negative value: %f", size)
	}
	
	// Size should be reasonable (less than 100GB for most systems)
	if size > 100*1024 {
		t.Logf("Warning: Very large journal size detected: %.1f MB", size)
	}
}

func TestCheckPersistentErrors(t *testing.T) {
	errors := checkPersistentErrors()
	
	// Should return a slice (might be empty)
	if errors == nil {
		t.Error("checkPersistentErrors returned nil, expected slice")
	}
	
	// If errors exist, they should be non-empty strings
	for i, err := range errors {
		if strings.TrimSpace(err) == "" {
			t.Errorf("Persistent error %d is empty or whitespace only", i)
		}
		
		// Should contain occurrence count
		if !strings.Contains(err, "occurred") && !strings.Contains(err, "times") {
			t.Errorf("Persistent error %d doesn't show occurrence count: %s", i, err)
		}
	}
}

func TestCheckLogRotation(t *testing.T) {
	issues := checkLogRotation()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkLogRotation returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Log rotation issue %d is empty or whitespace only", i)
		}
	}
}

func TestCheckFailedServices(t *testing.T) {
	services := checkFailedServices()
	
	// Should return a slice (might be empty)
	if services == nil {
		t.Error("checkFailedServices returned nil, expected slice")
	}
	
	// If services exist, they should be non-empty strings
	for i, service := range services {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Failed service %d is empty or whitespace only", i)
		}
		
		// Service names shouldn't contain spaces (systemd unit names)
		if strings.Contains(service, " ") {
			t.Errorf("Failed service %d contains spaces, might be malformed: %s", i, service)
		}
	}
}

func TestCheckCoreDumps(t *testing.T) {
	count := checkCoreDumps()
	
	// Should return a non-negative number
	if count < 0 {
		t.Errorf("checkCoreDumps returned negative value: %d", count)
	}
}

func TestCheckKernelIssues(t *testing.T) {
	issues := checkKernelIssues()
	
	// Should return a slice (might be empty)
	if issues == nil {
		t.Error("checkKernelIssues returned nil, expected slice")
	}
	
	// If issues exist, they should be non-empty strings
	for i, issue := range issues {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Kernel issue %d is empty or whitespace only", i)
		}
		
		// Should start with "Detected:" as per the implementation
		if !strings.HasPrefix(issue, "Detected:") {
			t.Errorf("Kernel issue %d doesn't start with 'Detected:': %s", i, issue)
		}
	}
}

func TestNormalizeErrorMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"2023-08-17 15:30:45 localhost kernel: error occurred",
			"[TIME] localhost kernel: error occurred",
		},
		{
			"Process [1234] failed",
			"Process [PID] failed",
		},
		{
			"Connection to 192.168.1.1 failed",
			"Connection to [IP] failed",
		},
		{
			"Device /dev/sda1 error",
			"Device [DEVICE] error",
		},
		{
			"pid 5678 segfaulted",
			"pid [NUM] segfaulted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeErrorMessage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeErrorMessage_LongMessage(t *testing.T) {
	// Test message truncation
	longMessage := strings.Repeat("a", 150)
	result := normalizeErrorMessage(longMessage)
	
	if len(result) > 103 { // 100 chars + "..."
		t.Errorf("Long message was not truncated properly: length %d", len(result))
	}
	
	if !strings.HasSuffix(result, "...") {
		t.Error("Truncated message should end with '...'")
	}
}

func TestLogDiagnosisFixValidation(t *testing.T) {
	diagnosis := DiagnoseLogIssues()
	
	// Verify that all fix IDs are unique
	fixIDs := make(map[string]bool)
	for i, fix := range diagnosis.Fixes {
		if fixIDs[fix.ID] {
			t.Errorf("Duplicate fix ID found: %s at index %d", fix.ID, i)
		}
		fixIDs[fix.ID] = true
	}
	
	// Common fix IDs that should always be present
	expectedFixes := []string{
		"show_system_overview",
	}
	
	for _, expectedID := range expectedFixes {
		if !fixIDs[expectedID] {
			t.Errorf("Expected fix ID '%s' not found", expectedID)
		}
	}
}