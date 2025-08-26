package registry

import (
	"fmt"
	"strings"
)

// Command represents a command in the registry
type Command struct {
	Name        string
	Summary     string
	Description string
	Usage       string
	Examples    []string
	Flags       []Flag
	ExitCodes   []ExitCode
	Aliases     []string
}

// Flag represents a command flag
type Flag struct {
	Name        string
	Short       string
	Long        string
	Description string
	Required    bool
	Default     string
	Type        string
}

// ExitCode represents a command exit code
type ExitCode struct {
	Code        int
	Description string
}

// Registry holds all registered commands
type Registry struct {
	commands map[string]*Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

// RegisterCommand registers a command in the registry
func (r *Registry) RegisterCommand(cmd *Command) {
	r.commands[cmd.Name] = cmd
}

// GetCommand retrieves a command by name
func (r *Registry) GetCommand(name string) (*Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// ListCommands returns all registered commands
func (r *Registry) ListCommands() []*Command {
	var cmds []*Command
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// SearchCommands searches for commands by partial name
func (r *Registry) SearchCommands(query string) []*Command {
	var results []*Command
	query = strings.ToLower(query)
	
	for _, cmd := range r.commands {
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Summary), query) {
			results = append(results, cmd)
		}
	}
	
	return results
}

// GenerateHelp generates help text for a command
func (r *Registry) GenerateHelp(cmdName string) string {
	cmd, exists := r.GetCommand(cmdName)
	if !exists {
		return fmt.Sprintf("Command '%s' not found", cmdName)
	}
	
	var help strings.Builder
	
	// Command header
	help.WriteString(fmt.Sprintf("%s - %s\n", cmd.Name, cmd.Summary))
	help.WriteString(strings.Repeat("=", len(cmd.Name)+len(cmd.Summary)+3))
	help.WriteString("\n\n")
	
	// Description
	if cmd.Description != "" {
		help.WriteString(cmd.Description)
		help.WriteString("\n\n")
	}
	
	// Usage
	if cmd.Usage != "" {
		help.WriteString("Usage:\n")
		help.WriteString(fmt.Sprintf("  %s\n\n", cmd.Usage))
	}
	
	// Flags
	if len(cmd.Flags) > 0 {
		help.WriteString("Flags:\n")
		for _, flag := range cmd.Flags {
			flagLine := fmt.Sprintf("  %s", flag.Long)
			if flag.Short != "" {
				flagLine = fmt.Sprintf("  %s, %s", flag.Short, flag.Long)
			}
			if flag.Required {
				flagLine += " (required)"
			}
			if flag.Default != "" {
				flagLine += fmt.Sprintf(" (default: %s)", flag.Default)
			}
			help.WriteString(flagLine)
			help.WriteString("\n")
			if flag.Description != "" {
				help.WriteString(fmt.Sprintf("      %s\n", flag.Description))
			}
		}
		help.WriteString("\n")
	}
	
	// Examples
	if len(cmd.Examples) > 0 {
		help.WriteString("Examples:\n")
		for _, example := range cmd.Examples {
			help.WriteString(fmt.Sprintf("  %s\n", example))
		}
		help.WriteString("\n")
	}
	
	// Exit codes
	if len(cmd.ExitCodes) > 0 {
		help.WriteString("Exit Codes:\n")
		for _, exitCode := range cmd.ExitCodes {
			help.WriteString(fmt.Sprintf("  %d - %s\n", exitCode.Code, exitCode.Description))
		}
		help.WriteString("\n")
	}
	
	// Aliases
	if len(cmd.Aliases) > 0 {
		help.WriteString("Aliases:\n")
		help.WriteString(fmt.Sprintf("  %s\n", strings.Join(cmd.Aliases, ", ")))
		help.WriteString("\n")
	}
	
	return help.String()
}

// GenerateUsage generates usage documentation
func (r *Registry) GenerateUsage() string {
	var usage strings.Builder
	
	usage.WriteString("# RedTriage Command Reference\n\n")
	usage.WriteString("This document describes all available commands in RedTriage.\n\n")
	
	// Group commands by category
	categories := map[string][]*Command{
		"Core Commands": {},
		"Collection":    {},
		"Analysis":      {},
		"Management":    {},
		"Utility":       {},
	}
	
	// Categorize commands
	for _, cmd := range r.commands {
		switch {
		case strings.Contains(cmd.Name, "collect") || strings.Contains(cmd.Name, "profile"):
			categories["Collection"] = append(categories["Collection"], cmd)
		case strings.Contains(cmd.Name, "findings") || strings.Contains(cmd.Name, "report"):
			categories["Analysis"] = append(categories["Analysis"], cmd)
		case strings.Contains(cmd.Name, "rules") || strings.Contains(cmd.Name, "config"):
			categories["Management"] = append(categories["Management"], cmd)
		case strings.Contains(cmd.Name, "help") || strings.Contains(cmd.Name, "banner") || strings.Contains(cmd.Name, "clear"):
			categories["Utility"] = append(categories["Utility"], cmd)
		default:
			categories["Core Commands"] = append(categories["Core Commands"], cmd)
		}
	}
	
	// Generate documentation for each category
	for category, cmds := range categories {
		if len(cmds) == 0 {
			continue
		}
		
		usage.WriteString(fmt.Sprintf("## %s\n\n", category))
		
		for _, cmd := range cmds {
			usage.WriteString(fmt.Sprintf("### %s\n\n", cmd.Name))
			usage.WriteString(fmt.Sprintf("%s\n\n", cmd.Summary))
			
			if cmd.Usage != "" {
				usage.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", cmd.Usage))
			}
			
			if len(cmd.Examples) > 0 {
				usage.WriteString("**Examples:**\n")
				for _, example := range cmd.Examples {
					usage.WriteString(fmt.Sprintf("- `%s`\n", example))
				}
				usage.WriteString("\n")
			}
		}
	}
	
	return usage.String()
}

// Global registry instance
var GlobalRegistry = NewRegistry()

// RegisterGlobalCommand registers a command in the global registry
func RegisterGlobalCommand(cmd *Command) {
	GlobalRegistry.RegisterCommand(cmd)
}

// GetGlobalCommand retrieves a command from the global registry
func GetGlobalCommand(name string) (*Command, bool) {
	return GlobalRegistry.GetCommand(name)
}

// SearchGlobalCommands searches for commands in the global registry
func SearchGlobalCommands(query string) []*Command {
	return GlobalRegistry.SearchCommands(query)
}

// GenerateGlobalHelp generates help for a command in the global registry
func GenerateGlobalHelp(cmdName string) string {
	return GlobalRegistry.GenerateHelp(cmdName)
}

// GenerateGlobalUsage generates usage documentation from the global registry
func GenerateGlobalUsage() string {
	return GlobalRegistry.GenerateUsage()
}
