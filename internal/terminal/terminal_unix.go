//go:build !windows
// +build !windows

package terminal

import (
	"os"
	"runtime"
)

// EnableVirtualTerminal enables virtual terminal sequences for Unix/Linux
func EnableVirtualTerminal() {
	// Unix/Linux terminals typically support virtual terminal sequences by default
	// Set TERM environment variable if not already set
	if os.Getenv("TERM") == "" {
		os.Setenv("TERM", "xterm-256color")
	}
}

// IsPowerShell returns false on Unix/Linux
func IsPowerShell() bool {
	return false
}

// IsCmd returns false on Unix/Linux
func IsCmd() bool {
	return false
}

// IsGitBash returns true if running in Git Bash on Unix/Linux
func IsGitBash() bool {
	return os.Getenv("SHELL") != "" && os.Getenv("SHELL") != "/bin/sh"
}

// GetTerminalType returns the terminal type for Unix/Linux
func GetTerminalType() string {
	if IsGitBash() {
		return "gitbash"
	}
	return "unix"
}

// EnablePowerShellFeatures enables PowerShell-specific features (no-op on Unix/Linux)
func EnablePowerShellFeatures() {
	// No PowerShell features on Unix/Linux
}

// EnableCmdFeatures enables CMD-specific features (no-op on Unix/Linux)
func EnableCmdFeatures() {
	// No CMD features on Unix/Linux
}

// EnableUnixFeatures enables Unix/Linux-specific terminal features
func EnableUnixFeatures() {
	// Enable enhanced Unix/Linux terminal features
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Enable additional color support for Unix terminals
		EnableVirtualTerminal()

		// Set Unix-specific environment variables if needed
		if os.Getenv("TERM") == "" {
			os.Setenv("TERM", "xterm-256color")
		}
	}
}
