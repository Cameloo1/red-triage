package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile          string
	platform         string
	outputDir        string
	timeout          int
	includeArtifacts []string
	excludeArtifacts []string
	sigmaRules       string
	dryRun           bool
	verbose          bool
	jsonLogs         bool
	allowNetwork     bool
)

var RootCmd = &cobra.Command{
	Use:   "RedTriage",
	Short: "RedTriage - Professional incident response triage tool",
	Long: `RedTriage is a professional, cross-platform incident response triage CLI tool.
It collects volatile and persistent artifacts, runs local detections, and packages
everything into a signed archive with a manifest and concise report.`,
	Version: "1.0.0",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if help was explicitly requested
		helpRequested := cmd.Flags().Changed("help")

		// Only show banner if not requesting help
		if !helpRequested {
			displayBanner()
		}

		// Always show help for root command
		cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate all persistent flags before any command runs
		return validatePersistentFlags()
	},
}

func displayBanner() {
	// Create color instances with better Windows compatibility
	redColor := color.New(color.FgRed, color.Bold)
	whiteColor := color.New(color.FgHiWhite, color.Bold)
	greyColor := color.New(color.FgHiBlack)

	// Grey line above
	greyColor.Println("┌─────────────────────────────────────────────────────────────────────────────┐")

	// Corrected REDTRIAGE ASCII Art - "RED" in red, "TRIAGE" in bright white with no spacing

	redColor.Print(" ██████╗ ███████╗██████╗ ")
	whiteColor.Println("████████╗██████╗ ██╗ █████╗ ███████╗███████╗")
	redColor.Print(" ██╔══██╗██╔════╝██╔══██║")
	whiteColor.Println("╚══██╔══╝██╔══██╗██║██╔══██╗██╔════╝██╔════╝")
	redColor.Print(" ██████╔╝█████╗  ██║  ██║")
	whiteColor.Println("   ██║   ██████╔╝██║███████║██║ ███╗█████╗ ")
	redColor.Print(" ██╔══██╗██╔══╝  ██║  ██║")
	whiteColor.Println("   ██║   ██╔══██╗██║██╔══██║██║  ██║██╔══╝")
	redColor.Print(" ██║  ██║███████╗██████╔╝")
	whiteColor.Println("   ██║   ██║  ██║██║██║  ██║███████║███████╗")
	redColor.Print(" ╚═╝  ╚═╝╚══════╝╚═════╝ ")
	whiteColor.Println("   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝╚══════╝")

	// Grey line below
	greyColor.Println("└─────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println()
}

func NewRootCmd() *cobra.Command {
	// Set a simple, clean help template for consistent formatting
	RootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	// Remove the usage template to avoid duplication
	RootCmd.SetUsageTemplate("")

	// Add a pre-run function to disable colors during help
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// If help is requested, disable color output for consistent formatting
		if cmd.Flags().Changed("help") || (len(args) > 0 && args[0] == "help") {
			// Disable color output during help display
			color.NoColor = true
		}
	}

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.redtriage.yml)")
	RootCmd.PersistentFlags().StringVar(&platform, "platform", "", "override platform detection (windows/linux)")
	RootCmd.PersistentFlags().StringVar(&outputDir, "output", "./redtriage-output", "output directory for triage results")
	RootCmd.PersistentFlags().IntVar(&timeout, "timeout", 300, "collection timeout in seconds")
	RootCmd.PersistentFlags().StringSliceVar(&includeArtifacts, "include", nil, "only collect specific artifacts")
	RootCmd.PersistentFlags().StringSliceVar(&excludeArtifacts, "exclude", nil, "exclude specific artifacts")
	RootCmd.PersistentFlags().StringVar(&sigmaRules, "sigma-rules", "", "path to Sigma rules directory")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be collected without actually collecting")
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	RootCmd.PersistentFlags().BoolVar(&jsonLogs, "json-logs", false, "output logs in JSON format")
	RootCmd.PersistentFlags().BoolVar(&allowNetwork, "allow-network", false, "allow network operations during collection")

	// Add subcommands
	RootCmd.AddCommand(collectCmd)
	RootCmd.AddCommand(enhancedCollectCmd)
	RootCmd.AddCommand(profileCmd)
	RootCmd.AddCommand(checkCmd)
	RootCmd.AddCommand(rulesCmd)
	RootCmd.AddCommand(findingsCmd)
	RootCmd.AddCommand(reportCmd)
	RootCmd.AddCommand(bundleCmd)
	RootCmd.AddCommand(verifyCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(diagCmd)
	RootCmd.AddCommand(healthCmd)

	return RootCmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		// Validate that the file exists
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: Config file not found: %s\n", cfgFile)
			fmt.Fprintf(os.Stderr, "Using default configuration...\n")
		}
	} else {
		// Search for config in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not determine home directory: %v\n", err)
			fmt.Fprintf(os.Stderr, "Using default configuration...\n")
			return
		}
		cfgFile = home + "/.redtriage.yml"

		// Check if config file exists
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Config file not found: %s\n", cfgFile)
			fmt.Fprintf(os.Stderr, "Using default configuration...\n")
		}
	}
}

// validatePersistentFlags validates all persistent flags before command execution
func validatePersistentFlags() error {
	// Validate platform flag
	if platform != "" {
		validPlatforms := []string{"windows", "linux"}
		valid := false
		for _, p := range validPlatforms {
			if platform == p {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid platform '%s'. Must be one of: %s", platform, strings.Join(validPlatforms, ", "))
		}
	}

	// Validate timeout
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", timeout)
	}

	// Validate output directory
	if outputDir != "" {
		if strings.Contains(outputDir, "..") || strings.Contains(outputDir, "//") {
			return fmt.Errorf("invalid output directory path: %s", outputDir)
		}
	}

	// Validate include/exclude artifacts
	if err := validateArtifactLists(includeArtifacts, "include"); err != nil {
		return err
	}
	if err := validateArtifactLists(excludeArtifacts, "exclude"); err != nil {
		return err
	}

	// Validate sigma rules path
	if sigmaRules != "" {
		if strings.Contains(sigmaRules, "..") || strings.Contains(sigmaRules, "//") {
			return fmt.Errorf("invalid sigma rules path: %s", sigmaRules)
		}
	}

	return nil
}

// validateArtifactLists validates artifact include/exclude lists
func validateArtifactLists(artifacts []string, flagName string) error {
	validArtifacts := []string{
		"processes", "services", "network", "logs", "files", "registry",
		"memory", "volatility", "timeline", "system", "users", "groups",
	}

	for _, artifact := range artifacts {
		// Check for empty or invalid values
		if artifact == "" || artifact == "help" {
			return fmt.Errorf("invalid artifact '%s' in %s flag. Must be one of: %s",
				artifact, flagName, strings.Join(validArtifacts, ", "))
		}

		valid := false
		for _, validArtifact := range validArtifacts {
			if artifact == validArtifact {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid artifact '%s' in %s flag. Must be one of: %s",
				artifact, flagName, strings.Join(validArtifacts, ", "))
		}
	}
	return nil
}
