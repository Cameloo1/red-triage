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
	// Enhanced terminal support for CMD
	terminal.EnableVirtualTerminal()
	
	// CMD-specific optimizations
	if runtime.GOOS == "windows" {
		// Enable additional Windows CMD features
		terminal.EnableCmdFeatures()
	}

	// Show CMD-optimized banner
	showCmdBanner()

	// Create and execute the root command
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "CMD Error: %v\n", err)
		os.Exit(1)
	}
}

func showCmdBanner() {
	fmt.Println("RedTriage Command Prompt Interface")
	fmt.Printf("Version: %s\n", version.GetShortVersion())
	fmt.Println("Professional Incident Response Triage Tool")
	fmt.Println("Optimized for Windows Command Prompt")
	fmt.Println("Enhanced compatibility with CMD environment")
	fmt.Println()
}
