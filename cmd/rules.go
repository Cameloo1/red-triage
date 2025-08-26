package cmd

import (
	"fmt"
	"strings"

	"github.com/redtriage/redtriage/detector"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage detection rules",
	Long: `Manage detection rule packs including listing, updating, and testing rules.
Supports both built-in heuristic rules and Sigma rules.`,
	Args: cobra.NoArgs,
	RunE: runRules,
}

var (
	rulesUpdate   bool
	rulesTest     bool
	rulesCategory string
)

func init() {
	rulesCmd.Flags().BoolVar(&rulesUpdate, "update", false, "Update rule packs from remote sources")
	rulesCmd.Flags().BoolVar(&rulesTest, "test", false, "Test rules against sample data")
	rulesCmd.Flags().StringVar(&rulesCategory, "category", "", "Filter rules by category")
}

func runRules(cmd *cobra.Command, args []string) error {
	// Validate inputs first
	if err := validateRulesInputs(); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	fmt.Println("Detection Rules Management")
	fmt.Println("=========================")

	// Initialize detector
	detector := detector.NewDetector()

	// List built-in rules
	fmt.Println("\nBuilt-in Rules:")
	fmt.Println("---------------")
	rules := detector.GetBuiltInRules()
	for i, rule := range rules {
		fmt.Printf("%d. %s (%s)\n", i+1, rule.Name, rule.Description)
		fmt.Printf("   Severity: %s\n", rule.Severity)
		fmt.Printf("   Category: %s\n", rule.Category)
		fmt.Println()
	}

	// Check Sigma rules if path provided
	if sigmaRules != "" {
		fmt.Printf("\nSigma Rules Path: %s\n", sigmaRules)
		fmt.Println("Note: Sigma rule integration is planned for future versions")
	} else {
		fmt.Println("\nNo Sigma rules path specified. Use --sigma-rules flag to specify a path.")
	}

	return nil
}

// validateRulesInputs validates all rules command inputs
func validateRulesInputs() error {
	// Validate category if specified
	if rulesCategory != "" {
		validCategories := []string{"process", "network", "file", "registry", "memory", "system", "malware", "persistence", "lateral_movement"}
		valid := false
		for _, c := range validCategories {
			if c == rulesCategory {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid category '%s'. Must be one of: %s", rulesCategory, strings.Join(validCategories, ", "))
		}
	}

	return nil
}
