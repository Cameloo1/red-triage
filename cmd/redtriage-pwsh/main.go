package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/redtriage/redtriage/cmd"
	"github.com/redtriage/redtriage/internal/terminal"
	"github.com/redtriage/redtriage/internal/version"
)

func main() {
	// Enhanced terminal support for PowerShell
	terminal.EnableVirtualTerminal()
	
	// PowerShell-specific optimizations
	if runtime.GOOS == "windows" {
		// Enable additional Windows terminal features
		terminal.EnablePowerShellFeatures()
	}

	// Show PowerShell-optimized banner
	showPowerShellBanner()

	// Create and execute the root command
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "PowerShell Error: %v\n", err)
		os.Exit(1)
	}
}

func showPowerShellBanner() {
	fmt.Println("RedTriage PowerShell Interface")
	fmt.Printf("Version: %s\n", version.GetShortVersion())
	fmt.Println("Professional Incident Response Triage Tool")
	fmt.Println("Optimized for PowerShell and Windows Terminal")
	fmt.Println("Enhanced color support and terminal features")
	fmt.Println()
}
