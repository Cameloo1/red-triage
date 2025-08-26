package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CommandValidator provides strict validation for CLI commands
type CommandValidator struct {
	strictMode bool
}

// NewCommandValidator creates a new command validator
func NewCommandValidator(strictMode bool) *CommandValidator {
	return &CommandValidator{
		strictMode: strictMode,
	}
}

// ValidateCommand validates command structure and arguments
func (cv *CommandValidator) ValidateCommand(command string, args []string, flags map[string]interface{}) error {
	// Validate command name
	if err := cv.validateCommandName(command); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Validate arguments
	if err := cv.validateArguments(args); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	// Validate flags
	if err := cv.validateFlags(flags); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	return nil
}

// validateCommandName validates the command name
func (cv *CommandValidator) validateCommandName(command string) error {
	if command == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	// Only allow alphanumeric characters, hyphens, and underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(command) {
		return fmt.Errorf("command name contains invalid characters: %s", command)
	}

	// Check for reserved commands
	reservedCommands := []string{"help", "version", "config", "init", "setup"}
	for _, reserved := range reservedCommands {
		if command == reserved {
			return fmt.Errorf("command name '%s' is reserved", command)
		}
	}

	return nil
}

// validateArguments validates command arguments
func (cv *CommandValidator) validateArguments(args []string) error {
	for i, arg := range args {
		if err := cv.validateArgument(arg, i); err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
	}
	return nil
}

// validateArgument validates a single argument
func (cv *CommandValidator) validateArgument(arg string, index int) error {
	if arg == "" {
		return fmt.Errorf("argument cannot be empty")
	}

	// Prevent directory traversal
	if strings.Contains(arg, "..") || strings.Contains(arg, "//") {
		return fmt.Errorf("argument contains invalid path characters: %s", arg)
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script>", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(arg), pattern) {
			return fmt.Errorf("argument contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// validateFlags validates command flags
func (cv *CommandValidator) validateFlags(flags map[string]interface{}) error {
	for flag, value := range flags {
		if err := cv.validateFlag(flag, value); err != nil {
			return fmt.Errorf("flag %s: %w", flag, err)
		}
	}
	return nil
}

// validateFlag validates a single flag
func (cv *CommandValidator) validateFlag(flag string, value interface{}) error {
	// Validate flag name
	if err := cv.validateFlagName(flag); err != nil {
		return err
	}

	// Validate flag value based on type
	switch v := value.(type) {
	case string:
		return cv.validateStringFlag(flag, v)
	case int:
		return cv.validateIntFlag(flag, v)
	case bool:
		return cv.validateBoolFlag(flag, v)
	case []string:
		return cv.validateStringSliceFlag(flag, v)
	default:
		return fmt.Errorf("unsupported flag type: %T", value)
	}
}

// validateFlagName validates flag name
func (cv *CommandValidator) validateFlagName(flag string) error {
	if flag == "" {
		return fmt.Errorf("flag name cannot be empty")
	}

	// Only allow alphanumeric characters, hyphens, and underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(flag) {
		return fmt.Errorf("flag name contains invalid characters: %s", flag)
	}

	return nil
}

// validateStringFlag validates string flag values
func (cv *CommandValidator) validateStringFlag(flag, value string) error {
	if value == "" {
		return fmt.Errorf("string flag value cannot be empty")
	}

	// Prevent directory traversal
	if strings.Contains(value, "..") || strings.Contains(value, "//") {
		return fmt.Errorf("string flag value contains invalid path characters: %s", value)
	}

	// Validate file paths if flag is path-related
	if cv.isPathFlag(flag) {
		return cv.validateFilePath(value)
	}

	return nil
}

// validateIntFlag validates integer flag values
func (cv *CommandValidator) validateIntFlag(flag string, value int) error {
	switch flag {
	case "timeout", "max-size", "max-files":
		if value <= 0 {
			return fmt.Errorf("flag %s must be positive, got %d", flag, value)
		}
	case "port":
		if value < 1 || value > 65535 {
			return fmt.Errorf("flag %s must be between 1 and 65535, got %d", flag, value)
		}
	}
	return nil
}

// validateBoolFlag validates boolean flag values
func (cv *CommandValidator) validateBoolFlag(flag string, value bool) error {
	// Boolean flags are always valid
	return nil
}

// validateStringSliceFlag validates string slice flag values
func (cv *CommandValidator) validateStringSliceFlag(flag string, values []string) error {
	for i, value := range values {
		if err := cv.validateStringFlag(flag, value); err != nil {
			return fmt.Errorf("slice element %d: %w", i+1, err)
		}
	}
	return nil
}

// isPathFlag checks if a flag is path-related
func (cv *CommandValidator) isPathFlag(flag string) bool {
	pathFlags := []string{
		"output", "input", "config", "rules", "template", "log", "cache",
		"temp", "work", "data", "backup", "archive", "export",
	}

	for _, pathFlag := range pathFlags {
		if strings.Contains(flag, pathFlag) {
			return true
		}
	}
	return false
}

// validateFilePath validates a file path
func (cv *CommandValidator) validateFilePath(path string) error {
	// Check for absolute path issues
	if filepath.IsAbs(path) {
		// On Windows, check for drive letter issues
		if strings.HasPrefix(path, "\\") {
			return fmt.Errorf("invalid absolute path: %s", path)
		}
	}

	// Check for directory traversal
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains invalid character '%s': %s", char, path)
		}
	}

	return nil
}

// ValidateExecutionEnvironment validates the execution environment
func (cv *CommandValidator) ValidateExecutionEnvironment() error {
	// Check working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if working directory is accessible
	if _, err := os.ReadDir(wd); err != nil {
		return fmt.Errorf("working directory is not accessible: %w", err)
	}

	// Check if we can write to working directory
	testFile := filepath.Join(wd, ".validation_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("working directory is not writable: %w", err)
	}
	os.Remove(testFile)

	// Check system time
	if err := cv.validateSystemTime(); err != nil {
		return fmt.Errorf("system time validation failed: %w", err)
	}

	return nil
}

// validateSystemTime validates system time
func (cv *CommandValidator) validateSystemTime() error {
	now := time.Now()
	
	// Check if time is reasonable (not too far in past or future)
	minTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxTime := time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)
	
	if now.Before(minTime) || now.After(maxTime) {
		return fmt.Errorf("system time appears incorrect: %v", now)
	}
	
	return nil
}

// ValidateFileAccess validates file access permissions
func (cv *CommandValidator) ValidateFileAccess(path string, operation string) error {
	switch operation {
	case "read":
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("cannot access file for reading: %w", err)
		}
	case "write":
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("cannot access directory for writing: %w", err)
		}
		// Test write permission
		testFile := filepath.Join(dir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("directory is not writable: %w", err)
		}
		os.Remove(testFile)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
	
	return nil
}
