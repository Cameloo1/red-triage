package collector

import (
	// No imports needed for this file
)

// EnhancedArtifact represents an enhanced collectable artifact with forensic capabilities
type EnhancedArtifact struct {
	Artifact
	ForensicType string            // Type of forensic artifact (memory, registry, file, etc.)
	Volatility   bool              // Whether this is volatile data that needs immediate collection
	Priority     int               // Collection priority (1=highest, 5=lowest)
	Dependencies []string          // Other artifacts this depends on
	Parameters  map[string]string // Collection parameters
}

// NewEnhancedArtifact creates a new enhanced artifact
func NewEnhancedArtifact(name, description, category, artifactType, forensicType string, priority int) EnhancedArtifact {
	return EnhancedArtifact{
		Artifact: Artifact{
			Name:        name,
			Description: description,
			Category:    category,
			Type:        artifactType,
			Platform:    "windows", // Focus on Windows initially
			Volatile:    false,
			Enabled:     true,
			Parameters:  make(map[string]string),
		},
		ForensicType: forensicType,
		Priority:     priority,
		Dependencies: make([]string, 0),
		Parameters:  make(map[string]string),
	}
}

// EnhancedArtifactRegistry contains all available enhanced artifacts
type EnhancedArtifactRegistry struct {
	artifacts map[string]EnhancedArtifact
}

// NewEnhancedArtifactRegistry creates a new enhanced artifact registry
func NewEnhancedArtifactRegistry() *EnhancedArtifactRegistry {
	registry := &EnhancedArtifactRegistry{
		artifacts: make(map[string]EnhancedArtifact),
	}
	
	// Register all enhanced artifacts
	registry.registerEnhancedArtifacts()
	
	return registry
}

// registerEnhancedArtifacts registers all enhanced artifact types
func (r *EnhancedArtifactRegistry) registerEnhancedArtifacts() {
	// System Artifacts (Priority 1 - Critical)
	memoryDump := NewEnhancedArtifact(
		"memory_dump",
		"Complete memory dump for analysis",
		"memory",
		"dump",
		"memory_analysis",
		1,
	)
	memoryDump.Volatile = true
	memoryDump.Parameters["format"] = "raw"
	memoryDump.Parameters["compression"] = "gzip"
	r.artifacts["memory_dump"] = memoryDump
	
	registryHives := NewEnhancedArtifact(
		"registry_hives",
		"Complete registry hives (SYSTEM, SOFTWARE, SAM, SECURITY)",
		"registry",
		"hive",
		"registry_analysis",
		1,
	)
	registryHives.Parameters["hives"] = "SYSTEM,SOFTWARE,SAM,SECURITY"
	registryHives.Parameters["backup"] = "true"
	r.artifacts["registry_hives"] = registryHives
	
	// File System Artifacts (Priority 2 - High)
	fileMetadata := NewEnhancedArtifact(
		"file_metadata",
		"File system metadata and timestamps",
		"filesystem",
		"metadata",
		"file_analysis",
		2,
	)
	fileMetadata.Parameters["directories"] = "C:\\Windows,C:\\Program Files,C:\\Users"
	fileMetadata.Parameters["include_hidden"] = "true"
	r.artifacts["file_metadata"] = fileMetadata
	
	prefetchFiles := NewEnhancedArtifact(
		"prefetch_files",
		"Windows Prefetch files for execution analysis",
		"filesystem",
		"prefetch",
		"execution_analysis",
		2,
	)
	prefetchFiles.Parameters["directory"] = "C:\\Windows\\Prefetch"
	prefetchFiles.Parameters["max_age"] = "30d"
	r.artifacts["prefetch_files"] = prefetchFiles
	
	usnJournal := NewEnhancedArtifact(
		"usn_journal",
		"USN Journal for file system change tracking",
		"filesystem",
		"usn",
		"timeline_analysis",
		2,
	)
	usnJournal.Parameters["max_entries"] = "10000"
	usnJournal.Parameters["include_deleted"] = "true"
	r.artifacts["usn_journal"] = usnJournal
	
	// Network Artifacts (Priority 2 - High)
	networkConnections := NewEnhancedArtifact(
		"network_connections",
		"Active network connections and routing tables",
		"network",
		"connection",
		"network_analysis",
		2,
	)
	networkConnections.Volatile = true
	networkConnections.Parameters["include_listening"] = "true"
	networkConnections.Parameters["include_processes"] = "true"
	r.artifacts["network_connections"] = networkConnections
	
	arpCache := NewEnhancedArtifact(
		"arp_cache",
		"ARP cache for network neighbor analysis",
		"network",
		"arp",
		"network_analysis",
		2,
	)
	arpCache.Volatile = true
	r.artifacts["arp_cache"] = arpCache
	
	dnsCache := NewEnhancedArtifact(
		"dns_cache",
		"DNS cache for domain resolution analysis",
		"network",
		"dns",
		"network_analysis",
		2,
	)
	dnsCache.Volatile = true
	r.artifacts["dns_cache"] = dnsCache
	
	// Execution Artifacts (Priority 2 - High)
	r.artifacts["scheduled_tasks"] = NewEnhancedArtifact(
		"scheduled_tasks",
		"Detailed scheduled task information",
		"execution",
		"task",
		"persistence_analysis",
		2,
	)
	r.artifacts["scheduled_tasks"].Parameters["include_disabled"] = "true"
	r.artifacts["scheduled_tasks"].Parameters["include_history"] = "true"
	
	r.artifacts["startup_items"] = NewEnhancedArtifact(
		"startup_items",
		"System startup items and autoruns",
		"execution",
		"startup",
		"persistence_analysis",
		2,
	)
	r.artifacts["startup_items"].Parameters["locations"] = "registry,startup_folders,services"
	
	processTree := NewEnhancedArtifact(
		"process_tree",
		"Complete process tree with parent-child relationships",
		"execution",
		"process_tree",
		"process_analysis",
		2,
	)
	processTree.Volatile = true
	processTree.Parameters["include_modules"] = "true"
	processTree.Parameters["include_handles"] = "true"
	r.artifacts["process_tree"] = processTree
	
	// Log Artifacts (Priority 3 - Medium)
	r.artifacts["event_logs"] = NewEnhancedArtifact(
		"event_logs",
		"Comprehensive Windows Event Logs",
		"logs",
		"event",
		"log_analysis",
		3,
	)
	r.artifacts["event_logs"].Parameters["logs"] = "Security,System,Application,Microsoft-Windows-Sysmon/Operational"
	r.artifacts["event_logs"].Parameters["max_age"] = "7d"
	r.artifacts["event_logs"].Parameters["include_evtx"] = "true"
	
	r.artifacts["powershell_logs"] = NewEnhancedArtifact(
		"powershell_logs",
		"PowerShell execution logs and command history",
		"logs",
		"powershell",
		"execution_analysis",
		3,
	)
	r.artifacts["powershell_logs"].Parameters["include_transcript"] = "true"
	r.artifacts["powershell_logs"].Parameters["include_modules"] = "true"
	
	r.artifacts["sysmon_logs"] = NewEnhancedArtifact(
		"sysmon_logs",
		"Sysmon logs for advanced monitoring",
		"logs",
		"sysmon",
		"advanced_monitoring",
		3,
	)
	r.artifacts["sysmon_logs"].Parameters["config"] = "default"
	r.artifacts["sysmon_logs"].Parameters["max_age"] = "30d"
	
	// Browser and Application Artifacts (Priority 3 - Medium)
	r.artifacts["browser_history"] = NewEnhancedArtifact(
		"browser_history",
		"Browser history, cache, and cookies",
		"application",
		"browser",
		"user_activity",
		3,
	)
	r.artifacts["browser_history"].Parameters["browsers"] = "chrome,firefox,edge,ie"
	r.artifacts["browser_history"].Parameters["include_cache"] = "true"
	r.artifacts["browser_history"].Parameters["include_cookies"] = "true"
	
	r.artifacts["email_clients"] = NewEnhancedArtifact(
		"email_clients",
		"Email client data and configurations",
		"application",
		"email",
		"communication_analysis",
		3,
	)
	r.artifacts["email_clients"].Parameters["clients"] = "outlook,thunderbird,mail_app"
	r.artifacts["email_clients"].Parameters["include_attachments"] = "false"
	
	// Hardware and Device Artifacts (Priority 4 - Low)
	r.artifacts["usb_devices"] = NewEnhancedArtifact(
		"usb_devices",
		"USB device history and registry entries",
		"hardware",
		"usb",
		"device_analysis",
		4,
	)
	r.artifacts["usb_devices"].Parameters["include_removed"] = "true"
	r.artifacts["usb_devices"].Parameters["include_serial_numbers"] = "true"
	
	r.artifacts["print_spooler"] = NewEnhancedArtifact(
		"print_spooler",
		"Print spooler data and job history",
		"hardware",
		"print",
		"activity_analysis",
		4,
	)
	r.artifacts["print_spooler"].Parameters["include_jobs"] = "true"
	r.artifacts["print_spooler"].Parameters["include_drivers"] = "true"
	
	// Cloud and Storage Artifacts (Priority 4 - Low)
	r.artifacts["cloud_storage"] = NewEnhancedArtifact(
		"cloud_storage",
		"Cloud storage artifacts and sync data",
		"storage",
		"cloud",
		"data_analysis",
		4,
	)
	r.artifacts["cloud_storage"].Parameters["providers"] = "onedrive,dropbox,google_drive"
	r.artifacts["cloud_storage"].Parameters["include_sync_status"] = "true"
	
	// Timeline and Correlation Artifacts (Priority 5 - Lowest)
	timelineData := NewEnhancedArtifact(
		"timeline_data",
		"Correlated timeline data from multiple sources",
		"timeline",
		"correlation",
		"timeline_analysis",
		5,
	)
	timelineData.Dependencies = []string{
		"file_metadata", "event_logs", "prefetch_files", "usn_journal",
	}
	timelineData.Parameters["format"] = "plaso"
	timelineData.Parameters["include_metadata"] = "true"
	r.artifacts["timeline_data"] = timelineData
}

// GetArtifact returns an artifact by name
func (r *EnhancedArtifactRegistry) GetArtifact(name string) (EnhancedArtifact, bool) {
	artifact, exists := r.artifacts[name]
	return artifact, exists
}

// GetAllArtifacts returns all registered artifacts
func (r *EnhancedArtifactRegistry) GetAllArtifacts() map[string]EnhancedArtifact {
	return r.artifacts
}

// GetArtifactsByPriority returns artifacts grouped by priority
func (r *EnhancedArtifactRegistry) GetArtifactsByPriority() map[int][]EnhancedArtifact {
	byPriority := make(map[int][]EnhancedArtifact)
	
	for _, artifact := range r.artifacts {
		priority := artifact.Priority
		byPriority[priority] = append(byPriority[priority], artifact)
	}
	
	return byPriority
}

// GetArtifactsByCategory returns artifacts grouped by category
func (r *EnhancedArtifactRegistry) GetArtifactsByCategory() map[string][]EnhancedArtifact {
	byCategory := make(map[string][]EnhancedArtifact)
	
	for _, artifact := range r.artifacts {
		category := artifact.Category
		byCategory[category] = append(byCategory[category], artifact)
	}
	
	return byCategory
}

// GetVolatileArtifacts returns all volatile artifacts
func (r *EnhancedArtifactRegistry) GetVolatileArtifacts() []EnhancedArtifact {
	var volatile []EnhancedArtifact
	
	for _, artifact := range r.artifacts {
		if artifact.Volatile {
			volatile = append(volatile, artifact)
		}
	}
	
	return volatile
}

// GetArtifactsByDependency returns artifacts that depend on a specific artifact
func (r *EnhancedArtifactRegistry) GetArtifactsByDependency(dependencyName string) []EnhancedArtifact {
	var dependent []EnhancedArtifact
	
	for _, artifact := range r.artifacts {
		for _, dep := range artifact.Dependencies {
			if dep == dependencyName {
				dependent = append(dependent, artifact)
				break
			}
		}
	}
	
	return dependent
}
