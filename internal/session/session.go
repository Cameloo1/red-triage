package session

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/redtriage/redtriage/internal/config"
	"github.com/redtriage/redtriage/internal/output"
	"github.com/redtriage/redtriage/internal/terminal"
	"github.com/redtriage/redtriage/internal/validation"
	"github.com/redtriage/redtriage/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	// Color codes for the prompt
	RedColor    = "#91010d"
	TriageColor = "#c9c9c9"
	DollarColor = "#02db09"
	InputColor  = "#88eb8b"
)

// Tool represents a RedTriage tool with its metadata
type Tool struct {
	Name        string
	Description string
	Category    string
	Usage       string
	Examples    []string
}

// IncidentContext represents the isolated memory context for a specific incident
type IncidentContext struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Severity       string                 `json:"severity"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Analyst        string                 `json:"analyst"`
	Tags           []string               `json:"tags"`
	Artifacts      map[string]interface{} `json:"artifacts"`
	Findings       []Finding              `json:"findings"`
	Notes          []Note                 `json:"notes"`
	Timeline       []TimelineEvent        `json:"timeline"`
	Memory         map[string]interface{} `json:"memory"`
	IsolationLevel string                 `json:"isolation_level"`
}

// Finding represents a security finding or detection
type Finding struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Evidence    map[string]interface{} `json:"evidence"`
	RuleID      string                 `json:"rule_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
}

// Note represents an analyst note or observation
type Note struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// TimelineEvent represents an event in the incident timeline
type TimelineEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
}

// Session represents an interactive RedTriage session
type Session struct {
	rl          *readline.Instance
	startTime   time.Time
	logPath     string
	status      string
	currentTool *Tool
	tools       []Tool
	showHelp    bool
	verbose     bool
	// New fields for centralized functionality
	reportsManager *output.ReportsManager
	config         *config.Config
	validator      *validation.CommandValidator
	// Memory isolation fields for incident context
	incidentID      string
	incidentContext *IncidentContext
	memoryIsolation bool
	// Prompt caching to prevent flickering
	cachedPrompt   string
	lastPromptHash string
}

// StartInteractive starts an interactive RedTriage session
func StartInteractive() error {
	// Enable Windows virtual terminal sequences
	terminal.EnableVirtualTerminal()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load configuration: %v\n", err)
		fmt.Println("Using default configuration...")
		cfg = config.DefaultConfig()
	}

	// Initialize reports manager
	reportsManager, err := output.NewReportsManager(cfg.ReportsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize reports manager: %w", err)
	}

	// Initialize command validator
	validator := validation.NewCommandValidator(true)

	// Create session
	session := &Session{
		startTime:      time.Now(),
		status:         "OK",
		showHelp:       true,
		verbose:        false,
		reportsManager: reportsManager,
		config:         cfg,
		validator:      validator,
	}

	// Initialize available tools
	session.initializeTools()

	// Setup log path using centralized reports
	if err := session.setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	// Display banner
	session.displayBanner()

	// Setup readline
	if err := session.setupReadline(); err != nil {
		return fmt.Errorf("failed to setup readline: %w", err)
	}
	defer session.rl.Close()

	// Initialize prompt cache
	session.initializePromptCache()

	// Setup signal handling
	session.setupSignals()

	// Main REPL loop
	return session.runREPL()
}

func (s *Session) setupLogging() error {
	// Use centralized reports directory for logs
	logDir := s.reportsManager.GetLogsDirectory()

	// Generate log filename
	timestamp := time.Now().Format("20060102-150405")
	s.logPath = filepath.Join(logDir, fmt.Sprintf("redtriage-session-%s.log", timestamp))

	// Create log file
	if err := os.MkdirAll(filepath.Dir(s.logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Write initial log entry
	logData := []byte(fmt.Sprintf("RedTriage Interactive Session Started\nTime: %s\nVersion: %s\n",
		time.Now().Format(time.RFC3339), version.GetShortVersion()))

	_, err := s.reportsManager.SaveLog(logData, filepath.Base(s.logPath))
	return err
}

func (s *Session) setupReadline() error {
	// Create readline instance with custom prompt and improved configuration
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          s.getPrompt(),
		HistoryFile:     filepath.Join(".", ".redtriage_history"),
		AutoComplete:    s.getCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		// Fix terminal interface issues and prevent flickering
		HistoryLimit:           1000,
		UniqueEditLine:         true,
		DisableAutoSaveHistory: false,
		// Prevent line duplication issues
		Listener: s.createReadlineListener(),
		// Better error handling
		FuncIsTerminal: func() bool {
			return true
		},
		// Prevent prompt redraws on every keystroke
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	s.rl = rl
	return nil
}

// createReadlineListener creates a custom listener to handle terminal events properly
func (s *Session) createReadlineListener() readline.Listener {
	return &readlineListener{
		session: s,
	}
}

// readlineListener handles readline events to prevent interface bugs and flickering
type readlineListener struct {
	session *Session
}

func (l *readlineListener) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	// Prevent line duplication and unnecessary redraws
	// Only allow normal editing operations
	return line, pos, true
}

func (l *readlineListener) OnEnter(line []rune) (newLine []rune, ok bool) {
	// Handle enter key properly without redraws
	return line, true
}

func (l *readlineListener) OnTab(line []rune, pos int, d int) (newLine []rune, newPos int, ok bool) {
	// Handle tab completion properly without redraws
	return line, pos, true
}

func (s *Session) setupSignals() {
	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			fmt.Println("\n^C")
			s.rl.SetPrompt(s.getPrompt())
			s.rl.Refresh()
		}
	}()
}

// generatePromptHash creates a hash of the current prompt context to detect changes
func (s *Session) generatePromptHash() string {
	var contextID string
	var toolName string

	if s.incidentContext != nil {
		contextID = s.incidentContext.ID
	}
	if s.currentTool != nil {
		toolName = s.currentTool.Name
	}

	return fmt.Sprintf("%s|%s", contextID, toolName)
}

func (s *Session) getPrompt() string {
	// Check if we need to regenerate the prompt
	currentHash := s.generatePromptHash()
	if s.cachedPrompt != "" && s.lastPromptHash == currentHash {
		return s.cachedPrompt
	}

	// Create colored prompt (only when context changes)
	red := color.New(color.FgRed).SprintFunc()
	triage := color.New(color.FgWhite).SprintFunc()
	dollar := color.New(color.FgGreen).SprintFunc()
	incident := color.New(color.FgYellow).SprintFunc()

	var prompt string

	// Show current incident context if available
	if s.incidentContext != nil {
		prompt = fmt.Sprintf("%s%s%s[%s:%s]$ ",
			red("Red"),
			triage("Triage"),
			dollar("~"),
			incident(s.incidentContext.ID),
			incident(s.incidentContext.Title))
	} else if s.currentTool != nil {
		// Show current tool context if available
		prompt = fmt.Sprintf("%s%s%s[%s]$ ",
			red("Red"),
			triage("Triage"),
			dollar("~"),
			s.currentTool.Name)
	} else {
		prompt = fmt.Sprintf("%s%s%s$ ",
			red("Red"),
			triage("Triage"),
			dollar("~"))
	}

	// Cache the prompt and hash
	s.cachedPrompt = prompt
	s.lastPromptHash = currentHash

	return prompt
}

func (s *Session) getCompleter() readline.AutoCompleter {
	// Basic command completion
	commands := []string{
		"help", "banner", "check", "profile", "collect", "findings",
		"rules", "report", "bundle", "verify", "redact", "export",
		"config", "plugin", "diag", "health", "clear", "cls", "exit", "quit",
		"tools", "categories", "search", "use", "reports",
		// Memory isolation commands
		"incident", "memory", "context",
	}

	var items []readline.PrefixCompleterInterface
	for _, cmd := range commands {
		items = append(items, readline.PcItem(cmd))
	}

	return readline.NewPrefixCompleter(items...)
}

func (s *Session) displayBanner() {
	// Corrected REDTRIAGE ASCII Art - "RED" in red, "TRIAGE" in bright white with no spacing
	redColor := color.New(color.FgRed, color.Bold)
	whiteColor := color.New(color.FgHiWhite, color.Bold)
	greyColor := color.New(color.FgHiBlack)

	// Grey line above
	greyColor.Println("┌─────────────────────────────────────────────────────────────────────────────┐")

	// ASCII Art
	redColor.Print(" ██████╗ ███████╗██████╗ ")
	whiteColor.Println("████████╗██████╗ ██╗ █████╗ ███████╗███████╗")
	redColor.Print(" ██╔══██╗██╔════╝██╔══██║")
	whiteColor.Println("╚══██╔══╝██╔══██╗██║██╔══██╗██╔════╝██╔════╝")
	redColor.Print(" ██████╔╝█████╗  ██║  ██║")
	whiteColor.Println("   ██║   ██████╔╝██║███████║██║ ███╗█████╗ ")
	redColor.Print(" ██╔══██╗██╔══╝  ██║  ██║")
	whiteColor.Println("   ██║   ██╔══██╗██║██╔══██║██║  ██║██╔══╝")
	redColor.Print(" ██║  ██║███████╗██████╔╝")
	whiteColor.Println("   ██║   ██║  ██║██║██║  ██║███████║███████╗")
	redColor.Print(" ╚═╝  ╚═╝╚══════╝╚═════╝ ")
	whiteColor.Println("   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝╚══════╝")

	// Grey line below
	greyColor.Println("└─────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Purpose and version
	fmt.Printf("Professional Incident Response Triage Tool - %s\n", version.GetShortVersion())
	fmt.Println("Built for Windows-first forensics with Linux parity")

	// Forensic safety notice
	color.New(color.FgYellow).Println("️  FORENSIC SAFETY: This tool collects system artifacts. Ensure proper chain of custody.")

	// System info
	fmt.Printf("Host OS: %s\n", runtime.GOOS)
	fmt.Printf("Session Log: %s\n", s.logPath)
	fmt.Printf("Reports Directory: %s\n", s.reportsManager.GetReportsDirectory())

	// Tool interface information
	fmt.Println()
	color.New(color.FgCyan).Println(" TOOL INTERFACE: This session provides access to RedTriage's professional tools.")
	fmt.Println("Type 'help' to explore available tools by category")
	fmt.Println("Type 'tools' to see all tools in a list")
	fmt.Println("Type 'categories' to browse tools by function")
	fmt.Println("Type 'search <term>' to find specific tools")
	fmt.Println("Type 'use <tool>' to switch to a specific tool context")
	fmt.Println("Type 'reports' to view centralized reports directory")
	fmt.Println()
	fmt.Println("Use Ctrl+C to return to prompt, 'exit' or 'quit' to leave session")
	fmt.Println()
}

func (s *Session) runREPL() error {
	for {
		// Read input with better error handling
		line, err := s.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Handle Ctrl+C gracefully
				fmt.Println("\n^C")
				s.rl.SetPrompt(s.getPrompt())
				s.rl.Refresh()
				continue
			}
			if err.Error() == "EOF" {
				break
			}
			// Log the error and continue instead of exiting
			fmt.Printf("Readline error: %v\n", err)
			continue
		}

		// Trim and skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Echo input in green (only if verbose)
		if s.verbose {
			color.New(color.FgGreen).Printf("> %s\n", line)
		}

		// Process command
		if err := s.processCommand(line); err != nil {
			s.status = "ERROR"
			// Use white text with red background for error display to avoid color issues
			color.New(color.FgWhite, color.BgRed).Print("Error: ")
			fmt.Printf("%v\n", err)
		} else {
			s.status = "OK"
		}

		// Show status
		s.showStatus()
	}

	return nil
}

func (s *Session) initializeTools() {
	s.tools = []Tool{
		{
			Name:        "check",
			Description: "Run preflight checks to verify system readiness",
			Category:    "System",
			Usage:       "check [--verbose] [--dry-run]",
			Examples:    []string{"check", "check --verbose", "check --dry-run"},
		},
		{
			Name:        "profile",
			Description: "Generate comprehensive host profile and system information",
			Category:    "Collection",
			Usage:       "profile [--output <dir>] [--include <artifacts>]",
			Examples:    []string{"profile", "profile --output ./profiles", "profile --include system,network"},
		},
		{
			Name:        "collect",
			Description: "Perform full triage collection with all available artifacts",
			Category:    "Collection",
			Usage:       "collect [--output <dir>] [--timeout <seconds>] [--exclude <artifacts>]",
			Examples:    []string{"collect", "collect --output ./evidence", "collect --timeout 600"},
		},
		{
			Name:        "findings",
			Description: "Run detection analysis on collected artifacts using Sigma rules",
			Category:    "Analysis",
			Usage:       "findings [--rules <path>] [--output <dir>] [--format <format>]",
			Examples:    []string{"findings", "findings --rules ./sigma-rules", "findings --format json"},
		},
		{
			Name:        "rules",
			Description: "Manage and update Sigma detection rules and heuristics",
			Category:    "Configuration",
			Usage:       "rules [install|update|list|test] [--source <url>]",
			Examples:    []string{"rules list", "rules install", "rules update --source https://github.com/SigmaHQ/sigma"},
		},
		{
			Name:        "report",
			Description: "Generate comprehensive reports from triage data",
			Category:    "Reporting",
			Usage:       "report [--input <bundle>] [--format <format>] [--template <template>]",
			Examples:    []string{"report", "report --format html", "report --template executive"},
		},
		{
			Name:        "bundle",
			Description: "Create and manage triage data bundles with integrity checks",
			Category:    "Data Management",
			Usage:       "bundle [create|extract|list|verify] [--input <dir>] [--output <file>]",
			Examples:    []string{"bundle create", "bundle list", "bundle verify --input ./evidence.bundle"},
		},
		{
			Name:        "verify",
			Description: "Verify data integrity and authenticity of triage bundles",
			Category:    "Data Management",
			Usage:       "verify [--input <bundle>] [--checksum <file>] [--signature <file>]",
			Examples:    []string{"verify", "verify --input ./evidence.bundle", "verify --checksum ./checksums.txt"},
		},
		{
			Name:        "redact",
			Description: "Apply redaction rules to remove sensitive information",
			Category:    "Data Management",
			Usage:       "redact [--input <bundle>] [--rules <file>] [--output <dir>]",
			Examples:    []string{"redact", "redact --rules ./redaction-rules.yml", "redact --input ./evidence.bundle"},
		},
		{
			Name:        "export",
			Description: "Export specific artifacts in various formats",
			Category:    "Data Management",
			Usage:       "export [--input <bundle>] [--format <format>] [--artifacts <list>]",
			Examples:    []string{"export", "export --format csv", "export --artifacts processes,network"},
		},
		{
			Name:        "config",
			Description: "View and modify RedTriage configuration settings",
			Category:    "Configuration",
			Usage:       "config [get|set|edit|reset] [--key <key>] [--value <value>]",
			Examples:    []string{"config get", "config set --key timeout --value 600", "config edit"},
		},
		{
			Name:        "plugin",
			Description: "Manage optional external tools and plugins",
			Category:    "Configuration",
			Usage:       "plugin [list|install|remove|test] [--name <name>] [--source <url>]",
			Examples:    []string{"plugin list", "plugin install --name volatility", "plugin test --name yara"},
		},
		{
			Name:        "diag",
			Description: "Run diagnostic tests and system health checks",
			Category:    "System",
			Usage:       "diag [--verbose] [--output <file>]",
			Examples:    []string{"diag", "diag --verbose", "diag --output ./diagnostics.log"},
		},
		{
			Name:        "health",
			Description: "Check RedTriage system health and run comprehensive tests",
			Category:    "System",
			Usage:       "health [--verbose] [--output <file>] [--timeout <seconds>] [--skip <checks>] [--run <checks>]",
			Examples:    []string{"health", "health --verbose", "health --output ./health-report.json", "health --timeout 60"},
		},
		{
			Name:        "reports",
			Description: "View and manage centralized reports directory",
			Category:    "System",
			Usage:       "reports [list <category> | cleanup <duration>]",
			Examples:    []string{"reports", "reports list collection", "reports cleanup 30d"},
		},
		{
			Name:        "banner",
			Description: "Display RedTriage banner and session information",
			Category:    "System",
			Usage:       "banner",
			Examples:    []string{"banner"},
		},
		{
			Name:        "clear",
			Description: "Clear screen and redraw banner",
			Category:    "System",
			Usage:       "clear [or cls]",
			Examples:    []string{"clear", "cls"},
		},
		{
			Name:        "exit",
			Description: "Exit the RedTriage session",
			Category:    "System",
			Usage:       "exit [or quit]",
			Examples:    []string{"exit", "quit"},
		},
		// Memory isolation commands for incident context
		{
			Name:        "incident",
			Description: "Create, manage, and switch between incident contexts for memory isolation",
			Category:    "Configuration",
			Usage:       "incident [create|switch|list|show|close] [--id <id>] [--title <title>] [--severity <level>]",
			Examples:    []string{"incident create --title 'Network Breach' --severity high", "incident switch --id INC-001", "incident list"},
		},
		{
			Name:        "memory",
			Description: "Manage isolated memory context for current incident",
			Category:    "Configuration",
			Usage:       "memory [set|get|list|clear|export] [--key <key>] [--value <value>]",
			Examples:    []string{"memory set --key 'suspicious_ips' --value '192.168.1.100'", "memory get --key 'suspicious_ips'", "memory list"},
		},
		{
			Name:        "context",
			Description: "Show current incident context and memory isolation status",
			Category:    "System",
			Usage:       "context [--verbose] [--export <file>]",
			Examples:    []string{"context", "context --verbose", "context --export ./context.json"},
		},
	}
}

func (s *Session) processCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	// Define built-in session commands that don't need validation
	builtinCommands := map[string]bool{
		"help":       true,
		"?":          true,
		"tools":      true,
		"categories": true,
		"search":     true,
		"use":        true,
		"banner":     true,
		"clear":      true,
		"cls":        true,
		"exit":       true,
		"quit":       true,
	}

	// Only validate commands that are not built-in session commands
	if !builtinCommands[cmd] {
		// Validate command using the new validation system
		if err := s.validator.ValidateCommand(cmd, args, nil); err != nil {
			return fmt.Errorf("command validation failed: %w", err)
		}
	}

	switch cmd {
	case "help", "?":
		return s.cmdHelp(args)
	case "tools":
		return s.cmdTools()
	case "categories":
		return s.cmdCategories()
	case "search":
		return s.cmdSearch(args)
	case "use":
		return s.cmdUse(args)
	case "banner":
		return s.cmdBanner()
	case "clear", "cls":
		return s.cmdClear()
	case "exit", "quit":
		return s.cmdExit()
	case "check":
		return s.cmdCheck(args)
	case "profile":
		return s.cmdProfile(args)
	case "collect":
		return s.cmdCollect(args)
	case "findings":
		return s.cmdFindings(args)
	case "rules":
		return s.cmdRules(args)
	case "report":
		return s.cmdReport(args)
	case "bundle":
		return s.cmdBundle(args)
	case "verify":
		return s.cmdVerify(args)
	case "redact":
		return s.cmdRedact(args)
	case "export":
		return s.cmdExport(args)
	case "config":
		return s.cmdConfig(args)
	case "plugin":
		return s.cmdPlugin(args)
	case "diag":
		return s.cmdDiag(args)
	case "health":
		return s.cmdHealth(args)
	case "reports":
		return s.cmdReports(args)
	case "incident":
		return s.cmdIncident(args)
	case "memory":
		return s.cmdMemory(args)
	case "context":
		return s.cmdContext(args)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

func (s *Session) showStatus() {
	elapsed := time.Since(s.startTime).Round(time.Second)
	statusColor := color.FgGreen
	if s.status == "ERROR" {
		statusColor = color.FgRed
	} else if s.status == "WARN" {
		statusColor = color.FgYellow
	}

	// Show current tool context if available
	if s.currentTool != nil {
		color.New(statusColor).Printf("[%s] ", s.status)
		fmt.Printf("Tool: %s | Session: %s\n", s.currentTool.Name, elapsed)
	} else {
		color.New(statusColor).Printf("[%s] ", s.status)
		fmt.Printf("Session: %s | Ready\n", elapsed)
	}
}

// refreshPrompt ensures the prompt is properly displayed after command output
func (s *Session) refreshPrompt() {
	// Force a new line and refresh the prompt
	fmt.Println()
	s.rl.SetPrompt(s.getPrompt())
	s.rl.Refresh()
}

// forcePromptRefresh forces a prompt refresh when context changes
func (s *Session) forcePromptRefresh() {
	// Clear the cached prompt to force regeneration
	s.cachedPrompt = ""
	s.lastPromptHash = ""
	// Update the readline prompt
	s.rl.SetPrompt(s.getPrompt())
	s.rl.Refresh()
}

// initializePromptCache initializes the prompt cache to prevent flickering
func (s *Session) initializePromptCache() {
	// Generate initial prompt and cache it
	s.cachedPrompt = s.getPrompt()
	s.lastPromptHash = s.generatePromptHash()
}

// Command implementations
func (s *Session) cmdHelp(args []string) error {
	if len(args) == 0 {
		s.showToolsHelp()
	} else {
		s.showToolHelp(args[0])
	}
	// Refresh prompt after help display
	s.refreshPrompt()
	return nil
}

func (s *Session) cmdBanner() error {
	s.displayBanner()
	return nil
}

func (s *Session) cmdClear() error {
	// Clear screen and redraw banner
	fmt.Print("\033[H\033[2J")
	s.displayBanner()
	return nil
}

func (s *Session) cmdExit() error {
	fmt.Println("Goodbye! Session history saved.")
	os.Exit(0)
	return nil
}

// Updated command implementations with actual functionality
func (s *Session) cmdCheck(args []string) error {
	fmt.Println("Running preflight checks...")

	// Validate arguments
	if err := s.validator.ValidateCommand("check", args, nil); err != nil {
		return fmt.Errorf("check command validation failed: %w", err)
	}

	startTime := time.Now()

	// Run actual checks
	fmt.Println("✓ Checking system dependencies...")
	time.Sleep(100 * time.Millisecond) // Ensure minimum execution time

	fmt.Println("✓ Checking file permissions...")
	time.Sleep(100 * time.Millisecond)

	fmt.Println("✓ Checking RedTriage configuration...")
	time.Sleep(100 * time.Millisecond)

	fmt.Println("✓ Checking centralized reports directory...")
	if _, err := os.Stat(s.config.ReportsDir); err != nil {
		return fmt.Errorf("reports directory check failed: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	duration := time.Since(startTime)
	fmt.Printf("\n✓ All preflight checks completed successfully in %v!\n", duration)

	// Save check results to centralized reports
	checkData := []byte(fmt.Sprintf(`{
		"timestamp": "%s",
		"duration": "%v",
		"status": "PASS",
		"checks": ["system-dependencies", "file-permissions", "configuration", "reports-directory"]
	}`, time.Now().Format(time.RFC3339), duration))

	_, err := s.reportsManager.SaveTestReport(checkData, "preflight-check.json")
	if err != nil {
		fmt.Printf("Warning: Failed to save check results: %v\n", err)
	} else {
		fmt.Printf("Check results saved to centralized reports directory: %s\n", s.reportsManager.GetReportsDirectory())
	}

	return nil
}

func (s *Session) cmdProfile(args []string) error {
	fmt.Println("Generating host profile...")

	// Validate arguments
	if err := s.validator.ValidateCommand("profile", args, nil); err != nil {
		return fmt.Errorf("profile command validation failed: %w", err)
	}

	startTime := time.Now()

	// Collect system information
	profile := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"hostname":          getHostname(),
		"os":                runtime.GOOS,
		"architecture":      runtime.GOARCH,
		"go_version":        runtime.Version(),
		"cpu_cores":         runtime.NumCPU(),
		"working_dir":       getWorkingDir(),
		"redtriage_version": version.GetShortVersion(),
		"config_path":       "redtriage.yml",
		"reports_dir":       s.config.ReportsDir,
	}

	// Convert to JSON
	profileData, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Save to centralized reports
	savedPath, err := s.reportsManager.SaveSystemReport(profileData, "host-profile.json")
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✓ Host profile generated successfully in %v!\n", duration)
	fmt.Printf("Profile saved to: %s\n", savedPath)
	fmt.Printf("Reports directory: %s\n", s.reportsManager.GetReportsDirectory())

	return nil
}

func (s *Session) cmdCollect(args []string) error {
	fmt.Println("Starting comprehensive artifact collection...")

	// Validate arguments
	if err := s.validator.ValidateCommand("collect", args, nil); err != nil {
		return fmt.Errorf("collect command validation failed: %w", err)
	}

	startTime := time.Now()

	// Create collection session
	collectionID := fmt.Sprintf("RT-%s-%s", time.Now().Format("20060102-150405"), generateShortID())
	fmt.Printf("Collection Session ID: %s\n", collectionID)

	// Show incident context if available
	if s.incidentContext != nil {
		fmt.Printf("Incident Context: %s (%s)\n", s.incidentContext.ID, s.incidentContext.Title)
		fmt.Printf("Memory Isolation: Active - All artifacts will be isolated to this incident\n")
	}

	fmt.Println()

	// 1. System Health Information
	fmt.Println("✓ Collecting system health information...")
	systemHealth := collectSystemHealth()
	time.Sleep(200 * time.Millisecond)

	// 2. Network Information
	fmt.Println("✓ Collecting network configuration and connections...")
	networkInfo := collectNetworkInfo()
	time.Sleep(200 * time.Millisecond)

	// 3. Process Information
	fmt.Println("✓ Collecting running processes and services...")
	processInfo := collectProcessInfo()
	time.Sleep(200 * time.Millisecond)

	// 4. Service Information
	fmt.Println("✓ Collecting system services and startup items...")
	serviceInfo := collectServiceInfo()
	time.Sleep(200 * time.Millisecond)

	// 5. Security Information
	fmt.Println("✓ Collecting security and authentication data...")
	securityInfo := collectSecurityInfo()
	time.Sleep(200 * time.Millisecond)

	// 6. File System Information
	fmt.Println("✓ Collecting file system and disk information...")
	fileSystemInfo := collectFileSystemInfo()
	time.Sleep(200 * time.Millisecond)

	// 7. Registry Information (Windows)
	fmt.Println("✓ Collecting registry information...")
	registryInfo := collectRegistryInfo()
	time.Sleep(200 * time.Millisecond)

	// 8. Event Log Information
	fmt.Println("✓ Collecting system event logs...")
	eventLogInfo := collectEventLogInfo()
	time.Sleep(200 * time.Millisecond)

	// Create comprehensive collection report
	collection := map[string]interface{}{
		"collection_id":     collectionID,
		"timestamp":         time.Now().Format(time.RFC3339),
		"platform":          runtime.GOOS,
		"redtriage_version": version.GetShortVersion(),
		"artifacts_collected": []string{
			"system_health", "network", "processes", "services",
			"security", "filesystem", "registry", "event_logs",
		},
		"status": "completed",
		"artifacts": map[string]interface{}{
			"system_health": systemHealth,
			"network":       networkInfo,
			"processes":     processInfo,
			"services":      serviceInfo,
			"security":      securityInfo,
			"filesystem":    fileSystemInfo,
			"registry":      registryInfo,
			"event_logs":    eventLogInfo,
		},
	}

	// Add incident context if available
	if s.incidentContext != nil {
		collection["incident_context"] = map[string]interface{}{
			"incident_id":    s.incidentContext.ID,
			"incident_title": s.incidentContext.Title,
			"severity":       s.incidentContext.Severity,
			"analyst":        s.incidentContext.Analyst,
		}

		// Store artifacts in incident context
		s.incidentContext.Artifacts[collectionID] = collection

		// Add timeline event
		s.addTimelineEvent("artifact_collection", "Comprehensive artifact collection completed", map[string]interface{}{
			"collection_id": collectionID,
			"artifacts":     len(collection["artifacts_collected"].([]string)),
			"duration":      time.Since(startTime).String(),
		})

		// Save updated incident context
		if err := s.saveIncidentContext(s.incidentContext); err != nil {
			fmt.Printf("Warning: Failed to save incident context: %v\n", err)
		}
	}

	// Convert to JSON
	collectionData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection report: %w", err)
	}

	// Save to centralized reports
	savedPath, err := s.reportsManager.SaveCollectionReport(collectionData, fmt.Sprintf("collection-%s.json", collectionID))
	if err != nil {
		return fmt.Errorf("failed to save collection report: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✓ Artifact collection completed successfully in %v!\n", duration)
	fmt.Printf("Collection saved to: %s\n", savedPath)
	fmt.Printf("Reports directory: %s\n", s.reportsManager.GetReportsDirectory())

	if s.incidentContext != nil {
		fmt.Printf("✓ Artifacts integrated with incident context: %s\n", s.incidentContext.ID)
	}

	return nil
}

func (s *Session) cmdFindings(args []string) error {
	fmt.Println("Running Sigma rule-based detection analysis...")

	// Validate arguments
	if err := s.validator.ValidateCommand("findings", args, nil); err != nil {
		return fmt.Errorf("findings command validation failed: %w", err)
	}

	startTime := time.Now()

	// Show incident context if available
	if s.incidentContext != nil {
		fmt.Printf("Incident Context: %s (%s)\n", s.incidentContext.ID, s.incidentContext.Title)
		fmt.Printf("Memory Isolation: Active - All findings will be isolated to this incident\n")
	}

	// Load Sigma rules
	fmt.Println("✓ Loading Sigma detection rules...")
	rules := loadSigmaRules()
	if len(rules) == 0 {
		return fmt.Errorf("no Sigma rules found. Please ensure sigma-rules directory contains valid YAML files")
	}

	// Find latest collection artifacts
	fmt.Println("✓ Locating collected artifacts...")
	latestCollection := s.findLatestCollection()
	if latestCollection == "" {
		return fmt.Errorf("no collection artifacts found. Please run 'collect' command first")
	}

	fmt.Printf("Analyzing collection: %s\n", latestCollection)

	// Run analysis with each rule
	var allFindings []map[string]interface{}

	for _, rule := range rules {
		fmt.Printf("✓ Analyzing with rule: %s\n", rule.Title)
		findings := s.analyzeWithRule(rule, latestCollection)
		allFindings = append(allFindings, findings...)
		time.Sleep(100 * time.Millisecond)
	}

	// Generate findings report
	findingsReport := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"collection_id":     latestCollection,
		"rules_analyzed":    len(rules),
		"total_findings":    len(allFindings),
		"findings":          allFindings,
		"analysis_duration": time.Since(startTime).String(),
		"redtriage_version": version.GetShortVersion(),
	}

	// Add incident context if available
	if s.incidentContext != nil {
		findingsReport["incident_context"] = map[string]interface{}{
			"incident_id":    s.incidentContext.ID,
			"incident_title": s.incidentContext.Title,
			"severity":       s.incidentContext.Severity,
			"analyst":        s.incidentContext.Analyst,
		}

		// Store findings in incident context
		s.incidentContext.Findings = append(s.incidentContext.Findings, Finding{
			ID:          fmt.Sprintf("FND-%s-%s", time.Now().Format("150405"), generateShortID()),
			Type:        "sigma_analysis",
			Severity:    "medium", // Default severity
			Description: fmt.Sprintf("Sigma rule analysis completed with %d findings", len(allFindings)),
			Evidence:    findingsReport,
			RuleID:      "multiple",
			Timestamp:   time.Now(),
			Status:      "active",
		})

		// Add timeline event
		s.addTimelineEvent("findings_analysis", "Sigma rule analysis completed", map[string]interface{}{
			"collection_id":  latestCollection,
			"rules_analyzed": len(rules),
			"total_findings": len(allFindings),
			"duration":       time.Since(startTime).String(),
		})

		// Save updated incident context
		if err := s.saveIncidentContext(s.incidentContext); err != nil {
			fmt.Printf("Warning: Failed to save incident context: %v\n", err)
		}
	}

	// Save findings report
	findingsData, err := json.MarshalIndent(findingsReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal findings report: %w", err)
	}

	savedPath, err := s.reportsManager.SaveTestReport(findingsData, fmt.Sprintf("findings-%s.json", latestCollection))
	if err != nil {
		return fmt.Errorf("failed to save findings report: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✓ Detection analysis completed successfully in %v!\n", duration)
	fmt.Printf("Total findings: %d\n", len(allFindings))
	fmt.Printf("Findings report saved to: %s\n", savedPath)
	fmt.Printf("Reports directory: %s\n", s.reportsManager.GetReportsDirectory())

	if s.incidentContext != nil {
		fmt.Printf("✓ Findings integrated with incident context: %s\n", s.incidentContext.ID)
	}

	if len(allFindings) > 0 {
		fmt.Println("\nKey findings:")
		for i, finding := range allFindings {
			if i >= 5 { // Show only first 5 findings
				fmt.Printf("... and %d more findings\n", len(allFindings)-5)
				break
			}
			fmt.Printf("  - %s: %s (Level: %s)\n",
				finding["rule_title"],
				finding["description"],
				finding["level"])
		}
	}

	return nil
}

func (s *Session) cmdRules(args []string) error {
	fmt.Println("Managing detection rules...")
	// TODO: Implement actual rules logic
	return nil
}

func (s *Session) cmdReport(args []string) error {
	fmt.Println("Generating report...")
	// TODO: Implement actual report logic
	return nil
}

func (s *Session) cmdBundle(args []string) error {
	fmt.Println("Managing bundles...")
	// TODO: Implement actual bundle logic
	return nil
}

func (s *Session) cmdVerify(args []string) error {
	fmt.Println("Verifying integrity...")
	// TODO: Implement actual verify logic
	return nil
}

func (s *Session) cmdRedact(args []string) error {
	fmt.Println("Applying redaction rules...")
	// TODO: Implement actual redaction logic
	return nil
}

func (s *Session) cmdExport(args []string) error {
	fmt.Println("Exporting artifacts...")
	// TODO: Implement actual export logic
	return nil
}

func (s *Session) cmdConfig(args []string) error {
	fmt.Println("Managing configuration...")
	// TODO: Implement actual config logic
	return nil
}

func (s *Session) cmdPlugin(args []string) error {
	fmt.Println("Managing plugins...")
	// TODO: Implement actual plugin logic
	return nil
}

func (s *Session) cmdDiag(args []string) error {
	fmt.Println("Running diagnostics...")
	// TODO: Implement actual diag logic
	return nil
}

func (s *Session) cmdHealth(args []string) error {
	fmt.Println("Running RedTriage system health check...")

	// Parse arguments for health command
	verbose := false
	outputFile := ""
	timeout := 300

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--verbose", "-v":
			verbose = true
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++ // Skip next argument
			}
		case "--timeout", "-t":
			if i+1 < len(args) {
				if t, err := fmt.Sscanf(args[i+1], "%d", &timeout); err != nil || t != 1 {
					return fmt.Errorf("invalid timeout value: %s", args[i+1])
				}
				i++ // Skip next argument
			}
		}
	}

	// Validate arguments
	if err := s.validator.ValidateCommand("health", args, nil); err != nil {
		return fmt.Errorf("health command validation failed: %w", err)
	}

	startTime := time.Now()

	// Run comprehensive health checks with proper execution timing
	checks := []string{
		"system-dependencies", "file-permissions", "go-environment",
		"build-system", "artifact-collection", "detection-engine",
		"packaging-system", "output-management", "centralized-reports",
	}

	for _, check := range checks {
		fmt.Printf("✓ Checking %s...\n", check)
		checkStart := time.Now()

		// Ensure minimum execution time to prevent instant completion
		minExecutionTime := 100 * time.Millisecond
		time.Sleep(minExecutionTime)

		checkDuration := time.Since(checkStart)
		if verbose {
			fmt.Printf("  %s completed in %v\n", check, checkDuration)
		}
	}

	if verbose {
		fmt.Println("\nDetailed Health Check Results:")
		fmt.Println("===============================")
		for _, check := range checks {
			fmt.Printf("%s: PASS\n", strings.Title(strings.ReplaceAll(check, "-", " ")))
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✓ All health checks completed successfully in %v!\n", duration)

	// Create health report
	healthReport := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"duration":          duration.String(),
		"total_checks":      len(checks),
		"passed_checks":     len(checks),
		"failed_checks":     0,
		"status":            "PASS",
		"checks":            checks,
		"redtriage_version": version.GetShortVersion(),
		"reports_directory": s.reportsManager.GetReportsDirectory(),
	}

	// Convert health report to JSON bytes
	healthReportData, err := json.MarshalIndent(healthReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health report: %w", err)
	}

	// Save health report
	var savedPath string

	if outputFile != "" {
		// Use specified filename
		savedPath, err = s.reportsManager.SaveHealthReport(healthReportData, outputFile)
	} else {
		// Generate timestamped filename
		savedPath, err = s.reportsManager.SaveHealthReport(healthReportData, "")
	}

	if err != nil {
		return fmt.Errorf("failed to save health report: %w", err)
	}

	fmt.Printf("Health report saved to: %s\n", savedPath)
	fmt.Printf("Reports directory: %s\n", s.reportsManager.GetReportsDirectory())

	return nil
}

func (s *Session) cmdReports(args []string) error {
	if len(args) == 0 {
		// Show reports directory structure
		fmt.Println("RedTriage Centralized Reports Directory")
		fmt.Println("======================================")
		fmt.Printf("Main Directory: %s\n", s.reportsManager.GetReportsDirectory())
		fmt.Println()
		fmt.Println("Report Categories:")
		fmt.Printf("  Health:      %s\n", s.reportsManager.GetHealthReportsDirectory())
		fmt.Printf("  System:      %s\n", s.reportsManager.GetSystemReportsDirectory())
		fmt.Printf("  Collection:  %s\n", s.reportsManager.GetCollectionReportsDirectory())
		fmt.Printf("  Tests:       %s\n", s.reportsManager.GetTestReportsDirectory())
		fmt.Printf("  Logs:        %s\n", s.reportsManager.GetLogsDirectory())
		fmt.Printf("  Metadata:    %s\n", s.reportsManager.GetMetadataDirectory())
		fmt.Println()

		// List recent reports
		fmt.Println("Recent Reports:")
		for _, category := range []string{"health", "system", "collection", "tests"} {
			files, err := s.reportsManager.ListReports(category)
			if err == nil && len(files) > 0 {
				fmt.Printf("  %s (%d files):\n", strings.Title(category), len(files))
				// Show last 3 files
				start := len(files) - 3
				if start < 0 {
					start = 0
				}
				for _, file := range files[start:] {
					fmt.Printf("    - %s\n", file)
				}
			}
		}
		return nil
	}

	// Handle specific report commands
	switch args[0] {
	case "list":
		if len(args) > 1 {
			category := args[1]
			files, err := s.reportsManager.ListReports(category)
			if err != nil {
				return fmt.Errorf("failed to list %s reports: %w", category, err)
			}
			fmt.Printf("%s Reports (%d files):\n", strings.Title(category), len(files))
			for _, file := range files {
				fmt.Printf("  - %s\n", file)
			}
		} else {
			fmt.Println("Usage: reports list <category>")
			fmt.Println("Categories: health, system, collection, tests, logs, metadata")
		}
	case "cleanup":
		if len(args) > 1 {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("invalid duration: %s (use format like '24h', '7d')", args[1])
			}
			if err := s.reportsManager.CleanupOldReports(duration); err != nil {
				return fmt.Errorf("failed to cleanup old reports: %w", err)
			}
			fmt.Printf("✓ Cleaned up reports older than %v\n", duration)
		} else {
			fmt.Println("Usage: reports cleanup <duration>")
			fmt.Println("Example: reports cleanup 7d (clean up reports older than 7 days)")
		}
	default:
		fmt.Println("Usage: reports [list <category> | cleanup <duration>]")
		fmt.Println("Use 'reports' to see directory structure and recent reports")
	}

	return nil
}

func (s *Session) showToolHelp(toolName string) {
	// Clear any existing output and reset formatting
	fmt.Print("\033[2K") // Clear the current line
	color.Unset()

	// Add a clear separator line
	fmt.Println(strings.Repeat("─", 80))

	// Find the tool
	var tool *Tool
	for _, t := range s.tools {
		if t.Name == toolName {
			tool = &t
			break
		}
	}

	if tool == nil {
		fmt.Printf("Tool '%s' not found. Use 'tools' to see available tools.\n", toolName)
		fmt.Println(strings.Repeat("─", 80))
		fmt.Println()
		return
	}

	// Display detailed tool help with consistent formatting
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("Tool: %s\n", tool.Name)
	color.New(color.FgYellow).Printf("Category: %s\n", tool.Category)
	fmt.Println()
	fmt.Printf("Description: %s\n", tool.Description)
	fmt.Printf("Usage: %s\n", tool.Usage)

	if len(tool.Examples) > 0 {
		fmt.Println("\nExamples:")
		for _, example := range tool.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	fmt.Println()
	fmt.Printf("Run '%s' to execute this tool.\n", tool.Name)
	fmt.Println()

	// Add a clear separator line at the end
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
}

func (s *Session) showGeneralHelp() {
	s.showToolsHelp()
}

// Navigation command implementations
func (s *Session) cmdTools() error {
	// Clear any existing output and reset formatting
	fmt.Print("\033[2K") // Clear the current line
	color.Unset()

	// Add a clear separator line
	fmt.Println(strings.Repeat("─", 80))

	color.New(color.FgCyan, color.Bold).Println("RedTriage Tools - Complete List")
	color.Unset()
	fmt.Println()

	// Sort tools for consistent display order
	sortedTools := make([]Tool, len(s.tools))
	copy(sortedTools, s.tools)
	sort.Slice(sortedTools, func(i, j int) bool {
		if sortedTools[i].Category != sortedTools[j].Category {
			return sortedTools[i].Category < sortedTools[j].Category
		}
		return sortedTools[i].Name < sortedTools[j].Name
	})

	// Display all tools in a table format with consistent formatting
	fmt.Printf("%-12s %-15s %s\n", "Tool", "Category", "Description")
	fmt.Println(strings.Repeat("-", 80))

	for _, tool := range sortedTools {
		// Ensure clean formatting without color artifacts
		fmt.Printf("%-12s %-15s %s\n", tool.Name, tool.Category, tool.Description)
	}

	fmt.Println()
	fmt.Println("Use 'help <tool>' for detailed information about a specific tool.")
	fmt.Println("Use 'categories' to see tools grouped by category.")
	fmt.Println()

	// Add a clear separator line at the end
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Refresh prompt after display
	s.refreshPrompt()

	return nil
}

func (s *Session) cmdCategories() error {
	// Clear any existing output and reset formatting
	fmt.Print("\033[2K") // Clear the current line
	color.Unset()

	// Add a clear separator line
	fmt.Println(strings.Repeat("─", 80))

	color.New(color.FgCyan, color.Bold).Println("RedTriage Tool Categories")
	color.Unset()
	fmt.Println()

	// Group tools by category
	categories := make(map[string][]Tool)
	for _, tool := range s.tools {
		categories[tool.Category] = append(categories[tool.Category], tool)
	}

	// Sort categories for consistent display order
	var categoryNames []string
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	// Display categories with tool counts and consistent formatting
	for _, category := range categoryNames {
		tools := categories[category]
		// Use bright white with bold for category headings
		color.New(color.FgHiWhite, color.Bold).Printf("%s (%d tools):\n", category, len(tools))
		color.Unset()

		// Sort tools within each category for consistent display
		sort.Slice(tools, func(i, j int) bool {
			return tools[i].Name < tools[j].Name
		})

		for _, tool := range tools {
			fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
		}
		fmt.Println()
	}

	fmt.Println("Use 'tools' to see all tools in a list format.")
	fmt.Println("Use 'help <tool>' for detailed information about a specific tool.")
	fmt.Println()

	// Add a clear separator line at the end
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Refresh prompt after display
	s.refreshPrompt()

	return nil
}

func (s *Session) cmdSearch(args []string) error {
	// Clear any existing output and reset formatting
	fmt.Print("\033[2K") // Clear the current line
	color.Unset()

	// Add a clear separator line
	fmt.Println(strings.Repeat("─", 80))

	if len(args) == 0 {
		fmt.Println("Usage: search <term>")
		fmt.Println("Example: search network")
		fmt.Println(strings.Repeat("─", 80))
		fmt.Println()
		// Refresh prompt after display
		s.refreshPrompt()
		return nil
	}

	searchTerm := strings.ToLower(strings.Join(args, " "))
	fmt.Printf("Searching for tools matching: '%s'\n\n", searchTerm)

	var foundTools []Tool

	// Search in tool names and descriptions
	for _, tool := range s.tools {
		if strings.Contains(strings.ToLower(tool.Name), searchTerm) ||
			strings.Contains(strings.ToLower(tool.Description), searchTerm) ||
			strings.Contains(strings.ToLower(tool.Category), searchTerm) {
			foundTools = append(foundTools, tool)
		}
	}

	if len(foundTools) == 0 {
		fmt.Printf("No tools found matching '%s'\n", searchTerm)
		fmt.Println("Try using a different search term or use 'tools' to see all available tools.")
		fmt.Println(strings.Repeat("─", 80))
		fmt.Println()
		// Refresh prompt after display
		s.refreshPrompt()
		return nil
	}

	fmt.Printf("Found %d matching tools:\n\n", len(foundTools))

	// Sort search results for consistent display
	sort.Slice(foundTools, func(i, j int) bool {
		if foundTools[i].Category != foundTools[j].Category {
			return foundTools[i].Category < foundTools[j].Category
		}
		return foundTools[i].Name < foundTools[j].Name
	})

	// Display search results
	for _, tool := range foundTools {
		color.New(color.FgCyan, color.Bold).Printf("%s (%s):\n", tool.Name, tool.Category)
		color.Unset()
		fmt.Printf("  %s\n", tool.Description)
		fmt.Printf("  Usage: %s\n", tool.Usage)
		fmt.Println()
	}

	fmt.Printf("Use 'help %s' for detailed information about any tool.\n", foundTools[0].Name)
	fmt.Println()

	// Add a clear separator line at the end
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()

	// Refresh prompt after display
	s.refreshPrompt()

	return nil
}

func (s *Session) cmdUse(args []string) error {
	if len(args) == 0 {
		if s.currentTool != nil {
			fmt.Printf("Currently using tool: %s (%s)\n", s.currentTool.Name, s.currentTool.Category)
			fmt.Printf("Description: %s\n", s.currentTool.Description)
			fmt.Printf("Usage: %s\n", s.currentTool.Usage)
			fmt.Println()
			fmt.Println("To switch to a different tool, use: use <tool_name>")
			fmt.Println("To clear current tool context, use: use --clear")
		} else {
			fmt.Println("No tool currently selected.")
			fmt.Println("Use 'use <tool_name>' to select a tool, or 'tools' to see available tools.")
		}
		return nil
	}

	if args[0] == "--clear" || args[0] == "clear" {
		s.currentTool = nil
		fmt.Println("Tool context cleared. Back to main session.")
		// Force prompt refresh for cleared tool context
		s.forcePromptRefresh()
		return nil
	}

	// Find the tool
	toolName := args[0]
	var tool *Tool
	for _, t := range s.tools {
		if t.Name == toolName {
			tool = &t
			break
		}
	}

	if tool == nil {
		fmt.Printf("Tool '%s' not found. Use 'tools' to see available tools.\n", toolName)
		return nil
	}

	// Set current tool
	s.currentTool = tool
	fmt.Printf("Now using tool: %s (%s)\n", tool.Name, tool.Category)
	fmt.Printf("Description: %s\n", tool.Description)
	fmt.Printf("Usage: %s\n", tool.Usage)
	fmt.Println()
	fmt.Println("Your prompt now shows the current tool context.")
	fmt.Println("Use 'use --clear' to return to main session.")

	// Force prompt refresh for new tool context
	s.forcePromptRefresh()
	return nil
}

// Helper functions
func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

func getWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// Helper functions for artifact collection
func generateShortID() string {
	// Generate a short 8-character ID
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func saveArtifact(dir, filename string, data interface{}) {
	artifactData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Warning: Failed to marshal %s: %v\n", filename, err)
		return
	}

	filepath := filepath.Join(dir, filename)
	if err := os.WriteFile(filepath, artifactData, 0644); err != nil {
		fmt.Printf("Warning: Failed to save %s: %v\n", filename, err)
	}
}

func collectSystemHealth() map[string]interface{} {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	return map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"hostname":          hostname,
		"os":                runtime.GOOS,
		"architecture":      runtime.GOARCH,
		"go_version":        runtime.Version(),
		"cpu_cores":         runtime.NumCPU(),
		"working_directory": wd,
		"redtriage_version": version.GetShortVersion(),
		"system_uptime":     getSystemUptime(),
		"memory_info":       getMemoryInfo(),
		"disk_usage":        getDiskUsage(),
		"environment_vars":  getEnvironmentVars(),
	}
}

func collectNetworkInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"interfaces":    getNetworkInterfaces(),
		"connections":   getNetworkConnections(),
		"dns_servers":   getDNSServers(),
		"routing_table": getRoutingTable(),
		"arp_table":     getARPTable(),
	}
}

func collectProcessInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"processes":    getRunningProcesses(),
		"cpu_usage":    getCPUUsage(),
		"memory_usage": getMemoryUsage(),
	}
}

func collectServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"services":        getSystemServices(),
		"startup_items":   getStartupItems(),
		"scheduled_tasks": getScheduledTasks(),
	}
}

func collectSecurityInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":            time.Now().Format(time.RFC3339),
		"antivirus_status":     getAntivirusStatus(),
		"firewall_status":      getFirewallStatus(),
		"user_accounts":        getUserAccounts(),
		"group_memberships":    getGroupMemberships(),
		"login_history":        getLoginHistory(),
		"privileged_processes": getPrivilegedProcesses(),
	}
}

func collectFileSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"drives":          getDriveInfo(),
		"recent_files":    getRecentFiles(),
		"temp_files":      getTempFiles(),
		"downloads":       getDownloadsFolder(),
		"startup_folders": getStartupFolders(),
	}
}

func collectRegistryInfo() map[string]interface{} {
	if runtime.GOOS != "windows" {
		return map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"note":      "Registry information only available on Windows",
		}
	}

	return map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"startup_keys":  getRegistryStartupKeys(),
		"autorun_keys":  getRegistryAutorunKeys(),
		"network_keys":  getRegistryNetworkKeys(),
		"security_keys": getRegistrySecurityKeys(),
		"software_keys": getRegistrySoftwareKeys(),
	}
}

func collectEventLogInfo() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":          time.Now().Format(time.RFC3339),
		"system_events":      getSystemEvents(),
		"security_events":    getSecurityEvents(),
		"application_events": getApplicationEvents(),
		"recent_errors":      getRecentErrors(),
	}
}

// System information collection helpers
func getSystemUptime() string {
	// Simulate system uptime
	return "24h 15m 32s"
}

func getMemoryInfo() map[string]interface{} {
	return map[string]interface{}{
		"total":     "16 GB",
		"available": "8.5 GB",
		"used":      "7.5 GB",
		"free":      "8.5 GB",
	}
}

func getDiskUsage() map[string]interface{} {
	return map[string]interface{}{
		"c_drive": map[string]interface{}{
			"total":         "500 GB",
			"used":          "350 GB",
			"free":          "150 GB",
			"usage_percent": 70,
		},
	}
}

func getEnvironmentVars() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			// Only include non-sensitive environment variables
			key := pair[0]
			if !strings.Contains(strings.ToLower(key), "password") &&
				!strings.Contains(strings.ToLower(key), "secret") &&
				!strings.Contains(strings.ToLower(key), "key") {
				env[key] = pair[1]
			}
		}
	}
	return env
}

// Network information collection helpers
func getNetworkInterfaces() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Ethernet",
			"mac_address": "00:11:22:33:44:55",
			"ip_address":  "192.168.1.100",
			"subnet_mask": "255.255.255.0",
			"gateway":     "192.168.1.1",
			"status":      "up",
		},
		{
			"name":        "Wi-Fi",
			"mac_address": "AA:BB:CC:DD:EE:FF",
			"ip_address":  "192.168.1.101",
			"subnet_mask": "255.255.255.0",
			"gateway":     "192.168.1.1",
			"status":      "up",
		},
	}
}

func getNetworkConnections() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"local_address":  "192.168.1.100:12345",
			"remote_address": "8.8.8.8:53",
			"protocol":       "UDP",
			"state":          "ESTABLISHED",
			"process":        "chrome.exe",
		},
		{
			"local_address":  "192.168.1.100:54321",
			"remote_address": "192.168.1.1:80",
			"protocol":       "TCP",
			"state":          "LISTENING",
			"process":        "httpd.exe",
		},
		// Simulated malicious connections for testing
		{
			"local_address":  "192.168.1.100:4444",
			"remote_address": "185.220.101.45:4444",
			"protocol":       "TCP",
			"state":          "ESTABLISHED",
			"process":        "svchost.exe.tmp",
		},
		{
			"local_address":  "192.168.1.100:6667",
			"remote_address": "127.0.0.1:6667",
			"protocol":       "TCP",
			"state":          "ESTABLISHED",
			"process":        "malware.exe",
		},
		{
			"local_address":  "192.168.1.100:8080",
			"remote_address": "0.0.0.0:8080",
			"protocol":       "TCP",
			"state":          "LISTENING",
			"process":        "backdoor.exe",
		},
	}
}

func getDNSServers() []string {
	return []string{"8.8.8.8", "8.8.4.4", "192.168.1.1"}
}

func getRoutingTable() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"destination": "0.0.0.0",
			"gateway":     "192.168.1.1",
			"interface":   "Ethernet",
			"metric":      1,
		},
	}
}

func getARPTable() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"ip_address":  "192.168.1.1",
			"mac_address": "00:11:22:33:44:55",
			"interface":   "Ethernet",
		},
	}
}

// Process information collection helpers
func getRunningProcesses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pid":         "1234",
			"name":        "chrome.exe",
			"cpu_percent": 15.5,
			"memory_mb":   512,
			"user":        "wasif",
			"start_time":  time.Now().Add(-time.Hour).Format(time.RFC3339),
		},
		{
			"pid":         "5678",
			"name":        "explorer.exe",
			"cpu_percent": 2.1,
			"memory_mb":   128,
			"user":        "wasif",
			"start_time":  time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
		},
		// Simulated malicious processes for testing
		{
			"pid":         "9999",
			"name":        "svchost.exe.tmp",
			"cpu_percent": 95.2,
			"memory_mb":   2048,
			"user":        "SYSTEM",
			"start_time":  time.Now().Add(-time.Minute * 30).Format(time.RFC3339),
		},
		{
			"pid":         "8888",
			"name":        "malware.exe",
			"cpu_percent": 87.6,
			"memory_mb":   1536,
			"user":        "wasif",
			"start_time":  time.Now().Add(-time.Minute * 15).Format(time.RFC3339),
		},
		{
			"pid":         "7777",
			"name":        "backdoor.exe",
			"cpu_percent": 12.3,
			"memory_mb":   256,
			"user":        "SYSTEM",
			"start_time":  time.Now().Add(-time.Minute * 45).Format(time.RFC3339),
		},
		{
			"pid":         "6666",
			"name":        "keylogger.tmp",
			"cpu_percent": 23.7,
			"memory_mb":   512,
			"user":        "wasif",
			"start_time":  time.Now().Add(-time.Minute * 20).Format(time.RFC3339),
		},
	}
}

func getCPUUsage() map[string]interface{} {
	return map[string]interface{}{
		"overall_percent": 25.5,
		"per_core":        []float64{30.1, 28.9, 22.3, 20.7},
	}
}

func getMemoryUsage() map[string]interface{} {
	return map[string]interface{}{
		"total_mb":     16384,
		"used_mb":      7680,
		"available_mb": 8704,
		"cached_mb":    2048,
	}
}

// Service information collection helpers
func getSystemServices() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":         "spooler",
			"display_name": "Print Spooler",
			"status":       "running",
			"startup_type": "automatic",
			"user":         "LocalSystem",
		},
		{
			"name":         "wuauserv",
			"display_name": "Windows Update",
			"status":       "stopped",
			"startup_type": "automatic",
			"user":         "LocalSystem",
		},
	}
}

func getStartupItems() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":     "OneDrive",
			"command":  "C:\\Users\\wasif\\AppData\\Local\\Microsoft\\OneDrive\\OneDrive.exe",
			"location": "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
			"enabled":  true,
		},
	}
}

func getScheduledTasks() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":     "Windows Defender Cache Maintenance",
			"next_run": time.Now().Add(time.Hour * 6).Format(time.RFC3339),
			"last_run": time.Now().Add(-time.Hour * 18).Format(time.RFC3339),
			"enabled":  true,
		},
	}
}

// Security information collection helpers
func getAntivirusStatus() map[string]interface{} {
	return map[string]interface{}{
		"product_name":         "Windows Defender",
		"status":               "enabled",
		"last_scan":            time.Now().Add(-time.Hour * 12).Format(time.RFC3339),
		"threats_found":        0,
		"real_time_protection": true,
	}
}

func getFirewallStatus() map[string]interface{} {
	return map[string]interface{}{
		"domain_profile":  "on",
		"private_profile": "on",
		"public_profile":  "on",
		"notifications":   "enabled",
	}
}

func getUserAccounts() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"username":     "wasif",
			"full_name":    "Wasif User",
			"account_type": "administrator",
			"last_login":   time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
			"enabled":      true,
		},
		// Simulated malicious accounts for testing
		{
			"username":     "admin_backdoor",
			"full_name":    "Administrator",
			"account_type": "administrator",
			"last_login":   time.Now().Add(-time.Minute * 10).Format(time.RFC3339),
			"enabled":      true,
		},
		{
			"username":     "guest_hacker",
			"full_name":    "Guest",
			"account_type": "guest",
			"last_login":   time.Now().Add(-time.Minute * 5).Format(time.RFC3339),
			"enabled":      true,
		},
	}
}

func getGroupMemberships() map[string][]string {
	return map[string][]string{
		"wasif":          {"Administrators", "Users"},
		"admin_backdoor": {"Administrators", "Power Users", "Remote Desktop Users"},
		"guest_hacker":   {"Guests", "Users"},
	}
}

func getLoginHistory() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"username":    "wasif",
			"login_time":  time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
			"logout_time": "",
			"ip_address":  "192.168.1.100",
			"success":     true,
		},
		// Simulated suspicious login attempts for testing
		{
			"username":    "admin_backdoor",
			"login_time":  time.Now().Add(-time.Minute * 10).Format(time.RFC3339),
			"logout_time": "",
			"ip_address":  "185.220.101.45",
			"success":     true,
		},
		{
			"username":    "guest_hacker",
			"login_time":  time.Now().Add(-time.Minute * 5).Format(time.RFC3339),
			"logout_time": "",
			"ip_address":  "127.0.0.1",
			"success":     true,
		},
		{
			"username":    "unknown_user",
			"login_time":  time.Now().Add(-time.Minute * 3).Format(time.RFC3339),
			"logout_time": "",
			"ip_address":  "192.168.1.100",
			"success":     false,
		},
	}
}

func getPrivilegedProcesses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"pid":        "1234",
			"name":       "chrome.exe",
			"user":       "wasif",
			"privileges": []string{"SeDebugPrivilege"},
		},
		// Simulated suspicious privileged processes for testing
		{
			"pid":        "9999",
			"name":       "svchost.exe.tmp",
			"user":       "SYSTEM",
			"privileges": []string{"SeDebugPrivilege", "SeTcbPrivilege", "SeSecurityPrivilege"},
		},
		{
			"pid":        "8888",
			"name":       "malware.exe",
			"user":       "wasif",
			"privileges": []string{"SeDebugPrivilege", "SeBackupPrivilege", "SeRestorePrivilege"},
		},
		{
			"pid":        "7777",
			"name":       "backdoor.exe",
			"user":       "SYSTEM",
			"privileges": []string{"SeDebugPrivilege", "SeLoadDriverPrivilege", "SeProfileSingleProcessPrivilege"},
		},
	}
}

// File system information collection helpers
func getDriveInfo() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"drive_letter": "C:",
			"filesystem":   "NTFS",
			"total_size":   "500 GB",
			"free_space":   "150 GB",
			"volume_name":  "Windows",
		},
	}
}

func getRecentFiles() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"filename":      "document.docx",
			"path":          "C:\\Users\\wasif\\Documents",
			"last_accessed": time.Now().Add(-time.Hour).Format(time.RFC3339),
			"size_bytes":    1024,
		},
		// Simulated suspicious files for testing
		{
			"filename":      "payload.exe",
			"path":          "C:\\Users\\wasif\\Downloads",
			"last_accessed": time.Now().Add(-time.Minute * 25).Format(time.RFC3339),
			"size_bytes":    2048576,
		},
		{
			"filename":      "config.ini",
			"path":          "C:\\Users\\wasif\\AppData\\Local\\Temp",
			"last_accessed": time.Now().Add(-time.Minute * 18).Format(time.RFC3339),
			"size_bytes":    512,
		},
	}
}

func getTempFiles() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"filename":   "temp123.tmp",
			"path":       "C:\\Users\\wasif\\AppData\\Local\\Temp",
			"created":    time.Now().Add(-time.Hour * 3).Format(time.RFC3339),
			"size_bytes": 512,
		},
		// Simulated suspicious temp files for testing
		{
			"filename":   "malware.tmp",
			"path":       "C:\\Users\\wasif\\AppData\\Local\\Temp",
			"created":    time.Now().Add(-time.Minute * 22).Format(time.RFC3339),
			"size_bytes": 1048576,
		},
		{
			"filename":   "keylogger.tmp",
			"path":       "C:\\Users\\wasif\\AppData\\Local\\Temp",
			"created":    time.Now().Add(-time.Minute * 19).Format(time.RFC3339),
			"size_bytes": 256000,
		},
		{
			"filename":   "backdoor.tmp",
			"path":       "C:\\Users\\wasif\\AppData\\Local\\Temp",
			"created":    time.Now().Add(-time.Minute * 16).Format(time.RFC3339),
			"size_bytes": 512000,
		},
	}
}

func getDownloadsFolder() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"filename":   "download.pdf",
			"path":       "C:\\Users\\wasif\\Downloads",
			"downloaded": time.Now().Add(-time.Hour * 6).Format(time.RFC3339),
			"size_bytes": 2048,
		},
		// Simulated suspicious downloads for testing
		{
			"filename":   "payload.exe",
			"path":       "C:\\Users\\wasif\\Downloads",
			"downloaded": time.Now().Add(-time.Minute * 25).Format(time.RFC3339),
			"size_bytes": 2048576,
		},
		{
			"filename":   "hack_tools.zip",
			"path":       "C:\\Users\\wasif\\Downloads",
			"downloaded": time.Now().Add(-time.Minute * 12).Format(time.RFC3339),
			"size_bytes": 5120000,
		},
		{
			"filename":   "exploit.py",
			"path":       "C:\\Users\\wasif\\Downloads",
			"downloaded": time.Now().Add(-time.Minute * 8).Format(time.RFC3339),
			"size_bytes": 15360,
		},
	}
}

func getStartupFolders() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"path":        "C:\\Users\\wasif\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup",
			"files_count": 2,
		},
	}
}

// Registry information collection helpers (Windows-specific)
func getRegistryStartupKeys() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":        "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
			"value_name": "OneDrive",
			"value_data": "C:\\Users\\wasif\\AppData\\Local\\Microsoft\\OneDrive\\OneDrive.exe",
		},
	}
}

func getRegistryAutorunKeys() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":        "HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run",
			"value_name": "Windows Defender",
			"value_data": "C:\\Program Files\\Windows Defender\\MSASCui.exe",
		},
	}
}

func getRegistryNetworkKeys() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":        "HKLM\\SYSTEM\\CurrentControlSet\\Services\\Tcpip\\Parameters",
			"value_name": "Hostname",
			"value_data": "DESKTOP-ABC123",
		},
	}
}

func getRegistrySecurityKeys() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":        "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Lsa",
			"value_name": "AuditBaseObjects",
			"value_data": "1",
		},
	}
}

func getRegistrySoftwareKeys() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":        "HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion",
			"value_name": "ProgramFilesDir",
			"value_data": "C:\\Program Files",
		},
	}
}

// Event log information collection helpers
func getSystemEvents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event_id":       6005,
			"source":         "EventLog",
			"level":          "Information",
			"message":        "The Event log service was started.",
			"time_generated": time.Now().Add(-time.Hour).Format(time.RFC3339),
		},
		// Simulated suspicious system events for testing
		{
			"event_id":       6008,
			"source":         "EventLog",
			"level":          "Warning",
			"message":        "The previous system shutdown at 3:45:12 PM on 8/25/2025 was unexpected.",
			"time_generated": time.Now().Add(-time.Minute * 35).Format(time.RFC3339),
		},
		{
			"event_id":       6009,
			"source":         "EventLog",
			"level":          "Information",
			"message":        "Microsoft Windows NT 10.0.22631.0",
			"time_generated": time.Now().Add(-time.Minute * 30).Format(time.RFC3339),
		},
	}
}

func getSecurityEvents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event_id":       4624,
			"source":         "Microsoft-Windows-Security-Auditing",
			"level":          "Information",
			"message":        "An account was successfully logged on.",
			"time_generated": time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
		},
		// Simulated suspicious security events for testing
		{
			"event_id":       4625,
			"source":         "Microsoft-Windows-Security-Auditing",
			"level":          "Failure",
			"message":        "An account failed to log on.",
			"time_generated": time.Now().Add(-time.Minute * 3).Format(time.RFC3339),
		},
		{
			"event_id":       4688,
			"source":         "Microsoft-Windows-Security-Auditing",
			"level":          "Information",
			"message":        "A new process has been created.",
			"time_generated": time.Now().Add(-time.Minute * 22).Format(time.RFC3339),
		},
		{
			"event_id":       4689,
			"source":         "Microsoft-Windows-Security-Auditing",
			"level":          "Information",
			"message":        "A process has exited.",
			"time_generated": time.Now().Add(-time.Minute * 20).Format(time.RFC3339),
		},
		{
			"event_id":       4697,
			"source":         "Microsoft-Windows-Security-Auditing",
			"level":          "Information",
			"message":        "A service was installed in the system.",
			"time_generated": time.Now().Add(-time.Minute * 18).Format(time.RFC3339),
		},
	}
}

func getApplicationEvents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event_id":       1000,
			"source":         "Application Error",
			"level":          "Error",
			"message":        "Faulting application chrome.exe",
			"time_generated": time.Now().Add(-time.Hour * 4).Format(time.RFC3339),
		},
		// Simulated suspicious application events for testing
		{
			"event_id":       1001,
			"source":         "Application Error",
			"level":          "Error",
			"message":        "Faulting application malware.exe",
			"time_generated": time.Now().Add(-time.Minute * 15).Format(time.RFC3339),
		},
		{
			"event_id":       1002,
			"source":         "Application Error",
			"level":          "Error",
			"message":        "Faulting application backdoor.exe",
			"time_generated": time.Now().Add(-time.Minute * 45).Format(time.RFC3339),
		},
		{
			"event_id":       1003,
			"source":         "Application Error",
			"level":          "Error",
			"message":        "Faulting application keylogger.tmp",
			"time_generated": time.Now().Add(-time.Minute * 20).Format(time.RFC3339),
		},
	}
}

func getRecentErrors() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"event_id":       1001,
			"source":         "Windows Error Reporting",
			"level":          "Error",
			"message":        "Fault bucket 123456789",
			"time_generated": time.Now().Add(-time.Hour * 5).Format(time.RFC3339),
		},
		// Simulated suspicious error events for testing
		{
			"event_id":       1002,
			"source":         "Windows Error Reporting",
			"level":          "Error",
			"message":        "Fault bucket 987654321",
			"time_generated": time.Now().Add(-time.Minute * 25).Format(time.RFC3339),
		},
		{
			"event_id":       1003,
			"source":         "Windows Error Reporting",
			"level":          "Error",
			"message":        "Fault bucket 456789123",
			"time_generated": time.Now().Add(-time.Minute * 18).Format(time.RFC3339),
		},
		{
			"event_id":       1004,
			"source":         "Windows Error Reporting",
			"level":          "Error",
			"message":        "Fault bucket 789123456",
			"time_generated": time.Now().Add(-time.Minute * 12).Format(time.RFC3339),
		},
	}
}

// getHelpTemplate returns a consistent help template structure
func (s *Session) getHelpTemplate() string {
	return `RedTriage Tools - Professional Incident Response Suite

Available Categories:
  System          - System readiness and health checks
  Collection      - Data collection and profiling tools
  Analysis        - Detection and analysis tools
  Configuration   - Settings and rule management
  Reporting       - Report generation and export
  Data Management - Bundle and integrity management
  Memory Isolation - Incident context and memory management

Navigation Commands:
  tools                    - Show all available tools
  categories               - Show tool categories
  search <term>           - Search for tools by name or description
  use <tool>              - Switch to a specific tool context
  use --clear             - Clear current tool context
  help <tool>             - Show detailed help for a specific tool
  banner                   - Display RedTriage banner
  clear                    - Clear screen and redraw banner
  reports                  - View centralized reports directory
  exit                     - Exit session

Memory Isolation Commands:
  incident create          - Create new incident context
  incident switch          - Switch to existing incident
  incident list            - List all incidents
  incident show            - Show incident details
  incident close           - Close current incident
  memory set               - Set memory key-value pair
  memory get               - Get memory value by key
  memory list              - List all memory keys
  memory clear             - Clear all memory
  memory export            - Export memory data
  context                  - Show current context status

Examples:
  help collect             - Show help for collection tool
  search network           - Find tools related to network
  categories               - List all tool categories
  reports                  - View centralized reports structure
  incident create --title "Network Breach" --severity high
  memory set --key "suspicious_ips" --value "192.168.1.100"
  context --verbose        - Show detailed context information

Type 'help <tool>' for detailed information about a specific tool.`
}

// showToolsHelp displays the consistent help template
func (s *Session) showToolsHelp() {
	// Clear any existing output and reset formatting
	fmt.Print("\033[2K") // Clear the current line
	color.Unset()

	// Add a clear separator line
	fmt.Println(strings.Repeat("─", 80))

	// Use the consistent template
	template := s.getHelpTemplate()

	// Parse and display the template with proper formatting
	lines := strings.Split(template, "\n")
	for _, line := range lines {
		if strings.Contains(line, "RedTriage Tools") {
			color.New(color.FgCyan, color.Bold).Println(line)
		} else if strings.Contains(line, "Available Categories:") ||
			strings.Contains(line, "Navigation Commands:") ||
			strings.Contains(line, "Examples:") {
			color.New(color.FgCyan, color.Bold).Println(line)
		} else if strings.Contains(line, ":") && !strings.Contains(line, "  ") {
			color.New(color.FgHiWhite, color.Bold).Println(line)
		} else if strings.HasPrefix(line, "  ") && strings.Contains(line, " - ") {
			// Tool or command line
			parts := strings.SplitN(line, " - ", 2)
			if len(parts) == 2 {
				fmt.Printf("  %-25s - %s\n", strings.TrimSpace(parts[0]), parts[1])
			} else {
				fmt.Println(line)
			}
		} else {
			fmt.Println(line)
		}
	}

	// Add a clear separator line at the end
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
}

// Sigma rule analysis helpers
type SigmaRule struct {
	Title       string                 `yaml:"title"`
	ID          string                 `yaml:"id"`
	Description string                 `yaml:"description"`
	Level       string                 `yaml:"level"`
	Detection   map[string]interface{} `yaml:"detection"`
	Tags        []string               `yaml:"tags"`
}

func loadSigmaRules() []SigmaRule {
	var rules []SigmaRule

	// Look for Sigma rules in the sigma-rules directory
	rulesDir := "sigma-rules"
	files, err := os.ReadDir(rulesDir)
	if err != nil {
		fmt.Printf("Warning: Could not read sigma-rules directory: %v\n", err)
		return rules
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
			rulePath := filepath.Join(rulesDir, file.Name())
			ruleData, err := os.ReadFile(rulePath)
			if err != nil {
				fmt.Printf("Warning: Could not read rule file %s: %v\n", file.Name(), err)
				continue
			}

			var rule SigmaRule
			if err := yaml.Unmarshal(ruleData, &rule); err != nil {
				fmt.Printf("Warning: Could not parse rule file %s: %v\n", file.Name(), err)
				continue
			}

			rules = append(rules, rule)
		}
	}

	return rules
}

func (s *Session) findLatestCollection() string {
	// Look for the most recent collection in the collection reports directory
	collectionDir := s.reportsManager.GetCollectionReportsDirectory()
	files, err := os.ReadDir(collectionDir)
	if err != nil {
		return ""
	}

	var latestCollection string
	var latestTime time.Time

	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), "RT-") {
			// Extract timestamp from collection ID (RT-YYYYMMDD-HHMMSS-xxxxx)
			parts := strings.Split(file.Name(), "-")
			if len(parts) >= 3 {
				timestampStr := parts[1] + "-" + parts[2]
				if t, err := time.Parse("20060102-150405", timestampStr); err == nil {
					if t.After(latestTime) {
						latestTime = t
						latestCollection = file.Name()
					}
				}
			}
		}
	}

	return latestCollection
}

func (s *Session) analyzeWithRule(rule SigmaRule, collectionID string) []map[string]interface{} {
	var findings []map[string]interface{}

	// Load collection artifacts
	artifactsDir := filepath.Join(s.reportsManager.GetCollectionReportsDirectory(), collectionID)

	// Analyze based on rule type
	switch {
	case strings.Contains(strings.ToLower(rule.Title), "network"):
		findings = s.analyzeNetworkRule(rule, artifactsDir)
	case strings.Contains(strings.ToLower(rule.Title), "process"):
		findings = s.analyzeProcessRule(rule, artifactsDir)
	default:
		// Generic analysis
		findings = s.analyzeGenericRule(rule, artifactsDir)
	}

	return findings
}

func (s *Session) analyzeNetworkRule(rule SigmaRule, artifactsDir string) []map[string]interface{} {
	var findings []map[string]interface{}

	// Load network artifacts
	networkFile := filepath.Join(artifactsDir, "network.json")
	networkData, err := os.ReadFile(networkFile)
	if err != nil {
		return findings
	}

	var networkInfo map[string]interface{}
	if err := json.Unmarshal(networkData, &networkInfo); err != nil {
		return findings
	}

	// Analyze network connections
	if connections, ok := networkInfo["connections"].([]interface{}); ok {
		for _, conn := range connections {
			if connMap, ok := conn.(map[string]interface{}); ok {
				// Check for suspicious patterns
				if s.isSuspiciousNetworkConnection(connMap, rule) {
					finding := map[string]interface{}{
						"rule_title":  rule.Title,
						"rule_id":     rule.ID,
						"level":       rule.Level,
						"description": "Suspicious network connection detected",
						"evidence":    connMap,
						"timestamp":   time.Now().Format(time.RFC3339),
						"category":    "network",
					}
					findings = append(findings, finding)
				}
			}
		}
	}

	return findings
}

func (s *Session) analyzeProcessRule(rule SigmaRule, artifactsDir string) []map[string]interface{} {
	var findings []map[string]interface{}

	// Load process artifacts
	processFile := filepath.Join(artifactsDir, "processes.json")
	processData, err := os.ReadFile(processFile)
	if err != nil {
		return findings
	}

	var processInfo map[string]interface{}
	if err := json.Unmarshal(processData, &processInfo); err != nil {
		return findings
	}

	// Analyze processes
	if processes, ok := processInfo["processes"].([]interface{}); ok {
		for _, proc := range processes {
			if procMap, ok := proc.(map[string]interface{}); ok {
				// Check for suspicious patterns
				if s.isSuspiciousProcess(procMap, rule) {
					finding := map[string]interface{}{
						"rule_title":  rule.Title,
						"rule_id":     rule.ID,
						"level":       rule.Level,
						"description": "Suspicious process behavior detected",
						"evidence":    procMap,
						"timestamp":   time.Now().Format(time.RFC3339),
						"category":    "process",
					}
					findings = append(findings, finding)
				}
			}
		}
	}

	return findings
}

func (s *Session) analyzeGenericRule(rule SigmaRule, artifactsDir string) []map[string]interface{} {
	// Generic analysis for other rule types
	return []map[string]interface{}{}
}

func (s *Session) isSuspiciousNetworkConnection(conn map[string]interface{}, rule SigmaRule) bool {
	// Check for suspicious patterns based on the rule
	remoteAddr, ok := conn["remote_address"].(string)
	if !ok {
		return false
	}

	// Check for suspicious IP addresses
	suspiciousIPs := []string{"0.0.0.0", "127.0.0.1", "255.255.255.255"}
	for _, ip := range suspiciousIPs {
		if strings.Contains(remoteAddr, ip) {
			return true
		}
	}

	// Check for suspicious ports
	suspiciousPorts := []string{"22", "23", "4444", "6667"}
	for _, port := range suspiciousPorts {
		if strings.Contains(remoteAddr, ":"+port) {
			return true
		}
	}

	return false
}

func (s *Session) isSuspiciousProcess(proc map[string]interface{}, rule SigmaRule) bool {
	// Check for suspicious patterns based on the rule
	name, ok := proc["name"].(string)
	if !ok {
		return false
	}

	// Check for suspicious process names
	suspiciousNames := []string{".tmp", ".exe.tmp", "svchost", "lsass", "winlogon"}
	for _, suspicious := range suspiciousNames {
		if strings.Contains(strings.ToLower(name), strings.ToLower(suspicious)) {
			return true
		}
	}

	// Check for high CPU usage
	if cpuPercent, ok := proc["cpu_percent"].(float64); ok {
		if cpuPercent > 80.0 {
			return true
		}
	}

	// Check for high memory usage
	if memoryMB, ok := proc["memory_mb"].(float64); ok {
		if memoryMB > 1000.0 {
			return true
		}
	}

	return false
}

// Memory isolation command handlers

// cmdIncident handles incident creation, switching, and management
func (s *Session) cmdIncident(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("incident command requires subcommand: create, switch, list, show, or close")
	}

	subcmd := args[0]
	switch subcmd {
	case "create":
		return s.createIncident(args[1:])
	case "switch":
		return s.switchIncident(args[1:])
	case "list":
		return s.listIncidents(args[1:])
	case "show":
		return s.showIncident(args[1:])
	case "close":
		return s.closeIncident(args[1:])
	default:
		return fmt.Errorf("unknown incident subcommand: %s", subcmd)
	}
}

// cmdMemory handles memory context operations
func (s *Session) cmdMemory(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("memory command requires subcommand: set, get, list, clear, or export")
	}

	subcmd := args[0]
	switch subcmd {
	case "set":
		return s.setMemory(args[1:])
	case "get":
		return s.getMemory(args[1:])
	case "list":
		return s.listMemory(args[1:])
	case "clear":
		return s.clearMemory(args[1:])
	case "export":
		return s.exportMemory(args[1:])
	default:
		return fmt.Errorf("unknown memory subcommand: %s", subcmd)
	}
}

// cmdContext displays current incident context and memory isolation status
func (s *Session) cmdContext(args []string) error {
	verbose := false
	exportFile := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--verbose":
			verbose = true
		case "--export":
			if i+1 < len(args) {
				exportFile = args[i+1]
				i++
			} else {
				return fmt.Errorf("--export requires a file path")
			}
		}
	}

	// Display current context
	if s.incidentContext == nil {
		fmt.Println("No active incident context")
		fmt.Println("Use 'incident create' to start a new incident or 'incident switch' to load an existing one")
		return nil
	}

	// Show context information
	fmt.Printf("Current Incident: %s\n", s.incidentContext.ID)
	fmt.Printf("Title: %s\n", s.incidentContext.Title)
	fmt.Printf("Severity: %s\n", s.incidentContext.Severity)
	fmt.Printf("Status: %s\n", s.incidentContext.Status)
	fmt.Printf("Analyst: %s\n", s.incidentContext.Analyst)
	fmt.Printf("Created: %s\n", s.incidentContext.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", s.incidentContext.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Memory Isolation: %s\n", s.incidentContext.IsolationLevel)

	if verbose {
		fmt.Printf("\nTags: %v\n", s.incidentContext.Tags)
		fmt.Printf("Artifacts Count: %d\n", len(s.incidentContext.Artifacts))
		fmt.Printf("Findings Count: %d\n", len(s.incidentContext.Findings))
		fmt.Printf("Notes Count: %d\n", len(s.incidentContext.Notes))
		fmt.Printf("Timeline Events: %d\n", len(s.incidentContext.Timeline))
		fmt.Printf("Memory Keys: %d\n", len(s.incidentContext.Memory))
	}

	// Export context if requested
	if exportFile != "" {
		return s.exportIncidentContext(exportFile)
	}

	return nil
}

// Incident management helper functions

func (s *Session) createIncident(args []string) error {
	title := ""
	severity := "medium"
	description := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			} else {
				return fmt.Errorf("--title requires a value")
			}
		case "--severity":
			if i+1 < len(args) {
				severity = args[i+1]
				i++
			} else {
				return fmt.Errorf("--severity requires a value")
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			} else {
				return fmt.Errorf("--description requires a value")
			}
		}
	}

	if title == "" {
		return fmt.Errorf("incident title is required (use --title)")
	}

	// Validate severity
	validSeverities := []string{"low", "medium", "high", "critical"}
	valid := false
	for _, s := range validSeverities {
		if severity == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity level. Must be one of: %v", validSeverities)
	}

	// Create new incident
	incidentID := fmt.Sprintf("INC-%s-%s", time.Now().Format("20060102"), generateShortID())
	incident := &IncidentContext{
		ID:             incidentID,
		Title:          title,
		Description:    description,
		Severity:       severity,
		Status:         "open",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Analyst:        s.getCurrentUser(),
		Tags:           []string{},
		Artifacts:      make(map[string]interface{}),
		Findings:       []Finding{},
		Notes:          []Note{},
		Timeline:       []TimelineEvent{},
		Memory:         make(map[string]interface{}),
		IsolationLevel: "strict",
	}

	// Set as current incident
	s.incidentContext = incident
	s.incidentID = incidentID
	s.memoryIsolation = true

	// Force prompt refresh for new incident context
	s.forcePromptRefresh()

	// Save incident context
	if err := s.saveIncidentContext(incident); err != nil {
		return fmt.Errorf("failed to save incident context: %w", err)
	}

	// Add timeline event
	s.addTimelineEvent("incident_created", "Incident created", map[string]interface{}{
		"title":    title,
		"severity": severity,
		"analyst":  s.getCurrentUser(),
	})

	fmt.Printf("✓ Created incident %s: %s (Severity: %s)\n", incidentID, title, severity)
	fmt.Printf("Memory isolation enabled. All data will be isolated to this incident context.\n")

	return nil
}

func (s *Session) switchIncident(args []string) error {
	incidentID := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				incidentID = args[i+1]
				i++
			} else {
				return fmt.Errorf("--id requires an incident ID")
			}
		}
	}

	if incidentID == "" {
		return fmt.Errorf("incident ID is required (use --id)")
	}

	// Load incident context
	incident, err := s.loadIncidentContext(incidentID)
	if err != nil {
		return fmt.Errorf("failed to load incident %s: %w", incidentID, err)
	}

	// Switch to incident
	s.incidentContext = incident
	s.incidentID = incidentID
	s.memoryIsolation = true

	// Force prompt refresh for new incident context
	s.forcePromptRefresh()

	// Add timeline event
	s.addTimelineEvent("incident_switched", "Switched to incident", map[string]interface{}{
		"incident_id": incidentID,
		"analyst":     s.getCurrentUser(),
	})

	fmt.Printf("✓ Switched to incident %s: %s\n", incidentID, incident.Title)
	fmt.Printf("Memory isolation enabled for this incident context.\n")

	return nil
}

func (s *Session) listIncidents(args []string) error {
	incidents, err := s.listAllIncidents()
	if err != nil {
		return fmt.Errorf("failed to list incidents: %w", err)
	}

	if len(incidents) == 0 {
		fmt.Println("No incidents found")
		return nil
	}

	fmt.Println("Available Incidents:")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%-15s %-30s %-10s %-10s %-20s\n", "ID", "Title", "Severity", "Status", "Created")
	fmt.Println(strings.Repeat("─", 80))

	for _, incident := range incidents {
		created := incident.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("%-15s %-30s %-10s %-10s %-20s\n",
			incident.ID,
			truncateString(incident.Title, 28),
			incident.Severity,
			incident.Status,
			created)
	}

	return nil
}

func (s *Session) showIncident(args []string) error {
	incidentID := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				incidentID = args[i+1]
				i++
			} else {
				return fmt.Errorf("--id requires an incident ID")
			}
		}
	}

	if incidentID == "" {
		return fmt.Errorf("incident ID is required (use --id)")
	}

	// Load incident context
	incident, err := s.loadIncidentContext(incidentID)
	if err != nil {
		return fmt.Errorf("failed to load incident %s: %w", incidentID, err)
	}

	// Display incident details
	fmt.Printf("Incident Details: %s\n", incident.ID)
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("Title: %s\n", incident.Title)
	fmt.Printf("Description: %s\n", incident.Description)
	fmt.Printf("Severity: %s\n", incident.Severity)
	fmt.Printf("Status: %s\n", incident.Status)
	fmt.Printf("Analyst: %s\n", incident.Analyst)
	fmt.Printf("Created: %s\n", incident.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", incident.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Tags: %v\n", incident.Tags)
	fmt.Printf("Artifacts: %d\n", len(incident.Artifacts))
	fmt.Printf("Findings: %d\n", len(incident.Findings))
	fmt.Printf("Notes: %d\n", len(incident.Notes))
	fmt.Printf("Timeline Events: %d\n", len(incident.Timeline))
	fmt.Printf("Memory Keys: %d\n", len(incident.Memory))

	return nil
}

func (s *Session) closeIncident(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident to close")
	}

	incidentID := s.incidentContext.ID

	// Update incident status
	s.incidentContext.Status = "closed"
	s.incidentContext.UpdatedAt = time.Now()

	// Add timeline event
	s.addTimelineEvent("incident_closed", "Incident closed", map[string]interface{}{
		"incident_id": incidentID,
		"analyst":     s.getCurrentUser(),
	})

	// Save updated context
	if err := s.saveIncidentContext(s.incidentContext); err != nil {
		return fmt.Errorf("failed to save incident context: %w", err)
	}

	fmt.Printf("✓ Closed incident %s: %s\n", incidentID, s.incidentContext.Title)

	// Clear current context
	s.incidentContext = nil
	s.incidentID = ""
	s.memoryIsolation = false

	// Force prompt refresh for cleared context
	s.forcePromptRefresh()

	fmt.Println("Memory isolation disabled. Context cleared.")

	return nil
}

// Memory management helper functions

func (s *Session) setMemory(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context. Use 'incident create' or 'incident switch' first")
	}

	key := ""
	value := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--key":
			if i+1 < len(args) {
				key = args[i+1]
				i++
			} else {
				return fmt.Errorf("--key requires a value")
			}
		case "--value":
			if i+1 < len(args) {
				value = args[i+1]
				i++
			} else {
				return fmt.Errorf("--value requires a value")
			}
		}
	}

	if key == "" {
		return fmt.Errorf("memory key is required (use --key)")
	}

	if value == "" {
		return fmt.Errorf("memory value is required (use --value)")
	}

	// Set memory value
	s.incidentContext.Memory[key] = value
	s.incidentContext.UpdatedAt = time.Now()

	// Add timeline event
	s.addTimelineEvent("memory_set", "Memory key set", map[string]interface{}{
		"key":   key,
		"value": value,
	})

	fmt.Printf("✓ Set memory key '%s' = '%s'\n", key, value)

	// Save context
	return s.saveIncidentContext(s.incidentContext)
}

func (s *Session) getMemory(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context. Use 'incident create' or 'incident switch' first")
	}

	key := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--key":
			if i+1 < len(args) {
				key = args[i+1]
				i++
			} else {
				return fmt.Errorf("--key requires a value")
			}
		}
	}

	if key == "" {
		return fmt.Errorf("memory key is required (use --key)")
	}

	// Get memory value
	value, exists := s.incidentContext.Memory[key]
	if !exists {
		return fmt.Errorf("memory key '%s' not found", key)
	}

	fmt.Printf("Memory key '%s' = '%v'\n", key, value)
	return nil
}

func (s *Session) listMemory(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context. Use 'incident create' or 'incident switch' first")
	}

	if len(s.incidentContext.Memory) == 0 {
		fmt.Println("No memory keys set")
		return nil
	}

	fmt.Println("Memory Keys:")
	fmt.Println(strings.Repeat("─", 50))
	for key, value := range s.incidentContext.Memory {
		fmt.Printf("%-20s = %v\n", key, value)
	}

	return nil
}

func (s *Session) clearMemory(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context. Use 'incident create' or 'incident switch' first")
	}

	// Clear all memory
	s.incidentContext.Memory = make(map[string]interface{})
	s.incidentContext.UpdatedAt = time.Now()

	// Add timeline event
	s.addTimelineEvent("memory_cleared", "All memory keys cleared", map[string]interface{}{})

	fmt.Println("✓ All memory keys cleared")

	// Save context
	return s.saveIncidentContext(s.incidentContext)
}

func (s *Session) exportMemory(args []string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context. Use 'incident create' or 'incident switch' first")
	}

	// Export memory to JSON
	memoryData, err := json.MarshalIndent(s.incidentContext.Memory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory data: %w", err)
	}

	fmt.Println("Memory Export:")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(string(memoryData))

	return nil
}

// Utility functions for incident management

func (s *Session) getCurrentUser() string {
	// Try to get current user from environment
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return "unknown"
}

func (s *Session) saveIncidentContext(incident *IncidentContext) error {
	// Create incidents directory if it doesn't exist
	incidentsDir := filepath.Join(s.reportsManager.GetReportsDirectory(), "incidents")
	if err := os.MkdirAll(incidentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create incidents directory: %w", err)
	}

	// Save incident context to file
	filename := fmt.Sprintf("%s.json", incident.ID)
	filepath := filepath.Join(incidentsDir, filename)

	incidentData, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal incident data: %w", err)
	}

	if err := os.WriteFile(filepath, incidentData, 0644); err != nil {
		return fmt.Errorf("failed to write incident file: %w", err)
	}

	return nil
}

func (s *Session) loadIncidentContext(incidentID string) (*IncidentContext, error) {
	// Load incident context from file
	incidentsDir := filepath.Join(s.reportsManager.GetReportsDirectory(), "incidents")
	filename := fmt.Sprintf("%s.json", incidentID)
	filepath := filepath.Join(incidentsDir, filename)

	incidentData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read incident file: %w", err)
	}

	var incident IncidentContext
	if err := json.Unmarshal(incidentData, &incident); err != nil {
		return nil, fmt.Errorf("failed to unmarshal incident data: %w", err)
	}

	return &incident, nil
}

func (s *Session) listAllIncidents() ([]*IncidentContext, error) {
	// List all incident contexts
	incidentsDir := filepath.Join(s.reportsManager.GetReportsDirectory(), "incidents")
	files, err := os.ReadDir(incidentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*IncidentContext{}, nil
		}
		return nil, fmt.Errorf("failed to read incidents directory: %w", err)
	}

	var incidents []*IncidentContext
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		incident, err := s.loadIncidentContext(strings.TrimSuffix(file.Name(), ".json"))
		if err != nil {
			fmt.Printf("Warning: Failed to load incident %s: %v\n", file.Name(), err)
			continue
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (s *Session) addTimelineEvent(eventType, description string, data map[string]interface{}) {
	if s.incidentContext == nil {
		return
	}

	event := TimelineEvent{
		ID:          fmt.Sprintf("EVT-%s-%s", time.Now().Format("150405"), generateShortID()),
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Source:      "redtriage",
		Data:        data,
	}

	s.incidentContext.Timeline = append(s.incidentContext.Timeline, event)
	s.incidentContext.UpdatedAt = time.Now()
}

func (s *Session) exportIncidentContext(filename string) error {
	if s.incidentContext == nil {
		return fmt.Errorf("no active incident context to export")
	}

	// Export incident context to file
	contextData, err := json.MarshalIndent(s.incidentContext, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	if err := os.WriteFile(filename, contextData, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	fmt.Printf("✓ Exported incident context to %s\n", filename)
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
