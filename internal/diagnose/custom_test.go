package diagnose

import (
	"strings"
	"testing"
)

func TestDiagnoseCustomIssue(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantKeywords []string
	}{
		{
			name:        "boot issue",
			description: "My system won't boot properly",
			wantKeywords: []string{"boot"},
		},
		{
			name:        "network problem",
			description: "Internet connection is not working",
			wantKeywords: []string{"network"},
		},
		{
			name:        "performance issue",
			description: "Computer is running very slow",
			wantKeywords: []string{"performance"},
		},
		{
			name:        "multiple keywords",
			description: "My computer is slow and network is not working",
			wantKeywords: []string{"performance", "network"},
		},
		{
			name:        "empty description",
			description: "",
			wantKeywords: []string{},
		},
		{
			name:        "no keywords",
			description: "Something is wrong but I don't know what",
			wantKeywords: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnosis := DiagnoseCustomIssue(tt.description)

			// Basic validation
			if diagnosis.Issue != "Custom Issue Diagnosis" {
				t.Errorf("Expected issue 'Custom Issue Diagnosis', got '%s'", diagnosis.Issue)
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

			// Check that detected keywords are mentioned in findings
			findingsText := strings.Join(diagnosis.Findings, " ")
			for _, keyword := range tt.wantKeywords {
				if !strings.Contains(strings.ToLower(findingsText), keyword) {
					t.Errorf("Expected keyword '%s' to be mentioned in findings", keyword)
				}
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    []string
	}{
		{
			name:        "boot keywords",
			description: "system won't boot at startup",
			expected:    []string{"boot"},
		},
		{
			name:        "network keywords",
			description: "internet connection and wifi problems",
			expected:    []string{"network"},
		},
		{
			name:        "performance keywords",
			description: "computer is slow and hangs frequently",
			expected:    []string{"performance"},
		},
		{
			name:        "disk keywords",
			description: "disk space is full and storage issues",
			expected:    []string{"disk"},
		},
		{
			name:        "multiple categories",
			description: "slow network connection and boot problems",
			expected:    []string{"boot", "network", "performance"},
		},
		{
			name:        "graphics keywords",
			description: "display resolution and screen problems",
			expected:    []string{"graphics"},
		},
		{
			name:        "audio keywords",
			description: "sound not working and speaker issues",
			expected:    []string{"audio"},
		},
		{
			name:        "package keywords",
			description: "software installation and apt problems",
			expected:    []string{"packages"},
		},
		{
			name:        "permission keywords",
			description: "access denied and permission issues",
			expected:    []string{"permissions"},
		},
		{
			name:        "hardware keywords",
			description: "usb device and hardware problems",
			expected:    []string{"hardware"},
		},
		{
			name:        "no keywords",
			description: "something is wrong",
			expected:    []string{},
		},
		{
			name:        "empty description",
			description: "",
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractKeywords(tt.description)

			// Check that all expected keywords are found
			expectedMap := make(map[string]bool)
			for _, keyword := range tt.expected {
				expectedMap[keyword] = true
			}

			resultMap := make(map[string]bool)
			for _, keyword := range result {
				resultMap[keyword] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected keyword '%s' not found in result %v", expected, result)
				}
			}

			// Allow for additional keywords to be found, but check that we don't miss any
			if len(result) < len(tt.expected) {
				t.Errorf("Expected at least %d keywords, got %d: %v", len(tt.expected), len(result), result)
			}
		})
	}
}

func TestGetKeywordBasedFixes(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		wantFixIDs []string
	}{
		{
			name:     "boot keywords",
			keywords: []string{"boot"},
			wantFixIDs: []string{"check_boot_issues"},
		},
		{
			name:     "network keywords",
			keywords: []string{"network"},
			wantFixIDs: []string{"diagnose_network"},
		},
		{
			name:     "performance keywords",
			keywords: []string{"performance"},
			wantFixIDs: []string{"check_performance"},
		},
		{
			name:     "multiple keywords",
			keywords: []string{"boot", "network"},
			wantFixIDs: []string{"check_boot_issues", "diagnose_network"},
		},
		{
			name:     "no keywords",
			keywords: []string{},
			wantFixIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixes := getKeywordBasedFixes(tt.keywords)

			// Create a map of fix IDs for easy lookup
			fixIDMap := make(map[string]bool)
			for _, fix := range fixes {
				fixIDMap[fix.ID] = true
			}

			// Check that all expected fix IDs are present
			for _, expectedID := range tt.wantFixIDs {
				if !fixIDMap[expectedID] {
					t.Errorf("Expected fix ID '%s' not found", expectedID)
				}
			}

			// Check that we got the expected number of fixes
			if len(fixes) != len(tt.wantFixIDs) {
				t.Errorf("Expected %d fixes, got %d", len(tt.wantFixIDs), len(fixes))
			}

			// Validate fix structure
			for i, fix := range fixes {
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
		})
	}
}

func TestGetGeneralTroubleshootingFixes(t *testing.T) {
	fixes := getGeneralTroubleshootingFixes()

	// Should return at least some fixes
	if len(fixes) == 0 {
		t.Error("Expected some general troubleshooting fixes")
	}

	// Expected fix IDs that should always be present
	expectedFixIDs := []string{
		"system_overview",
		"check_recent_changes",
		"basic_connectivity_test",
		"restart_common_services",
	}

	fixIDMap := make(map[string]bool)
	for _, fix := range fixes {
		fixIDMap[fix.ID] = true
	}

	for _, expectedID := range expectedFixIDs {
		if !fixIDMap[expectedID] {
			t.Errorf("Expected general fix ID '%s' not found", expectedID)
		}
	}

	// Validate fix structure
	for i, fix := range fixes {
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

func TestGetInformationGatheringFixes(t *testing.T) {
	fixes := getInformationGatheringFixes()

	// Should return at least some fixes
	if len(fixes) == 0 {
		t.Error("Expected some information gathering fixes")
	}

	// Expected fix IDs that should always be present
	expectedFixIDs := []string{
		"gather_system_info",
		"check_system_logs",
		"create_diagnostic_report",
	}

	fixIDMap := make(map[string]bool)
	for _, fix := range fixes {
		fixIDMap[fix.ID] = true
	}

	for _, expectedID := range expectedFixIDs {
		if !fixIDMap[expectedID] {
			t.Errorf("Expected info gathering fix ID '%s' not found", expectedID)
		}
	}

	// Validate fix structure
	for i, fix := range fixes {
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

	// Information gathering should generally be low risk
	for i, fix := range fixes {
		if fix.RiskLevel.String() != "Low" {
			t.Errorf("Information gathering fix %d should be low risk, got %s", i, fix.RiskLevel.String())
		}
	}
}

func TestGetTroubleshootingSuggestions(t *testing.T) {
	suggestions := GetTroubleshootingSuggestions()

	// Should return at least some suggestions
	if len(suggestions) == 0 {
		t.Error("Expected some troubleshooting suggestions")
	}

	// All suggestions should be non-empty
	for i, suggestion := range suggestions {
		if strings.TrimSpace(suggestion) == "" {
			t.Errorf("Suggestion %d is empty or whitespace only", i)
		}
	}

	// Should contain practical advice
	suggestionsText := strings.Join(suggestions, " ")
	expectedWords := []string{"restart", "check", "log", "system", "reboot", "error"}
	foundWords := 0

	for _, word := range expectedWords {
		if strings.Contains(strings.ToLower(suggestionsText), word) {
			foundWords++
		}
	}

	if foundWords < 3 {
		t.Errorf("Expected to find at least 3 practical troubleshooting words, found %d", foundWords)
	}
}

func TestCustomDiagnosisIntegration(t *testing.T) {
	// Test various real-world scenarios
	scenarios := []struct {
		description string
		expectKeywords bool
		expectFixes int // minimum number of fixes expected
	}{
		{
			description: "My computer won't start and I see a black screen",
			expectKeywords: true,
			expectFixes: 5, // boot fixes + general + info gathering
		},
		{
			description: "Internet is not working after system update",
			expectKeywords: true,
			expectFixes: 5, // network fixes + general + info gathering
		},
		{
			description: "Everything is running very slowly",
			expectKeywords: true,
			expectFixes: 5, // performance fixes + general + info gathering
		},
		{
			description: "I have a problem but don't know what it is",
			expectKeywords: false,
			expectFixes: 4, // just general + info gathering
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			diagnosis := DiagnoseCustomIssue(scenario.description)

			// Should have findings
			if len(diagnosis.Findings) == 0 {
				t.Error("Expected findings")
			}

			// Should have at least the minimum expected fixes
			if len(diagnosis.Fixes) < scenario.expectFixes {
				t.Errorf("Expected at least %d fixes, got %d", scenario.expectFixes, len(diagnosis.Fixes))
			}

			// Check keyword detection expectation
			findingsText := strings.Join(diagnosis.Findings, " ")
			hasKeywords := strings.Contains(strings.ToLower(findingsText), "detected keywords")

			if scenario.expectKeywords && !hasKeywords {
				t.Error("Expected keywords to be detected but none were found")
			}

			// All fixes should be properly formed
			for i, fix := range diagnosis.Fixes {
				if fix.ID == "" {
					t.Errorf("Fix %d has empty ID", i)
				}
				if len(fix.Commands) == 0 {
					t.Errorf("Fix %d has no commands", i)
				}
			}
		})
	}
}

func TestCustomDiagnosisFixUniqueness(t *testing.T) {
	// Test that fix IDs are unique within a diagnosis
	diagnosis := DiagnoseCustomIssue("My computer has boot and network problems")

	fixIDs := make(map[string]bool)
	for i, fix := range diagnosis.Fixes {
		if fixIDs[fix.ID] {
			t.Errorf("Duplicate fix ID found: %s at index %d", fix.ID, i)
		}
		fixIDs[fix.ID] = true
	}
}

func TestCustomDiagnosisRiskLevels(t *testing.T) {
	// Test that risk levels are appropriate
	diagnosis := DiagnoseCustomIssue("System problems")

	for _, fix := range diagnosis.Fixes {
		// Information gathering should be low risk
		if strings.Contains(fix.Description, "gather") || 
		   strings.Contains(fix.Description, "check") ||
		   strings.Contains(fix.Description, "examine") {
			if fix.RiskLevel.String() != "Low" {
				t.Errorf("Information gathering fix should be low risk: %s", fix.Title)
			}
		}

		// Service restarts should be medium risk
		if strings.Contains(fix.Description, "restart") || 
		   strings.Contains(fix.Description, "Restart") {
			if fix.RiskLevel.String() == "Low" {
				t.Errorf("Service restart should not be low risk: %s", fix.Title)
			}
		}
	}
}