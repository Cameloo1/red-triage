package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// OutputManager handles centralized output for all RedTriage commands
type OutputManager struct {
	outputDir    string
	logFile      *os.File
	outputFile   *os.File
	outputFormat string
	verbose      bool
	jsonOutput   bool
	startTime    time.Time
	commandName  string
	errors       []error
	warnings     []string
	results      []Result
}

// Result represents a command execution result
type Result struct {
	Type      string                 `json:"type" yaml:"type"`
	Status    string                 `json:"status" yaml:"status"`
	Message   string                 `json:"message" yaml:"message"`
	Data      interface{}            `json:"data,omitempty" yaml:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp" yaml:"timestamp"`
	Duration  time.Duration          `json:"duration" yaml:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty" yaml:"error,omitempty"`
	Warning   string                 `json:"warning,omitempty" yaml:"warning,omitempty"`
}

// NewOutputManager creates a new output manager
func NewOutputManager(commandName, outputDir, format string, verbose, jsonOutput bool) (*OutputManager, error) {
	om := &OutputManager{
		outputDir:    outputDir,
		outputFormat: format,
		verbose:      verbose,
		jsonOutput:   jsonOutput,
		startTime:    time.Now(),
		commandName:  commandName,
		errors:       make([]error, 0),
		warnings:     make([]string, 0),
		results:      make([]Result, 0),
	}

	// Ensure output directory exists
	if err := om.ensureOutputDir(); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Initialize log file
	if err := om.initLogFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}

	// Initialize output file if format is specified
	if format != "console" {
		if err := om.initOutputFile(); err != nil {
			return nil, fmt.Errorf("failed to initialize output file: %w", err)
		}
	}

	return om, nil
}

// ensureOutputDir ensures the output directory exists
func (om *OutputManager) ensureOutputDir() error {
	if om.outputDir == "" {
		return nil
	}

	// Check if directory exists and is writable
	if info, err := os.Stat(om.outputDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path exists but is not a directory: %s", om.outputDir)
		}
		// Check if directory is writable
		if info.Mode().Perm()&0200 == 0 {
			return fmt.Errorf("output directory is not writable: %s", om.outputDir)
		}
	} else if os.IsNotExist(err) {
		// Create directory with proper permissions
		if err := os.MkdirAll(om.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %s: %w", om.outputDir, err)
		}
	} else {
		return fmt.Errorf("failed to check output directory: %s: %w", om.outputDir, err)
	}

	return nil
}

// initLogFile initializes the log file
func (om *OutputManager) initLogFile() error {
	if om.outputDir == "" {
		return nil
	}

	logPath := filepath.Join(om.outputDir, fmt.Sprintf("%s-%s.log", om.commandName, time.Now().Format("20060102-150405")))
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %s: %w", logPath, err)
	}

	om.logFile = logFile
	return nil
}

// initOutputFile initializes the output file based on format
func (om *OutputManager) initOutputFile() error {
	if om.outputDir == "" || om.outputFormat == "console" {
		return nil
	}

	outputPath := filepath.Join(om.outputDir, fmt.Sprintf("%s-%s.%s", om.commandName, time.Now().Format("20060102-150405"), om.outputFormat))
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %s: %w", outputPath, err)
	}

	om.outputFile = outputFile
	return nil
}

// LogInfo logs an informational message
func (om *OutputManager) LogInfo(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Console output
	if om.verbose {
		color.New(color.FgBlue).Printf("[INFO] %s\n", formattedMessage)
	}

	// Log file output
	if om.logFile != nil {
		fmt.Fprintf(om.logFile, "[%s] [INFO] %s\n", timestamp, formattedMessage)
		om.logFile.Sync() // Ensure data is written immediately
	}

	// Add to results
	om.results = append(om.results, Result{
		Type:      "info",
		Status:    "success",
		Message:   formattedMessage,
		Timestamp: time.Now(),
		Duration:  time.Since(om.startTime),
	})
}

// LogWarning logs a warning message
func (om *OutputManager) LogWarning(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Console output
	color.New(color.FgYellow).Printf("[WARN] %s\n", formattedMessage)

	// Log file output
	if om.logFile != nil {
		fmt.Fprintf(om.logFile, "[%s] [WARN] %s\n", timestamp, formattedMessage)
		om.logFile.Sync() // Ensure data is written immediately
	}

	// Add to warnings
	om.warnings = append(om.warnings, formattedMessage)

	// Add to results
	om.results = append(om.results, Result{
		Type:      "warning",
		Status:    "warning",
		Message:   formattedMessage,
		Timestamp: time.Now(),
		Duration:  time.Since(om.startTime),
		Warning:   formattedMessage,
	})
}

// LogError logs an error message
func (om *OutputManager) LogError(err error, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Console output
	color.New(color.FgRed).Printf("[ERROR] %s: %v\n", formattedMessage, err)

	// Log file output
	if om.logFile != nil {
		fmt.Fprintf(om.logFile, "[%s] [ERROR] %s: %v\n", timestamp, formattedMessage, err)
		om.logFile.Sync() // Ensure data is written immediately
	}

	// Add to errors
	om.errors = append(om.errors, err)

	// Add to results
	om.results = append(om.results, Result{
		Type:      "error",
		Status:    "error",
		Message:   formattedMessage,
		Timestamp: time.Now(),
		Duration:  time.Since(om.startTime),
		Error:     err.Error(),
	})
}

// LogSuccess logs a success message
func (om *OutputManager) LogSuccess(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Console output
	color.New(color.FgGreen).Printf("[SUCCESS] %s\n", formattedMessage)

	// Log file output
	if om.logFile != nil {
		fmt.Fprintf(om.logFile, "[%s] [SUCCESS] %s\n", timestamp, formattedMessage)
		om.logFile.Sync() // Ensure data is written immediately
	}

	// Add to results
	om.results = append(om.results, Result{
		Type:      "success",
		Status:    "success",
		Message:   formattedMessage,
		Timestamp: time.Now(),
		Duration:  time.Since(om.startTime),
	})
}

// AddResult adds a custom result
func (om *OutputManager) AddResult(result Result) {
	om.results = append(om.results, result)
}

// WriteOutput writes the final output to the output file
func (om *OutputManager) WriteOutput() error {
	if om.outputFile == nil {
		return nil
	}

	var data []byte
	var err error

	switch om.outputFormat {
	case "json":
		data, err = json.MarshalIndent(om.results, "", "  ")
	case "yaml", "yml":
		data, err = yaml.Marshal(om.results)
	default:
		return fmt.Errorf("unsupported output format: %s", om.outputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	if _, err := om.outputFile.Write(data); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// PrintSummary prints a summary of the command execution
func (om *OutputManager) PrintSummary() {
	duration := time.Since(om.startTime)

	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("=== Command Execution Summary ===")
	fmt.Printf("Command: %s\n", om.commandName)
	fmt.Printf("Duration: %s\n", duration.Round(time.Millisecond))
	fmt.Printf("Results: %d\n", len(om.results))
	fmt.Printf("Warnings: %d\n", len(om.warnings))
	fmt.Printf("Errors: %d\n", len(om.errors))

	if om.outputDir != "" {
		fmt.Printf("Output Directory: %s\n", om.outputDir)
		if om.logFile != nil {
			fmt.Printf("Log File: %s\n", om.logFile.Name())
		}
		if om.outputFile != nil {
			fmt.Printf("Output File: %s\n", om.outputFile.Name())
		}
	}

	if len(om.warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range om.warnings {
			color.New(color.FgYellow).Printf("  ️  %s\n", warning)
		}
	}

	if len(om.errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range om.errors {
			color.New(color.FgRed).Printf("  ❌ %v\n", err)
		}
	}

	fmt.Println()
}

// Close closes all open files
func (om *OutputManager) Close() error {
	var errors []error

	if om.logFile != nil {
		if err := om.logFile.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close log file: %w", err))
		}
	}

	if om.outputFile != nil {
		if err := om.outputFile.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close output file: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred while closing: %v", errors)
	}

	return nil
}

// GetErrors returns all collected errors
func (om *OutputManager) GetErrors() []error {
	return om.errors
}

// HasErrors returns true if there are any errors
func (om *OutputManager) HasErrors() bool {
	return len(om.errors) > 0
}

// GetWarnings returns all collected warnings
func (om *OutputManager) GetWarnings() []string {
	return om.warnings
}

// GetResults returns all collected results
func (om *OutputManager) GetResults() []Result {
	return om.results
}
