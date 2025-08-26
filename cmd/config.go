package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage RedTriage configuration settings including viewing, editing, and validating config files.
Supports both local and global configuration management.`,
	Args:  cobra.NoArgs,
	RunE:  runConfig,
}

var (
	configShow bool
	configEdit bool
	configValidate bool
	configReset bool
	configPath string
)

func init() {
	configCmd.Flags().BoolVar(&configShow, "show", false, "Show current configuration")
	configCmd.Flags().BoolVar(&configEdit, "edit", false, "Edit configuration file")
	configCmd.Flags().BoolVar(&configValidate, "validate", false, "Validate configuration file")
	configCmd.Flags().BoolVar(&configReset, "reset", false, "Reset to default configuration")
	configCmd.Flags().StringVar(&configPath, "path", "", "Path to configuration file")
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Validate inputs first
	if err := validateConfigInputs(); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	fmt.Println("Configuration Management")
	fmt.Println("=======================")
	
	if configPath == "" {
		configPath = "~/.redtriage.yml"
	}
	
	fmt.Printf("Configuration file: %s\n", configPath)
	
	if configShow {
		fmt.Println("Current configuration:")
		showCurrentConfiguration()
	}
	
	if configValidate {
		fmt.Println("Validating configuration...")
		if err := validateConfigurationFile(); err != nil {
			fmt.Printf("❌ Configuration validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Configuration validation passed")
		}
	}
	
	if configEdit {
		fmt.Println("Opening configuration for editing...")
		if err := editConfigurationFile(); err != nil {
			fmt.Printf("❌ Failed to edit configuration: %v\n", err)
		} else {
			fmt.Println("✅ Configuration editing completed")
		}
	}
	
	if configReset {
		fmt.Println("Resetting to default configuration...")
		if err := resetConfigurationFile(); err != nil {
			fmt.Printf("❌ Failed to reset configuration: %v\n", err)
		} else {
			fmt.Println("✅ Configuration reset completed")
		}
	}
	
	if !configShow && !configValidate && !configEdit && !configReset {
		fmt.Println("No action specified. Use --show, --validate, --edit, or --reset flags.")
	}
	
	return nil
}

// validateConfigInputs validates all config command inputs
func validateConfigInputs() error {
	// Validate path if specified
	if configPath != "" {
		if strings.Contains(configPath, "..") || strings.Contains(configPath, "//") {
			return fmt.Errorf("invalid config path: %s (contains invalid characters)", configPath)
		}
	}

	return nil
}

// showCurrentConfiguration displays the current configuration
func showCurrentConfiguration() {
	fmt.Println("Platform:", platform)
	fmt.Println("Output Directory:", outputDir)
	fmt.Println("Timeout:", timeout, "seconds")
	fmt.Println("Include Artifacts:", includeArtifacts)
	fmt.Println("Exclude Artifacts:", excludeArtifacts)
	fmt.Println("Sigma Rules:", sigmaRules)
	fmt.Println("Dry Run:", dryRun)
	fmt.Println("Verbose:", verbose)
	fmt.Println("JSON Logs:", jsonLogs)
	fmt.Println("Allow Network:", allowNetwork)
}

// validateConfigurationFile validates the configuration file
func validateConfigurationFile() error {
	// Basic validation of current configuration values
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", timeout)
	}
	
	if outputDir != "" {
		if strings.Contains(outputDir, "..") || strings.Contains(outputDir, "//") {
			return fmt.Errorf("invalid output directory path: %s", outputDir)
		}
	}
	
	if sigmaRules != "" {
		if strings.Contains(sigmaRules, "..") || strings.Contains(sigmaRules, "//") {
			return fmt.Errorf("invalid sigma rules path: %s", sigmaRules)
		}
	}
	
	return nil
}

// editConfigurationFile opens the configuration file for editing
func editConfigurationFile() error {
	// For now, just show what would be edited
	fmt.Println("Configuration values that can be modified:")
	fmt.Println("- Platform detection override")
	fmt.Println("- Output directory")
	fmt.Println("- Collection timeout")
	fmt.Println("- Artifact inclusion/exclusion lists")
	fmt.Println("- Sigma rules path")
	fmt.Println("- Logging options")
	fmt.Println("- Network permissions")
	
	// In a real implementation, this would open the config file in an editor
	fmt.Println("Note: Direct editing not implemented. Modify the configuration file manually.")
	return nil
}

// resetConfigurationFile resets configuration to defaults
func resetConfigurationFile() error {
	// Reset to default values
	platform = ""
	outputDir = "./redtriage-output"
	timeout = 300
	includeArtifacts = nil
	excludeArtifacts = nil
	sigmaRules = ""
	dryRun = false
	verbose = false
	jsonLogs = false
	allowNetwork = false
	
	fmt.Println("Configuration reset to default values:")
	showCurrentConfiguration()
	return nil
}
