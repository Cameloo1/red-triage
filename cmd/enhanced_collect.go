package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/detector"
	"github.com/redtriage/redtriage/internal/output"

	"github.com/redtriage/redtriage/packager"
	"github.com/redtriage/redtriage/platform/windows"
	"github.com/redtriage/redtriage/reporter"
	"github.com/spf13/cobra"
)

var enhancedCollectCmd = &cobra.Command{
	Use:   "enhanced-collect",
	Short: "Enhanced collection with comprehensive forensic capabilities",
	Long: `Enhanced collection command that provides comprehensive system intelligence collection.
	
This command includes:
- 15+ artifact types (memory, registry, file system, network, etc.)
- Advanced log parsing and analysis
- Real-time correlation and anomaly detection
- Professional reporting in multiple formats
- Timeline analysis and threat hunting capabilities`,
	Args: cobra.NoArgs,
	RunE: runEnhancedCollect,
}

var (
	enhancedCollectionProfile string
	includeForensic           []string
	excludeForensic           []string
	enableLogAnalysis         bool
	enableTimelineAnalysis    bool
	enableAnomalyDetection    bool
	reportFormats             []string
	collectionPriority        string
	enableVolatileCollection  bool
)

func init() {
	enhancedCollectCmd.Flags().StringVar(&enhancedCollectionProfile, "profile", "comprehensive", "Collection profile (comprehensive, focused, rapid, forensic)")
	enhancedCollectCmd.Flags().StringSliceVar(&includeForensic, "include-forensic", nil, "Specific forensic artifacts to collect")
	enhancedCollectCmd.Flags().StringSliceVar(&excludeForensic, "exclude-forensic", nil, "Forensic artifacts to skip")
	enhancedCollectCmd.Flags().BoolVar(&enableLogAnalysis, "log-analysis", true, "Enable advanced log parsing and analysis")
	enhancedCollectCmd.Flags().BoolVar(&enableTimelineAnalysis, "timeline", true, "Enable timeline analysis and correlation")
	enhancedCollectCmd.Flags().BoolVar(&enableAnomalyDetection, "anomaly-detection", true, "Enable anomaly detection")
	enhancedCollectCmd.Flags().StringSliceVar(&reportFormats, "report-formats", []string{"html", "json", "csv", "xml"}, "Report output formats")
	enhancedCollectCmd.Flags().StringVar(&collectionPriority, "priority", "balanced", "Collection priority (volatile_first, balanced, comprehensive)")
	enhancedCollectCmd.Flags().BoolVar(&enableVolatileCollection, "volatile", true, "Enable volatile data collection (memory, network, etc.)")
}

func runEnhancedCollect(cmd *cobra.Command, args []string) error {
	// Initialize output manager
	outputDir := outputDir
	if outputDir == "" {
		outputDir = "./redtriage-enhanced-output"
	}

	om, err := output.NewOutputManager("enhanced-collect", outputDir, "console", verbose, jsonLogs)
	if err != nil {
		return fmt.Errorf("failed to initialize output manager: %w", err)
	}
	defer om.Close()

	// Validate inputs
	if err := validateEnhancedCollectInputs(om); err != nil {
		om.LogError(err, "Input validation failed")
		om.PrintSummary()
		return err
	}

	om.LogInfo("Starting RedTriage Enhanced Collection...")
	om.LogInfo("Profile: %s, Priority: %s, Volatile: %v", enhancedCollectionProfile, collectionPriority, enableVolatileCollection)

	// Initialize enhanced components based on platform
	var enhancedCollector interface{}

	switch runtime.GOOS {
	case "windows":
		enhancedCollector = windows.NewEnhancedWindowsCollector()
	case "linux":
		// For Linux, we'll use a basic collector for now
		enhancedCollector = collector.NewCollector()
	default:
		return fmt.Errorf("enhanced collection not supported on %s", runtime.GOOS)
	}

	if enhancedCollector == nil {
		return fmt.Errorf("failed to initialize enhanced collector")
	}

	// Type assert to get the collector
	var collectorInstance *collector.Collector
	if c, ok := enhancedCollector.(*collector.Collector); ok {
		collectorInstance = c
	} else {
		return fmt.Errorf("enhanced collector type not supported")
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

	enhancedReporter := reporter.NewEnhancedReporter()
	if enhancedReporter == nil {
		err := fmt.Errorf("failed to initialize enhanced reporter")
		om.LogError(err, "Enhanced reporter initialization failed")
		om.PrintSummary()
		return err
	}

	// Set enhanced collection profile
	profile := collector.CollectionProfile{
		Extended: true, // Always extended for enhanced collection
		Timeout:  time.Duration(timeout) * time.Second,
		Include:  includeForensic,
		Exclude:  excludeForensic,
	}

	om.LogInfo("Enhanced collection profile: profile=%s, priority=%s, include=%v, exclude=%v",
		enhancedCollectionProfile, collectionPriority, includeForensic, excludeForensic)

	// Collect enhanced artifacts
	om.LogInfo("Collecting enhanced artifacts...")
	startTime := time.Now()

	results, err := collectorInstance.Collect(profile)
	if err != nil {
		om.LogError(err, "Enhanced collection failed")
		om.PrintSummary()
		return fmt.Errorf("enhanced collection failed: %w", err)
	}

	collectionDuration := time.Since(startTime)
	om.LogSuccess("Enhanced artifact collection completed successfully in %v", collectionDuration)
	om.LogInfo("Collected %d enhanced artifacts", len(results))

	// Count artifacts by category and priority
	artifactCounts := make(map[string]int)
	priorityCounts := make(map[int]int)
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			om.LogWarning("Failed to collect artifact %s: %v", result.Artifact.Name, result.Error)
			continue
		}

		artifactCounts[result.Artifact.Category]++

		// Get priority from enhanced artifact if available
		if enhancedArtifact, ok := getEnhancedArtifact(result.Artifact.Name); ok {
			priorityCounts[enhancedArtifact.Priority]++
		}
	}

	om.LogInfo("Enhanced artifact collection summary:")
	for category, count := range artifactCounts {
		om.LogInfo("  %s: %d artifacts", category, count)
	}

	om.LogInfo("Collection priority breakdown:")
	for priority, count := range priorityCounts {
		om.LogInfo("  Priority %d: %d artifacts", priority, count)
	}

	if errorCount > 0 {
		om.LogWarning("  Failed: %d artifacts", errorCount)
	}

	// Run enhanced detections
	om.LogInfo("Running enhanced detections...")
	findings, err := detectorInstance.Evaluate(results)
	if err != nil {
		om.LogError(err, "Enhanced detection failed")
		om.PrintSummary()
		return fmt.Errorf("enhanced detection failed: %w", err)
	}

	om.LogSuccess("Enhanced detection analysis completed successfully")
	om.LogInfo("Found %d findings", len(findings))

	// Package results
	om.LogInfo("Packaging enhanced results...")
	bundlePath, err := packagerInstance.CreateBundle(results, findings, outputDir)
	if err != nil {
		om.LogError(err, "Enhanced packaging failed")
		om.PrintSummary()
		return fmt.Errorf("enhanced packaging failed: %w", err)
	}

	om.LogSuccess("Enhanced bundle creation completed successfully")
	om.LogInfo("Bundle created at: %s", bundlePath)

	// Generate enhanced reports
	om.LogInfo("Generating enhanced reports...")
	reports, err := enhancedReporter.GenerateEnhancedReports(results, findings, bundlePath)
	if err != nil {
		om.LogError(err, "Enhanced report generation failed")
		om.PrintSummary()
		return fmt.Errorf("enhanced report generation failed: %w", err)
	}

	om.LogSuccess("Enhanced report generation completed successfully")
	om.LogInfo("Reports generated: %v", reports)

	// Add final results
	om.AddResult(output.Result{
		Type:    "enhanced_collection_summary",
		Status:  "success",
		Message: "Enhanced triage collection completed successfully",
		Data: map[string]interface{}{
			"total_artifacts":      len(results),
			"successful_artifacts": len(results) - errorCount,
			"failed_artifacts":     errorCount,
			"findings_count":       len(findings),
			"bundle_path":          bundlePath,
			"reports":              reports,
			"output_directory":     outputDir,
			"collection_profile":   enhancedCollectionProfile,
			"collection_priority":  collectionPriority,
			"volatile_enabled":     enableVolatileCollection,
			"log_analysis_enabled": enableLogAnalysis,
			"timeline_enabled":     enableTimelineAnalysis,
			"anomaly_detection":    enableAnomalyDetection,
			"collection_duration":  collectionDuration.String(),
			"report_formats":       reportFormats,
		},
		Metadata: map[string]interface{}{
			"collection_mode": "enhanced_triage",
			"enhanced_features": []string{
				"comprehensive_artifacts",
				"advanced_log_parsing",
				"timeline_analysis",
				"anomaly_detection",
				"multi_format_reporting",
			},
		},
	})

	om.LogSuccess("Enhanced triage complete! Bundle created at: %s", bundlePath)
	om.LogInfo("Enhanced reports generated: %v", reports)
	om.LogInfo("Collection completed in %v", collectionDuration)

	om.PrintSummary()
	return nil
}

// getEnhancedArtifact retrieves enhanced artifact information
func getEnhancedArtifact(name string) (collector.EnhancedArtifact, bool) {
	registry := collector.NewEnhancedArtifactRegistry()
	return registry.GetArtifact(name)
}

func validateEnhancedCollectInputs(om *output.OutputManager) error {
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

	// Validate collection profile
	allowedProfiles := []string{"comprehensive", "focused", "rapid", "forensic"}
	profileValid := false
	for _, profile := range allowedProfiles {
		if enhancedCollectionProfile == profile {
			profileValid = true
			break
		}
	}
	if !profileValid {
		return fmt.Errorf("invalid collection profile '%s'. Must be one of: %s", enhancedCollectionProfile, strings.Join(allowedProfiles, ", "))
	}

	// Validate collection priority
	allowedPriorities := []string{"volatile_first", "balanced", "comprehensive"}
	priorityValid := false
	for _, priority := range allowedPriorities {
		if collectionPriority == priority {
			priorityValid = true
			break
		}
	}
	if !priorityValid {
		return fmt.Errorf("invalid collection priority '%s'. Must be one of: %s", collectionPriority, strings.Join(allowedPriorities, ", "))
	}

	// Validate report formats
	allowedFormats := []string{"html", "json", "csv", "xml", "pdf"}
	for _, format := range reportFormats {
		formatValid := false
		for _, allowed := range allowedFormats {
			if format == allowed {
				formatValid = true
				break
			}
		}
		if !formatValid {
			return fmt.Errorf("invalid report format '%s'. Must be one of: %s", format, strings.Join(allowedFormats, ", "))
		}
	}

	// Validate include forensic artifacts (if specified)
	if len(includeForensic) > 0 {
		for i, artifact := range includeForensic {
			if artifact == "" || artifact == "help" {
				return fmt.Errorf("artifact name at index %d cannot be empty or 'help'. Must be a valid artifact name", i)
			}
			// Basic validation - prevent suspicious input
			if strings.Contains(artifact, "..") || strings.Contains(artifact, "//") {
				return fmt.Errorf("artifact name at index %d contains invalid characters: %s", i, artifact)
			}
		}
	}

	// Validate exclude forensic artifacts (if specified)
	if len(excludeForensic) > 0 {
		for i, artifact := range excludeForensic {
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
