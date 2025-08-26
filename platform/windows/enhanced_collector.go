package windows

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// EnhancedWindowsCollector extends the basic Windows collector with forensic capabilities
type EnhancedWindowsCollector struct {
	*WindowsCollector
	artifactRegistry *collector.EnhancedArtifactRegistry
}

// NewEnhancedWindowsCollector creates a new enhanced Windows collector
func NewEnhancedWindowsCollector() *EnhancedWindowsCollector {
	return &EnhancedWindowsCollector{
		WindowsCollector: NewWindowsCollector(),
		artifactRegistry: collector.NewEnhancedArtifactRegistry(),
	}
}

// CollectEnhancedArtifacts collects artifacts based on priority and dependencies
func (e *EnhancedWindowsCollector) CollectEnhancedArtifacts(ctx context.Context, profile collector.CollectionProfile) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult
	
	// Collect volatile artifacts first (highest priority)
	volatileArtifacts := e.artifactRegistry.GetVolatileArtifacts()
	for _, artifact := range volatileArtifacts {
		if result, err := e.collectEnhancedArtifact(ctx, artifact); err == nil {
			results = append(results, result)
		} else {
			// Log error but continue with other artifacts
			fmt.Printf("Warning: Failed to collect volatile artifact %s: %v\n", artifact.Name, err)
		}
	}
	
	// Collect artifacts by priority
	byPriority := e.artifactRegistry.GetArtifactsByPriority()
	for priority := 1; priority <= 5; priority++ {
		if artifacts, exists := byPriority[priority]; exists {
			for _, artifact := range artifacts {
				// Skip if already collected (volatile artifacts)
				if artifact.Volatile {
					continue
				}
				
				// Check dependencies
				if e.checkDependencies(artifact, results) {
					if result, err := e.collectEnhancedArtifact(ctx, artifact); err == nil {
						results = append(results, result)
					} else {
						fmt.Printf("Warning: Failed to collect artifact %s: %v\n", artifact.Name, err)
					}
				}
			}
		}
	}
	
	return results, nil
}

// collectEnhancedArtifact collects a single enhanced artifact
func (e *EnhancedWindowsCollector) collectEnhancedArtifact(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.ForensicType {
	case "memory_analysis":
		return e.collectMemoryDump(ctx, artifact)
	case "registry_analysis":
		return e.collectRegistryHives(ctx, artifact)
	case "file_analysis":
		return e.collectFileMetadata(ctx, artifact)
	case "execution_analysis":
		return e.collectExecutionArtifacts(ctx, artifact)
	case "network_analysis":
		return e.collectNetworkArtifacts(ctx, artifact)
	case "log_analysis":
		return e.collectLogArtifacts(ctx, artifact)
	case "user_activity":
		return e.collectUserActivityArtifacts(ctx, artifact)
	case "device_analysis":
		return e.collectDeviceArtifacts(ctx, artifact)
	case "timeline_analysis":
		return e.collectTimelineData(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown forensic type: %s", artifact.ForensicType)
	}
}

// collectMemoryDump collects memory dump for analysis
func (e *EnhancedWindowsCollector) collectMemoryDump(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	// Note: This is a placeholder for memory dump collection
	// In a real implementation, this would use tools like DumpIt, WinPmem, or similar
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data: map[string]interface{}{
			"status":        "not_implemented",
			"message":       "Memory dump collection requires specialized tools",
			"recommendation": "Use DumpIt, WinPmem, or similar memory acquisition tools",
			"timestamp":     time.Now().Format(time.RFC3339),
		},
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "memory_analysis",
		},
		Size:     0,
		Checksum: "",
	}
	
	return result, nil
}

// collectRegistryHives collects registry hives for analysis
func (e *EnhancedWindowsCollector) collectRegistryHives(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var hiveData strings.Builder
	hives := strings.Split(artifact.Parameters["hives"], ",")
	
	for _, hive := range hives {
		hive = strings.TrimSpace(hive)
		hivePath := fmt.Sprintf("C:\\Windows\\System32\\config\\%s", hive)
		
		if info, err := os.Stat(hivePath); err == nil {
			hiveData.WriteString(fmt.Sprintf("=== %s Hive ===\n", hive))
			hiveData.WriteString(fmt.Sprintf("Path: %s\n", hivePath))
			hiveData.WriteString(fmt.Sprintf("Size: %d bytes\n", info.Size()))
			hiveData.WriteString(fmt.Sprintf("Modified: %s\n", info.ModTime().Format(time.RFC3339)))
			hiveData.WriteString("\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     hiveData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "registry_analysis",
		},
		Size:     int64(hiveData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectFileMetadata collects file system metadata
func (e *EnhancedWindowsCollector) collectFileMetadata(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var metadataData strings.Builder
	directories := strings.Split(artifact.Parameters["directories"], ",")
	
	for _, dir := range directories {
		dir = strings.TrimSpace(dir)
		if info, err := os.Stat(dir); err == nil {
			metadataData.WriteString(fmt.Sprintf("=== Directory: %s ===\n", dir))
			metadataData.WriteString(fmt.Sprintf("Exists: true\n"))
			metadataData.WriteString(fmt.Sprintf("Modified: %s\n", info.ModTime().Format(time.RFC3339)))
			metadataData.WriteString(fmt.Sprintf("Permissions: %s\n", info.Mode().String()))
			metadataData.WriteString("\n")
		} else {
			metadataData.WriteString(fmt.Sprintf("=== Directory: %s ===\n", dir))
			metadataData.WriteString(fmt.Sprintf("Exists: false\n"))
			metadataData.WriteString(fmt.Sprintf("Error: %v\n", err))
			metadataData.WriteString("\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     metadataData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "file_analysis",
		},
		Size:     int64(metadataData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectExecutionArtifacts collects execution-related artifacts
func (e *EnhancedWindowsCollector) collectExecutionArtifacts(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.Name {
	case "prefetch_files":
		return e.collectPrefetchFiles(ctx, artifact)
	case "scheduled_tasks":
		return e.collectScheduledTasks(ctx, artifact)
	case "startup_items":
		return e.collectStartupItems(ctx, artifact)
	case "process_tree":
		return e.collectProcessTree(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown execution artifact: %s", artifact.Name)
	}
}

// collectPrefetchFiles collects Windows Prefetch files
func (e *EnhancedWindowsCollector) collectPrefetchFiles(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	prefetchDir := artifact.Parameters["directory"]
	
	var prefetchData strings.Builder
	prefetchData.WriteString(fmt.Sprintf("=== Prefetch Files Directory: %s ===\n", prefetchDir))
	
	if entries, err := os.ReadDir(prefetchDir); err == nil {
		count := 0
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".pf") {
				count++
				if count <= 100 { // Limit output
					prefetchData.WriteString(fmt.Sprintf("File: %s\n", entry.Name()))
					if info, err := entry.Info(); err == nil {
						prefetchData.WriteString(fmt.Sprintf("  Size: %d bytes\n", info.Size()))
						prefetchData.WriteString(fmt.Sprintf("  Modified: %s\n", info.ModTime().Format(time.RFC3339)))
					}
				}
			}
		}
		prefetchData.WriteString(fmt.Sprintf("\nTotal .pf files: %d\n", count))
	} else {
		prefetchData.WriteString(fmt.Sprintf("Error reading directory: %v\n", err))
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     prefetchData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "execution_analysis",
		},
		Size:     int64(prefetchData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectNetworkArtifacts collects network-related artifacts
func (e *EnhancedWindowsCollector) collectNetworkArtifacts(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.Name {
	case "network_connections":
		return e.collectNetworkConnections(ctx, artifact)
	case "arp_cache":
		return e.collectARPCache(ctx, artifact)
	case "dns_cache":
		return e.collectDNSCache(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown network artifact: %s", artifact.Name)
	}
}

// collectNetworkConnections collects active network connections
func (e *EnhancedWindowsCollector) collectNetworkConnections(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var networkData strings.Builder
	
	// Get active connections with process information
	if output, err := exec.Command("netstat", "-ano").Output(); err == nil {
		networkData.WriteString("=== Active Network Connections ===\n")
		networkData.Write(output)
		networkData.WriteString("\n")
	}
	
	// Get listening ports
	if output, err := exec.Command("netstat", "-an").Output(); err == nil {
		networkData.WriteString("=== Listening Ports ===\n")
		networkData.Write(output)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     networkData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "network_analysis",
		},
		Size:     int64(networkData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectARPCache collects ARP cache
func (e *EnhancedWindowsCollector) collectARPCache(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	output, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect ARP cache: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "network_analysis",
		},
		Size:     int64(len(output)),
		Checksum: "",
	}
	
	return result, nil
}

// collectDNSCache collects DNS cache
func (e *EnhancedWindowsCollector) collectDNSCache(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	output, err := exec.Command("ipconfig", "/displaydns").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect DNS cache: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "network_analysis",
		},
		Size:     int64(len(output)),
		Checksum: "",
	}
	
	return result, nil
}

// collectLogArtifacts collects log-related artifacts
func (e *EnhancedWindowsCollector) collectLogArtifacts(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.Name {
	case "event_logs":
		return e.collectEventLogs(ctx, artifact)
	case "powershell_logs":
		return e.collectPowerShellLogs(ctx, artifact)
	case "sysmon_logs":
		return e.collectSysmonLogs(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown log artifact: %s", artifact.Name)
	}
}

// collectPowerShellLogs collects PowerShell logs
func (e *EnhancedWindowsCollector) collectPowerShellLogs(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var psData strings.Builder
	psData.WriteString("=== PowerShell Logs ===\n")
	
	// Check for PowerShell transcript logs
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		transcriptDir := filepath.Join(userProfile, "Documents", "WindowsPowerShell")
		if entries, err := os.ReadDir(transcriptDir); err == nil {
			psData.WriteString(fmt.Sprintf("Transcript directory: %s\n", transcriptDir))
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".txt") {
					psData.WriteString(fmt.Sprintf("Transcript: %s\n", entry.Name()))
				}
			}
		}
	}
	
	// Check PowerShell execution policy
	if output, err := exec.Command("powershell", "-Command", "Get-ExecutionPolicy").Output(); err == nil {
		psData.WriteString(fmt.Sprintf("Execution Policy: %s", strings.TrimSpace(string(output))))
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     psData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "log_analysis",
		},
		Size:     int64(psData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectUserActivityArtifacts collects user activity artifacts
func (e *EnhancedWindowsCollector) collectUserActivityArtifacts(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.Name {
	case "browser_history":
		return e.collectBrowserHistory(ctx, artifact)
	case "email_clients":
		return e.collectEmailClients(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown user activity artifact: %s", artifact.Name)
	}
}

// collectBrowserHistory collects browser history
func (e *EnhancedWindowsCollector) collectBrowserHistory(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var browserData strings.Builder
	browserData.WriteString("=== Browser History ===\n")
	
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		browsers := strings.Split(artifact.Parameters["browsers"], ",")
		for _, browser := range browsers {
			browser = strings.TrimSpace(browser)
			browserData.WriteString(fmt.Sprintf("Browser: %s\n", browser))
			
			// Check for common browser data locations
			browserPaths := map[string]string{
				"chrome":  filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Default"),
				"firefox": filepath.Join(userProfile, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles"),
				"edge":    filepath.Join(userProfile, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default"),
			}
			
			if path, exists := browserPaths[browser]; exists {
				if info, err := os.Stat(path); err == nil {
					browserData.WriteString(fmt.Sprintf("  Path: %s (exists)\n", path))
					browserData.WriteString(fmt.Sprintf("  Modified: %s\n", info.ModTime().Format(time.RFC3339)))
				} else {
					browserData.WriteString(fmt.Sprintf("  Path: %s (not found)\n", path))
				}
			}
			browserData.WriteString("\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     browserData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "user_activity",
		},
		Size:     int64(browserData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectDeviceArtifacts collects device-related artifacts
func (e *EnhancedWindowsCollector) collectDeviceArtifacts(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	switch artifact.Name {
	case "usb_devices":
		return e.collectUSBDevices(ctx, artifact)
	case "print_spooler":
		return e.collectPrintSpooler(ctx, artifact)
	default:
		return collector.ArtifactResult{}, fmt.Errorf("unknown device artifact: %s", artifact.Name)
	}
}

// collectUSBDevices collects USB device information
func (e *EnhancedWindowsCollector) collectUSBDevices(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var usbData strings.Builder
	usbData.WriteString("=== USB Devices ===\n")
	
	// Use WMI to get USB device information
	if output, err := exec.Command("wmic", "usbcontroller", "get", "name,deviceid", "/format:csv").Output(); err == nil {
		usbData.WriteString("USB Controllers:\n")
		usbData.Write(output)
		usbData.WriteString("\n")
	}
	
	// Get USB storage devices
	if output, err := exec.Command("wmic", "diskdrive", "where", "interfacetype='USB'", "get", "caption,size,serialnumber", "/format:csv").Output(); err == nil {
		usbData.WriteString("USB Storage Devices:\n")
		usbData.Write(output)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     usbData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "device_analysis",
		},
		Size:     int64(usbData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// collectTimelineData collects correlated timeline data
func (e *EnhancedWindowsCollector) collectTimelineData(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var timelineData strings.Builder
	timelineData.WriteString("=== Timeline Data ===\n")
	timelineData.WriteString("This artifact requires correlation of multiple data sources.\n")
	timelineData.WriteString("Dependencies: " + strings.Join(artifact.Dependencies, ", ") + "\n")
	timelineData.WriteString("Format: " + artifact.Parameters["format"] + "\n")
	timelineData.WriteString("Generated at: " + time.Now().Format(time.RFC3339) + "\n")
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     timelineData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "timeline_analysis",
		},
		Size:     int64(timelineData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

// checkDependencies checks if all dependencies for an artifact are satisfied
func (e *EnhancedWindowsCollector) checkDependencies(artifact collector.EnhancedArtifact, collectedResults []collector.ArtifactResult) bool {
	if len(artifact.Dependencies) == 0 {
		return true
	}
	
	collectedNames := make(map[string]bool)
	for _, result := range collectedResults {
		collectedNames[result.Artifact.Name] = true
	}
	
	for _, dependency := range artifact.Dependencies {
		if !collectedNames[dependency] {
			return false
		}
	}
	
	return true
}

// Helper methods for other execution artifacts
func (e *EnhancedWindowsCollector) collectScheduledTasks(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	// Enhanced scheduled task collection
	output, err := exec.Command("schtasks", "/query", "/fo", "csv", "/v").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect scheduled tasks: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "execution_analysis",
		},
		Size:     int64(len(output)),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectStartupItems(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var startupData strings.Builder
	startupData.WriteString("=== Startup Items ===\n")
	
	// Check common startup locations
	startupLocations := []string{
		os.Getenv("APPDATA") + "\\Microsoft\\Windows\\Start Menu\\Programs\\Startup",
		os.Getenv("PROGRAMDATA") + "\\Microsoft\\Windows\\Start Menu\\Programs\\Startup",
	}
	
	for _, location := range startupLocations {
		if entries, err := os.ReadDir(location); err == nil {
			startupData.WriteString(fmt.Sprintf("Location: %s\n", location))
			for _, entry := range entries {
				if info, err := entry.Info(); err == nil {
					startupData.WriteString(fmt.Sprintf("  %s (%d bytes, %s)\n", 
						entry.Name(), info.Size(), info.ModTime().Format(time.RFC3339)))
				}
			}
			startupData.WriteString("\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     startupData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "execution_analysis",
		},
		Size:     int64(startupData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectProcessTree(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	// Enhanced process tree collection
	output, err := exec.Command("tasklist", "/FO", "CSV", "/V", "/FI", "STATUS eq RUNNING").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect process tree: %w", err)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "execution_analysis",
		},
		Size:     int64(len(output)),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectEventLogs(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	// Enhanced event log collection
	var eventData strings.Builder
	logs := strings.Split(artifact.Parameters["logs"], ",")
	
	for _, logName := range logs {
		logName = strings.TrimSpace(logName)
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
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "log_analysis",
		},
		Size:     int64(eventData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectSysmonLogs(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	// Sysmon log collection
	var sysmonData strings.Builder
	sysmonData.WriteString("=== Sysmon Logs ===\n")
	
	// Check if Sysmon is installed and running
	if output, err := exec.Command("sc", "query", "SysmonDrv").Output(); err == nil {
		sysmonData.WriteString("Sysmon Driver Status:\n")
		sysmonData.Write(output)
		sysmonData.WriteString("\n")
	}
	
	// Try to get Sysmon events
	if events, err := exec.Command("wevtutil", "qe", "Microsoft-Windows-Sysmon/Operational", "/c:50", "/f:text").Output(); err == nil {
		sysmonData.WriteString("Recent Sysmon Events:\n")
		sysmonData.Write(events)
	} else {
		sysmonData.WriteString("Sysmon events not available or Sysmon not installed\n")
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     sysmonData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "log_analysis",
		},
		Size:     int64(sysmonData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectEmailClients(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var emailData strings.Builder
	emailData.WriteString("=== Email Clients ===\n")
	
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		clients := strings.Split(artifact.Parameters["clients"], ",")
		for _, client := range clients {
			client = strings.TrimSpace(client)
			emailData.WriteString(fmt.Sprintf("Client: %s\n", client))
			
			// Check for common email client locations
			clientPaths := map[string]string{
				"outlook":   filepath.Join(userProfile, "AppData", "Local", "Microsoft", "Outlook"),
				"thunderbird": filepath.Join(userProfile, "AppData", "Roaming", "Thunderbird", "Profiles"),
			}
			
			if path, exists := clientPaths[client]; exists {
				if info, err := os.Stat(path); err == nil {
					emailData.WriteString(fmt.Sprintf("  Path: %s (exists)\n", path))
					emailData.WriteString(fmt.Sprintf("  Modified: %s\n", info.ModTime().Format(time.RFC3339)))
				} else {
					emailData.WriteString(fmt.Sprintf("  Path: %s (not found)\n", path))
				}
			}
			emailData.WriteString("\n")
		}
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     emailData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "user_activity",
		},
		Size:     int64(emailData.Len()),
		Checksum: "",
	}
	
	return result, nil
}

func (e *EnhancedWindowsCollector) collectPrintSpooler(ctx context.Context, artifact collector.EnhancedArtifact) (collector.ArtifactResult, error) {
	var printData strings.Builder
	printData.WriteString("=== Print Spooler ===\n")
	
	// Get print spooler service status
	if output, err := exec.Command("sc", "query", "Spooler").Output(); err == nil {
		printData.WriteString("Spooler Service Status:\n")
		printData.Write(output)
		printData.WriteString("\n")
	}
	
	// Get printer information
	if output, err := exec.Command("wmic", "printer", "get", "name,portname,drivername", "/format:csv").Output(); err == nil {
		printData.WriteString("Installed Printers:\n")
		printData.Write(output)
	}
	
	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     printData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "enhanced_windows",
			Version:     e.version,
			Source:      "device_analysis",
		},
		Size:     int64(printData.Len()),
		Checksum: "",
	}
	
	return result, nil
}
