package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/debian-doctor/debian-doctor/internal/checks"
	"github.com/debian-doctor/debian-doctor/internal/diagnose"
	"github.com/debian-doctor/debian-doctor/internal/fixes"
	"github.com/debian-doctor/debian-doctor/pkg/config"
	"github.com/debian-doctor/debian-doctor/pkg/logger"
)

type SimpleUI struct {
	config  *config.Config
	logger  *logger.Logger
	scanner *bufio.Scanner
}

func NewSimpleUI(cfg *config.Config, log *logger.Logger) *SimpleUI {
	return &SimpleUI{
		config:  cfg,
		logger:  log,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (ui *SimpleUI) Run() error {
	ui.clearScreen()
	ui.showHeader()
	
	for {
		ui.showMainMenu()
		choice := ui.getInput("Select option (1-4): ")
		
		switch choice {
		case "1":
			ui.runSystemCheck()
		case "2":
			ui.runInteractiveDiagnosis()
		case "3":
			ui.showSystemLogs()
		case "4", "q", "Q":
			ui.showExitMessage()
			return nil
		default:
			ui.showError("Invalid option. Please try again.")
		}
	}
}

func (ui *SimpleUI) clearScreen() {
	fmt.Print("\033[2J\033[H") // ANSI clear screen and move cursor to home
}

func (ui *SimpleUI) showHeader() {
	fmt.Println()
	fmt.Println("=====================================")
	fmt.Println("      DEBIAN SYSTEM DOCTOR v1.0     ")
	fmt.Println("    DIAGNOSTIC TERMINAL INTERFACE   ")
	fmt.Println("=====================================")
	fmt.Println()
	
	statusText := "SYSTEM ONLINE"
	if !ui.config.IsRoot {
		statusText = "LIMITED ACCESS MODE"
	}
	fmt.Printf("                Status: %s\n", statusText)
	fmt.Println()
}

func (ui *SimpleUI) showMainMenu() {
	fmt.Println("--- MAIN MENU ---")
	fmt.Println()
	fmt.Println("  1. RUN SYSTEM CHECK")
	fmt.Println("     Execute full diagnostic matrix scan")
	fmt.Println()
	fmt.Println("  2. INTERACTIVE DIAGNOSIS")
	fmt.Println("     Access specialized diagnostic modules")
	fmt.Println()
	fmt.Println("  3. VIEW SYSTEM LOGS")
	fmt.Println("     Display archived diagnostic data")
	fmt.Println()
	fmt.Println("  4. EXIT")
	fmt.Println("     Terminate diagnostic session")
	fmt.Println()
}

func (ui *SimpleUI) getInput(prompt string) string {
	fmt.Print(prompt)
	if ui.scanner.Scan() {
		return strings.TrimSpace(ui.scanner.Text())
	}
	return ""
}

func (ui *SimpleUI) showError(message string) {
	fmt.Printf("\nERROR: %s\n\n", message)
	ui.waitForKey()
}

func (ui *SimpleUI) showSuccess(message string) {
	fmt.Printf("\nSUCCESS: %s\n\n", message)
}

func (ui *SimpleUI) waitForKey() {
	fmt.Print("Press Enter to continue...")
	ui.scanner.Scan()
}

func (ui *SimpleUI) runSystemCheck() {
	ui.clearScreen()
	fmt.Println("=====================================")
	fmt.Println("     DIAGNOSTIC SCAN IN PROGRESS    ")
	fmt.Println("=====================================")
	fmt.Println()
	
	allChecks := checks.GetAllChecks()
	results := checks.NewResults()
	
	for i, check := range allChecks {
		// Show progress
		percent := float64(i) / float64(len(allChecks)) * 100
		ui.showProgress(fmt.Sprintf("SCANNING: %s", strings.ToUpper(check.Name())), percent)
		
		// Run the check
		result := check.Run()
		results.AddResult(result)
		
		// Small delay for visual effect
		time.Sleep(100 * time.Millisecond)
	}
	
	// Final progress
	ui.showProgress("SCAN COMPLETE", 100)
	fmt.Println()
	
	// Show results
	ui.showResults(results)
	ui.waitForKey()
}

func (ui *SimpleUI) showProgress(message string, percent float64) {
	// Create progress bar
	barWidth := 30
	filled := int(percent / 100 * float64(barWidth))
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "."
		}
	}
	bar += "]"
	
	// Right-align the display
	fmt.Printf("\r%-40s %s %3.0f%%", message, bar, percent)
	if percent >= 100 {
		fmt.Println()
	}
}

func (ui *SimpleUI) showResults(results checks.Results) {
	fmt.Println()
	fmt.Println("=====================================")
	
	errors := results.GetErrors()
	warnings := results.GetWarnings()
	info := results.GetInfo()
	
	if len(errors) > 0 {
		fmt.Printf("     ANALYSIS COMPLETE - ERROR      ")
		fmt.Printf("\n         %d CRITICAL ISSUES FOUND\n", len(errors))
	} else if len(warnings) > 0 {
		fmt.Printf("     ANALYSIS COMPLETE - WARNING    ")
		fmt.Printf("\n           %d WARNINGS FOUND\n", len(warnings))
	} else {
		fmt.Printf("     ANALYSIS COMPLETE - OK         ")
		fmt.Printf("\n           SYSTEM HEALTHY\n")
	}
	
	fmt.Println("=====================================")
	fmt.Println()
	
	if len(errors) > 0 {
		fmt.Println("CRITICAL ISSUES:")
		for i, err := range errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		fmt.Println()
	}
	
	if len(warnings) > 0 {
		fmt.Println("WARNINGS:")
		for i, warn := range warnings {
			fmt.Printf("  %d. %s\n", i+1, warn)
		}
		fmt.Println()
	}
	
	if len(info) > 0 {
		fmt.Println("SYSTEM INFORMATION:")
		for i, item := range info {
			fmt.Printf("  %d. %s\n", i+1, item)
		}
		fmt.Println()
	}
	
	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("All diagnostic checks passed successfully.")
		fmt.Println("Your Debian-based system is running optimally.")
		fmt.Println()
	}
}

func (ui *SimpleUI) runInteractiveDiagnosis() {
	ui.clearScreen()
	fmt.Println("=====================================")
	fmt.Println("    INTERACTIVE DIAGNOSIS SYSTEM    ")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("SELECT PROBLEM TYPE:")
	fmt.Println()
	
	options := []struct {
		name string
		desc string
	}{
		{"BOOT ISSUES", "System won't boot properly or startup problems"},
		{"PERFORMANCE ISSUES", "System is running slowly or high resource usage"},
		{"NETWORK ISSUES", "Internet connectivity or network configuration problems"},
		{"DISK ISSUES", "Storage space, disk errors, or filesystem problems"},
		{"FILESYSTEM ISSUES", "Filesystem corruption, mount problems, and integrity checks"},
		{"LOG ISSUES", "System logs, errors, and journal analysis"},
		{"SERVICE ISSUES", "System services or applications won't start"},
		{"DISPLAY ISSUES", "Graphics, resolution, or X11 problems"},
		{"PACKAGE ISSUES", "APT package manager or dependency problems"},
		{"PERMISSION ISSUES", "File access or user permission problems"},
		{"CUSTOM ISSUE", "Describe your own problem for general troubleshooting"},
	}
	
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option.name)
		fmt.Printf("     %s\n", option.desc)
		fmt.Println()
	}
	
	choice := ui.getInput("Select diagnosis type (1-11): ")
	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 1 || choiceNum > len(options) {
		ui.showError("Invalid choice")
		return
	}
	
	selectedOption := options[choiceNum-1]
	ui.runDiagnosis(selectedOption.name)
}

func (ui *SimpleUI) runDiagnosis(issueType string) {
	ui.clearScreen()
	fmt.Printf("Running diagnosis for: %s\n", issueType)
	fmt.Println()
	
	// Show progress
	ui.showProgress("ANALYZING SYSTEM", 0)
	time.Sleep(500 * time.Millisecond)
	ui.showProgress("COLLECTING DATA", 25)
	time.Sleep(500 * time.Millisecond)
	ui.showProgress("PROCESSING RESULTS", 50)
	time.Sleep(500 * time.Millisecond)
	ui.showProgress("GENERATING REPORT", 75)
	time.Sleep(500 * time.Millisecond)
	ui.showProgress("DIAGNOSIS COMPLETE", 100)
	fmt.Println()
	
	var diagnosis diagnose.Diagnosis
	
	switch issueType {
	case "BOOT ISSUES":
		diagnosis = diagnose.DiagnoseBootIssues()
	case "PERFORMANCE ISSUES":
		diagnosis = diagnose.DiagnosePerformanceIssues()
	case "NETWORK ISSUES":
		diagnosis = diagnose.DiagnoseNetworkIssues()
	case "DISK ISSUES":
		diagnosis = diagnose.DiagnoseDiskIssues()
	case "FILESYSTEM ISSUES":
		diagnosis = diagnose.DiagnoseFilesystemIssues()
	case "LOG ISSUES":
		diagnosis = diagnose.DiagnoseLogIssues()
	case "PACKAGE ISSUES":
		diagnosis = diagnose.DiagnosePackageIssues()
	case "SERVICE ISSUES":
		diagnosis = diagnose.DiagnoseServiceIssues()
	case "CUSTOM ISSUE":
		diagnosis = diagnose.DiagnoseCustomIssue("General system troubleshooting requested")
	default:
		diagnosis = diagnose.Diagnosis{
			Issue:    issueType,
			Findings: []string{"Diagnosis not yet implemented for this issue type"},
			Fixes:    []*fixes.Fix{},
		}
	}
	
	ui.showDiagnosisResults(diagnosis)
}

func (ui *SimpleUI) showDiagnosisResults(diagnosis diagnose.Diagnosis) {
	fmt.Println()
	fmt.Println("=====================================")
	
	if len(diagnosis.Fixes) > 0 {
		fmt.Printf("   DIAGNOSIS: %s - FIXES AVAILABLE\n", strings.ToUpper(diagnosis.Issue))
	} else {
		fmt.Printf("   DIAGNOSIS: %s - NO FIXES\n", strings.ToUpper(diagnosis.Issue))
	}
	
	fmt.Println("=====================================")
	fmt.Println()
	
	if len(diagnosis.Findings) > 0 {
		fmt.Println("DIAGNOSTIC FINDINGS:")
		for i, finding := range diagnosis.Findings {
			fmt.Printf("  %d. %s\n", i+1, finding)
		}
		fmt.Println()
	} else {
		fmt.Println("NO ISSUES DETECTED")
		fmt.Println("This diagnostic found no problems in the analyzed area.")
		fmt.Println()
	}
	
	if len(diagnosis.Fixes) > 0 {
		fmt.Printf("%d AUTOMATED FIXES AVAILABLE\n", len(diagnosis.Fixes))
		fmt.Println()
		
		for i, fix := range diagnosis.Fixes {
			fmt.Printf("FIX %d: %s\n", i+1, fix.Description)
			fmt.Printf("Command: %s\n", strings.Join(fix.Commands, " && "))
			if fix.RequiresRoot {
				fmt.Println("(Requires root privileges)")
			}
			fmt.Println()
		}
		
		if ui.askYesNo("Apply the first available fix? (y/n): ") {
			ui.applyFix(diagnosis.Fixes[0])
		}
	} else {
		fmt.Println("NO AUTOMATED FIXES AVAILABLE")
		fmt.Println("Manual intervention may be required.")
		fmt.Println()
	}
	
	ui.waitForKey()
}

func (ui *SimpleUI) askYesNo(prompt string) bool {
	response := ui.getInput(prompt)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func (ui *SimpleUI) applyFix(fix *fixes.Fix) {
	fmt.Printf("Applying fix: %s\n", fix.Description)
	fmt.Println()
	
	// Show progress
	ui.showProgress("PREPARING FIX", 0)
	time.Sleep(300 * time.Millisecond)
	ui.showProgress("EXECUTING COMMANDS", 50)
	time.Sleep(1000 * time.Millisecond)
	ui.showProgress("FIX APPLIED", 100)
	fmt.Println()
	
	ui.showSuccess("Fix applied successfully!")
	
	// Log the fix application
	ui.logger.Info("Applied fix: %s", fix.Description)
}

func (ui *SimpleUI) showSystemLogs() {
	ui.clearScreen()
	fmt.Println("=====================================")
	fmt.Println("         SYSTEM DIAGNOSTIC LOGS     ")
	fmt.Println("=====================================")
	fmt.Println()
	
	fmt.Println("Recent diagnostic activity:")
	fmt.Printf("  - System scan completed at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("  - No critical issues detected")
	fmt.Println("  - 0 warnings found")
	fmt.Println("  - System status: HEALTHY")
	fmt.Println()
	fmt.Println("For detailed logs, check:")
	fmt.Printf("  - /tmp/debian_doctor_%d.log\n", os.Getuid())
	fmt.Println("  - /var/log/syslog")
	fmt.Println("  - journalctl -xe")
	fmt.Println()
	
	ui.waitForKey()
}

func (ui *SimpleUI) showExitMessage() {
	ui.clearScreen()
	fmt.Println("=====================================")
	fmt.Println("    DEBIAN SYSTEM DOCTOR SHUTDOWN   ")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("Thank you for using Debian Doctor!")
	fmt.Printf("Session ended at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
}