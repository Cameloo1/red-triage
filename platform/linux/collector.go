package linux

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// LinuxCollector implements ArtifactCollector for Linux systems
type LinuxCollector struct {
	version string
}

// NewLinuxCollector creates a new Linux collector
func NewLinuxCollector() *LinuxCollector {
	return &LinuxCollector{
		version: "1.0.0",
	}
}

// CollectHostProfile collects basic host information
func (l *LinuxCollector) CollectHostProfile(ctx context.Context) (*collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"host_profile",
		"Linux host profile information",
		"host",
		"command",
	)

	// Collect hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Collect OS information
	osInfo := l.getOSInfo()

	// Collect system information
	sysInfo := l.getSystemInfo()

	// Create host profile data
	profileData := map[string]interface{}{
		"hostname":        hostname,
		"os_info":         osInfo,
		"system_info":     sysInfo,
		"collection_time": time.Now().Format(time.RFC3339),
	}

	// Convert to string for size calculation
	profileStr := fmt.Sprintf("%v", profileData)

	result := &collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     profileData,
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "system",
		},
		Size:     int64(len(profileStr)),
		Checksum: l.calculateChecksum(profileStr),
	}

	return result, nil
}

// CollectBasicArtifacts collects basic system artifacts
func (l *LinuxCollector) CollectBasicArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult

	// Collect running processes
	if processes, err := l.collectProcesses(); err == nil {
		results = append(results, processes)
	}

	// Collect running services
	if services, err := l.collectServices(); err == nil {
		results = append(results, services)
	}

	// Collect network information
	if network, err := l.collectNetworkInfo(); err == nil {
		results = append(results, network)
	}

	// Collect system logs
	if logs, err := l.collectSystemLogs(); err == nil {
		results = append(results, logs)
	}

	return results, nil
}

// CollectExtendedArtifacts collects extended system artifacts
func (l *LinuxCollector) CollectExtendedArtifacts(ctx context.Context) ([]collector.ArtifactResult, error) {
	var results []collector.ArtifactResult

	// Collect cron jobs
	if cron, err := l.collectCronJobs(); err == nil {
		results = append(results, cron)
	}

	// Collect user accounts
	if users, err := l.collectUserAccounts(); err == nil {
		results = append(results, users)
	}

	// Collect installed packages
	if packages, err := l.collectInstalledPackages(); err == nil {
		results = append(results, packages)
	}

	return results, nil
}

// getOSInfo retrieves operating system information
func (l *LinuxCollector) getOSInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
					info[key] = value
				}
			}
		}
	}

	// Get kernel version
	if kernel, err := exec.Command("uname", "-r").Output(); err == nil {
		info["kernel"] = strings.TrimSpace(string(kernel))
	}

	return info
}

// getSystemInfo retrieves basic system information
func (l *LinuxCollector) getSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Get system architecture
	info["architecture"] = runtime.GOARCH

	// Get number of CPUs
	info["cpu_count"] = runtime.NumCPU()

	// Get memory info
	if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		info["memory_info"] = string(meminfo)
	}

	return info
}

// collectProcesses collects running process information
func (l *LinuxCollector) collectProcesses() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"running_processes",
		"Currently running processes",
		"process",
		"command",
	)

	// Use ps to get process information
	output, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return collector.ArtifactResult{}, fmt.Errorf("failed to collect processes: %w", err)
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     string(output),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "ps",
		},
		Size:     int64(len(output)),
		Checksum: l.calculateChecksum(string(output)),
	}

	return result, nil
}

// collectServices collects running service information
func (l *LinuxCollector) collectServices() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"running_services",
		"Currently running services",
		"service",
		"command",
	)

	var serviceData strings.Builder

	// Try systemctl if available
	if _, err := exec.LookPath("systemctl"); err == nil {
		if output, err := exec.Command("systemctl", "list-units", "--type=service", "--state=running").Output(); err == nil {
			serviceData.WriteString("=== Systemd Services ===\n")
			serviceData.Write(output)
			serviceData.WriteString("\n\n")
		}
	}

	// Try service command if available
	if _, err := exec.LookPath("service"); err == nil {
		if output, err := exec.Command("service", "--status-all").Output(); err == nil {
			serviceData.WriteString("=== Service Status ===\n")
			serviceData.Write(output)
		}
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     serviceData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "systemctl,service",
		},
		Size:     int64(serviceData.Len()),
		Checksum: l.calculateChecksum(serviceData.String()),
	}

	return result, nil
}

// collectNetworkInfo collects network configuration information
func (l *LinuxCollector) collectNetworkInfo() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"network_info",
		"Network configuration and connections",
		"network",
		"command",
	)

	var networkData strings.Builder

	// Get IP configuration
	if output, err := exec.Command("ip", "addr").Output(); err == nil {
		networkData.WriteString("=== IP Configuration ===\n")
		networkData.Write(output)
		networkData.WriteString("\n\n")
	}

	// Get network connections
	if output, err := exec.Command("netstat", "-tuln").Output(); err == nil {
		networkData.WriteString("=== Network Connections ===\n")
		networkData.Write(output)
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     networkData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "ip,netstat",
		},
		Size:     int64(networkData.Len()),
		Checksum: l.calculateChecksum(networkData.String()),
	}

	return result, nil
}

// collectSystemLogs collects system log entries
func (l *LinuxCollector) collectSystemLogs() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"system_logs",
		"System log entries",
		"log",
		"file",
	)

	var logData strings.Builder

	// Collect recent system messages
	if output, err := exec.Command("journalctl", "--no-pager", "-n", "100").Output(); err == nil {
		logData.WriteString("=== System Journal ===\n")
		logData.Write(output)
		logData.WriteString("\n\n")
	}

	// Collect recent auth log entries
	if authLog, err := os.ReadFile("/var/log/auth.log"); err == nil {
		lines := strings.Split(string(authLog), "\n")
		start := len(lines) - 50
		if start < 0 {
			start = 0
		}
		logData.WriteString("=== Recent Auth Log ===\n")
		for i := start; i < len(lines); i++ {
			logData.WriteString(lines[i] + "\n")
		}
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     logData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "journalctl,file",
		},
		Size:     int64(logData.Len()),
		Checksum: l.calculateChecksum(logData.String()),
	}

	return result, nil
}

// collectCronJobs collects cron job information
func (l *LinuxCollector) collectCronJobs() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"cron_jobs",
		"Scheduled cron jobs",
		"task",
		"file",
	)

	var cronData strings.Builder

	// Collect system crontab
	if output, err := exec.Command("crontab", "-l").Output(); err == nil {
		cronData.WriteString("=== System Crontab ===\n")
		cronData.Write(output)
		cronData.WriteString("\n\n")
	}

	// Collect user crontabs
	if users, err := exec.Command("cut", "-d:", "-f1", "/etc/passwd").Output(); err == nil {
		userList := strings.Split(strings.TrimSpace(string(users)), "\n")
		for _, user := range userList {
			if user != "" && user != "root" {
				if output, err := exec.Command("crontab", "-u", user, "-l").Output(); err == nil {
					cronData.WriteString(fmt.Sprintf("=== User Crontab (%s) ===\n", user))
					cronData.Write(output)
					cronData.WriteString("\n\n")
				}
			}
		}
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     cronData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "crontab",
		},
		Size:     int64(cronData.Len()),
		Checksum: l.calculateChecksum(cronData.String()),
	}

	return result, nil
}

// collectUserAccounts collects user account information
func (l *LinuxCollector) collectUserAccounts() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"user_accounts",
		"User account information",
		"user",
		"file",
	)

	var userData strings.Builder

	// Read /etc/passwd
	if passwd, err := os.ReadFile("/etc/passwd"); err == nil {
		userData.WriteString("=== User Accounts (/etc/passwd) ===\n")
		userData.Write(passwd)
		userData.WriteString("\n\n")
	}

	// Read /etc/group
	if group, err := os.ReadFile("/etc/group"); err == nil {
		userData.WriteString("=== Groups (/etc/group) ===\n")
		userData.Write(group)
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     userData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "file",
		},
		Size:     int64(userData.Len()),
		Checksum: l.calculateChecksum(userData.String()),
	}

	return result, nil
}

// collectInstalledPackages collects installed package information
func (l *LinuxCollector) collectInstalledPackages() (collector.ArtifactResult, error) {
	artifact := collector.NewBaseArtifact(
		"installed_packages",
		"Installed package information",
		"software",
		"command",
	)

	var packageData strings.Builder

	// Try different package managers
	if _, err := exec.LookPath("dpkg"); err == nil {
		if output, err := exec.Command("dpkg", "-l").Output(); err == nil {
			packageData.WriteString("=== Debian Packages ===\n")
			packageData.Write(output)
			packageData.WriteString("\n\n")
		}
	}

	if _, err := exec.LookPath("rpm"); err == nil {
		if output, err := exec.Command("rpm", "-qa").Output(); err == nil {
			packageData.WriteString("=== RPM Packages ===\n")
			packageData.Write(output)
		}
	}

	result := collector.ArtifactResult{
		Artifact: artifact.Artifact,
		Data:     packageData.String(),
		Metadata: collector.Metadata{
			CollectedAt: time.Now(),
			Collector:   "linux",
			Version:     l.version,
			Source:      "dpkg,rpm",
		},
		Size:     int64(packageData.Len()),
		Checksum: l.calculateChecksum(packageData.String()),
	}

	return result, nil
}

// calculateChecksum calculates SHA256 checksum for data
func (l *LinuxCollector) calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
