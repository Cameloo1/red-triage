package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// GenerateCaseID generates a unique case identifier
func GenerateCaseID() string {
	// Generate 8 random bytes
	bytes := make([]byte, 8)
	rand.Read(bytes)
	
	// Convert to hex and take first 16 characters
	hexStr := hex.EncodeToString(bytes)
	
	// Format: RT-YYYYMMDD-HHMMSS-XXXXXXXX
	now := time.Now()
	timestamp := now.Format("20060102-150405")
	
	return fmt.Sprintf("RT-%s-%s", timestamp, hexStr[:8])
}

// HasAdminPrivileges checks if the current process has administrator privileges
func HasAdminPrivileges() bool {
	if runtime.GOOS == "windows" {
		// Check if we can access the Windows directory
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	}
	
	// On Unix-like systems, check if effective user ID is 0
	return os.Geteuid() == 0
}

// GetFreeDiskSpace returns the free disk space in bytes for the given path
func GetFreeDiskSpace(path string) (int64, error) {
	// For Windows, we'll use a different approach since Statfs is not available
	if runtime.GOOS == "windows" {
		// Use Windows-specific disk space checking
		// For now, return a placeholder value
		// TODO: Implement proper Windows disk space checking
		return 1024 * 1024 * 1024 * 100, nil // 100GB placeholder
	}
	
	// For Unix-like systems, we would use syscall.Statfs
	// But since it's not available on Windows, we'll skip this for now
	return 0, fmt.Errorf("disk space checking not implemented for %s", runtime.GOOS)
}

// CheckClockSanity checks if the system clock appears to be correct
func CheckClockSanity() error {
	now := time.Now()
	
	// Check if time is reasonable (not too far in past or future)
	// Allow for some reasonable range around current time
	minTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxTime := time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)
	
	if now.Before(minTime) || now.After(maxTime) {
		return fmt.Errorf("system time appears incorrect: %v", now)
	}
	
	return nil
}

// IsToolAvailable checks if a command-line tool is available
func IsToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// ExecuteCommand executes a command with timeout
func ExecuteCommand(timeout time.Duration, command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.Output()
	
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	
	return string(output), nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirectory creates a directory and all parent directories
func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileHash returns the SHA256 hash of a file
func GetFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetFileModTime returns the modification time of a file
func GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// SafeFilename generates a safe filename by removing/replacing invalid characters
func SafeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"<", ">", ":", "\"", "|", "?", "*", "/", "\\"}
	result := filename
	
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// Remove leading/trailing spaces and dots
	result = strings.Trim(result, " .")
	
	// Ensure filename is not empty
	if result == "" {
		result = "unnamed"
	}
	
	return result
}

// CreateTempDir creates a temporary directory
func CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// CleanupTempDir removes a temporary directory
func CleanupTempDir(path string) error {
	return os.RemoveAll(path)
}

// GetPlatform returns the current platform
func GetPlatform() string {
	return runtime.GOOS
}

// GetArchitecture returns the current architecture
func GetArchitecture() string {
	return runtime.GOARCH
}

// IsWindows checks if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux checks if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
