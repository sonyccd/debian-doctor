package cmd

import (
	"fmt"
	"os"

	"github.com/debian-doctor/debian-doctor/internal/diagnose"
	"github.com/debian-doctor/debian-doctor/internal/tui"
	"github.com/debian-doctor/debian-doctor/pkg/config"
	"github.com/debian-doctor/debian-doctor/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	nonInteractive bool
	verbose        bool
	customIssue    string
)

var rootCmd = &cobra.Command{
	Use:   "debian-doctor",
	Short: "A comprehensive system diagnostic and troubleshooting tool for Debian-based systems",
	Long: `Debian Doctor performs automatic system health checks and provides 
interactive problem diagnosis with fix suggestions for Debian-based systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		if customIssue != "" {
			runCustomDiagnosis()
		} else if nonInteractive {
			runNonInteractiveMode()
		} else {
			runTUI()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Run in non-interactive mode")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&customIssue, "issue", "i", "", "Describe a custom issue for troubleshooting")
}

func runTUI() {
	// Set up configuration
	cfg := config.New()
	cfg.SetVerbose(verbose)
	cfg.SetNonInteractive(nonInteractive)
	
	// Set up logger
	log, err := logger.NewFromConfig(cfg)
	if err != nil {
		fmt.Printf("Error setting up logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()
	
	// Use simple text UI instead of Bubble Tea
	ui := tui.NewSimpleUI(cfg, log)
	if err := ui.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}
}

func runCustomDiagnosis() {
	fmt.Printf("CUSTOM ISSUE DIAGNOSIS\n")
	fmt.Printf("Issue: %s\n\n", customIssue)
	
	diagnosis := diagnose.DiagnoseCustomIssue(customIssue)
	
	// Display findings
	fmt.Println("ANALYSIS:")
	for _, finding := range diagnosis.Findings {
		fmt.Printf("  - %s\n", finding)
	}
	
	// Display troubleshooting suggestions
	fmt.Println("\nGENERAL TROUBLESHOOTING SUGGESTIONS:")
	suggestions := diagnose.GetTroubleshootingSuggestions()
	for i, suggestion := range suggestions {
		if i >= 5 { // Limit to first 5 suggestions
			break
		}
		fmt.Printf("  %d. %s\n", i+1, suggestion)
	}
	
	// Display fixes
	if len(diagnosis.Fixes) > 0 {
		fmt.Println("\nRECOMMENDED ACTIONS:")
		for i, fix := range diagnosis.Fixes {
			if i >= 10 { // Limit to first 10 fixes
				fmt.Printf("  ... and %d more (use interactive mode for full list)\n", len(diagnosis.Fixes)-10)
				break
			}
			fmt.Printf("\n  %d. %s\n", i+1, fix.Title)
			fmt.Printf("     %s\n", fix.Description)
			if len(fix.Commands) > 0 {
				fmt.Printf("     Command: %s\n", fix.Commands[0])
				if len(fix.Commands) > 1 {
					fmt.Printf("     (+ %d more commands)\n", len(fix.Commands)-1)
				}
			}
			fmt.Printf("     Risk Level: %s\n", fix.RiskLevel.String())
			if fix.RequiresRoot {
				fmt.Printf("     WARNING: Requires root privileges\n")
			}
		}
	}
	
	fmt.Println("\nTIP: Run 'debian-doctor' without flags for interactive mode with more options")
}

func runNonInteractiveMode() {
	fmt.Println("Running system checks...")
	// TODO: Implement non-interactive mode
}