package linux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// SimpleLinuxCollector implements ArtifactCollector for Linux systems
type SimpleLinuxCollector struct {
	version string
}

// NewSimpleLinuxCollector creates a new simple Linux collector
func NewSimpleLinuxCollector() *SimpleLinuxCollector {
	return &SimpleLinuxCollector{
		version: "1.0.0",
	}
}

// CollectHostProfile collects basic host information
func (lc *SimpleLinuxCollector) CollectHostProfile(ctx context.Context) (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"host_profile",
		"Linux host profile information",
		"host",
		"command",
	)
	
	// Collect hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	// Collect basic system info
	sysInfo := lc.getBasicSystemInfo()
	
	// Create host profile data
	profileData := map[string]interface{}{
		"hostname":        hostname,
		"platform":        "linux",
		"system_info":     sysInfo,
		"collection_time": time.Now().Format(time.RFC3339),
	}
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     profileData,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

// CollectBasicArtifacts collects basic system artifacts
func (lc *SimpleLinuxCollector) CollectBasicArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult
	
	// System information
	if sysResult, err := lc.collectSystemInfo(); err == nil {
		results = append(results, *sysResult)
	}
	
	// Process information
	if procResult, err := lc.collectProcessInfo(); err == nil {
		results = append(results, *procResult)
	}
	
	// Network information
	if netResult, err := lc.collectNetworkInfo(); err == nil {
		results = append(results, *netResult)
	}
	
	return results, nil
}

// CollectExtendedArtifacts collects extended system artifacts
func (lc *SimpleLinuxCollector) CollectExtendedArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult
	
	// Get basic artifacts first
	basicResults, err := lc.CollectBasicArtifacts(ctx)
	if err != nil {
		return nil, err
	}
	results = append(results, basicResults...)
	
	// File system information
	if fsResult, err := lc.collectFileSystemInfo(); err == nil {
		results = append(results, *fsResult)
	}
	
	// User information
	if userResult, err := lc.collectUserInfo(); err == nil {
		results = append(results, *userResult)
	}
	
	return results, nil
}

// Helper methods
func (lc *SimpleLinuxCollector) getBasicSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"go_version":   runtime.Version(),
	}
	
	// Try to get uname info
	if uname, err := exec.Command("uname", "-a").Output(); err == nil {
		info["uname"] = string(uname)
	}
	
	// Try to get hostname
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}
	
	return info
}

func (lc *SimpleLinuxCollector) collectSystemInfo() (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"system_info",
		"Linux system information",
		"system",
		"command",
	)
	
	data := lc.getBasicSystemInfo()
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     data,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

func (lc *SimpleLinuxCollector) collectProcessInfo() (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"process_info",
		"Linux process information",
		"process",
		"command",
	)
	
	var data interface{}
	
	// Try to get process list
	if ps, err := exec.Command("ps", "aux").Output(); err == nil {
		data = string(ps)
	} else {
		data = "Unable to collect process information"
	}
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     data,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

func (lc *SimpleLinuxCollector) collectNetworkInfo() (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"network_info",
		"Linux network information",
		"network",
		"command",
	)
	
	var data interface{}
	
	// Try to get network interfaces
	if ip, err := exec.Command("ip", "addr").Output(); err == nil {
		data = string(ip)
	} else {
		data = "Unable to collect network information"
	}
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     data,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

func (lc *SimpleLinuxCollector) collectFileSystemInfo() (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"filesystem_info",
		"Linux filesystem information",
		"filesystem",
		"command",
	)
	
	var data interface{}
	
	// Try to get disk usage
	if df, err := exec.Command("df", "-h").Output(); err == nil {
		data = string(df)
	} else {
		data = "Unable to collect filesystem information"
	}
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     data,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

func (lc *SimpleLinuxCollector) collectUserInfo() (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"user_info",
		"Linux user information",
		"users",
		"command",
	)
	
	var data interface{}
	
	// Try to get user list
	if users, err := exec.Command("cat", "/etc/passwd").Output(); err == nil {
		data = string(users)
	} else {
		data = "Unable to collect user information"
	}
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     data,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     lc.version,
			Source:      "system",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}
