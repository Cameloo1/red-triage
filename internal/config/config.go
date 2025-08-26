package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

// Config represents the RedTriage configuration
type Config struct {
	// General settings
	LogLevel        string `mapstructure:"log_level"`
	LogFormat       string `mapstructure:"log_format"`
	DefaultTimeout  string `mapstructure:"default_timeout"`
	MaxArtifactSize string `mapstructure:"max_artifact_size"`
	MaxLogSize      string `mapstructure:"max_log_size"`
	MaxLogAge       string `mapstructure:"max_log_age"`
	
	// Collection settings
	DetectionTimeout string `mapstructure:"detection_timeout"`
	MinSeverity     string `mapstructure:"min_severity"`
	CompressionLevel int    `mapstructure:"compression_level"`
	
	// Security settings
	ChecksumAlgorithm string `mapstructure:"checksum_algorithm"`
	RedactionEnabled  bool   `mapstructure:"redaction_enabled"`
	AllowNetwork      bool   `mapstructure:"allow_network"`
	
	// Artifact settings
	Artifacts map[string]ArtifactConfig `mapstructure:"artifacts"`
	
	// Platform-specific settings
	Platform string `mapstructure:"platform"`
	
	// Output settings
	DefaultOutputDir string `mapstructure:"default_output_dir"`
	ReportsDir       string `mapstructure:"reports_dir"`
	ReportFormats    []string `mapstructure:"report_formats"`
	
	// Rule settings
	SigmaRulesPath string `mapstructure:"sigma_rules_path"`
	CustomRulesPath string `mapstructure:"custom_rules_path"`
	
	// Session settings
	SaveHistory     bool   `mapstructure:"save_history"`
	HistoryFile     string `mapstructure:"history_file"`
	SessionLogPath  string `mapstructure:"session_log_path"`
	
	// Color settings
	ColorEnabled bool   `mapstructure:"color_enabled"`
	ColorMode    string `mapstructure:"color_mode"`
}

// ArtifactConfig represents configuration for a specific artifact type
type ArtifactConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Timeout string `mapstructure:"timeout"`
	MaxSize string `mapstructure:"max_size"`
}

// LoadConfig loads configuration from file or creates default if not found
func LoadConfig(configPath string) (*Config, error) {
	// For now, just return default config
	// TODO: Implement actual config file loading
	return DefaultConfig(), nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel:         "info",
		LogFormat:        "text",
		DefaultTimeout:   "20m",
		MaxArtifactSize:  "100MB",
		MaxLogSize:       "200MB",
		MaxLogAge:        "48h",
		DetectionTimeout: "5m",
		MinSeverity:      "medium",
		CompressionLevel: 6,
		ChecksumAlgorithm: "sha256",
		RedactionEnabled:  true,
		AllowNetwork:      false,
		Platform:          runtime.GOOS,
		DefaultOutputDir:  "./redtriage-output",
		ReportsDir:        "./redtriage-reports",
		ReportFormats:     []string{"md", "html", "json"},
		SaveHistory:       true,
		HistoryFile:       ".redtriage_history",
		SessionLogPath:    "./logs",
		ColorEnabled:      true,
		ColorMode:         "auto",
		Artifacts: map[string]ArtifactConfig{
			"processes": {
				Enabled: true,
				Timeout: "5m",
				MaxSize: "50MB",
			},
			"services": {
				Enabled: true,
				Timeout: "2m",
				MaxSize: "10MB",
			},
			"network": {
				Enabled: true,
				Timeout: "1m",
				MaxSize: "5MB",
			},
			"logs": {
				Enabled: true,
				Timeout: "10m",
				MaxSize: "200MB",
			},
		},
	}
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	config := DefaultConfig()
	
	// Set config file path
	viper.SetConfigName("redtriage")
	viper.SetConfigType("yml")
	
	// Search paths in order of preference
	searchPaths := []string{
		".", // Current directory
	}
	
	// Add platform-specific paths
	if runtime.GOOS == "windows" {
		programData := os.Getenv("PROGRAMDATA")
		if programData != "" {
			searchPaths = append(searchPaths, filepath.Join(programData, "RedTriage"))
		}
		// Add Windows user profile paths
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			searchPaths = append(searchPaths, userProfile)
		}
	} else {
		searchPaths = append(searchPaths, "/etc/redtriage")
		// Add Unix user home directory
		if home, err := os.UserHomeDir(); err == nil {
			searchPaths = append(searchPaths, home)
		}
	}
	
	// Add user home directory (cross-platform)
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, home)
	}
	
	// Add search paths
	for _, path := range searchPaths {
		viper.AddConfigPath(path)
	}
	
	// Environment variable prefix
	viper.SetEnvPrefix("REDTRIAGE")
	viper.AutomaticEnv()
	
	// Bind environment variables
	viper.BindEnv("log_level", "REDTRIAGE_LOG_LEVEL")
	viper.BindEnv("log_format", "REDTRIAGE_LOG_FORMAT")
	viper.BindEnv("default_timeout", "REDTRIAGE_DEFAULT_TIMEOUT")
	viper.BindEnv("allow_network", "REDTRIAGE_ALLOW_NETWORK")
	viper.BindEnv("color_enabled", "REDTRIAGE_COLOR_ENABLED")
	viper.BindEnv("platform", "REDTRIAGE_PLATFORM")
	viper.BindEnv("output_dir", "REDTRIAGE_OUTPUT_DIR")
	viper.BindEnv("reports_dir", "REDTRIAGE_REPORTS_DIR")
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is not an error, use defaults
		// Try to create a default config file in the current directory
		if err := config.Save("redtriage.yml"); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: Could not create default config file: %v\n", err)
		}
	}
	
	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Ensure output directories exist
	if err := config.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create output directories: %w", err)
	}
	
	return config, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Convert config to viper
	viper.Set("log_level", c.LogLevel)
	viper.Set("log_format", c.LogFormat)
	viper.Set("default_timeout", c.DefaultTimeout)
	viper.Set("max_artifact_size", c.MaxArtifactSize)
	viper.Set("max_log_size", c.MaxLogSize)
	viper.Set("max_log_age", c.MaxLogAge)
	viper.Set("detection_timeout", c.DetectionTimeout)
	viper.Set("min_severity", c.MinSeverity)
	viper.Set("compression_level", c.CompressionLevel)
	viper.Set("checksum_algorithm", c.ChecksumAlgorithm)
	viper.Set("redaction_enabled", c.RedactionEnabled)
	viper.Set("allow_network", c.AllowNetwork)
	viper.Set("platform", c.Platform)
	viper.Set("default_output_dir", c.DefaultOutputDir)
	viper.Set("reports_dir", c.ReportsDir)
	viper.Set("report_formats", c.ReportFormats)
	viper.Set("sigma_rules_path", c.SigmaRulesPath)
	viper.Set("custom_rules_path", c.CustomRulesPath)
	viper.Set("save_history", c.SaveHistory)
	viper.Set("history_file", c.HistoryFile)
	viper.Set("session_log_path", c.SessionLogPath)
	viper.Set("color_enabled", c.ColorEnabled)
	viper.Set("color_mode", c.ColorMode)
	viper.Set("artifacts", c.Artifacts)
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Write config file
	return viper.WriteConfigAs(path)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	// Validate log format
	validLogFormats := map[string]bool{
		"text": true, "json": true,
	}
	if !validLogFormats[c.LogFormat] {
		return fmt.Errorf("invalid log format: %s", c.LogFormat)
	}
	
	// Validate timeout formats
	if _, err := time.ParseDuration(c.DefaultTimeout); err != nil {
		return fmt.Errorf("invalid default timeout: %s", c.DefaultTimeout)
	}
	if _, err := time.ParseDuration(c.DetectionTimeout); err != nil {
		return fmt.Errorf("invalid detection timeout: %s", c.DetectionTimeout)
	}
	
	// Validate severity
	validSeverities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validSeverities[c.MinSeverity] {
		return fmt.Errorf("invalid minimum severity: %s", c.MinSeverity)
	}
	
	// Validate compression level
	if c.CompressionLevel < 0 || c.CompressionLevel > 9 {
		return fmt.Errorf("invalid compression level: %d (must be 0-9)", c.CompressionLevel)
	}
	
	// Validate checksum algorithm
	validAlgorithms := map[string]bool{
		"md5": true, "sha1": true, "sha256": true, "sha512": true,
	}
	if !validAlgorithms[c.ChecksumAlgorithm] {
		return fmt.Errorf("invalid checksum algorithm: %s", c.ChecksumAlgorithm)
	}
	
	// Validate platform
	validPlatforms := map[string]bool{
		"windows": true, "linux": true, "darwin": true,
	}
	if !validPlatforms[c.Platform] {
		return fmt.Errorf("invalid platform: %s", c.Platform)
	}
	
	return nil
}

// GetTimeout returns the timeout as a duration
func (c *Config) GetTimeout() time.Duration {
	duration, err := time.ParseDuration(c.DefaultTimeout)
	if err != nil {
		// Return default if parsing fails
		return 20 * time.Minute
	}
	return duration
}

// GetDetectionTimeout returns the detection timeout as a duration
func (c *Config) GetDetectionTimeout() time.Duration {
	duration, err := time.ParseDuration(c.DetectionTimeout)
	if err != nil {
		// Return default if parsing fails
		return 5 * time.Minute
	}
	return duration
}

// IsArtifactEnabled checks if a specific artifact type is enabled
func (c *Config) IsArtifactEnabled(artifactType string) bool {
	if artifact, exists := c.Artifacts[artifactType]; exists {
		return artifact.Enabled
	}
	return false
}

// GetArtifactTimeout returns the timeout for a specific artifact type
func (c *Config) GetArtifactTimeout(artifactType string) time.Duration {
	if artifact, exists := c.Artifacts[artifactType]; exists {
		if duration, err := time.ParseDuration(artifact.Timeout); err == nil {
			return duration
		}
	}
	// Return default timeout if not specified or invalid
	return c.GetTimeout()
}

// ensureDirectories ensures that all necessary output directories exist
func (c *Config) ensureDirectories() error {
	dirs := []string{
		c.DefaultOutputDir,
		c.ReportsDir,
		c.SessionLogPath,
	}
	
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}
