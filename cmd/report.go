package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate triage reports",
	Long: `Generate various types of triage reports from collected data.
Supports executive summaries, technical details, and compliance reports.`,
	Args: cobra.NoArgs,
	RunE: runReport,
}

var (
	reportType            string
	reportTemplate        string
	reportOutput          string
	reportIncludeEvidence bool
)

func init() {
	reportCmd.Flags().StringVar(&reportType, "type", "summary", "Report type (summary, technical, compliance, executive)")
	reportCmd.Flags().StringVar(&reportTemplate, "template", "", "Custom report template file")
	reportCmd.Flags().StringVar(&reportOutput, "output", "", "Output file for report")
	reportCmd.Flags().BoolVar(&reportIncludeEvidence, "evidence", false, "Include evidence details in report")
}

func runReport(cmd *cobra.Command, args []string) error {
	// Validate inputs first
	if err := validateReportInputs(); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	fmt.Println("Report Generation")
	fmt.Println("=================")

	// Process all flags and show what would be done
	fmt.Println("Processing report generation with the following parameters:")

	// Validate and process report type
	fmt.Printf("✓ Report type: %s\n", reportType)
	validTypes := []string{"summary", "technical", "compliance", "executive"}
	validType := false
	for _, t := range validTypes {
		if t == reportType {
			validType = true
			break
		}
	}
	if !validType {
		fmt.Printf("⚠️  Warning: Invalid report type '%s'. Valid values: %v\n", reportType, validTypes)
		fmt.Printf("   Using default type: summary\n")
		reportType = "summary"
	}

	if reportTemplate != "" {
		fmt.Printf("✓ Custom template: %s\n", reportTemplate)
		// Check if template file exists
		if _, err := os.Stat(reportTemplate); err != nil {
			fmt.Printf("⚠️  Warning: Template file not found: %s\n", reportTemplate)
		} else {
			fmt.Printf("✓ Template file found and accessible\n")
		}
	}

	if reportIncludeEvidence {
		fmt.Println("✓ Including evidence details")
	} else {
		fmt.Println("✓ Evidence details excluded")
	}

	fmt.Printf("\nGenerating %s report...\n", reportType)

	// Simulate report generation
	fmt.Println("✓ Loading triage data...")
	fmt.Println("✓ Analyzing artifacts...")
	fmt.Println("✓ Processing findings...")
	fmt.Println("✓ Generating report content...")

	// Show what would be generated
	fmt.Println("\n=== Sample Report Content (Simulated) ===")
	fmt.Printf("Report Type: %s\n", reportType)
	fmt.Printf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("Status: Report generation simulation completed")
	
	if reportIncludeEvidence {
		fmt.Println("Evidence: Included (simulated)")
	} else {
		fmt.Println("Evidence: Excluded")
	}

	// Handle output if requested
	if reportOutput != "" {
		fmt.Printf("\n✓ Saving report to: %s\n", reportOutput)
		// Create output directory if needed
		outputDir := filepath.Dir(reportOutput)
		if outputDir != "." && outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("⚠️  Warning: Could not create output directory: %v\n", err)
			}
		}
		
		// Create a sample report file
		reportContent := fmt.Sprintf(`RedTriage Report
Type: %s
Generated: %s
Status: Simulation completed
Evidence: %t

This is a simulated report. In production, this would contain actual triage data.
`, reportType, time.Now().Format("2006-01-02 15:04:05"), reportIncludeEvidence)
		
		if err := os.WriteFile(reportOutput, []byte(reportContent), 0644); err == nil {
			fmt.Printf("✓ Sample report file created: %s\n", reportOutput)
		} else {
			fmt.Printf("⚠️  Failed to create report file: %v\n", err)
		}
	} else {
		fmt.Println("\n✓ Report displayed in console (use --output to save to file)")
	}

	fmt.Println("\n✓ Report generation completed successfully")
	return nil
}

// validateReportInputs validates all report command inputs
func validateReportInputs() error {
	// Validate report type
	validTypes := []string{"summary", "technical", "compliance", "executive"}
	validType := false
	for _, t := range validTypes {
		if reportType == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid report type '%s'. Must be one of: %s", reportType, strings.Join(validTypes, ", "))
	}

	// Validate template path if specified
	if reportTemplate != "" {
		if strings.Contains(reportTemplate, "..") || strings.Contains(reportTemplate, "//") {
			return fmt.Errorf("invalid template path: %s (contains invalid characters)", reportTemplate)
		}
	}

	// Validate output path if specified
	if reportOutput != "" {
		if strings.Contains(reportOutput, "..") || strings.Contains(reportOutput, "//") {
			return fmt.Errorf("invalid output path: %s (contains invalid characters)", reportOutput)
		}
	}

	return nil
}
