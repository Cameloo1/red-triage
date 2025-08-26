package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/internal/output"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Collect and display host profile",
	Long: `Collect basic host profile information without creating a full archive or running detections.
Useful for quick system reconnaissance.`,
	Args: cobra.NoArgs,
	RunE: runProfile,
}

var (
	profileDetailed bool
	profileOutput   string
	profileFormat   string
)

func init() {
	profileCmd.Flags().BoolVar(&profileDetailed, "detailed", false, "Show detailed profile information")
	profileCmd.Flags().StringVar(&profileOutput, "output", "", "Output directory for profile data")
	profileCmd.Flags().StringVar(&profileFormat, "format", "text", "Output format (text, json, yaml)")
}

func runProfile(cmd *cobra.Command, args []string) error {
	// Initialize output manager
	outputDir := profileOutput
	if outputDir == "" {
		outputDir = "./redtriage-profile"
	}

	om, err := output.NewOutputManager("profile", outputDir, profileFormat, verbose, jsonLogs)
	if err != nil {
		return fmt.Errorf("failed to initialize output manager: %w", err)
	}
	defer om.Close()

	// Validate inputs
	if err := validateProfileInputs(om); err != nil {
		om.LogError(err, "Input validation failed")
		om.PrintSummary()
		return err
	}

	om.LogInfo("Starting host profile collection...")

	// Initialize collector
	collectorInstance := collector.NewCollector()
	if collectorInstance == nil {
		err := fmt.Errorf("failed to initialize collector")
		om.LogError(err, "Collector initialization failed")
		om.PrintSummary()
		return err
	}

	// Set collection profile for basic artifacts only
	profile := collector.CollectionProfile{
		Extended: profileDetailed,
		Timeout:  0, // No timeout for profile
	}

	om.LogInfo("Collecting host artifacts...")

	// Collect basic artifacts
	results, err := collectorInstance.Collect(profile)
	if err != nil {
		om.LogError(err, "Profile collection failed")
		om.PrintSummary()
		return fmt.Errorf("profile collection failed: %w", err)
	}

	om.LogSuccess("Profile collection completed successfully")

	// Display profile summary
	om.LogInfo("=== Host Profile Summary ===")
	hostArtifacts := 0
	systemArtifacts := 0
	networkArtifacts := 0
	processArtifacts := 0

	for _, result := range results {
		if result.Error != nil {
			om.LogWarning("Failed to collect artifact %s: %v", result.Artifact.Name, result.Error)
			continue
		}

		switch result.Artifact.Category {
		case "host":
			hostArtifacts++
			om.LogInfo(" %s: %s", result.Artifact.Name, result.Artifact.Description)
		case "system":
			systemArtifacts++
			if profileDetailed {
				om.LogInfo(" %s: %s", result.Artifact.Name, result.Artifact.Description)
			}
		case "network":
			networkArtifacts++
			if profileDetailed {
				om.LogInfo(" %s: %s", result.Artifact.Name, result.Artifact.Description)
			}
		case "process":
			processArtifacts++
			if profileDetailed {
				om.LogInfo(" %s: %s", result.Artifact.Name, result.Artifact.Description)
			}
		}
	}

	// Add summary results
	om.AddResult(output.Result{
		Type:    "summary",
		Status:  "success",
		Message: "Profile collection completed",
		Data: map[string]interface{}{
			"total_artifacts":   len(results),
			"host_artifacts":    hostArtifacts,
			"system_artifacts":  systemArtifacts,
			"network_artifacts": networkArtifacts,
			"process_artifacts": processArtifacts,
			"detailed_mode":     profileDetailed,
			"output_directory":  outputDir,
			"output_format":     profileFormat,
		},
		Metadata: map[string]interface{}{
			"collection_mode": "profile",
			"extended":        profileDetailed,
		},
	})

	om.LogSuccess("Profile complete! Collected %d artifacts", len(results))
	om.LogInfo("Host artifacts: %d, System artifacts: %d, Network artifacts: %d, Process artifacts: %d",
		hostArtifacts, systemArtifacts, networkArtifacts, processArtifacts)

	// Write output to file if requested
	if err := om.WriteOutput(); err != nil {
		om.LogWarning("Failed to write output file: %v", err)
	}

	om.PrintSummary()
	return nil
}

func validateProfileInputs(om *output.OutputManager) error {
	// Basic validation using simple approach

	// Validate output directory path if specified
	if profileOutput != "" {
		// Basic path validation - prevent directory traversal
		if strings.Contains(profileOutput, "..") || strings.Contains(profileOutput, "//") {
			return fmt.Errorf("invalid output directory path: %s (contains invalid characters)", profileOutput)
		}

		// Check if path is absolute and valid
		if filepath.IsAbs(profileOutput) {
			if _, err := filepath.Abs(profileOutput); err != nil {
				return fmt.Errorf("invalid absolute output directory path: %s", profileOutput)
			}
		}
	}

	// Validate format
	allowedFormats := []string{"text", "json", "yaml", "yml"}
	formatValid := false
	for _, format := range allowedFormats {
		if profileFormat == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("invalid format '%s'. Must be one of: %s", profileFormat, strings.Join(allowedFormats, ", "))
	}

	return nil
}
