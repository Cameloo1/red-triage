package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/redtriage/redtriage/cmd"
	"github.com/redtriage/redtriage/internal/session"
	"github.com/redtriage/redtriage/internal/terminal"
	"github.com/redtriage/redtriage/internal/version"
	"github.com/spf13/cobra"
)

var (
	interactive = flag.Bool("interactive", false, "Start interactive RedTriage session")
	versionFlag = flag.Bool("version", false, "Show version information")
	helpFlag    = flag.Bool("help", false, "Show help information")
)

func main() {
	// Enable Windows virtual terminal sequences for better color support
	terminal.EnableVirtualTerminal()

	// Parse command line flags
	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("RedTriage %s\n", version.GetShortVersion())
		fmt.Printf("Build Info: %s\n", version.GetBuildInfo())
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Show help if requested - use the full command help
	if *helpFlag {
		// Create and execute the root command with help
		rootCmd := cmd.NewRootCmd()
		
		// Disable color output for help to ensure consistent formatting
		rootCmd.SetHelpCommand(&cobra.Command{
			Use:    "help",
			Short:  "Help about any command",
			Hidden: true,
		})
		
		// Set help args and execute
		rootCmd.SetArgs([]string{"--help"})
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check if any non-flag arguments were provided
	// flag.NArg() returns the number of arguments remaining after flag parsing
	if flag.NArg() > 0 && !*interactive {
		// Command-line mode - arguments will be handled by the root command
		fmt.Println("RedTriage Command-Line Mode")
		fmt.Println("Use --interactive for interactive session")
		fmt.Println("Use --help for command options")
		os.Exit(0)
	}

	// Default to interactive mode if no non-flag arguments or if --interactive is specified
	if *interactive || flag.NArg() == 0 {
		fmt.Println("Starting RedTriage Interactive Session...")
		if err := session.StartInteractive(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
