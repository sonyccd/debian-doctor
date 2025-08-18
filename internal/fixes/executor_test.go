package fixes

import (
	"testing"

	"github.com/debian-doctor/debian-doctor/pkg/config"
	"github.com/debian-doctor/debian-doctor/pkg/logger"
)

func TestNewExecutor(t *testing.T) {
	cfg := config.New()
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	
	executor := NewExecutor(cfg, log)
	
	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}
	
	if executor.config != cfg {
		t.Error("Executor config not set correctly")
	}
	
	if executor.logger != log {
		t.Error("Executor logger not set correctly")
	}
}

func TestValidateFix(t *testing.T) {
	cfg := config.New()
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	executor := NewExecutor(cfg, log)

	tests := []struct {
		name    string
		fix     *Fix
		wantErr bool
	}{
		{
			name:    "nil fix",
			fix:     nil,
			wantErr: true,
		},
		{
			name: "empty title",
			fix: &Fix{
				Title:    "",
				Commands: []string{"echo test"},
			},
			wantErr: true,
		},
		{
			name: "no commands",
			fix: &Fix{
				Title:    "Test Fix",
				Commands: []string{},
			},
			wantErr: true,
		},
		{
			name: "dangerous command - rm -rf /",
			fix: &Fix{
				Title:    "Dangerous Fix",
				Commands: []string{"rm -rf /"},
			},
			wantErr: true,
		},
		{
			name: "dangerous command - dd",
			fix: &Fix{
				Title:    "Dangerous Fix",
				Commands: []string{"dd if=/dev/zero of=/dev/sda"},
			},
			wantErr: true,
		},
		{
			name: "valid fix",
			fix: &Fix{
				Title:    "Valid Fix",
				Commands: []string{"echo 'hello world'"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateFix(tt.fix)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRiskLevel(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
		color    string
	}{
		{RiskLow, "Low", "green"},
		{RiskMedium, "Medium", "yellow"},
		{RiskHigh, "High", "red"},
		{RiskCritical, "Critical", "magenta"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("RiskLevel.String() = %v, want %v", got, tt.expected)
			}
			
			if got := tt.level.Color(); got != tt.color {
				t.Errorf("RiskLevel.Color() = %v, want %v", got, tt.color)
			}
		})
	}
}

func TestGetCommonFixes(t *testing.T) {
	fixes := GetCommonFixes()
	
	if len(fixes) == 0 {
		t.Fatal("GetCommonFixes returned empty map")
	}

	// Test some expected fixes
	expectedFixes := []string{
		"update_package_cache",
		"clean_package_cache",
		"remove_orphaned_packages",
		"restart_networking",
		"flush_dns",
		"fix_broken_packages",
		"create_swap_file",
	}

	for _, fixID := range expectedFixes {
		fix, exists := fixes[fixID]
		if !exists {
			t.Errorf("Expected fix '%s' not found", fixID)
			continue
		}

		// Validate fix structure
		if fix.ID != fixID {
			t.Errorf("Fix ID mismatch: got %s, want %s", fix.ID, fixID)
		}
		
		if fix.Title == "" {
			t.Errorf("Fix '%s' has empty title", fixID)
		}
		
		if fix.Description == "" {
			t.Errorf("Fix '%s' has empty description", fixID)
		}
		
		if len(fix.Commands) == 0 {
			t.Errorf("Fix '%s' has no commands", fixID)
		}
	}
}

func TestFixValidation(t *testing.T) {
	fixes := GetCommonFixes()
	cfg := config.New()
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	executor := NewExecutor(cfg, log)

	// All common fixes should pass validation
	for fixID, fix := range fixes {
		t.Run("validate_"+fixID, func(t *testing.T) {
			err := executor.validateFix(fix)
			if err != nil {
				t.Errorf("Common fix '%s' failed validation: %v", fixID, err)
			}
		})
	}
}

func TestExecuteCommandValidation(t *testing.T) {
	cfg := config.New()
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	executor := NewExecutor(cfg, log)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "valid command",
			command: "echo test",
			wantErr: false,
		},
		{
			name:    "command with args",
			command: "ls -la /tmp",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.executeCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteFixPermissions(t *testing.T) {
	cfg := config.New()
	cfg.SetNonInteractive(true) // Avoid prompts in tests
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	executor := NewExecutor(cfg, log)

	// Test fix that requires root when not root
	fix := &Fix{
		Title:        "Root Required Fix",
		Description:  "Test fix requiring root",
		Commands:     []string{"echo test"},
		RequiresRoot: true,
		RiskLevel:    RiskLow,
	}

	// This should fail if not running as root
	err = executor.ExecuteFix(fix)
	if !cfg.IsRoot && err == nil {
		t.Error("Expected error when running root-required fix without root privileges")
	}
	if cfg.IsRoot && err != nil {
		t.Errorf("Unexpected error when running as root: %v", err)
	}
}