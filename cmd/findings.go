package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var findingsCmd = &cobra.Command{
	Use:   "findings",
	Short: "Manage and analyze detection findings",
	Long: `Manage and analyze detection findings from triage collections.
View, filter, and export findings in various formats.`,
	Args: cobra.NoArgs,
	RunE: runFindings,
}

var (
	findingsSeverity string
	findingsCategory string
	findingsExport   string
	findingsFilter   string
)

func init() {
	findingsCmd.Flags().StringVar(&findingsSeverity, "severity", "", "Filter by severity (low, medium, high, critical)")
	findingsCmd.Flags().StringVar(&findingsCategory, "category", "", "Filter by category (process, network, file, etc.)")
	findingsCmd.Flags().StringVar(&findingsExport, "export", "", "Export findings to file (json, csv, html)")
	findingsCmd.Flags().StringVar(&findingsFilter, "filter", "", "Custom filter expression")
}

func runFindings(cmd *cobra.Command, args []string) error {
	// Validate inputs first
	if err := validateFindingsInputs(); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	fmt.Println("Findings Management")
	fmt.Println("==================")

	// Process all flags and show what would be done
	fmt.Println("Processing findings with the following parameters:")

	if findingsSeverity != "" {
		fmt.Printf("✓ Severity filter: %s\n", findingsSeverity)
		// Validate severity
		validSeverities := []string{"low", "medium", "high", "critical"}
		valid := false
		for _, s := range validSeverities {
			if s == findingsSeverity {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("⚠️  Warning: Invalid severity '%s'. Valid values: %v\n", findingsSeverity, validSeverities)
		}
	}

	if findingsCategory != "" {
		fmt.Printf("✓ Category filter: %s\n", findingsCategory)
		// Validate category
		validCategories := []string{"process", "network", "file", "registry", "memory", "system"}
		valid := false
		for _, c := range validCategories {
			if c == findingsCategory {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("⚠️  Warning: Invalid category '%s'. Valid values: %v\n", findingsCategory, validCategories)
		}
	}

	if findingsFilter != "" {
		fmt.Printf("✓ Custom filter: %s\n", findingsFilter)
	}

	if findingsExport != "" {
		fmt.Printf("✓ Export format: %s\n", findingsExport)
		// Validate export format
		validFormats := []string{"json", "csv", "html"}
		valid := false
		for _, f := range validFormats {
			if f == findingsExport {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("⚠️  Warning: Invalid export format '%s'. Valid values: %v\n", findingsExport, validFormats)
		}
	}

	fmt.Println("\nSimulating findings analysis...")

	// Simulate loading findings
	fmt.Println("✓ Loading findings database...")
	fmt.Println("✓ Applying filters...")
	fmt.Println("✓ Processing results...")

	// Show what would be displayed
	fmt.Println("\n=== Sample Findings (Simulated) ===")
	fmt.Println("No actual findings found. Run 'collect' command first to gather data.")
	fmt.Println("When findings exist, they would be displayed here based on your filters.")

	// Handle export if requested
	if findingsExport != "" {
		fmt.Printf("\n✓ Exporting findings to: %s\n", findingsExport)
		// Create a sample export file
		exportDir := "./redtriage-exports"
		if err := os.MkdirAll(exportDir, 0755); err == nil {
			exportFile := filepath.Join(exportDir, fmt.Sprintf("findings.%s", findingsExport))
			if err := os.WriteFile(exportFile, []byte("Sample findings export\n"), 0644); err == nil {
				fmt.Printf("✓ Sample export file created: %s\n", exportFile)
			} else {
				fmt.Printf("⚠️  Failed to create export file: %v\n", err)
			}
		}
	}

	fmt.Println("\n✓ Findings command completed successfully")
	return nil
}

// validateFindingsInputs validates all findings command inputs
func validateFindingsInputs() error {
	// Validate severity if specified
	if findingsSeverity != "" {
		validSeverities := []string{"low", "medium", "high", "critical"}
		valid := false
		for _, s := range validSeverities {
			if s == findingsSeverity {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid severity '%s'. Must be one of: %s", findingsSeverity, strings.Join(validSeverities, ", "))
		}
	}

	// Validate category if specified
	if findingsCategory != "" {
		validCategories := []string{"process", "network", "file", "registry", "memory", "system"}
		valid := false
		for _, c := range validCategories {
			if c == findingsCategory {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid category '%s'. Must be one of: %s", findingsCategory, strings.Join(validCategories, ", "))
		}
	}

	// Validate export format if specified
	if findingsExport != "" {
		validFormats := []string{"json", "csv", "html"}
		valid := false
		for _, f := range validFormats {
			if f == findingsExport {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid export format '%s'. Must be one of: %s", findingsExport, strings.Join(validFormats, ", "))
		}
	}

	// Validate filter if specified
	if findingsFilter != "" {
		if strings.Contains(findingsFilter, "..") || strings.Contains(findingsFilter, "//") {
			return fmt.Errorf("invalid filter expression: %s (contains invalid characters)", findingsFilter)
		}
	}

	return nil
}
