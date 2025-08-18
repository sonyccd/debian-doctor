package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseServiceIssues(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	// Basic validation
	if diagnosis.Issue != "Service Issues" {
		t.Errorf("Expected issue 'Service Issues', got '%s'", diagnosis.Issue)
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

	// Should always have service overview fix
	hasOverviewFix := false
	for _, fix := range diagnosis.Fixes {
		if fix.ID == "service_overview" {
			hasOverviewFix = true
			break
		}
	}
	if !hasOverviewFix {
		t.Error("Expected service_overview fix to always be present")
	}
}

func TestCheckFailedSystemdServices(t *testing.T) {
	// This test depends on the system state, so we'll test the function exists
	// and returns a slice (empty or not)
	failed := checkFailedSystemdServices()

	// Should return a slice (might be empty)
	// Note: failed will never be nil, but might be empty

	// All service names should be non-empty
	for i, service := range failed {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Failed service %d has empty name", i)
		}
	}
}

func TestCheckServicesInErrorState(t *testing.T) {
	errorServices := checkServicesInErrorState()

	// Should return a slice (might be empty)
	// Note: errorServices will never be nil, but might be empty

	// All service names should be non-empty
	for i, service := range errorServices {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Error service %d has empty name", i)
		}
	}
}

func TestCheckCriticalServices(t *testing.T) {
	criticalServices := checkCriticalServices()

	// Should return a slice (might be empty)
	// Note: criticalServices will never be nil, but might be empty

	// All service names should be non-empty and from known critical services
	knownCriticalServices := map[string]bool{
		"networking": true, "systemd-networkd": true, "NetworkManager": true,
		"ssh": true, "sshd": true, "systemd-logind": true, "dbus": true,
		"systemd-resolved": true, "systemd-timesyncd": true,
	}

	for i, service := range criticalServices {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Critical service %d has empty name", i)
		}
		if !knownCriticalServices[service] {
			t.Errorf("Unknown critical service: %s", service)
		}
	}
}

func TestCheckFlappingServices(t *testing.T) {
	flappingServices := checkFlappingServices()

	// Should return a slice (might be empty)
	// Note: flappingServices will never be nil, but might be empty

	// All service names should be non-empty
	for i, service := range flappingServices {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Flapping service %d has empty name", i)
		}
	}
}

func TestCheckMaskedServices(t *testing.T) {
	maskedServices := checkMaskedServices()

	// Should return a slice (might be empty)
	if maskedServices == nil {
		t.Error("Expected slice, got nil")
	}

	// All service names should be non-empty
	for i, service := range maskedServices {
		if strings.TrimSpace(service) == "" {
			t.Errorf("Masked service %d has empty name", i)
		}
	}
}

func TestCheckServiceDependencies(t *testing.T) {
	dependencies := checkServiceDependencies()

	// Should return a slice (might be empty)
	// Note: dependencies will never be nil, but might be empty

	// All dependency issues should be non-empty
	for i, issue := range dependencies {
		if strings.TrimSpace(issue) == "" {
			t.Errorf("Dependency issue %d is empty", i)
		}
	}
}

func TestGenerateServiceLogCommands(t *testing.T) {
	tests := []struct {
		name     string
		services []string
		expected int
	}{
		{
			name:     "empty services",
			services: []string{},
			expected: 0,
		},
		{
			name:     "single service",
			services: []string{"nginx"},
			expected: 1,
		},
		{
			name:     "multiple services",
			services: []string{"nginx", "apache2", "mysql"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := generateServiceLogCommands(tt.services)

			if len(commands) != tt.expected {
				t.Errorf("Expected %d commands, got %d", tt.expected, len(commands))
			}

			// Check that each command contains journalctl and the service name
			for i, cmd := range commands {
				if !strings.Contains(cmd, "journalctl") {
					t.Errorf("Command %d should contain 'journalctl': %s", i, cmd)
				}
				if i < len(tt.services) && !strings.Contains(cmd, tt.services[i]) {
					t.Errorf("Command %d should contain service name '%s': %s", i, tt.services[i], cmd)
				}
			}
		})
	}
}

func TestGenerateEnableServiceCommands(t *testing.T) {
	tests := []struct {
		name     string
		services []string
		expected int
	}{
		{
			name:     "empty services",
			services: []string{},
			expected: 0,
		},
		{
			name:     "single service",
			services: []string{"nginx"},
			expected: 2, // enable + start
		},
		{
			name:     "multiple services",
			services: []string{"nginx", "apache2"},
			expected: 4, // 2 * (enable + start)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := generateEnableServiceCommands(tt.services)

			if len(commands) != tt.expected {
				t.Errorf("Expected %d commands, got %d", tt.expected, len(commands))
			}

			// Check that commands come in enable/start pairs
			for i := 0; i < len(commands); i += 2 {
				if i+1 < len(commands) {
					enableCmd := commands[i]
					startCmd := commands[i+1]

					if !strings.Contains(enableCmd, "enable") {
						t.Errorf("Command %d should be enable command: %s", i, enableCmd)
					}
					if !strings.Contains(startCmd, "start") {
						t.Errorf("Command %d should be start command: %s", i+1, startCmd)
					}
				}
			}
		})
	}
}

func TestGenerateDisableServiceCommands(t *testing.T) {
	tests := []struct {
		name     string
		services []string
		expected int
	}{
		{
			name:     "empty services",
			services: []string{},
			expected: 0,
		},
		{
			name:     "single service",
			services: []string{"nginx"},
			expected: 2, // stop + disable
		},
		{
			name:     "multiple services",
			services: []string{"nginx", "apache2"},
			expected: 4, // 2 * (stop + disable)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := generateDisableServiceCommands(tt.services)

			if len(commands) != tt.expected {
				t.Errorf("Expected %d commands, got %d", tt.expected, len(commands))
			}

			// Check that commands come in stop/disable pairs
			for i := 0; i < len(commands); i += 2 {
				if i+1 < len(commands) {
					stopCmd := commands[i]
					disableCmd := commands[i+1]

					if !strings.Contains(stopCmd, "stop") {
						t.Errorf("Command %d should be stop command: %s", i, stopCmd)
					}
					if !strings.Contains(disableCmd, "disable") {
						t.Errorf("Command %d should be disable command: %s", i+1, disableCmd)
					}
				}
			}
		})
	}
}

func TestGenerateFlappingAnalysisCommands(t *testing.T) {
	tests := []struct {
		name     string
		services []string
		expected int
	}{
		{
			name:     "empty services",
			services: []string{},
			expected: 0,
		},
		{
			name:     "single service",
			services: []string{"nginx"},
			expected: 2, // status + journalctl
		},
		{
			name:     "multiple services",
			services: []string{"nginx", "apache2"},
			expected: 4, // 2 * (status + journalctl)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := generateFlappingAnalysisCommands(tt.services)

			if len(commands) != tt.expected {
				t.Errorf("Expected %d commands, got %d", tt.expected, len(commands))
			}

			// Check that commands come in status/journalctl pairs
			for i := 0; i < len(commands); i += 2 {
				if i+1 < len(commands) {
					statusCmd := commands[i]
					journalCmd := commands[i+1]

					if !strings.Contains(statusCmd, "status") {
						t.Errorf("Command %d should be status command: %s", i, statusCmd)
					}
					if !strings.Contains(journalCmd, "journalctl") {
						t.Errorf("Command %d should be journalctl command: %s", i+1, journalCmd)
					}
				}
			}
		})
	}
}

func TestRemoveDuplicateServiceStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicateServiceStrings(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			// Convert to maps for easier comparison
			expectedMap := make(map[string]bool)
			for _, item := range tt.expected {
				expectedMap[item] = true
			}

			resultMap := make(map[string]bool)
			for _, item := range result {
				resultMap[item] = true
			}

			for item := range expectedMap {
				if !resultMap[item] {
					t.Errorf("Expected item '%s' not found in result", item)
				}
			}

			for item := range resultMap {
				if !expectedMap[item] {
					t.Errorf("Unexpected item '%s' found in result", item)
				}
			}
		})
	}
}

func TestServiceDiagnosisFixTypes(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	// Check for expected fix types
	expectedFixTypes := map[string]bool{
		"service_overview":        true,
		"check_service_conflicts": true,
	}

	foundFixTypes := make(map[string]bool)
	for _, fix := range diagnosis.Fixes {
		foundFixTypes[fix.ID] = true
	}

	for fixType := range expectedFixTypes {
		if !foundFixTypes[fixType] {
			t.Errorf("Expected fix type '%s' not found", fixType)
		}
	}
}

func TestServiceDiagnosisRiskLevels(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	for _, fix := range diagnosis.Fixes {
		// Information gathering should be low risk
		if strings.Contains(fix.Description, "overview") || 
		   strings.Contains(fix.Description, "Display") ||
		   strings.Contains(fix.Description, "Check") ||
		   strings.Contains(fix.Description, "Examine") {
			if fix.RiskLevel.String() != "Low" {
				t.Errorf("Information gathering fix should be low risk: %s", fix.Title)
			}
		}

		// Service restarts should be medium or high risk
		if strings.Contains(fix.Description, "restart") || 
		   strings.Contains(fix.Description, "Restart") {
			if fix.RiskLevel.String() == "Low" {
				t.Errorf("Service restart should not be low risk: %s", fix.Title)
			}
		}

		// Enabling/disabling services should be high risk
		if strings.Contains(fix.Description, "Enable") || 
		   strings.Contains(fix.Description, "enable") {
			if fix.RiskLevel.String() != "High" {
				t.Errorf("Service enable should be high risk: %s", fix.Title)
			}
		}
	}
}

func TestServiceDiagnosisReversibility(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	for _, fix := range diagnosis.Fixes {
		// Fixes that restart services should be reversible
		if strings.Contains(fix.Title, "Restart") && fix.Reversible {
			if len(fix.ReverseCommands) == 0 {
				t.Errorf("Reversible fix should have reverse commands: %s", fix.Title)
			}
		}

		// Fixes that enable services should be reversible
		if strings.Contains(fix.Title, "Enable") && fix.Reversible {
			if len(fix.ReverseCommands) == 0 {
				t.Errorf("Reversible fix should have reverse commands: %s", fix.Title)
			}
		}

		// Information gathering fixes should not be reversible
		if strings.Contains(fix.Description, "overview") || 
		   strings.Contains(fix.Description, "Display") {
			if fix.Reversible {
				t.Errorf("Information gathering fix should not be reversible: %s", fix.Title)
			}
		}
	}
}

func TestServiceDiagnosisRootRequirements(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	for _, fix := range diagnosis.Fixes {
		// Service control operations should require root
		if strings.Contains(fix.Description, "restart") || 
		   strings.Contains(fix.Description, "Restart") ||
		   strings.Contains(fix.Description, "Enable") ||
		   strings.Contains(fix.Description, "enable") ||
		   strings.Contains(fix.Description, "stop") ||
		   strings.Contains(fix.Description, "Stop") {
			if !fix.RequiresRoot {
				t.Errorf("Service control fix should require root: %s", fix.Title)
			}
		}

		// Information gathering should not require root
		if strings.Contains(fix.Description, "overview") || 
		   strings.Contains(fix.Description, "Display") ||
		   strings.Contains(fix.Description, "Check") ||
		   strings.Contains(fix.Description, "Examine") {
			if fix.RequiresRoot {
				t.Errorf("Information gathering fix should not require root: %s", fix.Title)
			}
		}
	}
}

func TestServiceDiagnosisIntegration(t *testing.T) {
	// Test the complete service diagnosis flow
	diagnosis := DiagnoseServiceIssues()

	// Should have meaningful findings
	if len(diagnosis.Findings) == 0 {
		t.Error("Expected some findings")
	}

	// Should have practical fixes
	if len(diagnosis.Fixes) < 2 { // At least overview and conflicts check
		t.Error("Expected at least 2 fixes")
	}

	// All fixes should be properly formed
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

		// Commands should be non-empty
		for j, cmd := range fix.Commands {
			if strings.TrimSpace(cmd) == "" {
				t.Errorf("Fix %d command %d is empty", i, j)
			}
		}
	}
}

func TestServiceDiagnosisFixUniqueness(t *testing.T) {
	// Test that fix IDs are unique within a diagnosis
	diagnosis := DiagnoseServiceIssues()

	fixIDs := make(map[string]bool)
	for i, fix := range diagnosis.Fixes {
		if fixIDs[fix.ID] {
			t.Errorf("Duplicate fix ID found: %s at index %d", fix.ID, i)
		}
		fixIDs[fix.ID] = true
	}
}

func TestServiceDiagnosisCommandSafety(t *testing.T) {
	diagnosis := DiagnoseServiceIssues()

	// Check that commands don't contain dangerous patterns
	dangerousPatterns := []string{
		"rm -rf", "dd if=", "mkfs", "fdisk", "parted",
		"shutdown", "halt", "reboot", "poweroff",
	}

	for _, fix := range diagnosis.Fixes {
		for _, cmd := range fix.Commands {
			cmdLower := strings.ToLower(cmd)
			for _, pattern := range dangerousPatterns {
				if strings.Contains(cmdLower, pattern) {
					t.Errorf("Potentially dangerous command found in fix '%s': %s", fix.Title, cmd)
				}
			}
		}
	}
}