package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/detector"
)

// Reporter represents the reporting engine
type Reporter struct {
	version string
}

// ReportInfo represents information about a generated report
type ReportInfo struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

// NewReporter creates a new reporter instance
func NewReporter() *Reporter {
	return &Reporter{
		version: "1.0.0",
	}
}

// GenerateReports generates all report types
func (r *Reporter) GenerateReports(artifacts []collector.ArtifactResult, findings []detector.Finding, bundlePath string) ([]ReportInfo, error) {
	var reports []ReportInfo
	
	// Get bundle directory
	bundleDir := strings.TrimSuffix(bundlePath, ".zip")
	reportsDir := filepath.Join(bundleDir, "reports")
	
	// Generate Markdown summary
	if summaryPath, err := r.generateMarkdownSummary(artifacts, findings, reportsDir); err == nil {
		if info, err := r.getReportInfo(summaryPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	// Generate HTML full report
	if htmlPath, err := r.generateHTMLReport(artifacts, findings, reportsDir); err == nil {
		if info, err := r.getReportInfo(htmlPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	// Generate Markdown findings report
	if findingsPath, err := r.generateFindingsReport(findings, reportsDir); err == nil {
		if info, err := r.getReportInfo(findingsPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	return reports, nil
}

// generateMarkdownSummary generates a concise Markdown summary
func (r *Reporter) generateMarkdownSummary(artifacts []collector.ArtifactResult, findings []detector.Finding, reportsDir string) (string, error) {
	summaryPath := filepath.Join(reportsDir, "summary.md")
	
	file, err := os.Create(summaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to create summary file: %w", err)
	}
	defer file.Close()
	
	// Write header
	fmt.Fprintf(file, "# RedTriage Summary Report\n\n")
	fmt.Fprintf(file, "**Generated:** %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Tool Version:** %s\n\n", r.version)
	
	// Write host profile
	if hostProfile := r.findHostProfile(artifacts); hostProfile != nil {
		fmt.Fprintf(file, "## Host Profile\n\n")
		if hostData, ok := hostProfile.Data.(map[string]interface{}); ok {
			if hostname, ok := hostData["hostname"].(string); ok {
				fmt.Fprintf(file, "**Hostname:** %s\n", hostname)
			}
			if osInfo, ok := hostData["os_info"].(map[string]interface{}); ok {
				if version, ok := osInfo["version"].(string); ok {
					fmt.Fprintf(file, "**OS Version:** %s\n", version)
				}
			}
		}
		fmt.Fprintf(file, "\n")
	}
	
	// Write artifacts summary
	fmt.Fprintf(file, "## Artifacts Collected\n\n")
	artifactSummary := r.summarizeArtifacts(artifacts)
	for category, count := range artifactSummary {
		fmt.Fprintf(file, "- **%s:** %d artifacts\n", category, count)
	}
	fmt.Fprintf(file, "\n")
	
	// Write findings summary
	fmt.Fprintf(file, "## Findings Summary\n\n")
	if len(findings) == 0 {
		fmt.Fprintf(file, "No findings detected.\n\n")
	} else {
		findingsBySeverity := r.groupFindingsBySeverity(findings)
		for severity, count := range findingsBySeverity {
			fmt.Fprintf(file, "- **%s:** %d findings\n", severity, count)
		}
		fmt.Fprintf(file, "\n")
		
		// List high and critical findings
		highFindings := r.filterFindingsBySeverity(findings, "high")
		if len(highFindings) > 0 {
			fmt.Fprintf(file, "### High Priority Findings\n\n")
			for _, finding := range highFindings {
				fmt.Fprintf(file, "- **%s:** %s\n", finding.RuleName, finding.Description)
			}
			fmt.Fprintf(file, "\n")
		}
	}
	
	// Write recommendations
	fmt.Fprintf(file, "## Recommendations\n\n")
	if len(findings) > 0 {
		fmt.Fprintf(file, "1. Review all findings for accuracy and context\n")
		fmt.Fprintf(file, "2. Investigate high and critical findings immediately\n")
		fmt.Fprintf(file, "3. Correlate findings with other evidence sources\n")
		fmt.Fprintf(file, "4. Document investigation steps and conclusions\n")
	} else {
		fmt.Fprintf(file, "1. No immediate threats detected\n")
		fmt.Fprintf(file, "2. Review collected artifacts for manual analysis\n")
		fmt.Fprintf(file, "3. Consider additional collection if needed\n")
	}
	
	return summaryPath, nil
}

// generateHTMLReport generates a comprehensive HTML report
func (r *Reporter) generateHTMLReport(artifacts []collector.ArtifactResult, findings []detector.Finding, reportsDir string) (string, error) {
	htmlPath := filepath.Join(reportsDir, "full_report.html")
	
	file, err := os.Create(htmlPath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML report: %w", err)
	}
	defer file.Close()
	
	// Write HTML header
	fmt.Fprintf(file, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RedTriage Full Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 30px; }
        .section { margin-bottom: 30px; }
        .finding { border-left: 4px solid #ddd; padding-left: 15px; margin: 10px 0; }
        .finding.high { border-left-color: #ff6b6b; }
        .finding.medium { border-left-color: #feca57; }
        .finding.low { border-left-color: #48dbfb; }
        .artifact { background: #f9f9f9; padding: 10px; margin: 5px 0; border-radius: 3px; }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>RedTriage Full Report</h1>
        <p><strong>Generated:</strong> %s</p>
        <p><strong>Tool Version:</strong> %s</p>
    </div>
`, time.Now().Format(time.RFC3339), r.version)
	
	// Write host profile section
	fmt.Fprintf(file, `<div class="section">
    <h2>Host Profile</h2>`)
	if hostProfile := r.findHostProfile(artifacts); hostProfile != nil {
		if hostData, ok := hostProfile.Data.(map[string]interface{}); ok {
			fmt.Fprintf(file, `<table>
        <tr><th>Property</th><th>Value</th></tr>`)
			if hostname, ok := hostData["hostname"].(string); ok {
				fmt.Fprintf(file, `<tr><td>Hostname</td><td>%s</td></tr>`, hostname)
			}
			if osInfo, ok := hostData["os_info"].(map[string]interface{}); ok {
				if version, ok := osInfo["version"].(string); ok {
					fmt.Fprintf(file, `<tr><td>OS Version</td><td>%s</td></tr>`, version)
				}
			}
			fmt.Fprintf(file, `</table>`)
		}
	}
	fmt.Fprintf(file, `</div>`)
	
	// Write artifacts section
	fmt.Fprintf(file, `<div class="section">
    <h2>Collected Artifacts</h2>
    <table>
        <tr><th>Name</th><th>Category</th><th>Type</th><th>Size</th><th>Description</th></tr>`)
	for _, artifact := range artifacts {
		fmt.Fprintf(file, `<tr>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%d bytes</td>
            <td>%s</td>
        </tr>`, artifact.Artifact.Name, artifact.Artifact.Category, artifact.Artifact.Type, artifact.Size, artifact.Artifact.Description)
	}
	fmt.Fprintf(file, `</table></div>`)
	
	// Write findings section
	fmt.Fprintf(file, `<div class="section">
    <h2>Detection Findings</h2>`)
	if len(findings) == 0 {
		fmt.Fprintf(file, `<p>No findings detected.</p>`)
	} else {
		for _, finding := range findings {
			severityClass := strings.ToLower(finding.Severity)
			fmt.Fprintf(file, `<div class="finding %s">
                <h3>%s</h3>
                <p><strong>Rule:</strong> %s</p>
                <p><strong>Severity:</strong> %s</p>
                <p><strong>Category:</strong> %s</p>
                <p><strong>Description:</strong> %s</p>`, severityClass, finding.RuleName, finding.RuleID, finding.Severity, finding.Category, finding.Description)
			
			if len(finding.Evidence) > 0 {
				fmt.Fprintf(file, `<p><strong>Evidence:</strong></p><ul>`)
				for _, evidence := range finding.Evidence {
					fmt.Fprintf(file, `<li>%s: %s (Confidence: %.1f%%)</li>`, evidence.Type, evidence.Description, evidence.Confidence*100)
				}
				fmt.Fprintf(file, `</ul>`)
			}
			
			fmt.Fprintf(file, `</div>`)
		}
	}
	fmt.Fprintf(file, `</div>`)
	
	// Write footer
	fmt.Fprintf(file, `
    <div class="section">
        <h2>Report Information</h2>
        <p>This report was generated by RedTriage, a professional incident response triage tool.</p>
        <p>For questions or support, please refer to the RedTriage documentation.</p>
    </div>
</body>
</html>`)
	
	return htmlPath, nil
}

// generateFindingsReport generates a detailed Markdown findings report
func (r *Reporter) generateFindingsReport(findings []detector.Finding, reportsDir string) (string, error) {
	findingsPath := filepath.Join(reportsDir, "findings.md")
	
	file, err := os.Create(findingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to create findings report: %w", err)
	}
	defer file.Close()
	
	// Write header
	fmt.Fprintf(file, "# RedTriage Findings Report\n\n")
	fmt.Fprintf(file, "**Generated:** %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Total Findings:** %d\n\n", len(findings))
	
	if len(findings) == 0 {
		fmt.Fprintf(file, "No findings detected during this triage collection.\n")
		return findingsPath, nil
	}
	
	// Group findings by severity
	findingsBySeverity := r.groupFindingsBySeverity(findings)
	
	// Write findings by severity
	for _, severity := range []string{"critical", "high", "medium", "low"} {
		if count, exists := findingsBySeverity[severity]; exists && count > 0 {
			severityFindings := r.filterFindingsBySeverity(findings, severity)
			
			fmt.Fprintf(file, "## %s Severity Findings (%d)\n\n", strings.Title(severity), count)
			
			for i, finding := range severityFindings {
				fmt.Fprintf(file, "### %d. %s\n\n", i+1, finding.RuleName)
				fmt.Fprintf(file, "- **Rule ID:** %s\n", finding.RuleID)
				fmt.Fprintf(file, "- **Category:** %s\n", finding.Category)
				fmt.Fprintf(file, "- **Description:** %s\n", finding.Description)
				fmt.Fprintf(file, "- **Timestamp:** %s\n", finding.Timestamp.Format(time.RFC3339))
				
				if len(finding.Tags) > 0 {
					fmt.Fprintf(file, "- **Tags:** %s\n", strings.Join(finding.Tags, ", "))
				}
				
				if len(finding.Evidence) > 0 {
					fmt.Fprintf(file, "- **Evidence:**\n")
					for _, evidence := range finding.Evidence {
						fmt.Fprintf(file, "  - %s: %s (Confidence: %.1f%%)\n", evidence.Type, evidence.Description, evidence.Confidence*100)
					}
				}
				
				fmt.Fprintf(file, "\n")
			}
		}
	}
	
	// Write summary statistics
	fmt.Fprintf(file, "## Summary Statistics\n\n")
	fmt.Fprintf(file, "| Severity | Count |\n")
	fmt.Fprintf(file, "|----------|-------|\n")
	for _, severity := range []string{"critical", "high", "medium", "low"} {
		if count, exists := findingsBySeverity[severity]; exists {
			fmt.Fprintf(file, "| %s | %d |\n", strings.Title(severity), count)
		}
	}
	
	return findingsPath, nil
}

// Helper methods

// findHostProfile finds the host profile artifact
func (r *Reporter) findHostProfile(artifacts []collector.ArtifactResult) *collector.ArtifactResult {
	for _, artifact := range artifacts {
		if artifact.Artifact.Name == "host_profile" {
			return &artifact
		}
	}
	return nil
}

// summarizeArtifacts summarizes artifacts by category
func (r *Reporter) summarizeArtifacts(artifacts []collector.ArtifactResult) map[string]int {
	summary := make(map[string]int)
	for _, artifact := range artifacts {
		category := artifact.Artifact.Category
		summary[category]++
	}
	return summary
}

// groupFindingsBySeverity groups findings by severity
func (r *Reporter) groupFindingsBySeverity(findings []detector.Finding) map[string]int {
	grouped := make(map[string]int)
	for _, finding := range findings {
		grouped[finding.Severity]++
	}
	return grouped
}

// filterFindingsBySeverity filters findings by minimum severity
func (r *Reporter) filterFindingsBySeverity(findings []detector.Finding, minSeverity string) []detector.Finding {
	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	
	minLevel := severityLevels[minSeverity]
	if minLevel == 0 {
		minLevel = 1
	}
	
	var filtered []detector.Finding
	for _, finding := range findings {
		if level := severityLevels[finding.Severity]; level >= minLevel {
			filtered = append(filtered, finding)
		}
	}
	
	// Sort by severity (highest first)
	sort.Slice(filtered, func(i, j int) bool {
		return severityLevels[filtered[i].Severity] > severityLevels[filtered[j].Severity]
	})
	
	return filtered
}

// getReportInfo gets information about a generated report
func (r *Reporter) getReportInfo(reportPath string) (ReportInfo, error) {
	info, err := os.Stat(reportPath)
	if err != nil {
		return ReportInfo{}, err
	}
	
	// Determine report type from extension
	ext := filepath.Ext(reportPath)
	var reportType string
	switch ext {
	case ".md":
		reportType = "markdown"
	case ".html":
		reportType = "html"
	default:
		reportType = "unknown"
	}
	
	return ReportInfo{
		Type: reportType,
		Path: reportPath,
		Size: info.Size(),
		// TODO: Calculate checksum
		Checksum: "",
	}, nil
}
