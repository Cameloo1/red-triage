package logging

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// WindowsEventLogParser parses Windows Event Log format
type WindowsEventLogParser struct{}

func (p *WindowsEventLogParser) ParseLine(line string) (*LogEntry, error) {
	// Windows Event Log format: EventID, Level, Source, Time, Message
	// Example: 4624,Information,Security,2024-01-01T12:00:00.000Z,An account was successfully logged on.
	
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid Windows Event Log format")
	}
	
	eventID := strings.TrimSpace(parts[0])
	level := strings.TrimSpace(parts[1])
	source := strings.TrimSpace(parts[2])
	timestampStr := strings.TrimSpace(parts[3])
	message := strings.TrimSpace(parts[4])
	
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp = time.Now() // Fallback to current time
	}
	
	// Determine severity based on level
	severity := p.getSeverityFromLevel(level)
	
	// Extract additional information from message
	user, process, ip := p.extractInfoFromMessage(message)
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    source,
		Level:     level,
		Message:   message,
		EventID:   eventID,
		Category:  p.getCategoryFromEventID(eventID),
		User:      user,
		Process:   process,
		IPAddress: ip,
		RawData:   line,
		Metadata:  make(map[string]interface{}),
		Severity:  severity,
		Tags:      p.getTagsFromEventID(eventID),
	}
	
	return entry, nil
}

func (p *WindowsEventLogParser) GetFormatName() string {
	return "windows_event"
}

func (p *WindowsEventLogParser) IsCompatible(line string) bool {
	// Check if line matches Windows Event Log format
	parts := strings.Split(line, ",")
	return len(parts) >= 5 && p.isValidEventID(parts[0])
}

func (p *WindowsEventLogParser) getSeverityFromLevel(level string) int {
	switch strings.ToLower(level) {
	case "critical", "error":
		return 5
	case "warning":
		return 4
	case "information":
		return 2
	case "verbose", "debug":
		return 1
	default:
		return 3
	}
}

func (p *WindowsEventLogParser) getCategoryFromEventID(eventID string) string {
	// Map common Event IDs to categories
	eventIDMap := map[string]string{
		"4624": "login",
		"4625": "login_failure",
		"4634": "logout",
		"4688": "process_creation",
		"4689": "process_termination",
		"4697": "service_installation",
		"4698": "scheduled_task",
		"4700": "scheduled_task_creation",
		"4701": "scheduled_task_deletion",
		"4702": "scheduled_task_modification",
	}
	
	if category, exists := eventIDMap[eventID]; exists {
		return category
	}
	return "system"
}

func (p *WindowsEventLogParser) getTagsFromEventID(eventID string) []string {
	// Map common Event IDs to tags
	eventIDMap := map[string][]string{
		"4624": {"authentication", "success"},
		"4625": {"authentication", "failure"},
		"4688": {"process", "creation"},
		"4697": {"service", "installation"},
		"4698": {"scheduled_task", "creation"},
	}
	
	if tags, exists := eventIDMap[eventID]; exists {
		return tags
	}
	return []string{"system"}
}

func (p *WindowsEventLogParser) isValidEventID(eventID string) bool {
	// Check if event ID is numeric
	if _, err := strconv.Atoi(strings.TrimSpace(eventID)); err != nil {
		return false
	}
	return true
}

func (p *WindowsEventLogParser) extractInfoFromMessage(message string) (user, process, ip string) {
	// Extract user information
	if userMatch := regexp.MustCompile(`(?i)user.*?:\s*([^\s,]+)`).FindStringSubmatch(message); len(userMatch) > 1 {
		user = userMatch[1]
	}
	
	// Extract process information
	if processMatch := regexp.MustCompile(`(?i)process.*?:\s*([^\s,]+)`).FindStringSubmatch(message); len(processMatch) > 1 {
		process = processMatch[1]
	}
	
	// Extract IP address
	if ipMatch := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`).FindString(message); ipMatch != "" {
		ip = ipMatch
	}
	
	return user, process, ip
}

// SysmonLogParser parses Sysmon log format
type SysmonLogParser struct{}

func (p *SysmonLogParser) ParseLine(line string) (*LogEntry, error) {
	// Sysmon format: EventID, Time, Process, Command, etc.
	// Example: 1,2024-01-01T12:00:00.000Z,notepad.exe,C:\Windows\System32\notepad.exe,1234
	
	parts := strings.Split(line, ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid Sysmon format")
	}
	
	eventID := strings.TrimSpace(parts[0])
	timestampStr := strings.TrimSpace(parts[1])
	process := strings.TrimSpace(parts[2])
	command := strings.TrimSpace(parts[3])
	
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp = time.Now()
	}
	
	// Determine severity and category based on event ID
	severity, category := p.getSeverityAndCategory(eventID)
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    "sysmon",
		Level:     "information",
		Message:   fmt.Sprintf("Sysmon Event %s: %s", eventID, command),
		EventID:   eventID,
		Category:  category,
		Process:   process,
		Command:   command,
		RawData:   line,
		Metadata:  make(map[string]interface{}),
		Severity:  severity,
		Tags:      p.getTagsFromEventID(eventID),
	}
	
	return entry, nil
}

func (p *SysmonLogParser) GetFormatName() string {
	return "sysmon"
}

func (p *SysmonLogParser) IsCompatible(line string) bool {
	parts := strings.Split(line, ",")
	return len(parts) >= 4 && p.isValidSysmonEventID(parts[0])
}

func (p *SysmonLogParser) getSeverityAndCategory(eventID string) (int, string) {
	// Map Sysmon Event IDs to severity and category
	eventIDMap := map[string]struct {
		severity int
		category string
	}{
		"1":  {3, "process_creation"},
		"2":  {3, "file_time_change"},
		"3":  {4, "network_connection"},
		"4":  {3, "service_state_change"},
		"5":  {4, "process_termination"},
		"6":  {4, "driver_load"},
		"7":  {4, "image_load"},
		"8":  {4, "create_remote_thread"},
		"9":  {4, "raw_access_read"},
		"10": {4, "process_access"},
		"11": {4, "file_create"},
		"12": {4, "registry_event"},
		"13": {4, "registry_event"},
		"14": {4, "registry_event"},
		"15": {4, "file_create_stream_hash"},
		"16": {4, "service_configuration_change"},
		"17": {4, "pipe_created"},
		"18": {4, "pipe_created"},
		"19": {4, "wmi_event"},
		"20": {4, "wmi_event"},
		"21": {4, "wmi_event"},
		"22": {4, "dns_query"},
		"23": {4, "file_delete"},
		"24": {4, "clipboard_change"},
		"25": {4, "process_tampering"},
		"26": {4, "file_delete_detected"},
		"27": {4, "file_block_executable"},
		"28": {4, "file_block_executable"},
	}
	
	if info, exists := eventIDMap[eventID]; exists {
		return info.severity, info.category
	}
	
	return 3, "sysmon"
}

func (p *SysmonLogParser) getTagsFromEventID(eventID string) []string {
	// Map Sysmon Event IDs to tags
	eventIDMap := map[string][]string{
		"1":  {"process", "creation"},
		"3":  {"network", "connection"},
		"5":  {"process", "termination"},
		"8":  {"process", "injection"},
		"10": {"process", "access"},
		"11": {"file", "creation"},
		"12": {"registry", "modification"},
		"22": {"dns", "query"},
		"23": {"file", "deletion"},
		"25": {"process", "tampering"},
	}
	
	if tags, exists := eventIDMap[eventID]; exists {
		return tags
	}
	
	return []string{"sysmon"}
}

func (p *SysmonLogParser) isValidSysmonEventID(eventID string) bool {
	if _, err := strconv.Atoi(strings.TrimSpace(eventID)); err != nil {
		return false
	}
	return true
}

// PowerShellLogParser parses PowerShell log format
type PowerShellLogParser struct{}

func (p *PowerShellLogParser) ParseLine(line string) (*LogEntry, error) {
	// PowerShell format: Time, Level, Message, Command
	// Example: 2024-01-01T12:00:00.000Z,Information,Command executed,Get-Process
	
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid PowerShell log format")
	}
	
	timestampStr := strings.TrimSpace(parts[0])
	level := strings.TrimSpace(parts[1])
	message := strings.TrimSpace(parts[2])
	
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp = time.Now()
	}
	
	// Extract command if present
	command := ""
	if len(parts) > 3 {
		command = strings.TrimSpace(parts[3])
	}
	
	// Determine severity
	severity := p.getSeverityFromLevel(level)
	
	// Check for suspicious commands
	tags := p.getTagsFromCommand(command)
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    "powershell",
		Level:     level,
		Message:   message,
		Category:  "powershell",
		Command:   command,
		RawData:   line,
		Metadata:  make(map[string]interface{}),
		Severity:  severity,
		Tags:      tags,
	}
	
	return entry, nil
}

func (p *PowerShellLogParser) GetFormatName() string {
	return "powershell"
}

func (p *PowerShellLogParser) IsCompatible(line string) bool {
	parts := strings.Split(line, ",")
	return len(parts) >= 3 && p.isValidTimestamp(parts[0])
}

func (p *PowerShellLogParser) getSeverityFromLevel(level string) int {
	switch strings.ToLower(level) {
	case "error":
		return 4
	case "warning":
		return 3
	case "information":
		return 2
	case "verbose", "debug":
		return 1
	default:
		return 2
	}
}

func (p *PowerShellLogParser) getTagsFromCommand(command string) []string {
	tags := []string{"powershell"}
	
	// Check for suspicious commands
	suspiciousPatterns := []string{
		"invoke-expression", "iex", "downloadstring", "webclient",
		"net.webclient", "system.net.webclient", "invoke-webrequest",
		"start-process", "start-job", "invoke-command",
	}
	
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(command), pattern) {
			tags = append(tags, "suspicious", "execution")
			break
		}
	}
	
	return tags
}

func (p *PowerShellLogParser) isValidTimestamp(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, strings.TrimSpace(timestamp))
	return err == nil
}

// GenericLogParser parses generic text log format
type GenericLogParser struct{}

func (p *GenericLogParser) ParseLine(line string) (*LogEntry, error) {
	// Generic format: try to extract timestamp and message
	// Example: 2024-01-01 12:00:00 [INFO] Application started
	
	// Try to extract timestamp
	timestamp := time.Now()
	message := line
	
	// Common timestamp patterns
	timestampPatterns := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000Z",
		"Jan 2 15:04:05",
		"02/01/2006 15:04:05",
	}
	
	for _, pattern := range timestampPatterns {
		if idx := strings.Index(line, " "); idx != -1 {
			timeStr := line[:idx]
			if parsed, err := time.Parse(pattern, timeStr); err == nil {
				timestamp = parsed
				message = line[idx+1:]
				break
			}
		}
	}
	
	// Extract level if present
	level := "information"
	if strings.Contains(strings.ToUpper(message), "[ERROR]") {
		level = "error"
	} else if strings.Contains(strings.ToUpper(message), "[WARN]") {
		level = "warning"
	} else if strings.Contains(strings.ToUpper(message), "[DEBUG]") {
		level = "debug"
	}
	
	// Determine severity
	severity := p.getSeverityFromLevel(level)
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    "generic",
		Level:     level,
		Message:   message,
		Category:  "system",
		RawData:   line,
		Metadata:  make(map[string]interface{}),
		Severity:  severity,
		Tags:      []string{"generic"},
	}
	
	return entry, nil
}

func (p *GenericLogParser) GetFormatName() string {
	return "generic"
}

func (p *GenericLogParser) IsCompatible(line string) bool {
	// Generic parser is always compatible as fallback
	return true
}

func (p *GenericLogParser) getSeverityFromLevel(level string) int {
	switch strings.ToLower(level) {
	case "error":
		return 4
	case "warning":
		return 3
	case "information":
		return 2
	case "debug", "verbose":
		return 1
	default:
		return 2
	}
}

// JSONLogParser parses JSON log format
type JSONLogParser struct{}

func (p *JSONLogParser) ParseLine(line string) (*LogEntry, error) {
	// JSON format: {"timestamp": "...", "level": "...", "message": "..."}
	
	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Extract fields
	timestamp := time.Now()
	if ts, exists := jsonData["timestamp"]; exists {
		if tsStr, ok := ts.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				timestamp = parsed
			}
		}
	}
	
	level := "information"
	if lvl, exists := jsonData["level"]; exists {
		if lvlStr, ok := lvl.(string); ok {
			level = lvlStr
		}
	}
	
	message := ""
	if msg, exists := jsonData["message"]; exists {
		if msgStr, ok := msg.(string); ok {
			message = msgStr
		}
	}
	
	// Extract additional fields
	user := ""
	if u, exists := jsonData["user"]; exists {
		if uStr, ok := u.(string); ok {
			user = uStr
		}
	}
	
	process := ""
	if proc, exists := jsonData["process"]; exists {
		if procStr, ok := proc.(string); ok {
			process = procStr
		}
	}
	
	// Determine severity
	severity := p.getSeverityFromLevel(level)
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    "json",
		Level:     level,
		Message:   message,
		Category:  "application",
		User:      user,
		Process:   process,
		RawData:   line,
		Metadata:  jsonData,
		Severity:  severity,
		Tags:      []string{"json"},
	}
	
	return entry, nil
}

func (p *JSONLogParser) GetFormatName() string {
	return "json"
}

func (p *JSONLogParser) IsCompatible(line string) bool {
	// Check if line starts with { and ends with }
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}")
}

func (p *JSONLogParser) getSeverityFromLevel(level string) int {
	switch strings.ToLower(level) {
	case "fatal", "critical", "error":
		return 5
	case "warning":
		return 4
	case "info", "information":
		return 2
	case "debug", "trace":
		return 1
	default:
		return 3
	}
}
