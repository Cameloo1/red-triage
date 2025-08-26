package main

import (
	"fmt"
	"os"

	"github.com/redtriage/redtriage/cmd"
	"github.com/redtriage/redtriage/internal/terminal"
	"github.com/redtriage/redtriage/internal/version"
)

func main() {
	// Enable Windows virtual terminal sequences for better color support
	terminal.EnableVirtualTerminal()

	// Show banner
	showBanner()

	// Create and execute the root command
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showBanner() {
	fmt.Println("RedTriage Command-Line Interface")
	fmt.Printf("Version: %s\n", version.GetShortVersion())
	fmt.Println("Professional Incident Response Triage Tool")
	fmt.Println("Built for Windows-first forensics with Linux parity")
	fmt.Println()
}
