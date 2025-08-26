# RedTriage Technical Documentation
*Last Updated: August 2025*

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Multi-CLI Architecture](#multi-cli-architecture)
3. [Health Command Implementation](#health-command-implementation)
4. [Enhanced Features](#enhanced-features)
5. [Configuration Management](#configuration-management)
6. [Testing Framework](#testing-framework)
7. [Build System](#build-system)
8. [Output Management](#output-management)
9. [Critical Fixes](#critical-fixes)
10. [Version History](#version-history)

## Architecture Overview

RedTriage follows a modular architecture designed for extensibility and reliability across multiple platforms and terminal types.

### Core Components
- **Collection Engine**: Handles artifact collection from various system sources
- **Detection Engine**: Processes Sigma rules and custom detection logic
- **Output Management**: Manages report generation and file organization
- **Platform Abstraction**: Provides cross-platform compatibility
- **Terminal Interface**: Supports multiple CLI implementations

### Directory Structure
```
redtriage/
├── cmd/                    # Command implementations
│   ├── redtriage/         # Main CLI implementation
│   ├── redtriage-cmd/     # Windows CMD compatibility
│   ├── redtriage-pwsh/    # PowerShell compatibility
│   ├── redtriage-bash/    # Bash compatibility
│   └── redtriage-cli/     # Generic CLI implementation
├── internal/               # Internal packages
│   ├── terminal/          # Terminal interface
│   ├── output/            # Output management
│   ├── logging/           # Logging system
│   ├── validation/        # Input validation
│   ├── version/           # Version management
│   ├── session/           # Session management
│   ├── registry/          # Registry operations
│   └── config/            # Configuration management
├── sigma-rules/           # Detection rules
├── build/                 # Build outputs
└── redtriage-reports/     # Centralized reports directory
```

## Multi-CLI Architecture

### Terminal-Specific Implementations

#### Windows CMD (redtriage-cmd.exe)
- Optimized for Windows Command Prompt
- Enhanced error handling for Windows-specific operations
- Registry collection and Windows Event Log support
- Process enumeration using Windows API

#### PowerShell (redtriage-pwsh.exe)
- Native PowerShell integration
- Enhanced output formatting for PowerShell
- Support for PowerShell-specific artifacts
- Interactive mode with PowerShell cmdlets

#### Bash (redtriage-bash.exe)
- Cross-platform bash compatibility
- Unix-style command line arguments
- Enhanced file system operations
- Process and network analysis

#### Generic CLI (redtriage-cli.exe)
- Platform-agnostic implementation
- Standard command line interface
- Consistent behavior across platforms
- Fallback for unsupported terminal types

### Build System Integration

The build system generates platform-specific executables:
- **Windows**: `.exe` files with Windows-specific optimizations
- **Linux/macOS**: ELF/Mach-O binaries with Unix compatibility
- **Cross-compilation**: Support for building on different platforms

## Health Command Implementation

### Health Check Categories

#### System Health
- **Process Analysis**: Running processes, services, and system state
- **Memory Analysis**: Memory usage, allocation, and potential issues
- **File System**: Disk space, file integrity, and permissions
- **Network**: Network connectivity, routing, and configuration

#### Build System Health
- **Go Environment**: Go version, modules, and dependencies
- **Build Tools**: Compiler, linker, and build utilities
- **Dependencies**: Module versions and compatibility
- **Build Outputs**: Generated binaries and artifacts

#### Test Suite Health
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Performance Tests**: Load and stress testing
- **Security Tests**: Vulnerability and security testing

### Health Check Implementation

```go
type HealthChecker struct {
    SystemChecks    []HealthCheck
    BuildChecks     []HealthCheck
    TestChecks      []HealthCheck
    Output          io.Writer
    Verbose         bool
}

type HealthCheck struct {
    Name        string
    Description string
    Run         func() HealthResult
    Required    bool
}

type HealthResult struct {
    Status      HealthStatus
    Message     string
    Details     map[string]interface{}
    Duration    time.Duration
}
```

### Health Check Results

Health checks return standardized results:
- **PASS**: Check completed successfully
- **FAIL**: Check failed with errors
- **WARNING**: Check completed with warnings
- **SKIP**: Check was skipped
- **ERROR**: Check encountered an error

## Enhanced Features

### Advanced Collection Capabilities

#### Extended Artifact Collection
- **Process Information**: Detailed process trees and dependencies
- **Network Analysis**: Active connections, routing tables, and DNS
- **File System Forensics**: Metadata, timestamps, and integrity
- **Registry Analysis**: Windows registry collection and analysis
- **Memory Analysis**: Volatile memory collection and analysis
- **Log Analysis**: System logs, security events, and applications

#### Intelligent Collection
- **Adaptive Timeouts**: Dynamic timeout adjustment based on system load
- **Resource Monitoring**: Real-time resource usage tracking
- **Incremental Updates**: Delta collection for changed artifacts
- **Parallel Processing**: Multi-threaded collection for performance

### Detection Engine Enhancements

#### Sigma Rule Support
- **Rule Validation**: Syntax and logic validation
- **Rule Compilation**: Optimized rule execution
- **Custom Rules**: User-defined detection logic
- **Rule Management**: Rule versioning and updates

#### Advanced Detection
- **Behavioral Analysis**: Pattern recognition and anomaly detection
- **Threat Intelligence**: Integration with external threat feeds
- **Risk Scoring**: Automated risk assessment and prioritization
- **False Positive Reduction**: Machine learning-based filtering

## Configuration Management

### Configuration File Structure

```yaml
# RedTriage Configuration
version: "1.0"
environment: "production"

# Collection settings
collection:
  timeout: "10m"
  max_size: "1GB"
  compression: "gzip"
  checksums: true
  
# Detection settings
detection:
  rules_path: "./sigma-rules"
  min_severity: "medium"
  timeout: "5m"
  parallel: true
  
# Output settings
output:
  format: ["html", "json", "markdown"]
  directory: "./redtriage-reports"
  retention: "30d"
  encryption: false
  
# Security settings
security:
  redaction: true
  sensitive_patterns:
    - "password"
    - "api_key"
    - "token"
  allow_network: false
```

### Environment-Specific Configuration

- **Development**: Verbose logging, debug mode, test data
- **Staging**: Production-like settings, limited resources
- **Production**: Optimized performance, minimal logging, security focus

## Testing Framework

### Test Categories

#### Unit Tests
- **Component Testing**: Individual function and method testing
- **Mock Testing**: Isolated testing with mocked dependencies
- **Edge Case Testing**: Boundary conditions and error scenarios
- **Performance Testing**: Function execution time and memory usage

#### Integration Tests
- **End-to-End Testing**: Complete workflow validation
- **API Testing**: Interface and contract validation
- **Cross-Platform Testing**: Platform compatibility validation
- **Dependency Testing**: External dependency integration

#### System Tests
- **Health Checks**: System health validation
- **Command Validation**: CLI command and flag testing
- **Output Validation**: Report generation and format validation
- **Performance Validation**: Load and stress testing

### Test Automation

#### Automated Test Execution
```bash
# Run all tests
go test ./...

# Run specific test categories
go test -tags=unit ./...
go test -tags=integration ./...
go test -tags=system ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
```

#### Continuous Integration
- **Automated Builds**: Triggered on code changes
- **Test Execution**: Automated test suite execution
- **Quality Gates**: Minimum coverage and performance requirements
- **Deployment**: Automated deployment on successful tests

## Build System

### Multi-Platform Build Support

#### Windows Build
```batch
@echo off
set GOOS=windows
set GOARCH=amd64
go build -o build/redtriage-cmd.exe ./cmd/redtriage-cmd
go build -o build/redtriage-pwsh.exe ./cmd/redtriage-pwsh
go build -o build/redtriage-cli.exe ./cmd/redtriage-cli
```

#### Linux/macOS Build
```bash
#!/bin/bash
# Build for multiple platforms
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    export GOOS GOARCH
    
    # Build main executable
    go build -o "build/redtriage-$GOOS-$GOARCH" ./cmd/redtriage
    
    # Build terminal-specific versions
    go build -o "build/redtriage-bash-$GOOS-$GOARCH" ./cmd/redtriage-bash
    go build -o "build/redtriage-cli-$GOOS-$GOARCH" ./cmd/redtriage-cli
done
```

#### PowerShell Build
```powershell
# PowerShell build script
$platforms = @("windows/amd64", "linux/amd64", "darwin/amd64")

foreach ($platform in $platforms) {
    $parts = $platform -split "/"
    $env:GOOS = $parts[0]
    $env:GOARCH = $parts[1]
    
    # Build executables
    go build -o "build/redtriage-$env:GOOS-$env:GOARCH.exe" ./cmd/redtriage
    go build -o "build/redtriage-pwsh-$env:GOOS-$env:GOARCH.exe" ./cmd/redtriage-pwsh
}
```

### Build Optimization

#### Compiler Optimizations
- **Dead Code Elimination**: Remove unused code and functions
- **Inlining**: Optimize function calls for performance
- **Cross-Compilation**: Build for target platforms
- **Stripping**: Remove debug symbols for production builds

#### Dependency Management
- **Module Pinning**: Lock dependency versions
- **Vendor Directory**: Include dependencies in source
- **Minimal Imports**: Reduce binary size and attack surface
- **Security Scanning**: Scan dependencies for vulnerabilities

## Output Management

### Centralized Report Directory

All reports are consolidated into a single `redtriage-reports/` directory:

```
redtriage-reports/
├── system-usage/           # System usage reports
├── redtriage-usage/        # RedTriage usage reports
├── health/                 # Health check reports
├── tests/                  # Test execution reports
├── artifacts/              # Collected artifacts
├── findings/               # Detection results
└── summary/                # Consolidated summaries
```

### Report Formats

#### HTML Reports
- **Interactive Navigation**: Tabbed interface and search
- **Rich Formatting**: Charts, graphs, and visual elements
- **Responsive Design**: Mobile and desktop compatibility
- **Export Options**: PDF and print support

#### JSON Reports
- **Machine Readable**: Structured data for automation
- **API Integration**: REST API and webhook support
- **Data Analysis**: Integration with analysis tools
- **Version Control**: Schema versioning and compatibility

#### Markdown Reports
- **Documentation**: Human-readable documentation
- **Version Control**: Git-friendly format
- **Collaboration**: Easy sharing and editing
- **Conversion**: Convert to other formats

### Output Validation

#### Report Integrity
- **Checksum Verification**: SHA-256 file integrity
- **Schema Validation**: JSON schema compliance
- **Format Validation**: Output format verification
- **Content Validation**: Data completeness and accuracy

## Critical Fixes

### Health Command Fixes

#### Memory Analysis Issues
- **Fixed**: Memory collection timeouts on large systems
- **Solution**: Implemented adaptive timeout based on system memory
- **Result**: Reliable memory analysis across all system sizes

#### Process Enumeration
- **Fixed**: Process list truncation on Windows
- **Solution**: Implemented pagination for large process lists
- **Result**: Complete process enumeration without truncation

#### File System Operations
- **Fixed**: Permission errors on restricted directories
- **Solution**: Added graceful fallback for inaccessible locations
- **Result**: Robust file system analysis with error handling

### CLI Compatibility Fixes

#### Windows CMD Issues
- **Fixed**: Special character handling in command arguments
- **Solution**: Enhanced argument parsing and escaping
- **Result**: Reliable command execution in CMD environment

#### PowerShell Integration
- **Fixed**: Output formatting inconsistencies
- **Solution**: Standardized output formatting across all terminals
- **Result**: Consistent user experience across platforms

#### Cross-Platform Compatibility
- **Fixed**: Path separator issues between platforms
- **Solution**: Implemented platform-agnostic path handling
- **Result**: Seamless operation across Windows, Linux, and macOS

## Version History

### Version 1.0.0 (2024-08-25)
- **Initial Release**: Core functionality and basic CLI
- **Multi-Platform Support**: Windows, Linux, and macOS
- **Basic Collection**: Process, network, and file system artifacts
- **Health Checks**: System health validation

### Version 1.1.0 (2024-08-26)
- **Enhanced Collection**: Extended artifact collection capabilities
- **Detection Engine**: Sigma rule support and threat detection
- **Multi-CLI Support**: Terminal-specific implementations
- **Advanced Reporting**: HTML, JSON, and Markdown reports

### Version 1.2.0 (2024-08-27)
- **Health Command**: Comprehensive system health validation
- **Build System**: Multi-platform build automation
- **Testing Framework**: Automated test execution and validation
- **Output Management**: Centralized report directory structure

### Version 1.3.0 (2024-08-28)
- **Critical Fixes**: Memory analysis, process enumeration, and CLI compatibility
- **Performance Optimization**: Parallel processing and resource management
- **Security Enhancements**: Input validation and output sanitization
- **Documentation**: Comprehensive technical documentation

---

*This document consolidates all technical information from various summary files into a single, comprehensive reference.*
