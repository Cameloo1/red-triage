package reporter

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/detector"
	"github.com/redtriage/redtriage/internal/logging"
)

// EnhancedReporter provides comprehensive reporting capabilities
type EnhancedReporter struct {
	*Reporter
	logParser *logging.LogParser
}

// ReportTemplate defines a report template
type ReportTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // executive, technical, timeline, network, user, security, compliance
	Format      string            `json:"format"` // html, pdf, json, csv, xml
	Enabled     bool              `json:"enabled"`
	Parameters  map[string]string `json:"parameters"`
}

// ReportData contains all data needed for report generation
type ReportData struct {
	Artifacts     []collector.ArtifactResult `json:"artifacts"`
	Findings      []detector.Finding         `json:"findings"`
	LogAnalysis   []logging.LogAnalysisResult `json:"log_analysis"`
	Timeline      []logging.TimelineEvent    `json:"timeline"`
	Anomalies     []logging.Anomaly          `json:"anomalies"`
	Metadata      map[string]interface{}     `json:"metadata"`
	CollectionInfo CollectionInfo            `json:"collection_info"`
}

// CollectionInfo contains information about the collection process
type CollectionInfo struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	Platform     string    `json:"platform"`
	Collector    string    `json:"collector"`
	Version      string    `json:"version"`
	TotalArtifacts int     `json:"total_artifacts"`
	TotalFindings int      `json:"total_findings"`
	TotalLogs    int       `json:"total_logs"`
}

// NewEnhancedReporter creates a new enhanced reporter
func NewEnhancedReporter() *EnhancedReporter {
	return &EnhancedReporter{
		Reporter:  NewReporter(),
		logParser: logging.NewLogParser(),
	}
}

// GenerateEnhancedReports generates comprehensive reports in multiple formats
func (er *EnhancedReporter) GenerateEnhancedReports(artifacts []collector.ArtifactResult, findings []detector.Finding, bundlePath string) ([]ReportInfo, error) {
	var reports []ReportInfo
	
	// Prepare report data
	reportData := er.prepareReportData(artifacts, findings)
	
	// Get bundle directory
	bundleDir := strings.TrimSuffix(bundlePath, ".zip")
	reportsDir := filepath.Join(bundleDir, "reports")
	
	// Generate reports in different formats
	formats := []string{"html", "json", "csv", "xml"}
	
	for _, format := range formats {
		if reportPath, err := er.generateReportInFormat(reportData, format, reportsDir); err == nil {
			if info, err := er.getReportInfo(reportPath); err == nil {
				reports = append(reports, info)
			}
		}
	}
	
	// Generate specialized reports
	if executivePath, err := er.generateExecutiveSummary(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(executivePath); err == nil {
			reports = append(reports, info)
		}
	}
	
	if technicalPath, err := er.generateTechnicalReport(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(technicalPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	if timelinePath, err := er.generateTimelineReport(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(timelinePath); err == nil {
			reports = append(reports, info)
		}
	}
	
	if networkPath, err := er.generateNetworkReport(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(networkPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	if userPath, err := er.generateUserActivityReport(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(userPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	if securityPath, err := er.generateSecurityReport(reportData, reportsDir); err == nil {
		if info, err := er.getReportInfo(securityPath); err == nil {
			reports = append(reports, info)
		}
	}
	
	return reports, nil
}

// prepareReportData prepares all data needed for report generation
func (er *EnhancedReporter) prepareReportData(artifacts []collector.ArtifactResult, findings []detector.Finding) ReportData {
	// Analyze logs if available
	var logAnalysis []logging.LogAnalysisResult
	var timeline []logging.TimelineEvent
	var anomalies []logging.Anomaly
	
	// Process log artifacts
	for _, artifact := range artifacts {
		if artifact.Artifact.Category == "log" {
			if logData, ok := artifact.Data.(string); ok {
				// Create temporary log file for parsing
				if tempFile, err := er.createTempLogFile(logData); err == nil {
					defer os.Remove(tempFile.Name())
					defer tempFile.Close()
					
					if entries, err := er.logParser.ParseLogFile(tempFile.Name()); err == nil {
						// Analyze logs
						analysis := er.logParser.AnalyzeLogs(entries)
						logAnalysis = append(logAnalysis, analysis...)
						
						// Generate timeline
						timeline = append(timeline, er.logParser.GenerateTimeline(entries)...)
						
						// Detect anomalies
						anomalies = append(anomalies, er.logParser.DetectAnomalies(entries)...)
					}
				}
			}
		}
	}
	
	// Prepare collection info
	collectionInfo := CollectionInfo{
		StartTime:      time.Now().Add(-time.Hour), // Estimate
		EndTime:        time.Now(),
		Duration:       "1 hour", // Estimate
		Platform:       "windows",
		Collector:      "enhanced_windows",
		Version:        "2.0.0",
		TotalArtifacts: len(artifacts),
		TotalFindings:  len(findings),
		TotalLogs:      len(logAnalysis),
	}
	
	return ReportData{
		Artifacts:      artifacts,
		Findings:       findings,
		LogAnalysis:    logAnalysis,
		Timeline:       timeline,
		Anomalies:      anomalies,
		Metadata:       make(map[string]interface{}),
		CollectionInfo: collectionInfo,
	}
}

// createTempLogFile creates a temporary log file for parsing
func (er *EnhancedReporter) createTempLogFile(content string) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "redtriage_log_*.tmp")
	if err != nil {
		return nil, err
	}
	
	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}
	
	tempFile.Seek(0, 0)
	return tempFile, nil
}

// generateReportInFormat generates a report in the specified format
func (er *EnhancedReporter) generateReportInFormat(data ReportData, format, reportsDir string) (string, error) {
	var reportPath string
	var err error
	
	switch format {
	case "html":
		reportPath, err = er.generateHTMLReport(data, reportsDir)
	case "json":
		reportPath, err = er.generateJSONReport(data, reportsDir)
	case "csv":
		reportPath, err = er.generateCSVReport(data, reportsDir)
	case "xml":
		reportPath, err = er.generateXMLReport(data, reportsDir)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
	
	return reportPath, err
}

// generateHTMLReport generates a comprehensive HTML report
func (er *EnhancedReporter) generateHTMLReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "comprehensive_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML report: %w", err)
	}
	defer file.Close()
	
	// Write HTML header with modern styling
	fmt.Fprintf(file, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RedTriage Comprehensive Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; background: #f8f9fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; margin-bottom: 30px; border-radius: 10px; }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .section { background: white; margin-bottom: 30px; padding: 25px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .section h2 { color: #2c3e50; margin-bottom: 20px; padding-bottom: 10px; border-bottom: 3px solid #3498db; }
        .finding { border-left: 4px solid #ddd; padding: 15px; margin: 15px 0; background: #f8f9fa; border-radius: 0 5px 5px 0; }
        .finding.high { border-left-color: #e74c3c; background: #fdf2f2; }
        .finding.medium { border-left-color: #f39c12; background: #fef9e7; }
        .finding.low { border-left-color: #3498db; background: #f0f8ff; }
        .finding.critical { border-left-color: #8e44ad; background: #f4f1f7; }
        .artifact { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border: 1px solid #e9ecef; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .stat-card { background: white; padding: 20px; text-align: center; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .stat-number { font-size: 2em; font-weight: bold; color: #3498db; }
        .stat-label { color: #7f8c8d; margin-top: 5px; }
        .timeline { margin: 20px 0; }
        .timeline-event { background: white; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #3498db; }
        .timeline-time { color: #7f8c8d; font-size: 0.9em; }
        .table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        .table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .table th { background-color: #f8f9fa; font-weight: 600; }
        .table tr:hover { background-color: #f5f5f5; }
        .severity-badge { padding: 4px 8px; border-radius: 12px; font-size: 0.8em; font-weight: bold; }
        .severity-critical { background: #8e44ad; color: white; }
        .severity-high { background: #e74c3c; color: white; }
        .severity-medium { background: #f39c12; color: white; }
        .severity-low { background: #3498db; color: white; }
        .chart-container { margin: 20px 0; height: 300px; background: #f8f9fa; border-radius: 5px; display: flex; align-items: center; justify-content: center; color: #7f8c8d; }
        .footer { text-align: center; margin-top: 40px; padding: 20px; color: #7f8c8d; border-top: 1px solid #e9ecef; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 RedTriage Comprehensive Report</h1>
            <p>Professional Incident Response & Digital Forensics Analysis</p>
        </div>
        
        <div class="section">
            <h2>📊 Executive Summary</h2>
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">%d</div>
                    <div class="stat-label">Artifacts Collected</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">%d</div>
                    <div class="stat-label">Findings Detected</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">%d</div>
                    <div class="stat-label">Log Entries Analyzed</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">%d</div>
                    <div class="stat-label">Anomalies Found</div>
                </div>
            </div>
            
            <p><strong>Collection Period:</strong> %s to %s</p>
            <p><strong>Platform:</strong> %s</p>
            <p><strong>Tool Version:</strong> %s</p>
        </div>
        
        <div class="section">
            <h2>🚨 Critical Findings</h2>`, 
		data.CollectionInfo.TotalArtifacts,
		data.CollectionInfo.TotalFindings,
		data.CollectionInfo.TotalLogs,
		len(data.Anomalies),
		data.CollectionInfo.StartTime.Format("2006-01-02 15:04:05"),
		data.CollectionInfo.EndTime.Format("2006-01-02 15:04:05"),
		data.CollectionInfo.Platform,
		data.CollectionInfo.Version)
	
	// Write critical findings
	criticalFindings := er.filterFindingsBySeverity(data.Findings, "critical")
	if len(criticalFindings) > 0 {
		for _, finding := range criticalFindings {
			fmt.Fprintf(file, `
            <div class="finding critical">
                <h3>🚨 %s</h3>
                <p><strong>Rule ID:</strong> %s</p>
                <p><strong>Category:</strong> %s</p>
                <p><strong>Description:</strong> %s</p>
                <p><strong>Evidence:</strong></p>
                <ul>`, finding.RuleName, finding.RuleID, finding.Category, finding.Description)
			
			for _, evidence := range finding.Evidence {
				fmt.Fprintf(file, `<li>%s: %s (Confidence: %.1f%%)</li>`, evidence.Type, evidence.Description, evidence.Confidence*100)
			}
			
			fmt.Fprintf(file, `</ul></div>`)
		}
	} else {
		fmt.Fprintf(file, `<p>✅ No critical findings detected.</p>`)
	}
	
	fmt.Fprintf(file, `</div>
        
        <div class="section">
            <h2>📋 All Findings</h2>
            <table class="table">
                <thead>
                    <tr>
                        <th>Rule</th>
                        <th>Category</th>
                        <th>Severity</th>
                        <th>Description</th>
                        <th>Evidence Count</th>
                    </tr>
                </thead>
                <tbody>`)
	
	for _, finding := range data.Findings {
		severityClass := fmt.Sprintf("severity-%s", strings.ToLower(finding.Severity))
		fmt.Fprintf(file, `
                    <tr>
                        <td><strong>%s</strong></td>
                        <td>%s</td>
                        <td><span class="severity-badge %s">%s</span></td>
                        <td>%s</td>
                        <td>%d</td>
                    </tr>`, 
			finding.RuleName, finding.Category, severityClass, finding.Severity, finding.Description, len(finding.Evidence))
	}
	
	fmt.Fprintf(file, `
                </tbody>
            </table>
        </div>
        
        <div class="section">
            <h2>⏰ Timeline Analysis</h2>
            <div class="timeline">`)
	
	// Sort timeline events by timestamp
	sort.Slice(data.Timeline, func(i, j int) bool {
		return data.Timeline[i].Timestamp.Before(data.Timeline[j].Timestamp)
	})
	
	for _, event := range data.Timeline {
		fmt.Fprintf(file, `
                <div class="timeline-event">
                    <div class="timeline-time">%s</div>
                    <div><strong>%s</strong> - %s</div>
                    <div>Source: %s | Type: %s</div>
                </div>`, 
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type, event.Description, event.Source, event.Type)
	}
	
	fmt.Fprintf(file, `
            </div>
        </div>
        
        <div class="section">
            <h2>🔍 Anomaly Detection</h2>`)
	
	if len(data.Anomalies) > 0 {
		for _, anomaly := range data.Anomalies {
			fmt.Fprintf(file, `
            <div class="finding">
                <h3>⚠️ %s</h3>
                <p><strong>Type:</strong> %s</p>
                <p><strong>Description:</strong> %s</p>
                <p><strong>Evidence:</strong> %s</p>
                <p><strong>Severity:</strong> %d</p>
            </div>`, 
				anomaly.Type, anomaly.Type, anomaly.Description, anomaly.Evidence, anomaly.Severity)
		}
	} else {
		fmt.Fprintf(file, `<p>✅ No anomalies detected.</p>`)
	}
	
	fmt.Fprintf(file, `
        </div>
        
        <div class="section">
            <h2>📁 Collected Artifacts</h2>
            <table class="table">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Category</th>
                        <th>Type</th>
                        <th>Size</th>
                        <th>Description</th>
                    </tr>
                </thead>
                <tbody>`)
	
	for _, artifact := range data.Artifacts {
		fmt.Fprintf(file, `
                    <tr>
                        <td>%s</td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>%d bytes</td>
                        <td>%s</td>
                    </tr>`, 
			artifact.Artifact.Name, artifact.Artifact.Category, artifact.Artifact.Type, artifact.Size, artifact.Artifact.Description)
	}
	
	fmt.Fprintf(file, `
                </tbody>
            </table>
        </div>
        
        <div class="footer">
            <p>Report generated by RedTriage v%s on %s</p>
            <p>Professional Incident Response & Digital Forensics Tool</p>
        </div>
    </div>
</body>
</html>`, 
		data.CollectionInfo.Version, time.Now().Format("2006-01-02 15:04:05"))
	
	return reportPath, nil
}

// generateJSONReport generates a JSON report
func (er *EnhancedReporter) generateJSONReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "comprehensive_report.json")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON report: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return "", fmt.Errorf("failed to encode JSON: %w", err)
	}
	
	return reportPath, nil
}

// generateCSVReport generates a CSV report
func (er *EnhancedReporter) generateCSVReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "comprehensive_report.csv")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV report: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write findings
	if err := writer.Write([]string{"Findings Report"}); err != nil {
		return "", err
	}
	if err := writer.Write([]string{"Rule Name", "Category", "Severity", "Description", "Evidence Count"}); err != nil {
		return "", err
	}
	
	for _, finding := range data.Findings {
		if err := writer.Write([]string{
			finding.RuleName,
			finding.Category,
			finding.Severity,
			finding.Description,
			fmt.Sprintf("%d", len(finding.Evidence)),
		}); err != nil {
			return "", err
		}
	}
	
	// Write artifacts
	if err := writer.Write([]string{""}); err != nil {
		return "", err
	}
	if err := writer.Write([]string{"Artifacts Report"}); err != nil {
		return "", err
	}
	if err := writer.Write([]string{"Name", "Category", "Type", "Size", "Description"}); err != nil {
		return "", err
	}
	
	for _, artifact := range data.Artifacts {
		if err := writer.Write([]string{
			artifact.Artifact.Name,
			artifact.Artifact.Category,
			artifact.Artifact.Type,
			fmt.Sprintf("%d", artifact.Size),
			artifact.Artifact.Description,
		}); err != nil {
			return "", err
		}
	}
	
	return reportPath, nil
}

// generateXMLReport generates an XML report
func (er *EnhancedReporter) generateXMLReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "comprehensive_report.xml")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create XML report: %w", err)
	}
	defer file.Close()
	
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return "", fmt.Errorf("failed to encode XML: %w", err)
	}
	
	return reportPath, nil
}

// generateExecutiveSummary generates an executive summary report
func (er *EnhancedReporter) generateExecutiveSummary(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "executive_summary.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create executive summary: %w", err)
	}
	defer file.Close()
	
	// Generate executive summary HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Executive Summary - RedTriage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .finding { margin: 10px 0; padding: 10px; border-left: 4px solid #ddd; }
        .critical { border-left-color: #e74c3c; }
        .high { border-left-color: #f39c12; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Executive Summary</h1>
        <p>RedTriage Incident Response Report</p>
    </div>
    
    <div class="summary">
        <h2>Key Findings</h2>
        <p>Total Artifacts: %d</p>
        <p>Total Findings: %d</p>
        <p>Critical Issues: %d</p>
        <p>High Priority Issues: %d</p>
    </div>
</body>
</html>`, 
		data.CollectionInfo.TotalArtifacts,
		data.CollectionInfo.TotalFindings,
		len(er.filterFindingsBySeverity(data.Findings, "critical")),
		len(er.filterFindingsBySeverity(data.Findings, "high")))
	
	return reportPath, nil
}

// generateTechnicalReport generates a technical deep-dive report
func (er *EnhancedReporter) generateTechnicalReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "technical_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create technical report: %w", err)
	}
	defer file.Close()
	
	// Generate technical report HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Technical Report - RedTriage Report</title>
    <style>
        body { font-family: monospace; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .technical { margin: 20px 0; }
        pre { background: #f8f8f8; padding: 10px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Technical Deep-Dive Report</h1>
        <p>Detailed technical analysis and evidence</p>
    </div>
    
    <div class="technical">
        <h2>Technical Details</h2>
        <p>This report contains %d artifacts and %d findings.</p>
        <p>Platform: %s</p>
        <p>Collector: %s</p>
    </div>
</body>
</html>`, 
		data.CollectionInfo.TotalArtifacts,
		data.CollectionInfo.TotalFindings,
		data.CollectionInfo.Platform,
		data.CollectionInfo.Collector)
	
	return reportPath, nil
}

// generateTimelineReport generates a timeline report
func (er *EnhancedReporter) generateTimelineReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "timeline_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create timeline report: %w", err)
	}
	defer file.Close()
	
	// Generate timeline report HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Timeline Report - RedTriage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .timeline { margin: 20px 0; }
        .event { margin: 10px 0; padding: 10px; border-left: 4px solid #3498db; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Timeline Analysis Report</h1>
        <p>Chronological sequence of events</p>
    </div>
    
    <div class="timeline">
        <h2>Timeline Events (%d total)</h2>`, len(data.Timeline))
	
	// Sort timeline events
	sort.Slice(data.Timeline, func(i, j int) bool {
		return data.Timeline[i].Timestamp.Before(data.Timeline[j].Timestamp)
	})
	
	for _, event := range data.Timeline {
		fmt.Fprintf(file, `
        <div class="event">
            <strong>%s</strong> - %s<br>
            Source: %s | Type: %s
        </div>`, 
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Description, event.Source, event.Type)
	}
	
	fmt.Fprintf(file, `
    </div>
</body>
</html>`)
	
	return reportPath, nil
}

// generateNetworkReport generates a network analysis report
func (er *EnhancedReporter) generateNetworkReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "network_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create network report: %w", err)
	}
	defer file.Close()
	
	// Generate network report HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Network Analysis Report - RedTriage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .network { margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Network Analysis Report</h1>
        <p>Network connections and activity analysis</p>
    </div>
    
    <div class="network">
        <h2>Network Analysis</h2>
        <p>This report analyzes network-related artifacts and findings.</p>
        <p>Total artifacts: %d</p>
        <p>Network findings: %d</p>
    </div>
</body>
</html>`, 
		data.CollectionInfo.TotalArtifacts,
		len(er.filterFindingsByCategory(data.Findings, "network")))
	
	return reportPath, nil
}

// generateUserActivityReport generates a user activity report
func (er *EnhancedReporter) generateUserActivityReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "user_activity_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create user activity report: %w", err)
	}
	defer file.Close()
	
	// Generate user activity report HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>User Activity Report - RedTriage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .user { margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>User Activity Analysis Report</h1>
        <p>User behavior and activity patterns</p>
    </div>
    
    <div class="user">
        <h2>User Activity Analysis</h2>
        <p>This report analyzes user-related artifacts and findings.</p>
        <p>Total artifacts: %d</p>
        <p>User-related findings: %d</p>
    </div>
</body>
</html>`, 
		data.CollectionInfo.TotalArtifacts,
		len(er.filterFindingsByCategory(data.Findings, "user")))
	
	return reportPath, nil
}

// generateSecurityReport generates a security incident report
func (er *EnhancedReporter) generateSecurityReport(data ReportData, reportsDir string) (string, error) {
	reportPath := filepath.Join(reportsDir, "security_report.html")
	
	file, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create security report: %w", err)
	}
	defer file.Close()
	
	// Generate security report HTML
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Security Incident Report - RedTriage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .security { margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Security Incident Report</h1>
        <p>Security findings and incident analysis</p>
    </div>
    
    <div class="security">
        <h2>Security Analysis</h2>
        <p>This report contains security-related findings and analysis.</p>
        <p>Total findings: %d</p>
        <p>Critical security issues: %d</p>
        <p>High security issues: %d</p>
    </div>
</body>
</html>`, 
		data.CollectionInfo.TotalFindings,
		len(er.filterFindingsBySeverity(data.Findings, "critical")),
		len(er.filterFindingsBySeverity(data.Findings, "high")))
	
	return reportPath, nil
}

// Helper methods for filtering findings
func (er *EnhancedReporter) filterFindingsBySeverity(findings []detector.Finding, minSeverity string) []detector.Finding {
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
	
	return filtered
}

func (er *EnhancedReporter) filterFindingsByCategory(findings []detector.Finding, category string) []detector.Finding {
	var filtered []detector.Finding
	for _, finding := range findings {
		if finding.Category == category {
			filtered = append(filtered, finding)
		}
	}
	
	return filtered
}
