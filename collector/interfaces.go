package collector

import (
	"context"
	"runtime"
	"time"
)

// ArtifactCollector defines the interface for platform-specific artifact collection
type ArtifactCollector interface {
	// CollectHostProfile collects basic host information
	CollectHostProfile(ctx context.Context) (*ArtifactResult, error)
	
	// CollectBasicArtifacts collects basic system artifacts
	CollectBasicArtifacts(ctx context.Context) ([]ArtifactResult, error)
	
	// CollectExtendedArtifacts collects extended system artifacts
	CollectExtendedArtifacts(ctx context.Context) ([]ArtifactResult, error)
}

// CollectionProfile defines what artifacts to collect
type CollectionProfile struct {
	Extended bool          // Whether to collect extended artifacts
	Timeout  time.Duration // Collection timeout
	Include  []string      // Specific artifacts to include
	Exclude  []string      // Specific artifacts to exclude
}

// ArtifactResult represents the result of collecting a single artifact
type ArtifactResult struct {
	Artifact   Artifact      // The artifact definition
	Data       interface{}   // The collected data
	Metadata   Metadata      // Collection metadata
	Error      error         // Any error that occurred during collection
	Size       int64         // Size of the collected data
	Checksum   string        // SHA256 checksum of the data
}

// Artifact represents a collectable artifact
type Artifact struct {
	Name        string            // Unique name for the artifact
	Description string            // Human-readable description
	Category    string            // Category (host, process, network, etc.)
	Type        string            // Type (file, registry, command, etc.)
	Platform    string            // Platform requirement (windows, linux, all)
	Volatile    bool              // Whether this is volatile data
	Size        int64             // Expected size
	Timeout     time.Duration     // Collection timeout
	Enabled     bool              // Whether this artifact is enabled
	Parameters  map[string]string // Additional parameters
}

// Metadata contains information about the collection process
type Metadata struct {
	CollectedAt time.Time         // When the artifact was collected
	Collector   string            // Which collector was used
	Duration    time.Time         // How long collection took
	Source      string            // Source of the data
	Version     string            // Version of the collector
	Tags        map[string]string // Additional metadata tags
}

// BaseArtifact provides common functionality for artifacts
type BaseArtifact struct {
	Artifact
}

// NewBaseArtifact creates a new base artifact
func NewBaseArtifact(name, description, category, artifactType string) BaseArtifact {
	return BaseArtifact{
		Artifact: Artifact{
			Name:        name,
			Description: description,
			Category:    category,
			Type:        artifactType,
			Platform:    "all",
			Volatile:    false,
			Enabled:     true,
			Parameters:  make(map[string]string),
		},
	}
}

// Collector orchestrates artifact collection across platforms
type Collector struct {
	platformCollector ArtifactCollector
	profile           CollectionProfile
}

// NewCollector creates a new collector instance with proper platform detection
func NewCollector() *Collector {
	collector := &Collector{
		profile: CollectionProfile{
			Extended: false,
			Timeout:  5 * time.Minute,
		},
	}
	
	// Initialize platform-specific collector using factory
	collector.initializePlatformCollector()
	
	return collector
}

// initializePlatformCollector initializes the appropriate platform collector
func (c *Collector) initializePlatformCollector() {
	// Use platform factory to create appropriate collector
	factory := NewPlatformFactory()
	c.platformCollector = factory.CreateCollector()
}

// Collect collects artifacts based on the collection profile
func (c *Collector) Collect(profile CollectionProfile) ([]ArtifactResult, error) {
	c.profile = profile
	
	var results []ArtifactResult
	
	// Check if platform collector is available
	if c.platformCollector == nil {
		// Fallback to mock collector if factory failed
		factory := NewPlatformFactory()
		c.platformCollector = factory.CreateCollector()
	}
	
	// Collect host profile
	if hostResult, err := c.platformCollector.CollectHostProfile(context.Background()); err == nil {
		results = append(results, *hostResult)
	}
	
	// Collect basic artifacts
	if basicResults, err := c.platformCollector.CollectBasicArtifacts(context.Background()); err == nil {
		results = append(results, basicResults...)
	}
	
	// Collect extended artifacts if requested
	if profile.Extended {
		if extendedResults, err := c.platformCollector.CollectExtendedArtifacts(context.Background()); err == nil {
			results = append(results, extendedResults...)
		}
	}
	
	return results, nil
}

// SetPlatformCollector sets the platform-specific collector
func (c *Collector) SetPlatformCollector(collector ArtifactCollector) {
	c.platformCollector = collector
}

// GetPlatformCollector returns the current platform collector
func (c *Collector) GetPlatformCollector() ArtifactCollector {
	return c.platformCollector
}

// GetPlatform returns the current platform
func (c *Collector) GetPlatform() string {
	return runtime.GOOS
}
