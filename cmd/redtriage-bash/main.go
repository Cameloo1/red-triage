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
	// Enhanced terminal support for Linux/Bash
	terminal.EnableVirtualTerminal()
	
	// Linux/Bash-specific optimizations
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Enable additional Unix/Linux features
		terminal.EnableUnixFeatures()
	}

	// Show Linux/Bash-optimized banner
	showUnixBanner()

	// Create and execute the root command
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Unix Error: %v\n", err)
		os.Exit(1)
	}
}

func showUnixBanner() {
	fmt.Println("RedTriage Unix/Linux Interface")
	fmt.Printf("Version: %s\n", version.GetShortVersion())
	fmt.Println("Professional Incident Response Triage Tool")
	fmt.Println("Optimized for Linux, macOS, and Bash environments")
	fmt.Println("Enhanced Unix compatibility and terminal features")
	fmt.Println()
}
