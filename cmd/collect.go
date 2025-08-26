package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/detector"
	"github.com/redtriage/redtriage/internal/output"

	"github.com/redtriage/redtriage/packager"
	"github.com/redtriage/redtriage/reporter"
	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect artifacts and create triage bundle",
	Long: `Collect system artifacts, run detections, and package everything into a triage bundle.
This is the main command for incident response triage.`,
	Args: cobra.NoArgs,
	RunE: runCollect,
}

var (
	extendedCollection bool
	includeSpecific    []string
	excludeSpecific    []string
	compressionType    string
	createChecksums    bool
)

func init() {
	// Set consistent help template for the collect command
	collectCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	collectCmd.Flags().BoolVar(&extendedCollection, "extended", false, "Collect extended artifacts (more comprehensive)")
	collectCmd.Flags().StringSliceVar(&includeSpecific, "artifacts", nil, "Specific artifacts to collect")
	collectCmd.Flags().StringSliceVar(&excludeSpecific, "skip", nil, "Artifacts to skip")
	collectCmd.Flags().StringVar(&compressionType, "compression", "zip", "Compression type (zip, tar.gz, none)")
	collectCmd.Flags().BoolVar(&createChecksums, "checksums", true, "Create checksums for collected artifacts")
}

func runCollect(cmd *cobra.Command, args []string) error {
	// Initialize output manager
	outputDir := outputDir
	if outputDir == "" {
		outputDir = "./redtriage-output"
	}

	om, err := output.NewOutputManager("collect", outputDir, "console", verbose, jsonLogs)
	if err != nil {
		return fmt.Errorf("failed to initialize output manager: %w", err)
	}
	defer om.Close()

	// Validate inputs
	if err := validateCollectInputs(om); err != nil {
		om.LogError(err, "Input validation failed")
		om.PrintSummary()
		return err
	}

	om.LogInfo("Starting RedTriage collection...")

	// Initialize components
	collectorInstance := collector.NewCollector()
	if collectorInstance == nil {
		err := fmt.Errorf("failed to initialize collector")
		om.LogError(err, "Collector initialization failed")
		om.PrintSummary()
		return err
	}

	detectorInstance := detector.NewDetector()
	if detectorInstance == nil {
		err := fmt.Errorf("failed to initialize detector")
		om.LogError(err, "Detector initialization failed")
		om.PrintSummary()
		return err
	}

	packagerInstance := packager.NewPackager()
	if packagerInstance == nil {
		err := fmt.Errorf("failed to initialize packager")
		om.LogError(err, "Packager initialization failed")
		om.PrintSummary()
		return err
	}

	reporterInstance := reporter.NewReporter()
	if reporterInstance == nil {
		err := fmt.Errorf("failed to initialize reporter")
		om.LogError(err, "Reporter initialization failed")
		om.PrintSummary()
		return err
	}

	// Set collection profile
	profile := collector.CollectionProfile{
		Extended: extendedCollection,
		Timeout:  time.Duration(timeout) * time.Second,
		Include:  includeSpecific,
		Exclude:  excludeSpecific,
	}

	om.LogInfo("Collection profile: extended=%v, timeout=%s, include=%v, exclude=%v",
		extendedCollection, profile.Timeout, includeSpecific, excludeSpecific)

	// Collect artifacts
	om.LogInfo("Collecting artifacts...")
	results, err := collectorInstance.Collect(profile)
	if err != nil {
		om.LogError(err, "Collection failed")
		om.PrintSummary()
		return fmt.Errorf("collection failed: %w", err)
	}

	om.LogSuccess("Artifact collection completed successfully")
	om.LogInfo("Collected %d artifacts", len(results))

	// Count artifacts by category
	artifactCounts := make(map[string]int)
	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
			om.LogWarning("Failed to collect artifact %s: %v", result.Artifact.Name, result.Error)
			continue
		}
		artifactCounts[result.Artifact.Category]++
	}

	om.LogInfo("Artifact collection summary:")
	for category, count := range artifactCounts {
		om.LogInfo("  %s: %d artifacts", category, count)
	}
	if errorCount > 0 {
		om.LogWarning("  Failed: %d artifacts", errorCount)
	}

	// Run detections
	om.LogInfo("Running detections...")
	findings, err := detectorInstance.Evaluate(results)
	if err != nil {
		om.LogError(err, "Detection failed")
		om.PrintSummary()
		return fmt.Errorf("detection failed: %w", err)
	}

	om.LogSuccess("Detection analysis completed successfully")
	om.LogInfo("Found %d findings", len(findings))

	// Package results
	om.LogInfo("Packaging results...")
	bundlePath, err := packagerInstance.CreateBundle(results, findings, outputDir)
	if err != nil {
		om.LogError(err, "Packaging failed")
		om.PrintSummary()
		return fmt.Errorf("packaging failed: %w", err)
	}

	om.LogSuccess("Bundle creation completed successfully")
	om.LogInfo("Bundle created at: %s", bundlePath)

	// Generate reports
	om.LogInfo("Generating reports...")
	reports, err := reporterInstance.GenerateReports(results, findings, bundlePath)
	if err != nil {
		om.LogError(err, "Report generation failed")
		om.PrintSummary()
		return fmt.Errorf("report generation failed: %w", err)
	}

	om.LogSuccess("Report generation completed successfully")
	om.LogInfo("Reports generated: %v", reports)

	// Add final results
	om.AddResult(output.Result{
		Type:    "collection_summary",
		Status:  "success",
		Message: "Triage collection completed successfully",
		Data: map[string]interface{}{
			"total_artifacts":      len(results),
			"successful_artifacts": len(results) - errorCount,
			"failed_artifacts":     errorCount,
			"findings_count":       len(findings),
			"bundle_path":          bundlePath,
			"reports":              reports,
			"output_directory":     outputDir,
			"extended_collection":  extendedCollection,
			"timeout":              timeout,
		},
		Metadata: map[string]interface{}{
			"collection_mode": "full_triage",
			"compression":     compressionType,
			"checksums":       createChecksums,
		},
	})

	om.LogSuccess("Triage complete! Bundle created at: %s", bundlePath)
	om.LogInfo("Reports generated: %v", reports)

	om.PrintSummary()
	return nil
}

func validateCollectInputs(om *output.OutputManager) error {
	// Basic validation using simple approach

	// Validate output directory path if specified
	if outputDir != "" {
		// Basic path validation - prevent directory traversal
		if strings.Contains(outputDir, "..") || strings.Contains(outputDir, "//") {
			return fmt.Errorf("invalid output directory path: %s (contains invalid characters)", outputDir)
		}

		// Check if path is absolute and valid
		if filepath.IsAbs(outputDir) {
			if _, err := filepath.Abs(outputDir); err != nil {
				return fmt.Errorf("invalid absolute output directory path: %s", outputDir)
			}
		}
	}

	// Validate timeout
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", timeout)
	}

	// Validate compression type
	allowedCompression := []string{"zip", "tar.gz", "none"}
	compressionValid := false
	for _, comp := range allowedCompression {
		if compressionType == comp {
			compressionValid = true
			break
		}
	}
	if !compressionValid {
		return fmt.Errorf("invalid compression type '%s'. Must be one of: %s", compressionType, strings.Join(allowedCompression, ", "))
	}

	// Validate include artifacts (if specified)
	if len(includeSpecific) > 0 {
		for i, artifact := range includeSpecific {
			if artifact == "" || artifact == "help" {
				return fmt.Errorf("artifact name at index %d cannot be empty or 'help'. Must be a valid artifact name", i)
			}
			// Basic validation - prevent suspicious input
			if strings.Contains(artifact, "..") || strings.Contains(artifact, "//") {
				return fmt.Errorf("artifact name at index %d contains invalid characters: %s", i, artifact)
			}
		}
	}

	// Validate exclude artifacts (if specified)
	if len(excludeSpecific) > 0 {
		for i, artifact := range excludeSpecific {
			if artifact == "" || artifact == "help" {
				return fmt.Errorf("artifact name at index %d cannot be empty or 'help'. Must be a valid artifact name", i)
			}
			// Basic validation - prevent suspicious input
			if strings.Contains(artifact, "..") || strings.Contains(artifact, "//") {
				return fmt.Errorf("artifact name at index %d contains invalid characters: %s", i, artifact)
			}
		}
	}

	return nil
}
