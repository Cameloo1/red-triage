package windows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// WindowsCollector implements ArtifactCollector for Windows systems
type WindowsCollector struct {
	version string
}

// NewWindowsCollector creates a new Windows collector
func NewWindowsCollector() *WindowsCollector {
	return &WindowsCollector{
		version: "1.0.0",
	}
}

// CollectHostProfile collects basic host information
func (w *WindowsCollector) CollectHostProfile(ctx context.Context) (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"host_profile",
		"Windows host profile information",
		"host",
		"command",
	)
	
	// Collect hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	// Collect OS information
	osInfo := w.getOSInfo()
	
	// Collect system information
	sysInfo := w.getSystemInfo()
	
	// Create host profile data
	profileData := map[string]interface{}{
		"hostname":     hostname,
		"os_info":      osInfo,
		"system_info":  sysInfo,
		"collection_time": time.Now().Format(time.RFC3339),
	}
	
	// Convert to JSON string for size calculation
	profileStr := fmt.Sprintf("%v", profileData)
	
	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     profileData,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "system",
		},
		Size:     int64(len(profileStr)),
		Checksum: w.calculateChecksum(profileStr),
	}
	
	return result, nil
}

// CollectBasicArtifacts collects basic system artifacts
func (w *WindowsCollector) CollectBasicArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult
	
	// Collect running processes
	if processes, err := w.collectProcesses(); err == nil {
		results = append(results, processes)
	}
	
	// Collect running services
	if services, err := w.collectServices(); err == nil {
		results = append(results, services)
	}
	
	// Collect scheduled tasks
	if tasks, err := w.collectScheduledTasks(); err == nil {
		results = append(results, tasks)
	}
	
	// Collect network information
	if network, err := w.collectNetworkInfo(); err == nil {
		results = append(results, network)
	}
	
	// Collect event logs
	if events, err := w.collectEventLogs(); err == nil {
		results = append(results, events)
	}
	
	return results, nil
}

// CollectExtendedArtifacts collects extended system artifacts
func (w *WindowsCollector) CollectExtendedArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult
	
	// Collect autoruns
	if autoruns, err := w.collectAutoruns(); err == nil {
		results = append(results, autoruns)
	}
	
	// Collect execution traces
	if traces, err := w.collectExecutionTraces(); err == nil {
		results = append(results, traces)
	}
	
	// Collect installed software
	if software, err := w.collectInstalledSoftware(); err == nil {
		results = append(results, software)
	}
	
	return results, nil
}

// getOSInfo retrieves operating system information
func (w *WindowsCollector) getOSInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	// Get Windows version
	if version, err := exec.Command("ver").Output(); err == nil {
		info["version"] = strings.TrimSpace(string(version))
	}
	
	// Get Windows build info
	if build, err := exec.Command("wmic", "os", "get", "BuildNumber", "/value").Output(); err == nil {
		info["build"] = strings.TrimSpace(string(build))
	}
	
	// Get Windows edition
	if edition, err := exec.Command("wmic", "os", "get", "Caption", "/value").Output(); err == nil {
		info["edition"] = strings.TrimSpace(string(edition))
	}
	
	return info
}

// getSystemInfo retrieves basic system information
func (w *WindowsCollector) getSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	// Get system architecture
	info["architecture"] = runtime.GOARCH
	
	// Get number of CPUs
	info["cpu_count"] = runtime.NumCPU()
	
	// Get memory info (basic)
	info["memory_info"] = "Available via WMI"
	
	return info
}

// collectProcesses collects running process information
func (w *WindowsCollector) collectProcesses() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"running_processes",
		"Currently running processes",
		"process",
		"command",
	)
	
	// Use tasklist to get process information
	output, err := exec.Command("tasklist", "/FO", "CSV", "/V").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect processes: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "tasklist",
		},
		Size:     int64(len(output)),
		Checksum: w.calculateChecksum(string(output)),
	}
	
	return result, nil
}

// collectServices collects running service information
func (w *WindowsCollector) collectServices() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"running_services",
		"Currently running services",
		"service",
		"command",
	)
	
	// Use sc query to get service information
	output, err := exec.Command("sc", "query", "type=", "state=", "all").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect services: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "sc",
		},
		Size:     int64(len(output)),
		Checksum: w.calculateChecksum(string(output)),
	}
	
	return result, nil
}

// collectScheduledTasks collects scheduled task information
func (w *WindowsCollector) collectScheduledTasks() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"scheduled_tasks",
		"Scheduled tasks",
		"task",
		"command",
	)
	
	// Use schtasks to get scheduled task information
	output, err := exec.Command("schtasks", "/query", "/fo", "csv", "/v").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect scheduled tasks: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "schtasks",
		},
		Size:     int64(len(output)),
		Checksum: w.calculateChecksum(string(output)),
	}
	
	return result, nil
}

// collectNetworkInfo collects network configuration information
func (w *WindowsCollector) collectNetworkInfo() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"network_info",
		"Network configuration and connections",
		"network",
		"command",
	)
	
	// Use ipconfig and netstat to get network information
	var networkData strings.Builder
	
	// Get IP configuration
	if ipconfig, err := exec.Command("ipconfig", "/all").Output(); err == nil {
		networkData.WriteString("=== IP Configuration ===\n")
		networkData.Write(ipconfig)
		networkData.WriteString("\n\n")
	}
	
	// Get network connections
	if netstat, err := exec.Command("netstat", "-an").Output(); err == nil {
		networkData.WriteString("=== Network Connections ===\n")
		networkData.Write(netstat)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     networkData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "ipconfig,netstat",
		},
		Size:     int64(networkData.Len()),
		Checksum: w.calculateChecksum(networkData.String()),
	}
	
	return result, nil
}

// collectEventLogs collects recent event log entries
func (w *WindowsCollector) collectEventLogs() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"event_logs",
		"Recent event log entries",
		"log",
		"command",
	)
	
	// Use wevtutil to get recent events from key logs
	var eventData strings.Builder
	
	logs := []string{"System", "Security", "Application"}
	for _, logName := range logs {
		if events, err := exec.Command("wevtutil", "qe", logName, "/c:100", "/f:text").Output(); err == nil {
			eventData.WriteString(fmt.Sprintf("=== %s Log ===\n", logName))
			eventData.Write(events)
			eventData.WriteString("\n\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     eventData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "wevtutil",
		},
		Size:     int64(eventData.Len()),
		Checksum: w.calculateChecksum(eventData.String()),
	}
	
	return result, nil
}

// collectAutoruns collects autorun entries
func (w *WindowsCollector) collectAutoruns() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"autoruns",
		"System autorun entries",
		"autorun",
		"registry",
	)
	
	// Implement autorun collection from registry
	var autorunData strings.Builder
	
	// Common autorun registry locations
	autorunKeys := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`,
	}
	
	for _, key := range autorunKeys {
		// Use reg query to get autorun entries
		if output, err := exec.Command("reg", "query", key).Output(); err == nil {
			autorunData.WriteString(fmt.Sprintf("=== %s ===\n", key))
			autorunData.Write(output)
			autorunData.WriteString("\n\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     autorunData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "registry",
		},
		Size:     int64(autorunData.Len()),
		Checksum: w.calculateChecksum(autorunData.String()),
	}
	
	return result, nil
}

// collectExecutionTraces collects execution trace information
func (w *WindowsCollector) collectExecutionTraces() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"execution_traces",
		"Process execution traces",
		"trace",
		"file",
	)
	
	// Implement execution trace collection
	var traceData strings.Builder
	
	// Collect Prefetch files info
	prefetchDir := `C:\Windows\Prefetch`
	if entries, err := os.ReadDir(prefetchDir); err == nil {
		traceData.WriteString("=== Prefetch Files ===\n")
		count := 0
		for _, entry := range entries {
			if count < 50 { // Limit to first 50 entries
				traceData.WriteString(fmt.Sprintf("%s\n", entry.Name()))
				count++
			}
		}
		traceData.WriteString(fmt.Sprintf("\nTotal Prefetch files: %d\n", len(entries)))
	}
	
	// Collect recent file access info
	traceData.WriteString("\n=== Recent File Access ===\n")
	if recent, err := exec.Command("dir", "/O:D", "/T:W", "%USERPROFILE%\\Recent", "/B").Output(); err == nil {
		traceData.Write(recent)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     traceData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "file_system",
		},
		Size:     int64(traceData.Len()),
		Checksum: w.calculateChecksum(traceData.String()),
	}
	
	return result, nil
}

// collectInstalledSoftware collects installed software information
func (w *WindowsCollector) collectInstalledSoftware() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"installed_software",
		"Installed software list",
		"software",
		"command",
	)
	
	// Use wmic to get installed software information
	output, err := exec.Command("wmic", "product", "get", "name,version,vendor", "/format:csv").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect installed software: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "windows",
			Version:     w.version,
			Source:      "wmic",
		},
		Size:     int64(len(output)),
		Checksum: w.calculateChecksum(string(output)),
	}
	
	return result, nil
}

// calculateChecksum calculates SHA256 checksum for data
func (w *WindowsCollector) calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
