package logging

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	EventID     string                 `json:"event_id,omitempty"`
	Category    string                 `json:"category,omitempty"`
	User        string                 `json:"user,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	Process     string                 `json:"process,omitempty"`
	Command     string                 `json:"command,omitempty"`
	RawData     string                 `json:"raw_data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Severity    int                    `json:"severity"` // 1=low, 5=critical
	Tags        []string               `json:"tags"`
}

// LogParser represents the enhanced log parsing engine
type LogParser struct {
	parsers map[string]LogFormatParser
	rules   []LogAnalysisRule
}

// LogFormatParser defines the interface for parsing different log formats
type LogFormatParser interface {
	ParseLine(line string) (*LogEntry, error)
	GetFormatName() string
	IsCompatible(line string) bool
}

// LogAnalysisRule defines a rule for log analysis
type LogAnalysisRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pattern     string   `json:"pattern"`
	Severity    int      `json:"severity"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Action      string   `json:"action"` // alert, block, log, etc.
}

// LogAnalysisResult represents the result of log analysis
type LogAnalysisResult struct {
	Rule        LogAnalysisRule `json:"rule"`
	Entry       LogEntry        `json:"entry"`
	Confidence  float64         `json:"confidence"`
	Timestamp   time.Time       `json:"timestamp"`
	Description string          `json:"description"`
}

// NewLogParser creates a new enhanced log parser
func NewLogParser() *LogParser {
	parser := &LogParser{
		parsers: make(map[string]LogFormatParser),
		rules:   make([]LogAnalysisRule, 0),
	}
	
	// Register built-in parsers
	parser.registerBuiltInParsers()
	
	// Load built-in analysis rules
	parser.loadBuiltInRules()
	
	return parser
}

// registerBuiltInParsers registers all built-in log format parsers
func (lp *LogParser) registerBuiltInParsers() {
	// Windows Event Log parser
	lp.parsers["windows_event"] = &WindowsEventLogParser{}
	
	// Sysmon parser
	lp.parsers["sysmon"] = &SysmonLogParser{}
	
	// PowerShell parser
	lp.parsers["powershell"] = &PowerShellLogParser{}
	
	// Generic text log parser
	lp.parsers["generic"] = &GenericLogParser{}
	
	// JSON log parser
	lp.parsers["json"] = &JSONLogParser{}
}

// loadBuiltInRules loads built-in log analysis rules
func (lp *LogParser) loadBuiltInRules() {
	builtInRules := []LogAnalysisRule{
		{
			ID:          "LOG001",
			Name:        "Failed Login Attempts",
			Description: "Detects multiple failed login attempts",
			Pattern:     `(?i)failed.*log.*in|log.*in.*failed|authentication.*failed`,
			Severity:    3,
			Category:    "authentication",
			Tags:        []string{"login", "authentication", "brute_force"},
			Action:      "alert",
		},
		{
			ID:          "LOG002",
			Name:        "Privilege Escalation",
			Description: "Detects privilege escalation attempts",
			Pattern:     `(?i)privilege.*escalation|elevation.*privilege|runas|sudo`,
			Severity:    4,
			Category:    "privilege_escalation",
			Tags:        []string{"privilege", "escalation", "security"},
			Action:      "alert",
		},
		{
			ID:          "LOG003",
			Name:        "Suspicious PowerShell Commands",
			Description: "Detects suspicious PowerShell commands",
			Pattern:     `(?i)invoke.*expression|iex|downloadstring|webclient|net\.webclient`,
			Severity:    4,
			Category:    "powershell",
			Tags:        []string{"powershell", "malware", "execution"},
			Action:      "alert",
		},
		{
			ID:          "LOG004",
			Name:        "Process Injection",
			Description: "Detects process injection attempts",
			Pattern:     `(?i)process.*injection|inject.*process|createremotethread`,
			Severity:    5,
			Category:    "process_injection",
			Tags:        []string{"injection", "malware", "process"},
			Action:      "alert",
		},
		{
			ID:          "LOG005",
			Name:        "Network Scanning",
			Description: "Detects network scanning activity",
			Pattern:     `(?i)port.*scan|network.*scan|nmap|masscan`,
			Severity:    3,
			Category:    "network",
			Tags:        []string{"scanning", "network", "reconnaissance"},
			Action:      "log",
		},
	}
	
	lp.rules = append(lp.rules, builtInRules...)
}

// ParseLogFile parses a log file and returns parsed entries
func (lp *LogParser) ParseLogFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()
	
	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	// Determine log format from first few lines
	format := lp.detectLogFormat(file)
	parser, exists := lp.parsers[format]
	if !exists {
		parser = lp.parsers["generic"] // Fallback to generic parser
	}
	
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		
		if entry, err := parser.ParseLine(line); err == nil {
			entry.Source = filePath
			entry.Metadata["line_number"] = lineNumber
			entries = append(entries, *entry)
		}
	}
	
	return entries, scanner.Err()
}

// detectLogFormat detects the log format from the file content
func (lp *LogParser) detectLogFormat(file *os.File) string {
	// Reset file pointer
	file.Seek(0, 0)
	
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, 10)
	
	// Read first 10 lines to determine format
	for i := 0; i < 10 && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}
	
	// Check each parser for compatibility
	for format, parser := range lp.parsers {
		for _, line := range lines {
			if parser.IsCompatible(line) {
				return format
			}
		}
	}
	
	return "generic" // Default fallback
}

// AnalyzeLogs analyzes parsed log entries using defined rules
func (lp *LogParser) AnalyzeLogs(entries []LogEntry) []LogAnalysisResult {
	var results []LogAnalysisResult
	
	for _, entry := range entries {
		for _, rule := range lp.rules {
			if match := lp.applyRule(rule, entry); match != nil {
				results = append(results, *match)
			}
		}
	}
	
	return results
}

// applyRule applies a single analysis rule to a log entry
func (lp *LogParser) applyRule(rule LogAnalysisRule, entry LogEntry) *LogAnalysisResult {
	// Check if the rule pattern matches the entry
	matched, confidence := lp.matchPattern(rule.Pattern, entry)
	if !matched {
		return nil
	}
	
	// Create analysis result
	result := &LogAnalysisResult{
		Rule:        rule,
		Entry:       entry,
		Confidence:  confidence,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Rule '%s' matched: %s", rule.Name, rule.Description),
	}
	
	return result
}

// matchPattern checks if a pattern matches a log entry
func (lp *LogParser) matchPattern(pattern string, entry LogEntry) (bool, float64) {
	// Check message field first
	if matched, confidence := lp.matchString(pattern, entry.Message); matched {
		return true, confidence
	}
	
	// Check command field
	if matched, confidence := lp.matchString(pattern, entry.Command); matched {
		return true, confidence
	}
	
	// Check raw data
	if matched, confidence := lp.matchString(pattern, entry.RawData); matched {
		return true, confidence
	}
	
	return false, 0.0
}

// matchString performs pattern matching on a string
func (lp *LogParser) matchString(pattern, text string) (bool, float64) {
	if pattern == "" || text == "" {
		return false, 0.0
	}
	
	// Try regex first
	if regex, err := regexp.Compile(pattern); err == nil {
		if regex.MatchString(text) {
			// Calculate confidence based on match quality
			confidence := 0.8
			if strings.Contains(strings.ToLower(text), strings.ToLower(pattern)) {
				confidence = 0.9
			}
			return true, confidence
		}
	}
	
	// Fallback to simple string matching
	if strings.Contains(strings.ToLower(text), strings.ToLower(pattern)) {
		return true, 0.7
	}
	
	return false, 0.0
}

// GenerateTimeline generates a timeline from log entries
func (lp *LogParser) GenerateTimeline(entries []LogEntry) []TimelineEvent {
	var timeline []TimelineEvent
	
	for _, entry := range entries {
		event := TimelineEvent{
			Timestamp: entry.Timestamp,
			Source:    entry.Source,
			Type:      entry.Category,
			Description: entry.Message,
			Severity:  entry.Severity,
			User:      entry.User,
			Process:   entry.Process,
			IPAddress: entry.IPAddress,
			Tags:      entry.Tags,
		}
		timeline = append(timeline, event)
	}
	
	// Sort timeline by timestamp
	// This would be implemented with a proper sort
	return timeline
}

// DetectAnomalies detects anomalies in log entries
func (lp *LogParser) DetectAnomalies(entries []LogEntry) []Anomaly {
	var anomalies []Anomaly
	
	// Group entries by user, process, IP, etc.
	userActivity := make(map[string][]LogEntry)
	processActivity := make(map[string][]LogEntry)
	ipActivity := make(map[string][]LogEntry)
	
	for _, entry := range entries {
		if entry.User != "" {
			userActivity[entry.User] = append(userActivity[entry.User], entry)
		}
		if entry.Process != "" {
			processActivity[entry.Process] = append(processActivity[entry.Process], entry)
		}
		if entry.IPAddress != "" {
			ipActivity[entry.IPAddress] = append(ipActivity[entry.IPAddress], entry)
		}
	}
	
	// Detect unusual patterns
	anomalies = append(anomalies, lp.detectUnusualUserActivity(userActivity)...)
	anomalies = append(anomalies, lp.detectUnusualProcessActivity(processActivity)...)
	anomalies = append(anomalies, lp.detectUnusualIPActivity(ipActivity)...)
	
	return anomalies
}

// detectUnusualUserActivity detects unusual user behavior
func (lp *LogParser) detectUnusualUserActivity(userActivity map[string][]LogEntry) []Anomaly {
	var anomalies []Anomaly
	
	for user, entries := range userActivity {
		// Check for unusual login times
		loginCount := 0
		for _, entry := range entries {
			if strings.Contains(strings.ToLower(entry.Message), "login") {
				loginCount++
			}
		}
		
		if loginCount > 10 { // Threshold for unusual activity
			anomaly := Anomaly{
				Type:        "unusual_user_activity",
				Description: fmt.Sprintf("User %s has %d login events", user, loginCount),
				Severity:    3,
				Timestamp:   time.Now(),
				Evidence:    fmt.Sprintf("User: %s, Login count: %d", user, loginCount),
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies
}

// detectUnusualProcessActivity detects unusual process behavior
func (lp *LogParser) detectUnusualProcessActivity(processActivity map[string][]LogEntry) []Anomaly {
	var anomalies []Anomaly
	
	for process, entries := range processActivity {
		// Check for unusual process execution patterns
		if len(entries) > 100 { // Threshold for unusual activity
			anomaly := Anomaly{
				Type:        "unusual_process_activity",
				Description: fmt.Sprintf("Process %s has %d log entries", process, len(entries)),
				Severity:    2,
				Timestamp:   time.Now(),
				Evidence:    fmt.Sprintf("Process: %s, Entry count: %d", process, len(entries)),
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies
}

// detectUnusualIPActivity detects unusual IP address activity
func (lp *LogParser) detectUnusualIPActivity(ipActivity map[string][]LogEntry) []Anomaly {
	var anomalies []Anomaly
	
	for ip, entries := range ipActivity {
		// Check for unusual IP activity
		if len(entries) > 50 { // Threshold for unusual activity
			anomaly := Anomaly{
				Type:        "unusual_ip_activity",
				Description: fmt.Sprintf("IP %s has %d log entries", ip, len(entries)),
				Severity:    3,
				Timestamp:   time.Now(),
				Evidence:    fmt.Sprintf("IP: %s, Entry count: %d", ip, len(entries)),
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies
}

// AddRule adds a custom analysis rule
func (lp *LogParser) AddRule(rule LogAnalysisRule) {
	lp.rules = append(lp.rules, rule)
}

// GetRules returns all analysis rules
func (lp *LogParser) GetRules() []LogAnalysisRule {
	return lp.rules
}

// TimelineEvent represents a timeline event
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    int       `json:"severity"`
	User        string    `json:"user,omitempty"`
	Process     string    `json:"process,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	Tags        []string  `json:"tags"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    int       `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
	Evidence    string    `json:"evidence"`
	Confidence  float64   `json:"confidence"`
}
