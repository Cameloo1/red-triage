package output

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReportsManager handles centralized report storage and organization
type ReportsManager struct {
	reportsDir string
	config     *ReportsConfig
}

// ReportsConfig defines the structure for organizing reports
type ReportsConfig struct {
	HealthReportsDir    string
	SystemReportsDir    string
	CollectionReportsDir string
	TestReportsDir      string
	LogsDir             string
	MetadataDir         string
}

// NewReportsManager creates a new reports manager
func NewReportsManager(reportsDir string) (*ReportsManager, error) {
	rm := &ReportsManager{
		reportsDir: reportsDir,
		config: &ReportsConfig{
			HealthReportsDir:    filepath.Join(reportsDir, "health"),
			SystemReportsDir:    filepath.Join(reportsDir, "system"),
			CollectionReportsDir: filepath.Join(reportsDir, "collection"),
			TestReportsDir:      filepath.Join(reportsDir, "tests"),
			LogsDir:             filepath.Join(reportsDir, "logs"),
			MetadataDir:         filepath.Join(reportsDir, "metadata"),
		},
	}

	// Create all necessary directories
	if err := rm.createDirectoryStructure(); err != nil {
		return nil, fmt.Errorf("failed to create reports directory structure: %w", err)
	}

	return rm, nil
}

// createDirectoryStructure creates all necessary subdirectories
func (rm *ReportsManager) createDirectoryStructure() error {
	dirs := []string{
		rm.reportsDir,
		rm.config.HealthReportsDir,
		rm.config.SystemReportsDir,
		rm.config.CollectionReportsDir,
		rm.config.TestReportsDir,
		rm.config.LogsDir,
		rm.config.MetadataDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SaveHealthReport saves a health check report
func (rm *ReportsManager) SaveHealthReport(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("health-report-%s.json", timestamp)
	}

	filepath := filepath.Join(rm.config.HealthReportsDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save health report: %w", err)
	}

	return filepath, nil
}

// SaveSystemReport saves a system profile report
func (rm *ReportsManager) SaveSystemReport(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("system-profile-%s.json", timestamp)
	}

	filepath := filepath.Join(rm.config.SystemReportsDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save system report: %w", err)
	}

	return filepath, nil
}

// SaveCollectionReport saves a collection report
func (rm *ReportsManager) SaveCollectionReport(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("collection-report-%s.json", timestamp)
	}

	filepath := filepath.Join(rm.config.CollectionReportsDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save collection report: %w", err)
	}

	return filepath, nil
}

// SaveTestReport saves a test report
func (rm *ReportsManager) SaveTestReport(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("test-report-%s.json", timestamp)
	}

	filepath := filepath.Join(rm.config.TestReportsDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save test report: %w", err)
	}

	return filepath, nil
}

// SaveLog saves a log file
func (rm *ReportsManager) SaveLog(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("redtriage-%s.log", timestamp)
	}

	filepath := filepath.Join(rm.config.LogsDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save log: %w", err)
	}

	return filepath, nil
}

// SaveMetadata saves metadata information
func (rm *ReportsManager) SaveMetadata(data []byte, filename string) (string, error) {
	if filename == "" {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("metadata-%s.json", timestamp)
	}

	filepath := filepath.Join(rm.config.MetadataDir, filename)
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	return filepath, nil
}

// GetReportsDirectory returns the main reports directory
func (rm *ReportsManager) GetReportsDirectory() string {
	return rm.reportsDir
}

// GetHealthReportsDirectory returns the health reports directory
func (rm *ReportsManager) GetHealthReportsDirectory() string {
	return rm.config.HealthReportsDir
}

// GetSystemReportsDirectory returns the system reports directory
func (rm *ReportsManager) GetSystemReportsDirectory() string {
	return rm.config.SystemReportsDir
}

// GetCollectionReportsDirectory returns the collection reports directory
func (rm *ReportsManager) GetCollectionReportsDirectory() string {
	return rm.config.CollectionReportsDir
}

// GetTestReportsDirectory returns the test reports directory
func (rm *ReportsManager) GetTestReportsDirectory() string {
	return rm.config.TestReportsDir
}

// GetLogsDirectory returns the logs directory
func (rm *ReportsManager) GetLogsDirectory() string {
	return rm.config.LogsDir
}

// GetMetadataDirectory returns the metadata directory
func (rm *ReportsManager) GetMetadataDirectory() string {
	return rm.config.MetadataDir
}

// ListReports lists all reports in a specific category
func (rm *ReportsManager) ListReports(category string) ([]string, error) {
	var dir string
	switch category {
	case "health":
		dir = rm.config.HealthReportsDir
	case "system":
		dir = rm.config.SystemReportsDir
	case "collection":
		dir = rm.config.CollectionReportsDir
	case "tests":
		dir = rm.config.TestReportsDir
	case "logs":
		dir = rm.config.LogsDir
	case "metadata":
		dir = rm.config.MetadataDir
	default:
		return nil, fmt.Errorf("unknown report category: %s", category)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// CleanupOldReports removes reports older than the specified duration
func (rm *ReportsManager) CleanupOldReports(olderThan time.Duration) error {
	categories := []string{
		rm.config.HealthReportsDir,
		rm.config.SystemReportsDir,
		rm.config.CollectionReportsDir,
		rm.config.TestReportsDir,
		rm.config.LogsDir,
		rm.config.MetadataDir,
	}

	for _, dir := range categories {
		if err := rm.cleanupDirectory(dir, olderThan); err != nil {
			return fmt.Errorf("failed to cleanup directory %s: %w", dir, err)
		}
	}

	return nil
}

// cleanupDirectory removes files older than the specified duration from a directory
func (rm *ReportsManager) cleanupDirectory(dir string, olderThan time.Duration) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filepath); err != nil {
				// Log error but continue with other files
				fmt.Printf("Warning: failed to remove old file %s: %v\n", filepath, err)
			}
		}
	}

	return nil
}
