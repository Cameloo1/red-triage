package linux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redtriage/redtriage/collector"
)

// EnhancedLinuxCollector provides comprehensive forensic collection for Linux
type EnhancedLinuxCollector struct {
	baseDir string
}

// NewEnhancedLinuxCollector creates a new enhanced Linux collector
func NewEnhancedLinuxCollector() *EnhancedLinuxCollector {
	return &EnhancedLinuxCollector{
		baseDir: "/tmp/redtriage-enhanced",
	}
}

// CollectEnhancedArtifacts implements comprehensive artifact collection for Linux
func (elc *EnhancedLinuxCollector) CollectEnhancedArtifacts(ctx context.Context, profile collector.CollectionProfile) ([]collector.CollectionResult, error) {
	var results []collector.CollectionResult

	// Create base directory
	if err := os.MkdirAll(elc.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Collect volatile data first (if enabled)
	if profile.Extended {
		if volatileResults, err := elc.collectVolatileData(results); err == nil {
			results = volatileResults
		}
	}

	// Collect system artifacts
	if sysResults, err := elc.collectSystemArtifacts(results); err == nil {
		results = sysResults
	}

	// Collect network artifacts
	if netResults, err := elc.collectNetworkArtifacts(results); err == nil {
		results = netResults
	}

	// Collect file system artifacts
	if fsResults, err := elc.collectFileSystemArtifacts(results); err == nil {
		results = fsResults
	}

	// Collect process artifacts
	if procResults, err := elc.collectProcessArtifacts(results); err == nil {
		results = procResults
	}

	// Collect user artifacts
	if userResults, err := elc.collectUserArtifacts(results); err == nil {
		results = userResults
	}

	// Collect service artifacts
	if svcResults, err := elc.collectServiceArtifacts(results); err == nil {
		results = svcResults
	}

	// Collect log artifacts
	if logResults, err := elc.collectLogArtifacts(results); err == nil {
		results = logResults
	}

	// Collect timeline artifacts
	if timelineResults, err := elc.collectTimelineArtifacts(results); err == nil {
		results = timelineResults
	}

	return results, nil
}

// collectVolatileData collects volatile system data
func (elc *EnhancedLinuxCollector) collectVolatileData(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// Memory information
	if memInfo, err := exec.Command("cat", "/proc/meminfo").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "memory_info",
			Category:    "memory",
			Description: "Current memory state",
			Path:        filepath.Join(elc.baseDir, "memory_info.txt"),
		}

		if err := os.WriteFile(artifact.Path, memInfo, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(memInfo)),
			})
		}
	}

	// Load average
	if loadAvg, err := exec.Command("cat", "/proc/loadavg").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "load_average",
			Category:    "system",
			Description: "System load average",
			Path:        filepath.Join(elc.baseDir, "load_average.txt"),
		}

		if err := os.WriteFile(artifact.Path, loadAvg, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(loadAvg)),
			})
		}
	}

	// Current processes (detailed)
	if psOutput, err := exec.Command("ps", "auxf").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "process_tree",
			Category:    "process",
			Description: "Detailed process tree",
			Path:        filepath.Join(elc.baseDir, "process_tree.txt"),
		}

		if err := os.WriteFile(artifact.Path, psOutput, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(psOutput)),
			})
		}
	}

	// Network connections
	if netstat, err := exec.Command("netstat", "-tuln").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "network_connections",
			Category:    "network",
			Description: "Active network connections",
			Path:        filepath.Join(elc.baseDir, "network_connections.txt"),
		}

		if err := os.WriteFile(artifact.Path, netstat, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(netstat)),
			})
		}
	}

	return results, nil
}

// collectSystemArtifacts collects comprehensive system information
func (elc *EnhancedLinuxCollector) collectSystemArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// System information
	commands := map[string]string{
		"uname":        "uname -a",
		"hostname":     "hostname",
		"uptime":       "uptime",
		"cpuinfo":      "cat /proc/cpuinfo",
		"version":      "cat /proc/version",
		"lsb_release":  "lsb_release -a",
		"os_release":   "cat /etc/os-release",
		"kernel_cmdline": "cat /proc/cmdline",
		"interrupts":   "cat /proc/interrupts",
		"modules":      "lsmod",
		"dmesg":        "dmesg",
	}

	for name, cmd := range commands {
		if output, err := exec.Command("sh", "-c", cmd).Output(); err == nil {
			artifact := &collector.Artifact{
				Name:        fmt.Sprintf("system_%s", name),
				Category:    "system",
				Description: fmt.Sprintf("System %s information", name),
				Path:        filepath.Join(elc.baseDir, fmt.Sprintf("system_%s.txt", name)),
			}

			if err := os.WriteFile(artifact.Path, output, 0644); err == nil {
				results = append(results, collector.CollectionResult{
					Artifact: artifact,
					Success:  true,
					Size:     int64(len(output)),
				})
			}
		}
	}

	return results, nil
}

// collectNetworkArtifacts collects comprehensive network information
func (elc *EnhancedLinuxCollector) collectNetworkArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// Network interfaces
	if ipAddr, err := exec.Command("ip", "addr").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "network_interfaces",
			Category:    "network",
			Description: "Network interface configuration",
			Path:        filepath.Join(elc.baseDir, "network_interfaces.txt"),
		}

		if err := os.WriteFile(artifact.Path, ipAddr, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(ipAddr)),
			})
		}
	}

	// Routing table
	if ipRoute, err := exec.Command("ip", "route").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "routing_table",
			Category:    "network",
			Description: "Network routing table",
			Path:        filepath.Join(elc.baseDir, "routing_table.txt"),
		}

		if err := os.WriteFile(artifact.Path, ipRoute, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(ipRoute)),
			})
		}
	}

	// ARP table
	if arp, err := exec.Command("ip", "neigh").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "arp_table",
			Category:    "network",
			Description: "ARP table",
			Path:        filepath.Join(elc.baseDir, "arp_table.txt"),
		}

		if err := os.WriteFile(artifact.Path, arp, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(arp)),
			})
		}
	}

	// Network statistics
	if netstat, err := exec.Command("netstat", "-i").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "network_statistics",
			Category:    "network",
			Description: "Network interface statistics",
			Path:        filepath.Join(elc.baseDir, "network_statistics.txt"),
		}

		if err := os.WriteFile(artifact.Path, netstat, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(netstat)),
			})
		}
	}

	return results, nil
}

// collectFileSystemArtifacts collects file system information
func (elc *EnhancedLinuxCollector) collectFileSystemArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// Disk usage
	if df, err := exec.Command("df", "-h").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "disk_usage",
			Category:    "filesystem",
			Description: "Disk usage information",
			Path:        filepath.Join(elc.baseDir, "disk_usage.txt"),
		}

		if err := os.WriteFile(artifact.Path, df, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(df)),
			})
		}
	}

	// Mount points
	if mount, err := exec.Command("mount").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "mount_points",
			Category:    "filesystem",
			Description: "Mounted file systems",
			Path:        filepath.Join(elc.baseDir, "mount_points.txt"),
		}

		if err := os.WriteFile(artifact.Path, mount, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(mount)),
			})
		}
	}

	// Inode usage
	if dfi, err := exec.Command("df", "-i").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "inode_usage",
			Category:    "filesystem",
			Description: "Inode usage information",
			Path:        filepath.Join(elc.baseDir, "inode_usage.txt"),
		}

		if err := os.WriteFile(artifact.Path, dfi, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(dfi)),
			})
		}
	}

	// File system types
	if fstypes, err := exec.Command("blkid").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "filesystem_types",
			Category:    "filesystem",
			Description: "File system types and UUIDs",
			Path:        filepath.Join(elc.baseDir, "filesystem_types.txt"),
		}

		if err := os.WriteFile(artifact.Path, fstypes, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(fstypes)),
			})
		}
	}

	return results, nil
}

// collectProcessArtifacts collects process information
func (elc *EnhancedLinuxCollector) collectProcessArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// Process list with full details
	if ps, err := exec.Command("ps", "aux").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "process_list",
			Category:    "process",
			Description: "Complete process list",
			Path:        filepath.Join(elc.baseDir, "process_list.txt"),
		}

		if err := os.WriteFile(artifact.Path, ps, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(ps)),
			})
		}
	}

	// Process tree
	if pstree, err := exec.Command("pstree", "-p").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "process_tree",
			Category:    "process",
			Description: "Process tree with PIDs",
			Path:        filepath.Join(elc.baseDir, "process_tree.txt"),
		}

		if err := os.WriteFile(artifact.Path, pstree, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(pstree)),
			})
		}
	}

	// Open files
	if lsof, err := exec.Command("lsof").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "open_files",
			Category:    "process",
			Description: "Open files by processes",
			Path:        filepath.Join(elc.baseDir, "open_files.txt"),
		}

		if err := os.WriteFile(artifact.Path, lsof, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(lsof)),
			})
		}
	}

	return results, nil
}

// collectUserArtifacts collects user account information
func (elc *EnhancedLinuxCollector) collectUserArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// User accounts
	if passwd, err := exec.Command("cat", "/etc/passwd").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "user_accounts",
			Category:    "users",
			Description: "User account information",
			Path:        filepath.Join(elc.baseDir, "user_accounts.txt"),
		}

		if err := os.WriteFile(artifact.Path, passwd, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(passwd)),
			})
		}
	}

	// Group information
	if group, err := exec.Command("cat", "/etc/group").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "group_information",
			Category:    "users",
			Description: "Group information",
			Path:        filepath.Join(elc.baseDir, "group_information.txt"),
		}

		if err := os.WriteFile(artifact.Path, group, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(group)),
			})
		}
	}

	// Currently logged in users
	if who, err := exec.Command("who").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "logged_in_users",
			Category:    "users",
			Description: "Currently logged in users",
			Path:        filepath.Join(elc.baseDir, "logged_in_users.txt"),
		}

		if err := os.WriteFile(artifact.Path, who, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(who)),
			})
		}
	}

	// Last login information
	if last, err := exec.Command("last").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "last_logins",
			Category:    "users",
			Description: "Last login information",
			Path:        filepath.Join(elc.baseDir, "last_logins.txt"),
		}

		if err := os.WriteFile(artifact.Path, last, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(last)),
			})
		}
	}

	return results, nil
}

// collectServiceArtifacts collects service information
func (elc *EnhancedLinuxCollector) collectServiceArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// Systemd services
	if systemctl, err := exec.Command("systemctl", "list-units", "--type=service", "--state=running").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "running_services",
			Category:    "services",
			Description: "Running systemd services",
			Path:        filepath.Join(elc.baseDir, "running_services.txt"),
		}

		if err := os.WriteFile(artifact.Path, systemctl, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(systemctl)),
			})
		}
	}

	// Failed services
	if failed, err := exec.Command("systemctl", "list-units", "--type=service", "--state=failed").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "failed_services",
			Category:    "services",
			Description: "Failed systemd services",
			Path:        filepath.Join(elc.baseDir, "failed_services.txt"),
		}

		if err := os.WriteFile(artifact.Path, failed, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(failed)),
			})
		}
	}

	// Cron jobs
	if crontab, err := exec.Command("crontab", "-l").Output(); err == nil {
		artifact := &collector.Artifact{
			Name:        "cron_jobs",
			Category:    "services",
			Description: "User cron jobs",
			Path:        filepath.Join(elc.baseDir, "cron_jobs.txt"),
		}

		if err := os.WriteFile(artifact.Path, crontab, 0644); err == nil {
			results = append(results, collector.CollectionResult{
				Artifact: artifact,
				Success:  true,
				Size:     int64(len(crontab)),
			})
		}
	}

	return results, nil
}

// collectLogArtifacts collects system log files
func (elc *EnhancedLinuxCollector) collectLogArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	logFiles := []string{
		"/var/log/syslog",
		"/var/log/auth.log",
		"/var/log/kern.log",
		"/var/log/dmesg",
		"/var/log/messages",
		"/var/log/secure",
	}

	for _, logPath := range logFiles {
		if _, err := os.Stat(logPath); err == nil {
			artifact := &collector.Artifact{
				Name:        fmt.Sprintf("log_%s", filepath.Base(logPath)),
				Category:    "logs",
				Description: fmt.Sprintf("System log: %s", logPath),
				Path:        filepath.Join(elc.baseDir, fmt.Sprintf("log_%s", filepath.Base(logPath))),
			}

			// Copy log file with size limits
			if err := elc.copyLogFile(logPath, artifact.Path); err == nil {
				if stat, err := os.Stat(artifact.Path); err == nil {
					results = append(results, collector.CollectionResult{
						Artifact: artifact,
						Success:  true,
						Size:     stat.Size(),
					})
				}
			}
		}
	}

	return results, nil
}

// collectTimelineArtifacts collects timeline information
func (elc *EnhancedLinuxCollector) collectTimelineArtifacts(results []collector.CollectionResult) ([]collector.CollectionResult, error) {
	// File access times in common directories
	dirs := []string{"/home", "/tmp", "/var/log", "/etc"}
	
	var timeline strings.Builder
	timeline.WriteString("=== File Timeline Information ===\n\n")

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			timeline.WriteString(fmt.Sprintf("--- %s ---\n", dir))
			
			// Use find to get file access times
			cmd := exec.Command("find", dir, "-type", "f", "-printf", "%T@ %p\n", "-atime", "-7")
			if output, err := cmd.Output(); err == nil {
				timeline.WriteString(string(output))
			}
			timeline.WriteString("\n")
		}
	}

	artifact := &collector.Artifact{
		Name:        "file_timeline",
		Category:    "timeline",
		Description: "File access timeline information",
		Path:        filepath.Join(elc.baseDir, "file_timeline.txt"),
	}

	if err := os.WriteFile(artifact.Path, []byte(timeline.String()), 0644); err == nil {
		results = append(results, collector.CollectionResult{
			Artifact: artifact,
			Success:  true,
			Size:     int64(len(timeline.String())),
		})
	}

	return results, nil
}

// copyLogFile copies a log file with size limits
func (elc *EnhancedLinuxCollector) copyLogFile(src, dst string) error {
	// Check source file size
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Limit log file size to 10MB
	maxSize := int64(10 * 1024 * 1024)
	if stat.Size() > maxSize {
		// Use tail to get last 10MB
		cmd := exec.Command("tail", "-c", strconv.FormatInt(maxSize, 10), src)
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		return os.WriteFile(dst, output, 0644)
	}

	// Copy entire file
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// GetPlatform returns the platform identifier
func (elc *EnhancedLinuxCollector) GetPlatform() string {
	return "linux"
}

// GetCapabilities returns the collector capabilities
func (elc *EnhancedLinuxCollector) GetCapabilities() []string {
	return []string{
		"volatile_data",
		"system_artifacts",
		"network_artifacts",
		"filesystem_artifacts",
		"process_artifacts",
		"user_artifacts",
		"service_artifacts",
		"log_artifacts",
		"timeline_artifacts",
		"enhanced_collection",
	}
}
