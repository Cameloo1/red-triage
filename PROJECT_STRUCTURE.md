# RedTriage Project Structure

This document describes the organization and structure of the RedTriage project.

## Directory Overview

```
redtriage/
├── cmd/                    # Command-line interface implementations
│   ├── root.go            # Root command and main entry point
│   ├── health.go          # Health check command
│   ├── collect.go         # Artifact collection command
│   ├── check.go           # Preflight check command
│   ├── profile.go         # Host profiling command
│   ├── report.go          # Report generation command
│   ├── config.go          # Configuration management command
│   ├── bundle.go          # Bundle management command
│   ├── findings.go        # Findings management command
│   ├── rules.go           # Detection rules command
│   ├── verify.go          # Verification command
│   ├── diag.go            # Diagnostics command
│   ├── enhanced_collect.go # Enhanced collection command
│   ├── redtriage/         # Main CLI implementation
│   ├── redtriage-cmd/     # Windows CMD compatibility
│   ├── redtriage-pwsh/    # PowerShell compatibility
│   ├── redtriage-bash/    # Bash compatibility
│   └── redtriage-cli/     # Generic CLI implementation
├── internal/               # Internal packages (not exported)
│   ├── terminal/          # Terminal interface abstractions
│   ├── output/            # Output management and formatting
│   ├── logging/           # Logging system and configuration
│   ├── validation/        # Input validation and sanitization
│   ├── version/           # Version management and information
│   ├── session/           # Session management and state
│   ├── registry/          # Registry operations (Windows)
│   └── config/            # Configuration parsing and management
├── pkg/                   # Public packages (exported)
│   ├── collector/         # Artifact collection engine
│   ├── detector/          # Threat detection engine
│   ├── packager/          # Data packaging and archiving
│   ├── reporter/          # Report generation system
│   ├── utils/             # Utility functions and helpers
│   ├── platform/          # Platform-specific implementations
│   └── integrations/      # Third-party integrations
├── scripts/               # Build and automation scripts
│   ├── build.sh           # Linux/macOS build script
│   ├── build.bat          # Windows build script
│   ├── build.ps1          # PowerShell build script
│   ├── run_tests.ps1      # Comprehensive test runner
│   ├── test_all_cli.ps1   # Multi-CLI test suite
│   └── run_tests.bat      # Windows test runner
├── docs/                  # Project documentation
│   ├── TECHNICAL_DOCUMENTATION.md
│   ├── TESTING_AND_DEPLOYMENT.md
│   ├── COMPREHENSIVE_TEST_REPORT.md
│   └── CLEANUP_SUMMARY.md
├── sigma-rules/           # Detection rules and signatures
├── build/                 # Build outputs and artifacts
├── redtriage-reports/     # Test results and reports
├── redtriage-output/      # Collection outputs
├── logs/                  # Application logs
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── redtriage.yml          # Configuration file
├── README.md              # Project overview
├── CONTRIBUTING.md        # Contribution guidelines
├── CHANGELOG.md           # Version history
├── LICENSE                # MIT License
├── .gitignore            # Git ignore rules
└── PROJECT_STRUCTURE.md   # This file
```

## Package Organization

### Command Layer (`cmd/`)

The command layer contains all CLI command implementations:

- **Root Commands**: Main command structure and entry points
- **CLI Implementations**: Platform-specific CLI versions
- **Command Logic**: Business logic for each command

### Internal Packages (`internal/`)

Internal packages are not exported and contain implementation details:

- **Terminal**: Terminal interface abstractions
- **Output**: Output management and formatting
- **Logging**: Logging system configuration
- **Validation**: Input validation and sanitization
- **Version**: Version management
- **Session**: Session state management
- **Registry**: Windows registry operations
- **Config**: Configuration management

### Public Packages (`pkg/`)

Public packages are exported and can be used by external applications:

- **Collector**: Artifact collection engine
- **Detector**: Threat detection engine
- **Packager**: Data packaging and archiving
- **Reporter**: Report generation system
- **Utils**: Utility functions
- **Platform**: Platform-specific implementations
- **Integrations**: Third-party integrations

### Scripts (`scripts/`)

Build and automation scripts:

- **Build Scripts**: Platform-specific build automation
- **Test Scripts**: Test execution and automation
- **Utility Scripts**: Development and maintenance tools

### Documentation (`docs/`)

Project documentation:

- **Technical Documentation**: Architecture and implementation details
- **Testing Guide**: Testing procedures and guidelines
- **Test Reports**: Comprehensive testing results
- **Project History**: Development and cleanup summaries

## Build System

### Multi-Platform Builds

The project supports building on multiple platforms:

- **Windows**: PowerShell and Batch scripts
- **Linux/macOS**: Bash scripts
- **Cross-Platform**: Go cross-compilation

### Build Outputs

Build outputs are organized in the `build/` directory:

- **Executables**: Platform-specific binaries
- **Reports**: Build validation reports
- **Logs**: Build process logs

## Testing Structure

### Test Organization

- **Unit Tests**: Individual component testing
- **Integration Tests**: Component interaction testing
- **System Tests**: End-to-end workflow testing
- **Performance Tests**: Load and stress testing

### Test Automation

- **PowerShell**: Primary test automation
- **Batch**: Windows test automation
- **Go Tests**: Unit and integration tests

## Configuration Management

### Configuration Files

- **redtriage.yml**: Main configuration file
- **Environment Variables**: Runtime configuration
- **Command Line Flags**: Override configuration

### Configuration Hierarchy

1. Default values
2. Configuration file
3. Environment variables
4. Command line flags

## Development Workflow

### Local Development

1. Clone repository
2. Install dependencies (`go mod download`)
3. Run tests (`go test ./...`)
4. Build project (`./scripts/build.sh`)

### Testing Workflow

1. Unit tests (`go test`)
2. Integration tests (`./scripts/run_tests.ps1`)
3. System tests (`./scripts/test_all_cli.ps1`)
4. Validation and reporting

### Build Workflow

1. Code validation
2. Test execution
3. Build compilation
4. Output validation
5. Package creation

## File Naming Conventions

### Go Files

- **Commands**: `command.go` (e.g., `health.go`)
- **Packages**: `package.go` (e.g., `collector.go`)
- **Tests**: `package_test.go` (e.g., `collector_test.go`)

### Scripts

- **Build**: `build.{sh|bat|ps1}`
- **Test**: `run_tests.{sh|bat|ps1}`
- **Utility**: `utility.{sh|bat|ps1}`

### Documentation

- **User**: `README.md`, `CONTRIBUTING.md`
- **Technical**: `docs/TECHNICAL_*.md`
- **Testing**: `docs/TESTING_*.md`
- **Project**: `CHANGELOG.md`, `PROJECT_STRUCTURE.md`

## Dependencies

### Go Modules

- **Go Version**: 1.21+
- **Module Path**: `github.com/redtriage/redtriage`
- **Dependencies**: Minimal external dependencies

### External Tools

- **Build Tools**: Make, PowerShell, Bash
- **Testing Tools**: Go test framework
- **Documentation**: Markdown, Go documentation

## Quality Assurance

### Code Quality

- **Formatting**: `gofmt` compliance
- **Linting**: `golangci-lint` validation
- **Testing**: Comprehensive test coverage
- **Documentation**: Complete API documentation

### Testing Quality

- **Coverage**: 90%+ test coverage
- **Validation**: Cross-platform compatibility
- **Automation**: Automated test execution
- **Reporting**: Comprehensive test reports

## Deployment

### Release Process

1. **Development**: Feature development and testing
2. **Testing**: Comprehensive test execution
3. **Validation**: Quality assurance and validation
4. **Release**: Version tagging and documentation
5. **Distribution**: Binary distribution and updates

### Distribution

- **Source Code**: GitHub repository
- **Binaries**: GitHub releases
- **Documentation**: GitHub wiki and docs
- **Support**: GitHub issues and discussions

---

This structure ensures a clean, organized, and maintainable codebase that follows Go best practices and provides a professional development experience.
