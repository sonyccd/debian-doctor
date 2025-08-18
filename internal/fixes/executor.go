package fixes

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/debian-doctor/debian-doctor/pkg/config"
	"github.com/debian-doctor/debian-doctor/pkg/logger"
)

// Fix represents a system fix that can be executed
type Fix struct {
	ID          string   // Unique identifier for the fix
	Title       string   // Human-readable title
	Description string   // Detailed description of what the fix does
	Commands    []string // Shell commands to execute
	RequiresRoot bool    // Whether the fix requires root privileges
	Reversible  bool     // Whether the fix can be undone
	ReverseCommands []string // Commands to reverse the fix (if reversible)
	RiskLevel   RiskLevel // Risk assessment
}

// RiskLevel indicates the safety level of a fix
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

func (r RiskLevel) String() string {
	switch r {
		case RiskLow:
			return "Low"
		case RiskMedium:
			return "Medium"
		case RiskHigh:
			return "High"
		case RiskCritical:
			return "Critical"
	}
	return "Unknown"
}

func (r RiskLevel) Color() string {
	switch r {
		case RiskLow:
			return "green"
		case RiskMedium:
			return "yellow"
		case RiskHigh:
			return "red"
		case RiskCritical:
			return "magenta"
	}
	return "white"
}

// Executor handles the execution of system fixes
type Executor struct {
	config *config.Config
	logger *logger.Logger
}

// NewExecutor creates a new fix executor
func NewExecutor(cfg *config.Config, log *logger.Logger) *Executor {
	return &Executor{
		config: cfg,
		logger: log,
	}
}

// ExecuteFix executes a fix with user confirmation and safety checks
func (e *Executor) ExecuteFix(fix *Fix) error {
	// Validate fix
	if err := e.validateFix(fix); err != nil {
		return fmt.Errorf("fix validation failed: %w", err)
	}

	// Check permissions
	if fix.RequiresRoot && !e.config.IsRoot {
		return fmt.Errorf("fix '%s' requires root privileges", fix.Title)
	}

	// Show fix details and get confirmation
	if !e.config.NonInteractive {
		if !e.confirmExecution(fix) {
			e.logger.Info("Fix execution cancelled by user")
			return nil
		}
	}

	// Execute the fix
	e.logger.Info(fmt.Sprintf("Executing fix: %s", fix.Title))
	
	for i, cmd := range fix.Commands {
		e.logger.Info(fmt.Sprintf("Running command %d/%d: %s", i+1, len(fix.Commands), cmd))
		
		if err := e.executeCommand(cmd); err != nil {
			e.logger.Error(fmt.Sprintf("Command failed: %s", err))
			
			// If this is not the first command, offer to reverse
			if i > 0 && fix.Reversible {
				if e.offerReverse(fix, i) {
					e.reverseFix(fix, i-1)
				}
			}
			
			return fmt.Errorf("fix execution failed at command %d: %w", i+1, err)
		}
	}

	e.logger.Info(fmt.Sprintf("Fix '%s' executed successfully", fix.Title))
	return nil
}

// validateFix performs safety checks on a fix
func (e *Executor) validateFix(fix *Fix) error {
	if fix == nil {
		return fmt.Errorf("fix is nil")
	}
	
	if fix.Title == "" {
		return fmt.Errorf("fix title is required")
	}
	
	if len(fix.Commands) == 0 {
		return fmt.Errorf("fix has no commands")
	}

	// Check for dangerous commands
	dangerousPatterns := []string{
		"rm -rf /",
		"dd if=",
		"mkfs",
		"fdisk",
		"parted",
		"> /dev/",
	}

	for _, cmd := range fix.Commands {
		for _, pattern := range dangerousPatterns {
			if strings.Contains(strings.ToLower(cmd), pattern) {
				return fmt.Errorf("dangerous command detected: %s", cmd)
			}
		}
	}

	return nil
}

// confirmExecution shows fix details and asks for user confirmation
func (e *Executor) confirmExecution(fix *Fix) bool {
	fmt.Printf("\n🔧 Fix Details:\n")
	fmt.Printf("Title: %s\n", fix.Title)
	fmt.Printf("Description: %s\n", fix.Description)
	fmt.Printf("Risk Level: %s\n", fix.RiskLevel.String())
	fmt.Printf("Requires Root: %t\n", fix.RequiresRoot)
	fmt.Printf("Reversible: %t\n", fix.Reversible)
	
	fmt.Printf("\nCommands to execute:\n")
	for i, cmd := range fix.Commands {
		fmt.Printf("  %d. %s\n", i+1, cmd)
	}

	if fix.RiskLevel >= RiskHigh {
		fmt.Printf("\n⚠️  WARNING: This is a %s risk operation!\n", fix.RiskLevel.String())
		fmt.Printf("Please review the commands carefully before proceeding.\n")
	}

	fmt.Printf("\nDo you want to proceed? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	return response == "y" || response == "yes"
}

// executeCommand runs a single shell command
func (e *Executor) executeCommand(cmdStr string) error {
	// Split command into parts
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Run command
	err := cmd.Run()
	if err != nil {
		// Check if it's an exit error to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("command exited with code %d", status.ExitStatus())
			}
		}
		return err
	}

	return nil
}

// offerReverse asks if the user wants to reverse partially executed changes
func (e *Executor) offerReverse(fix *Fix, failedAt int) bool {
	if !fix.Reversible {
		return false
	}

	fmt.Printf("\n❌ Fix failed at step %d.\n", failedAt+1)
	fmt.Printf("This fix is reversible. Do you want to undo the changes made so far? (y/N): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	return response == "y" || response == "yes"
}

// reverseFix undoes changes made by a partially executed fix
func (e *Executor) reverseFix(fix *Fix, lastExecutedStep int) {
	e.logger.Info(fmt.Sprintf("Reversing fix '%s' up to step %d", fix.Title, lastExecutedStep+1))
	
	// Execute reverse commands in reverse order
	for i := lastExecutedStep; i >= 0; i-- {
		if i < len(fix.ReverseCommands) {
			cmd := fix.ReverseCommands[i]
			e.logger.Info(fmt.Sprintf("Reversing step %d: %s", i+1, cmd))
			
			if err := e.executeCommand(cmd); err != nil {
				e.logger.Error(fmt.Sprintf("Failed to reverse step %d: %s", i+1, err))
			}
		}
	}
	
	e.logger.Info("Fix reversal completed")
}

// GetCommonFixes returns a collection of commonly used fixes
func GetCommonFixes() map[string]*Fix {
	return map[string]*Fix{
		"update_package_cache": {
			ID:          "update_package_cache",
			Title:       "Update Package Cache",
			Description: "Updates the APT package cache to refresh available package information",
			Commands:    []string{"apt-get update"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   RiskLow,
		},
		"clean_package_cache": {
			ID:          "clean_package_cache", 
			Title:       "Clean Package Cache",
			Description: "Removes cached package files to free disk space",
			Commands:    []string{"apt-get clean", "apt-get autoclean"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   RiskLow,
		},
		"remove_orphaned_packages": {
			ID:          "remove_orphaned_packages",
			Title:       "Remove Orphaned Packages",
			Description: "Removes packages that were automatically installed but are no longer needed",
			Commands:    []string{"apt-get autoremove -y"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   RiskMedium,
		},
		"restart_networking": {
			ID:          "restart_networking",
			Title:       "Restart Network Service",
			Description: "Restarts the networking service to resolve connection issues",
			Commands:    []string{"systemctl restart networking"},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"systemctl start networking"},
			RiskLevel:   RiskMedium,
		},
		"flush_dns": {
			ID:          "flush_dns",
			Title:       "Flush DNS Cache",
			Description: "Clears the DNS resolver cache to fix name resolution issues",
			Commands:    []string{"systemctl restart systemd-resolved"},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{"systemctl start systemd-resolved"},
			RiskLevel:   RiskLow,
		},
		"fix_broken_packages": {
			ID:          "fix_broken_packages",
			Title:       "Fix Broken Packages",
			Description: "Attempts to fix broken package dependencies",
			Commands:    []string{"apt-get -f install", "dpkg --configure -a"},
			RequiresRoot: true,
			Reversible:  false,
			RiskLevel:   RiskMedium,
		},
		"create_swap_file": {
			ID:          "create_swap_file",
			Title:       "Create Swap File (1GB)",
			Description: "Creates a 1GB swap file to improve system performance when memory is low",
			Commands: []string{
				"fallocate -l 1G /swapfile",
				"chmod 600 /swapfile",
				"mkswap /swapfile",
				"swapon /swapfile",
				"echo '/swapfile none swap sw 0 0' >> /etc/fstab",
			},
			RequiresRoot: true,
			Reversible:  true,
			ReverseCommands: []string{
				"swapoff /swapfile",
				"rm /swapfile",
				"sed -i '/\\/swapfile/d' /etc/fstab",
			},
			RiskLevel: RiskMedium,
		},
	}
}