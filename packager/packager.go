package packager

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/redtriage/redtriage/collector"
	"github.com/redtriage/redtriage/detector"
	"github.com/redtriage/redtriage/utils"
)

// Packager represents the packaging engine
type Packager struct {
	version string
}

// BundleManifest represents the manifest for a triage bundle
type BundleManifest struct {
	CaseID        string                 `json:"case_id"`
	ToolVersion   string                 `json:"tool_version"`
	CollectionTime time.Time             `json:"collection_time"`
	HostInfo      map[string]interface{} `json:"host_info"`
	Artifacts     []ArtifactInfo         `json:"artifacts"`
	Findings      []FindingInfo          `json:"findings"`
	Configuration map[string]interface{} `json:"configuration"`
	RedactionRules []string              `json:"redaction_rules"`
	Checksums     map[string]string      `json:"checksums"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ArtifactInfo represents information about a collected artifact
type ArtifactInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Type        string                 `json:"type"`
	Size        int64                  `json:"size"`
	Checksum    string                 `json:"checksum"`
	CollectedAt time.Time              `json:"collected_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// FindingInfo represents information about a detection finding
type FindingInfo struct {
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	Severity    string                 `json:"severity"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Evidence    []EvidenceInfo         `json:"evidence"`
	Tags        []string               `json:"tags"`
	Timestamp   time.Time              `json:"timestamp"`
}

// EvidenceInfo represents information about evidence
type EvidenceInfo struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Value       string                 `json:"value"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
}

// NewPackager creates a new packager instance
func NewPackager() *Packager {
	return &Packager{
		version: "1.0.0",
	}
}

// CreateBundle creates a triage bundle with collected artifacts and findings
func (p *Packager) CreateBundle(artifacts []collector.ArtifactResult, findings []detector.Finding, outputDir string) (string, error) {
	// Generate case ID
	caseID := utils.GenerateCaseID()
	
	// Create bundle directory
	bundleDir := filepath.Join(outputDir, fmt.Sprintf("redtriage-%s", caseID))
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bundle directory: %w", err)
	}
	
	// Create subdirectories
	artifactsDir := filepath.Join(bundleDir, "artifacts")
	findingsDir := filepath.Join(bundleDir, "findings")
	reportsDir := filepath.Join(bundleDir, "reports")
	
	for _, dir := range []string{artifactsDir, findingsDir, reportsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create subdirectory %s: %w", dir, err)
		}
	}
	
	// Copy artifacts to bundle
	artifactInfos, err := p.copyArtifacts(artifacts, artifactsDir)
	if err != nil {
		return "", fmt.Errorf("failed to copy artifacts: %w", err)
	}
	
	// Write findings to bundle
	findingInfos, err := p.writeFindings(findings, findingsDir)
	if err != nil {
		return "", fmt.Errorf("failed to write findings: %w", err)
	}
	
	// Create manifest
	manifest, err := p.createManifest(caseID, artifactInfos, findingInfos)
	if err != nil {
		return "", fmt.Errorf("failed to create manifest: %w", err)
	}
	
	// Write manifest
	manifestPath := filepath.Join(bundleDir, "manifest.json")
	if err := p.writeManifest(manifest, manifestPath); err != nil {
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}
	
	// Write checksums file
	checksumsPath := filepath.Join(bundleDir, "checksums.txt")
	if err := p.writeChecksums(manifest.Checksums, checksumsPath); err != nil {
		return "", fmt.Errorf("failed to write checksums: %w", err)
	}
	
	// Create ZIP archive
	zipPath := bundleDir + ".zip"
	if err := p.createZipArchive(bundleDir, zipPath); err != nil {
		return "", fmt.Errorf("failed to create ZIP archive: %w", err)
	}
	
	// Calculate final checksum
	finalChecksum, err := utils.GetFileHash(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate final checksum: %w", err)
	}
	
	// Update manifest with final checksum
	manifest.Checksums["bundle.zip"] = finalChecksum
	if err := p.writeManifest(manifest, manifestPath); err != nil {
		return "", fmt.Errorf("failed to update manifest: %w", err)
	}
	
	return zipPath, nil
}

// copyArtifacts copies artifacts to the bundle directory
func (p *Packager) copyArtifacts(artifacts []collector.ArtifactResult, artifactsDir string) ([]ArtifactInfo, error) {
	var artifactInfos []ArtifactInfo
	
	for _, artifact := range artifacts {
		// Create safe filename
		safeName := utils.SafeFilename(artifact.Artifact.Name)
		artifactPath := filepath.Join(artifactsDir, safeName+".txt")
		
		// Convert artifact data to string and write to file
		var dataStr string
		switch v := artifact.Data.(type) {
		case string:
			dataStr = v
		default:
			// Convert to JSON for complex data
			if jsonData, err := json.MarshalIndent(v, "", "  "); err == nil {
				dataStr = string(jsonData)
			} else {
				dataStr = fmt.Sprintf("%v", v)
			}
		}
		
		// Write artifact data
		if err := os.WriteFile(artifactPath, []byte(dataStr), 0644); err != nil {
			return nil, fmt.Errorf("failed to write artifact %s: %w", artifact.Artifact.Name, err)
		}
		
		// Calculate checksum
		checksum, err := utils.GetFileHash(artifactPath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate checksum for %s: %w", artifact.Artifact.Name, err)
		}
		
		// Create artifact info
		artifactInfo := ArtifactInfo{
			Name:        artifact.Artifact.Name,
			Description: artifact.Artifact.Description,
			Category:    artifact.Artifact.Category,
			Type:        artifact.Artifact.Type,
			Size:        int64(len(dataStr)),
			Checksum:    checksum,
			CollectedAt: artifact.Metadata.CollectedAt,
			Metadata:    map[string]interface{}{},
		}
		
		artifactInfos = append(artifactInfos, artifactInfo)
	}
	
	return artifactInfos, nil
}

// writeFindings writes findings to the bundle directory
func (p *Packager) writeFindings(findings []detector.Finding, findingsDir string) ([]FindingInfo, error) {
	var findingInfos []FindingInfo
	
	// Write findings summary
	findingsPath := filepath.Join(findingsDir, "findings.json")
	findingsData, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal findings: %w", err)
	}
	
	if err := os.WriteFile(findingsPath, findingsData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write findings: %w", err)
	}
	
	// Convert findings to FindingInfo
	for _, finding := range findings {
		// Convert evidence
		var evidenceInfos []EvidenceInfo
		for _, evidence := range finding.Evidence {
			evidenceInfo := EvidenceInfo{
				Type:        evidence.Type,
				Source:      evidence.Source,
				Value:       evidence.Value,
				Description: evidence.Description,
				Confidence:  evidence.Confidence,
			}
			evidenceInfos = append(evidenceInfos, evidenceInfo)
		}
		
		findingInfo := FindingInfo{
			RuleID:      finding.RuleID,
			RuleName:    finding.RuleName,
			Severity:    finding.Severity,
			Category:    finding.Category,
			Description: finding.Description,
			Evidence:    evidenceInfos,
			Tags:        finding.Tags,
			Timestamp:   finding.Timestamp,
		}
		
		findingInfos = append(findingInfos, findingInfo)
	}
	
	return findingInfos, nil
}

// createManifest creates the bundle manifest
func (p *Packager) createManifest(caseID string, artifacts []ArtifactInfo, findings []FindingInfo) (*BundleManifest, error) {
	// Get hostname for host info
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	// Create checksums map
	checksums := make(map[string]string)
	
	// Add artifact checksums
	for _, artifact := range artifacts {
		checksums[artifact.Name] = artifact.Checksum
	}
	
	// Add findings checksum
	if findingsData, err := json.Marshal(findings); err == nil {
		findingsHash := sha256.Sum256(findingsData)
		checksums["findings"] = fmt.Sprintf("%x", findingsHash)
	}
	
	manifest := &BundleManifest{
		CaseID:        caseID,
		ToolVersion:   p.version,
		CollectionTime: time.Now(),
		HostInfo: map[string]interface{}{
			"hostname": hostname,
			"platform": "windows", // TODO: Detect platform
		},
		Artifacts:     artifacts,
		Findings:      findings,
		Configuration: make(map[string]interface{}),
		RedactionRules: []string{},
		Checksums:     checksums,
		Metadata: map[string]interface{}{
			"created_by": "RedTriage",
			"created_at": time.Now().Format(time.RFC3339),
		},
	}
	
	return manifest, nil
}

// writeManifest writes the manifest to a file
func (p *Packager) writeManifest(manifest *BundleManifest, path string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// writeChecksums writes the checksums to a file
func (p *Packager) writeChecksums(checksums map[string]string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create checksums file: %w", err)
	}
	defer file.Close()
	
	for name, checksum := range checksums {
		if _, err := fmt.Fprintf(file, "%s  %s\n", checksum, name); err != nil {
			return fmt.Errorf("failed to write checksum: %w", err)
		}
	}
	
	return nil
}

// createZipArchive creates a ZIP archive of the bundle directory
func (p *Packager) createZipArchive(sourceDir, zipPath string) error {
	zipfile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipfile.Close()
	
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		
		// Create file in ZIP
		file, err := archive.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to create file in ZIP: %w", err)
		}
		
		// Open source file
		sourceFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer sourceFile.Close()
		
		// Copy file contents
		_, err = io.Copy(file, sourceFile)
		if err != nil {
			return fmt.Errorf("failed to copy file contents: %w", err)
		}
		
		return nil
	})
}
