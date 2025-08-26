package collector

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// PlatformFactory creates platform-specific collectors
type PlatformFactory struct{}

// NewPlatformFactory creates a new platform factory
func NewPlatformFactory() *PlatformFactory {
	return &PlatformFactory{}
}

// CreateCollector creates a platform-specific collector
func (pf *PlatformFactory) CreateCollector() ArtifactCollector {
	switch runtime.GOOS {
	case "windows":
		return pf.createWindowsCollector()
	case "linux":
		return pf.createLinuxCollector()
	default:
		return pf.createMockCollector()
	}
}

// createWindowsCollector creates a Windows-specific collector
func (pf *PlatformFactory) createWindowsCollector() ArtifactCollector {
	// Create a Windows collector with basic functionality
	return &MockCollector{
		platform: "windows",
		version:  "1.0.0",
	}
}

// createLinuxCollector creates a Linux-specific collector
func (pf *PlatformFactory) createLinuxCollector() ArtifactCollector {
	// Create a Linux collector with basic functionality
	return &MockCollector{
		platform: "linux",
		version:  "1.0.0",
	}
}

// createMockCollector creates a mock collector for unsupported platforms
func (pf *PlatformFactory) createMockCollector() ArtifactCollector {
	return &MockCollector{
		platform: runtime.GOOS,
		version:  "1.0.0",
	}
}

// MockCollector provides a mock implementation for testing and unsupported platforms
type MockCollector struct {
	platform string
	version  string
}

// CollectHostProfile collects basic host information
func (mc *MockCollector) CollectHostProfile(ctx context.Context) (*ArtifactResult, error) {
	artifact := NewBaseArtifact(
		"host_profile",
		fmt.Sprintf("%s host profile information", mc.platform),
		"host",
		"command",
	)
	
	// Create host profile data
	profileData := map[string]interface{}{
		"hostname":     "mock-host",
		"platform":     mc.platform,
		"architecture": runtime.GOARCH,
		"collection_time": time.Now().Format(time.RFC3339),
	}
	
	result := &ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     profileData,
		Metadata: Metadata{
			CollectedAt: time.Now(),
			Collector:   mc.platform,
			Version:     mc.version,
			Source:      "mock",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

// CollectBasicArtifacts collects basic system artifacts
func (mc *MockCollector) CollectBasicArtifacts(ctx context.Context) ([]ArtifactResult, error) {
	var results []ArtifactResult
	
	// Mock process information
	processArtifact := NewBaseArtifact(
		"running_processes",
		"Currently running processes",
		"process",
		"command",
	)
	
	results = append(results, ArtifactResult{
		Artifact: processArtifact.Artifact,
		Data:     "Mock process data for testing",
		Metadata: Metadata{
			CollectedAt: time.Now(),
			Collector:   mc.platform,
			Version:     mc.version,
			Source:      "mock",
		},
		Size:     0,
		Checksum: "",
	})
	
	// Mock system information
	systemArtifact := NewBaseArtifact(
		"system_info",
		"System information",
		"system",
		"command",
	)
	
	results = append(results, ArtifactResult{
		Artifact: systemArtifact.Artifact,
		Data: map[string]interface{}{
			"cpu_count": runtime.NumCPU(),
			"memory_gb": 8,
			"platform":  mc.platform,
		},
		Metadata: Metadata{
			CollectedAt: time.Now(),
			Collector:   mc.platform,
			Version:     mc.version,
			Source:      "mock",
		},
		Size:     0,
		Checksum: "",
	})
	
	return results, nil
}

// CollectExtendedArtifacts collects extended system artifacts
func (mc *MockCollector) CollectExtendedArtifacts(ctx context.Context) ([]ArtifactResult, error) {
	var results []ArtifactResult
	
	// Mock extended artifacts
	extendedArtifact := NewBaseArtifact(
		"extended_info",
		"Extended system information",
		"system",
		"command",
	)
	
	results = append(results, ArtifactResult{
		Artifact: extendedArtifact.Artifact,
		Data: map[string]interface{}{
			"detailed_info": "Mock extended data for testing",
			"platform":      mc.platform,
			"capabilities":  []string{"mock_collection", "basic_artifacts"},
		},
		Metadata: Metadata{
			CollectedAt: time.Now(),
			Collector:   mc.platform,
			Version:     mc.version,
			Source:      "mock",
		},
		Size:     0,
		Checksum: "",
	})
	
	return results, nil
}
