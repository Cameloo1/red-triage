//go:build windows
// +build windows

package terminal

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
	"strings"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	// Windows console mode constants
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	ENABLE_PROCESSED_INPUT             = 0x0001
	ENABLE_LINE_INPUT                  = 0x0002
	ENABLE_ECHO_INPUT                  = 0x0004
	ENABLE_WINDOW_INPUT                = 0x0008
	ENABLE_MOUSE_INPUT                 = 0x0010
	ENABLE_INSERT_MODE                 = 0x0020
	ENABLE_QUICK_EDIT_MODE            = 0x0040
	ENABLE_EXTENDED_FLAGS             = 0x0080
)

// EnableVirtualTerminal enables Windows virtual terminal sequences
func EnableVirtualTerminal() {
	if runtime.GOOS != "windows" {
		return
	}

	// Get console handle
	handle := syscall.Handle(os.Stdout.Fd())

	// Get current console mode
	var mode uint32
	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))

	// Enable virtual terminal processing
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING

	// Set new console mode
	procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	
	// Also try to enable for stderr
	if stderrHandle := syscall.Handle(os.Stderr.Fd()); stderrHandle != 0 {
		var stderrMode uint32
		procGetConsoleMode.Call(uintptr(stderrHandle), uintptr(unsafe.Pointer(&stderrMode)))
		stderrMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
		procSetConsoleMode.Call(uintptr(stderrHandle), uintptr(stderrMode))
	}
}

// IsColorSupported checks if the terminal supports colors
func IsColorSupported() bool {
	// Check if we're in a terminal
	if !isTerminal() {
		return false
	}

	// Check for color support
	if runtime.GOOS == "windows" {
		// Windows: check if virtual terminal sequences are supported
		return checkWindowsColorSupport()
	}

	// Unix-like systems: check TERM environment variable
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// checkWindowsColorSupport checks if Windows supports colors
func checkWindowsColorSupport() bool {
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	return (mode & ENABLE_VIRTUAL_TERMINAL_PROCESSING) != 0
}

// GetColorMode returns the color mode supported by the terminal
func GetColorMode() string {
	if !IsColorSupported() {
		return "none"
	}

	// Check for truecolor support
	if checkTrueColorSupport() {
		return "truecolor"
	}

	// Check for 256-color support
	if check256ColorSupport() {
		return "256"
	}

	return "basic"
}

// checkTrueColorSupport checks if truecolor is supported
func checkTrueColorSupport() bool {
	// This is a simplified check - in practice, you'd want to check
	// the COLORTERM environment variable and other indicators
	colorTerm := os.Getenv("COLORTERM")
	return colorTerm == "truecolor" || colorTerm == "24bit"
}

// check256ColorSupport checks if 256 colors are supported
func check256ColorSupport() bool {
	term := os.Getenv("TERM")
	// Common terminals that support 256 colors
	supportedTerms := []string{
		"xterm-256color", "screen-256color", "tmux-256color",
		"rxvt-unicode-256color", "linux-256color", "xterm-termite",
		"windows-256color", "cygwin-256color", "mintty-256color",
	}
	
	for _, supported := range supportedTerms {
		if term == supported {
			return true
		}
	}
	
	return false
}

// MapTrueColorTo256 maps truecolor hex values to 256-color approximations
func MapTrueColorTo256(hexColor string) int {
	// This is a simplified mapping - you'd want a more sophisticated algorithm
	switch hexColor {
	case "#91010d": // Red
		return 88
	case "#c9c9c9": // Triage
		return 252
	case "#02db09": // Dollar
		return 40
	case "#88eb8b": // Input
		return 120
	default:
		return 7 // Default to white
	}
}

// IsPowerShell checks if the current terminal is PowerShell
func IsPowerShell() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// Check for PowerShell-specific environment variables
	psModulePath := os.Getenv("PSModulePath")
	psVersionTable := os.Getenv("PSVersionTable")
	
	return psModulePath != "" || psVersionTable != ""
}

// IsCmd checks if the current terminal is Command Prompt
func IsCmd() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// Check for CMD-specific environment variables
	cmdLine := os.Getenv("CMDCMDLINE")
	comSpec := os.Getenv("COMSPEC")
	
	return cmdLine != "" || (comSpec != "" && strings.Contains(strings.ToLower(comSpec), "cmd.exe"))
}

// IsGitBash checks if the current terminal is Git Bash
func IsGitBash() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// Check for Git Bash-specific environment variables
	gitBash := os.Getenv("GIT_BASH")
	msys := os.Getenv("MSYSTEM")
	
	return gitBash != "" || (msys != "" && strings.Contains(strings.ToLower(msys), "msys"))
}

// GetTerminalType returns the type of terminal being used
func GetTerminalType() string {
	if runtime.GOOS != "windows" {
		// Unix-like systems
		term := os.Getenv("TERM")
		if term != "" {
			return term
		}
		return "unknown"
	}
	
	// Windows systems
	if IsPowerShell() {
		return "powershell"
	}
	if IsCmd() {
		return "cmd"
	}
	if IsGitBash() {
		return "gitbash"
	}
	
	return "windows"
}

// EnablePowerShellFeatures enables PowerShell-specific terminal features
func EnablePowerShellFeatures() {
	// Enable enhanced PowerShell terminal features
	if IsPowerShell() {
		// Enable additional color support for PowerShell
		EnableVirtualTerminal()
		
		// Set PowerShell-specific environment variables if needed
		if os.Getenv("POWERSHELL_TELEMETRY_OPTOUT") == "" {
			os.Setenv("POWERSHELL_TELEMETRY_OPTOUT", "1")
		}
	}
}

// EnableCmdFeatures enables CMD-specific terminal features
func EnableCmdFeatures() {
	// Enable enhanced CMD terminal features
	if IsCmd() {
		// Enable additional color support for CMD
		EnableVirtualTerminal()
		
		// Set CMD-specific environment variables if needed
		if os.Getenv("CMDCMDLINE") == "" {
			os.Setenv("CMDCMDLINE", "redtriage")
		}
	}
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
