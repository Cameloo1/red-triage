package detector

import (
	"fmt"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// Detector represents the detection engine
type Detector struct {
	rules []Rule
}

// Rule represents a detection rule
type Rule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Logic       string   `json:"logic"`
	Enabled     bool     `json:"enabled"`
}

// Finding represents a detection finding
type Finding struct {
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	Severity    string                 `json:"severity"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Evidence    []Evidence             `json:"evidence"`
	Tags        []string               `json:"tags"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Evidence represents evidence supporting a finding
type Evidence struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Value       string                 `json:"value"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewDetector creates a new detector instance
func NewDetector() *Detector {
	detector := &Detector{
		rules: make([]Rule, 0),
	}
	
	// Load built-in rules
	detector.loadBuiltInRules()
	
	return detector
}

// loadBuiltInRules loads the built-in heuristic detection rules
func (d *Detector) loadBuiltInRules() {
	builtInRules := []Rule{
		{
			ID:          "RT001",
			Name:        "Suspicious Process Names",
			Description: "Detects processes with suspicious names that may indicate malware",
			Severity:    "medium",
			Category:    "process",
			Tags:        []string{"malware", "process", "naming"},
			Logic:       "Process names containing suspicious patterns",
			Enabled:     true,
		},
		{
			ID:          "RT002",
			Name:        "Unusual Network Connections",
			Description: "Detects unusual or suspicious network connections",
			Severity:    "medium",
			Category:    "network",
			Tags:        []string{"network", "connection", "suspicious"},
			Logic:       "Network connections to suspicious IPs or unusual ports",
			Enabled:     true,
		},
		{
			ID:          "RT003",
			Name:        "Suspicious Scheduled Tasks",
			Description: "Detects suspicious scheduled tasks that may indicate persistence",
			Severity:    "high",
			Category:    "persistence",
			Tags:        []string{"persistence", "scheduled_task", "malware"},
			Logic:       "Scheduled tasks with suspicious commands or timing",
			Enabled:     true,
		},
		{
			ID:          "RT004",
			Name:        "Unusual Service Names",
			Description: "Detects services with unusual names that may indicate malware",
			Severity:    "medium",
			Category:    "service",
			Tags:        []string{"malware", "service", "naming"},
			Logic:       "Service names containing suspicious patterns",
			Enabled:     true,
		},
		{
			ID:          "RT005",
			Name:        "Suspicious Event Log Patterns",
			Description: "Detects suspicious patterns in event logs",
			Severity:    "low",
			Category:    "log",
			Tags:        []string{"log", "event", "pattern"},
			Logic:       "Event log entries matching suspicious patterns",
			Enabled:     true,
		},
	}
	
	d.rules = append(d.rules, builtInRules...)
}

// Evaluate runs detections against collected artifacts
func (d *Detector) Evaluate(artifacts []collector.ArtifactResult) ([]Finding, error) {
	var findings []Finding
	
	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}
		
		// Apply rule logic based on category
		switch rule.Category {
		case "process":
			if finding := d.evaluateProcessRule(rule, artifacts); finding != nil {
				findings = append(findings, *finding)
			}
		case "network":
			if finding := d.evaluateNetworkRule(rule, artifacts); finding != nil {
				findings = append(findings, *finding)
			}
		case "persistence":
			if finding := d.evaluatePersistenceRule(rule, artifacts); finding != nil {
				findings = append(findings, *finding)
			}
		case "service":
			if finding := d.evaluateServiceRule(rule, artifacts); finding != nil {
				findings = append(findings, *finding)
			}
		case "log":
			if finding := d.evaluateLogRule(rule, artifacts); finding != nil {
				findings = append(findings, *finding)
			}
		}
	}
	
	return findings, nil
}

// evaluateProcessRule evaluates process-related rules
func (d *Detector) evaluateProcessRule(rule Rule, artifacts []collector.ArtifactResult) *Finding {
	// Look for process artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "process" {
			// Check for suspicious process names
			if strings.Contains(strings.ToLower(artifact.Data.(string)), "suspicious") {
				return &Finding{
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Category:    rule.Category,
					Description: fmt.Sprintf("Suspicious process detected: %s", rule.Description),
					Evidence: []Evidence{
						{
							Type:        "process_name",
							Source:      artifact.Artifact.Name,
							Value:       "suspicious_process",
							Description: "Process name contains suspicious pattern",
							Confidence:  0.7,
						},
					},
					Tags:      rule.Tags,
					Timestamp: time.Now(),
				}
			}
		}
	}
	
	return nil
}

// evaluateNetworkRule evaluates network-related rules
func (d *Detector) evaluateNetworkRule(rule Rule, artifacts []collector.ArtifactResult) *Finding {
	// Look for network artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "network" {
			// Check for suspicious network connections
			if strings.Contains(strings.ToLower(artifact.Data.(string)), "suspicious") {
				return &Finding{
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Category:    rule.Category,
					Description: fmt.Sprintf("Suspicious network activity detected: %s", rule.Description),
					Evidence: []Evidence{
						{
							Type:        "network_connection",
							Source:      artifact.Artifact.Name,
							Value:       "suspicious_connection",
							Description: "Network connection matches suspicious pattern",
							Confidence:  0.6,
						},
					},
					Tags:      rule.Tags,
					Timestamp: time.Now(),
				}
			}
		}
	}
	
	return nil
}

// evaluatePersistenceRule evaluates persistence-related rules
func (d *Detector) evaluatePersistenceRule(rule Rule, artifacts []collector.ArtifactResult) *Finding {
	// Look for persistence artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "task" {
			// Check for suspicious scheduled tasks
			if strings.Contains(strings.ToLower(artifact.Data.(string)), "suspicious") {
				return &Finding{
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Category:    rule.Category,
					Description: fmt.Sprintf("Suspicious persistence mechanism detected: %s", rule.Description),
					Evidence: []Evidence{
						{
							Type:        "scheduled_task",
							Source:      artifact.Artifact.Name,
							Value:       "suspicious_task",
							Description: "Scheduled task matches suspicious pattern",
							Confidence:  0.8,
						},
					},
					Tags:      rule.Tags,
					Timestamp: time.Now(),
				}
			}
		}
	}
	
	return nil
}

// evaluateServiceRule evaluates service-related rules
func (d *Detector) evaluateServiceRule(rule Rule, artifacts []collector.ArtifactResult) *Finding {
	// Look for service artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "service" {
			// Check for suspicious service names
			if strings.Contains(strings.ToLower(artifact.Data.(string)), "suspicious") {
				return &Finding{
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Category:    rule.Category,
					Description: fmt.Sprintf("Suspicious service detected: %s", rule.Description),
					Evidence: []Evidence{
						{
							Type:        "service_name",
							Source:      artifact.Artifact.Name,
							Value:       "suspicious_service",
							Description: "Service name contains suspicious pattern",
							Confidence:  0.7,
						},
					},
					Tags:      rule.Tags,
					Timestamp: time.Now(),
				}
			}
		}
	}
	
	return nil
}

// evaluateLogRule evaluates log-related rules
func (d *Detector) evaluateLogRule(rule Rule, artifacts []collector.ArtifactResult) *Finding {
	// Look for log artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "log" {
			// Check for suspicious log patterns
			if strings.Contains(strings.ToLower(artifact.Data.(string)), "suspicious") {
				return &Finding{
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					Category:    rule.Category,
					Description: fmt.Sprintf("Suspicious log pattern detected: %s", rule.Description),
					Evidence: []Evidence{
						{
							Type:        "log_pattern",
							Source:      artifact.Artifact.Name,
							Value:       "suspicious_pattern",
							Description: "Log entry matches suspicious pattern",
							Confidence:  0.5,
						},
					},
					Tags:      rule.Tags,
					Timestamp: time.Now(),
				}
			}
		}
	}
	
	return nil
}

// GetBuiltInRules returns the built-in detection rules
func (d *Detector) GetBuiltInRules() []Rule {
	return d.rules
}

// AddRule adds a custom detection rule
func (d *Detector) AddRule(rule Rule) {
	d.rules = append(d.rules, rule)
}

// EnableRule enables a specific rule by ID
func (d *Detector) EnableRule(ruleID string) error {
	for i, rule := range d.rules {
		if rule.ID == ruleID {
			d.rules[i].Enabled = true
			return nil
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}

// DisableRule disables a specific rule by ID
func (d *Detector) DisableRule(ruleID string) error {
	for i, rule := range d.rules {
		if rule.ID == ruleID {
			d.rules[i].Enabled = false
			return nil
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}

// FilterFindingsBySeverity filters findings by minimum severity level
func FilterFindingsBySeverity(findings []Finding, minSeverity string) []Finding {
	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	
	minLevel := severityLevels[minSeverity]
	if minLevel == 0 {
		minLevel = 1 // Default to low
	}
	
	var filtered []Finding
	for _, finding := range findings {
		if level := severityLevels[finding.Severity]; level >= minLevel {
			filtered = append(filtered, finding)
		}
	}
	
	return filtered
}
