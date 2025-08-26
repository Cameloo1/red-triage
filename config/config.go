package config

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// General settings
	LogLevel    string `mapstructure:"log_level"`
	LogFormat   string `mapstructure:"log_format"`
	LogFile     string `mapstructure:"log_file"`
	
	// Collection settings
	CollectionTimeout int    `mapstructure:"collection_timeout"`
	MaxArtifactSize  int64  `mapstructure:"max_artifact_size"`
	TempDir          string `mapstructure:"temp_dir"`
	
	// Detection settings
	RulesDir      string `mapstructure:"rules_dir"`
	SigmaRulesDir string `mapstructure:"sigma_rules_dir"`
	
	// Packaging settings
	OutputDir     string `mapstructure:"output_dir"`
	Compression   string `mapstructure:"compression"`
	ChecksumAlgo  string `mapstructure:"checksum_algorithm"`
	
	// Redaction settings
	RedactionEnabled bool     `mapstructure:"redaction_enabled"`
	RedactionPatterns []string `mapstructure:"redaction_patterns"`
	
	// Platform-specific settings
	Platform string `mapstructure:"platform"`
	
	// Network settings
	AllowNetwork bool `mapstructure:"allow_network"`
	
	// Artifact-specific settings
	Artifacts map[string]ArtifactConfig `mapstructure:"artifacts"`
}

// ArtifactConfig represents configuration for a specific artifact type
type ArtifactConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MaxSize     int64  `mapstructure:"max_size"`
	Timeout     int    `mapstructure:"timeout"`
	OutputPath  string `mapstructure:"output_path"`
}

// Load loads configuration from files and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("redtriage")
	viper.SetConfigType("yml")
	
	// Set default values
	setDefaults()
	
	// Add config paths
	addConfigPaths()
	
	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}
	
	// Create config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate and set platform
	if err := config.validateAndSetPlatform(); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")
	viper.SetDefault("collection_timeout", 300)
	viper.SetDefault("max_artifact_size", 100*1024*1024) // 100MB
	viper.SetDefault("compression", "zip")
	viper.SetDefault("checksum_algorithm", "sha256")
	viper.SetDefault("redaction_enabled", true)
	viper.SetDefault("allow_network", false)
}

// addConfigPaths adds search paths for configuration files
func addConfigPaths() {
	// Current directory
	viper.AddConfigPath(".")
	
	// User home directory
	if home, err := homedir.Dir(); err == nil {
		viper.AddConfigPath(home)
	}
	
	// Windows-specific paths
	if runtime.GOOS == "windows" {
		viper.AddConfigPath("C:\\ProgramData\\RedTriage")
		viper.AddConfigPath("C:\\Users\\%USERNAME%\\AppData\\Local\\RedTriage")
	}
	
	// Linux-specific paths
	if runtime.GOOS == "linux" {
		viper.AddConfigPath("/etc/redtriage")
		viper.AddConfigPath("/usr/local/etc/redtriage")
	}
}

// validateAndSetPlatform validates and sets the platform configuration
func (c *Config) validateAndSetPlatform() error {
	if c.Platform == "" {
		c.Platform = runtime.GOOS
	}
	
	// Validate platform
	switch c.Platform {
	case "windows", "linux":
		// Valid platforms
	default:
		return fmt.Errorf("unsupported platform: %s", c.Platform)
	}
	
	return nil
}

// GetArtifactConfig returns configuration for a specific artifact
func (c *Config) GetArtifactConfig(artifactName string) ArtifactConfig {
	if config, exists := c.Artifacts[artifactName]; exists {
		return config
	}
	
	// Return default configuration
	return ArtifactConfig{
		Enabled: true,
		MaxSize: c.MaxArtifactSize,
		Timeout: c.CollectionTimeout,
	}
}

// IsArtifactEnabled checks if a specific artifact is enabled
func (c *Config) IsArtifactEnabled(artifactName string) bool {
	config := c.GetArtifactConfig(artifactName)
	return config.Enabled
}
