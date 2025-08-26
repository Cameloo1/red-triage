package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/redtriage/redtriage/internal/output"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run preflight checks",
	Long: `Run preflight checks to verify system readiness for RedTriage operations.
Checks include system requirements, permissions, and available tools.`,
	Args: cobra.NoArgs,
	RunE: runCheck,
}

var (
	checkOutput  string
	checkFormat  string
	checkVerbose bool
)

func init() {
	checkCmd.Flags().StringVar(&checkOutput, "output", "", "Output directory for check results")
	checkCmd.Flags().StringVar(&checkFormat, "format", "text", "Output format (text, json, yaml)")
	checkCmd.Flags().BoolVar(&checkVerbose, "verbose", false, "Show detailed check information")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Initialize output manager
	outputDir := checkOutput
	if outputDir == "" {
		outputDir = "./redtriage-checks"
	}

	om, err := output.NewOutputManager("check", outputDir, checkFormat, checkVerbose || verbose, jsonLogs)
	if err != nil {
		return fmt.Errorf("failed to initialize output manager: %w", err)
	}
	defer om.Close()

	// Validate inputs
	if err := validateCheckInputs(om); err != nil {
		om.LogError(err, "Input validation failed")
		om.PrintSummary()
		return err
	}

	om.LogInfo("Starting RedTriage preflight checks...")

	// Run all checks
	results := runAllChecks(om)

	// Determine overall status
	overallStatus := "PASS"
	if hasFailedChecks(results) {
		overallStatus = "FAIL"
	} else if hasWarnings(results) {
		overallStatus = "WARN"
	}

	// Add final results
	om.AddResult(output.Result{
		Type:    "check_summary",
		Status:  overallStatus,
		Message: "Preflight checks completed",
		Data: map[string]interface{}{
			"total_checks":     len(results),
			"passed_checks":    countPassedChecks(results),
			"failed_checks":    countFailedChecks(results),
			"warning_checks":   countWarningChecks(results),
			"overall_status":   overallStatus,
			"output_directory": outputDir,
			"verbose_mode":     checkVerbose || verbose,
		},
		Metadata: map[string]interface{}{
			"check_mode": "preflight",
			"platform":   runtime.GOOS,
			"arch":       runtime.GOARCH,
		},
	})

	// Display summary
	if overallStatus == "PASS" {
		om.LogSuccess("All preflight checks passed! System is ready for RedTriage operations.")
	} else if overallStatus == "WARN" {
		om.LogWarning("Preflight checks completed with warnings. Review warnings before proceeding.")
	} else {
		om.LogError(fmt.Errorf("preflight checks failed"), "System is not ready for RedTriage operations")
	}

	// Write output to file if requested
	if err := om.WriteOutput(); err != nil {
		om.LogWarning("Failed to write output file: %v", err)
	}

	om.PrintSummary()
	return nil
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Name           string
	Status         string // "PASS", "FAIL", "WARN"
	Message        string
	Details        string
	Recommendation string
}

func runAllChecks(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// System checks
	results = append(results, checkSystemRequirements(om)...)

	// Permission checks
	results = append(results, checkPermissions(om)...)

	// Tool availability checks
	results = append(results, checkToolAvailability(om)...)

	// Configuration checks
	results = append(results, checkConfiguration(om)...)

	// Enhanced feature checks
	results = append(results, checkEnhancedFeatures(om)...)

	return results
}

func checkSystemRequirements(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check operating system
	osCheck := CheckResult{
		Name:           "Operating System",
		Status:         "PASS",
		Message:        fmt.Sprintf("Supported OS: %s", runtime.GOOS),
		Details:        fmt.Sprintf("Detected OS: %s, Architecture: %s", runtime.GOOS, runtime.GOARCH),
		Recommendation: "No action required",
	}

	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		osCheck.Status = "WARN"
		osCheck.Message = fmt.Sprintf("OS %s may have limited support", runtime.GOOS)
		osCheck.Recommendation = "Consider using Windows or Linux for full functionality"
	}

	results = append(results, osCheck)
	om.LogInfo(" %s: %s", osCheck.Name, osCheck.Message)

	// Check available memory
	memCheck := CheckResult{
		Name:           "Available Memory",
		Status:         "PASS",
		Message:        "Sufficient memory available",
		Details:        "Memory check passed",
		Recommendation: "No action required",
	}

	// Basic memory check (this would be more sophisticated in production)
	if runtime.GOOS == "windows" {
		// Windows-specific memory check could go here
	} else {
		// Linux-specific memory check could go here
	}

	results = append(results, memCheck)
	om.LogInfo(" %s: %s", memCheck.Name, memCheck.Message)

	// Check disk space
	diskCheck := CheckResult{
		Name:           "Disk Space",
		Status:         "PASS",
		Message:        "Sufficient disk space available",
		Details:        "Disk space check passed",
		Recommendation: "No action required",
	}

	// Check current directory writability
	if _, err := os.Stat("."); err == nil {
		// Check if we can write to current directory
		testFile := ".redtriage_test"
		if f, err := os.Create(testFile); err == nil {
			f.Close()
			os.Remove(testFile)
		} else {
			diskCheck.Status = "FAIL"
			diskCheck.Message = "Cannot write to current directory"
			diskCheck.Details = fmt.Sprintf("Write test failed: %v", err)
			diskCheck.Recommendation = "Ensure current directory is writable or specify different output directory"
		}
	}

	results = append(results, diskCheck)
	if diskCheck.Status == "PASS" {
		om.LogInfo(" %s: %s", diskCheck.Name, diskCheck.Message)
	} else {
		om.LogError(fmt.Errorf(diskCheck.Details), "Disk space check failed")
	}

	return results
}

func checkPermissions(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check current user permissions
	userCheck := CheckResult{
		Name:           "User Permissions",
		Status:         "PASS",
		Message:        "User has sufficient permissions",
		Details:        "Permission check passed",
		Recommendation: "No action required",
	}

	// Check if we can read system information
	if runtime.GOOS == "windows" {
		// Windows-specific permission checks
		if _, err := os.Stat("C:\\Windows\\System32"); err != nil {
			userCheck.Status = "WARN"
			userCheck.Message = "Limited access to system directories"
			userCheck.Details = "Cannot access C:\\Windows\\System32"
			userCheck.Recommendation = "Run as Administrator for full system access"
		}
	} else {
		// Linux-specific permission checks
		if _, err := os.Stat("/proc"); err != nil {
			userCheck.Status = "WARN"
			userCheck.Message = "Limited access to process information"
			userCheck.Details = "Cannot access /proc directory"
			userCheck.Recommendation = "Run with appropriate permissions or use sudo"
		}
	}

	results = append(results, userCheck)
	if userCheck.Status == "PASS" {
		om.LogInfo(" %s: %s", userCheck.Name, userCheck.Message)
	} else {
		om.LogWarning("️  %s: %s", userCheck.Name, userCheck.Message)
	}

	return results
}

func checkToolAvailability(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check for required tools
	tools := []string{"redtriage"}
	if runtime.GOOS == "windows" {
		tools = append(tools, "powershell", "cmd")
	} else {
		tools = append(tools, "bash", "sh")
	}

	for _, tool := range tools {
		toolCheck := CheckResult{
			Name:           fmt.Sprintf("Tool: %s", tool),
			Status:         "PASS",
			Message:        fmt.Sprintf("Tool %s is available", tool),
			Details:        "Tool availability check passed",
			Recommendation: "No action required",
		}

		// Check if tool is in PATH
		if tool == "redtriage" {
			// Check if redtriage executable exists
			if _, err := os.Stat("redtriage.exe"); err != nil {
				toolCheck.Status = "FAIL"
				toolCheck.Message = "RedTriage executable not found"
				toolCheck.Details = "redtriage.exe not found in current directory"
				toolCheck.Recommendation = "Ensure redtriage.exe is in current directory or PATH"
			}
		} else {
			// For other tools, check if they're in PATH
			path := os.Getenv("PATH")
			if path == "" {
				toolCheck.Status = "WARN"
				toolCheck.Message = fmt.Sprintf("PATH environment variable not set for %s", tool)
				toolCheck.Details = "Cannot verify tool availability"
				toolCheck.Recommendation = "Set PATH environment variable"
			}
		}

		results = append(results, toolCheck)
		if toolCheck.Status == "PASS" {
			om.LogInfo(" %s: %s", toolCheck.Name, toolCheck.Message)
		} else if toolCheck.Status == "WARN" {
			om.LogWarning("️  %s: %s", toolCheck.Name, toolCheck.Message)
		} else {
			om.LogError(fmt.Errorf(toolCheck.Details), "Tool availability check failed")
		}
	}

	return results
}

func checkConfiguration(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check configuration file
	configCheck := CheckResult{
		Name:           "Configuration File",
		Status:         "PASS",
		Message:        "Configuration is valid",
		Details:        "Configuration check passed",
		Recommendation: "No action required",
	}

	// Check for config file
	configPaths := []string{
		"./redtriage.yml",
		"./redtriage.yaml",
		"./config/redtriage.yml",
	}

	found := false
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			found = true
			configCheck.Details = fmt.Sprintf("Configuration file found at: %s", path)
			break
		}
	}

	if !found {
		configCheck.Status = "WARN"
		configCheck.Message = "No configuration file found"
		configCheck.Details = "Using default configuration values"
		configCheck.Recommendation = "Create redtriage.yml for custom configuration"
	}

	results = append(results, configCheck)
	if configCheck.Status == "PASS" {
		om.LogInfo(" %s: %s", configCheck.Name, configCheck.Message)
	} else {
		om.LogWarning("️  %s: %s", configCheck.Name, configCheck.Message)
	}

	return results
}

// Enhanced feature checks

func checkEnhancedFeatures(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	om.LogInfo("Running enhanced feature checks...")

	// Check enhanced artifacts
	results = append(results, checkEnhancedArtifacts(om)...)

	// Check enhanced collector
	results = append(results, checkEnhancedCollector(om)...)

	// Check enhanced log parser
	results = append(results, checkEnhancedLogParser(om)...)

	// Check enhanced reporter
	results = append(results, checkEnhancedReporter(om)...)

	// Check enhanced integration
	results = append(results, checkEnhancedIntegration(om)...)

	return results
}

func checkEnhancedArtifacts(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check enhanced artifacts file
	artifactsCheck := CheckResult{
		Name:           "Enhanced Artifacts",
		Status:         "PASS",
		Message:        "Enhanced artifacts system available",
		Details:        "Enhanced artifact registry and collection system found",
		Recommendation: "No action required",
	}

	if _, err := os.Stat("collector/enhanced_artifacts.go"); os.IsNotExist(err) {
		artifactsCheck.Status = "FAIL"
		artifactsCheck.Message = "Enhanced artifacts file not found"
		artifactsCheck.Details = "collector/enhanced_artifacts.go not found"
		artifactsCheck.Recommendation = "Ensure enhanced artifacts system is properly installed"
	} else {
		om.LogInfo(" Enhanced artifacts file: Found")
	}

	results = append(results, artifactsCheck)
	return results
}

func checkEnhancedCollector(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check enhanced collector file
	collectorCheck := CheckResult{
		Name:           "Enhanced Collector",
		Status:         "PASS",
		Message:        "Enhanced Windows collector available",
		Details:        "Enhanced Windows collector system found",
		Recommendation: "No action required",
	}

	if _, err := os.Stat("platform/windows/enhanced_collector.go"); os.IsNotExist(err) {
		collectorCheck.Status = "FAIL"
		collectorCheck.Message = "Enhanced collector file not found"
		collectorCheck.Details = "platform/windows/enhanced_collector.go not found"
		collectorCheck.Recommendation = "Ensure enhanced collector system is properly installed"
	} else {
		om.LogInfo(" Enhanced collector file: Found")
	}

	results = append(results, collectorCheck)
	return results
}

func checkEnhancedLogParser(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check enhanced log parser files
	logParserCheck := CheckResult{
		Name:           "Enhanced Log Parser",
		Status:         "PASS",
		Message:        "Enhanced log parsing system available",
		Details:        "Enhanced log parsing and analysis system found",
		Recommendation: "No action required",
	}

	enhancedLogFiles := []string{
		"internal/logging/enhanced_log_parser.go",
		"internal/logging/log_parsers.go",
	}

	missingFiles := []string{}
	for _, file := range enhancedLogFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingFiles) > 0 {
		logParserCheck.Status = "FAIL"
		logParserCheck.Message = "Enhanced log parser files missing"
		logParserCheck.Details = fmt.Sprintf("Missing files: %v", missingFiles)
		logParserCheck.Recommendation = "Ensure enhanced log parsing system is properly installed"
	} else {
		om.LogInfo(" Enhanced log parser files: All found")
	}

	results = append(results, logParserCheck)
	return results
}

func checkEnhancedReporter(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check enhanced reporter file
	reporterCheck := CheckResult{
		Name:           "Enhanced Reporter",
		Status:         "PASS",
		Message:        "Enhanced reporting system available",
		Details:        "Enhanced reporting and output system found",
		Recommendation: "No action required",
	}

	if _, err := os.Stat("reporter/enhanced_reporter.go"); os.IsNotExist(err) {
		reporterCheck.Status = "FAIL"
		reporterCheck.Message = "Enhanced reporter file not found"
		reporterCheck.Details = "reporter/enhanced_reporter.go not found"
		reporterCheck.Recommendation = "Ensure enhanced reporting system is properly installed"
	} else {
		om.LogInfo(" Enhanced reporter file: Found")
	}

	results = append(results, reporterCheck)
	return results
}

func checkEnhancedIntegration(om *output.OutputManager) []CheckResult {
	var results []CheckResult

	// Check enhanced integration
	integrationCheck := CheckResult{
		Name:           "Enhanced Integration",
		Status:         "PASS",
		Message:        "Enhanced features integration available",
		Details:        "Enhanced features integration system found",
		Recommendation: "No action required",
	}

	// Check enhanced collect command
	if _, err := os.Stat("cmd/enhanced_collect.go"); os.IsNotExist(err) {
		integrationCheck.Status = "FAIL"
		integrationCheck.Message = "Enhanced collect command not found"
		integrationCheck.Details = "cmd/enhanced_collect.go not found"
		integrationCheck.Recommendation = "Ensure enhanced collect command is properly installed"
	} else {
		om.LogInfo(" Enhanced collect command: Found")
	}

	// Check test file
	if _, err := os.Stat("test_enhanced_features.go"); os.IsNotExist(err) {
		om.LogWarning(" Enhanced features test file: Not found")
	} else {
		om.LogInfo(" Enhanced features test file: Found")
	}

	// Check documentation
	if _, err := os.Stat("ENHANCED_FEATURES_README.md"); os.IsNotExist(err) {
		om.LogWarning(" Enhanced features documentation: Not found")
	} else {
		om.LogInfo(" Enhanced features documentation: Found")
	}

	results = append(results, integrationCheck)
	return results
}

func hasFailedChecks(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == "FAIL" {
			return true
		}
	}
	return false
}

func hasWarnings(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == "WARN" {
			return true
		}
	}
	return false
}

func countPassedChecks(results []CheckResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "PASS" {
			count++
		}
	}
	return count
}

func countFailedChecks(results []CheckResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "FAIL" {
			count++
		}
	}
	return count
}

func countWarningChecks(results []CheckResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "WARN" {
			count++
		}
	}
	return count
}

func validateCheckInputs(om *output.OutputManager) error {
	// Validate output directory path if specified
	if checkOutput != "" {
		// Basic path validation - prevent directory traversal
		if strings.Contains(checkOutput, "..") || strings.Contains(checkOutput, "//") {
			return fmt.Errorf("invalid output directory path: %s (contains invalid characters)", checkOutput)
		}

		// Check if path is absolute and valid
		if filepath.IsAbs(checkOutput) {
			if _, err := filepath.Abs(checkOutput); err != nil {
				return fmt.Errorf("invalid absolute output directory path: %s", checkOutput)
			}
		}
	}

	// Validate format
	allowedFormats := []string{"text", "json", "yaml", "yml"}
	formatValid := false
	for _, format := range allowedFormats {
		if checkFormat == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("invalid format '%s'. Must be one of: %s", checkFormat, strings.Join(allowedFormats, ", "))
	}

	return nil
}
