# RedTriage Testing and Deployment Guide
*Last Updated: August 2025*

## Table of Contents
1. [Testing Overview](#testing-overview)
2. [Test Categories](#test-categories)
3. [Automated Testing](#automated-testing)
4. [Manual Testing](#manual-testing)
5. [Build System](#build-system)
6. [Deployment](#deployment)
7. [Troubleshooting](#troubleshooting)
8. [Performance Testing](#performance-testing)
9. [Security Testing](#security-testing)
10. [Continuous Integration](#continuous-integration)

## Testing Overview

RedTriage implements a comprehensive testing strategy to ensure reliability, performance, and security across all platforms and terminal types.

### Testing Philosophy
- **Comprehensive Coverage**: Test all commands, flags, and functionality
- **Cross-Platform Validation**: Ensure consistent behavior across Windows, Linux, and macOS
- **Terminal Compatibility**: Validate all CLI implementations (CMD, PowerShell, Bash)
- **Performance Validation**: Verify acceptable performance under various conditions
- **Security Verification**: Ensure secure operation and data handling

### Testing Environment
- **Development**: Local development and unit testing
- **Staging**: Production-like environment for integration testing
- **Production**: Live environment for final validation
- **CI/CD**: Automated testing in continuous integration pipeline

## Test Categories

### Unit Tests
Unit tests focus on individual components and functions:

#### Core Components
- **Collection Engine**: Artifact collection functionality
- **Detection Engine**: Threat detection and rule processing
- **Output Management**: Report generation and file handling
- **Validation**: Input validation and error handling
- **Configuration**: Configuration file parsing and management

#### Platform-Specific Tests
- **Windows**: Registry operations, Windows API calls
- **Linux**: File system operations, system calls
- **macOS**: Property list handling, system extensions

#### Terminal Interface Tests
- **Command Parsing**: Argument parsing and validation
- **Output Formatting**: Consistent output across terminals
- **Error Handling**: Graceful error handling and user feedback

### Integration Tests
Integration tests validate component interactions:

#### End-to-End Workflows
- **Complete Collection**: Full artifact collection workflow
- **Health Checks**: System health validation process
- **Report Generation**: Complete report generation workflow
- **Configuration Management**: Configuration loading and validation

#### Cross-Platform Integration
- **Platform Detection**: Automatic platform detection
- **Terminal Selection**: Appropriate terminal interface selection
- **Path Handling**: Cross-platform path compatibility
- **File Operations**: Platform-agnostic file operations

### System Tests
System tests validate complete system operation:

#### Health Validation
- **System Health**: System resource validation
- **Build Health**: Build system validation
- **Test Health**: Test suite validation
- **Performance Health**: Performance baseline validation

#### Command Validation
- **CLI Commands**: All command and flag combinations
- **Output Formats**: All supported output formats
- **Error Scenarios**: Error handling and recovery
- **Edge Cases**: Boundary conditions and limits

## Automated Testing

### Test Automation Framework

#### Go Test Integration
```go
// Example test structure
func TestHealthCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected string
        shouldFail bool
    }{
        {
            name:     "basic health check",
            args:     []string{"health"},
            expected: "PASS",
            shouldFail: false,
        },
        {
            name:     "verbose health check",
            args:     []string{"health", "--verbose"},
            expected: "PASS",
            shouldFail: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runCommand(tt.args)
            if tt.shouldFail && result.Success {
                t.Errorf("Expected failure but got success")
            }
            if !tt.shouldFail && !result.Success {
                t.Errorf("Expected success but got failure: %s", result.Error)
            }
        })
    }
}
```

#### PowerShell Test Scripts
```powershell
# Comprehensive CLI testing
param(
    [switch]$SkipTimeoutTests,
    [switch]$SkipBuildTests,
    [string]$OutputDir = "./comprehensive-test-results"
)

# Test all command combinations
$commands = @("health", "collect", "check", "profile", "report")
$flags = @("--verbose", "--output", "--help", "--version")

foreach ($command in $commands) {
    foreach ($flag in $flags) {
        $testName = "$command-$flag"
        Write-Host "Testing: $testName"
        
        try {
            $result = & "./redtriage-pwsh.exe" $command $flag
            $status = if ($result) { "PASS" } else { "FAIL" }
        } catch {
            $status = "ERROR"
        }
        
        # Record results
        Add-Content -Path "$OutputDir/results.txt" -Value "$testName: $status"
    }
}
```

#### Batch Test Scripts
```batch
@echo off
setlocal enabledelayedexpansion

set OUTPUT_DIR=comprehensive-test-results
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

echo Running comprehensive CLI tests...
echo Test Results > %OUTPUT_DIR%\results.txt

for %%c in (health collect check profile report) do (
    for %%f in (--verbose --output --help --version) do (
        echo Testing: %%c %%f
        redtriage-cmd.exe %%c %%f >nul 2>&1
        if !errorlevel! equ 0 (
            echo %%c %%f: PASS >> %OUTPUT_DIR%\results.txt
        ) else (
            echo %%c %%f: FAIL >> %OUTPUT_DIR%\results.txt
        )
    )
)

echo Tests completed. Results saved to %OUTPUT_DIR%\results.txt
```

### Test Execution

#### Automated Test Runner
```go
// test_runner.go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

type TestResult struct {
    Name      string
    Status    string
    Duration  time.Duration
    Output    string
    Error     string
}

func runComprehensiveTests() []TestResult {
    var results []TestResult
    
    // Test all CLI versions
    cliVersions := []string{"redtriage-cmd", "redtriage-pwsh", "redtriage-bash", "redtriage-cli"}
    
    for _, cli := range cliVersions {
        if _, err := os.Stat(cli + ".exe"); err == nil {
            result := testCLIVersion(cli)
            results = append(results, result)
        }
    }
    
    return results
}

func testCLIVersion(cliName string) TestResult {
    start := time.Now()
    
    // Test basic functionality
    cmd := exec.Command(cliName, "health")
    output, err := cmd.CombinedOutput()
    
    duration := time.Since(start)
    status := "PASS"
    if err != nil {
        status = "FAIL"
    }
    
    return TestResult{
        Name:     cliName,
        Status:   status,
        Duration: duration,
        Output:   string(output),
        Error:    func() string { if err != nil { return err.Error() }; return "" }(),
    }
}
```

#### Test Result Aggregation
```go
func aggregateTestResults(results []TestResult) {
    // Create output directory
    outputDir := "comprehensive-test-results"
    os.MkdirAll(outputDir, 0755)
    
    // Generate summary report
    summary := generateSummaryReport(results)
    os.WriteFile(filepath.Join(outputDir, "summary.txt"), []byte(summary), 0644)
    
    // Generate detailed results
    details := generateDetailedReport(results)
    os.WriteFile(filepath.Join(outputDir, "detailed-results.txt"), []byte(details), 0644)
    
    // Generate JSON report
    jsonData, _ := json.MarshalIndent(results, "", "  ")
    os.WriteFile(filepath.Join(outputDir, "results.json"), jsonData, 0644)
}
```

## Manual Testing

### Manual Test Procedures

#### Health Command Testing
1. **Basic Health Check**
   ```bash
   redtriage health
   ```
   - Verify output format
   - Check for any error messages
   - Validate health status

2. **Verbose Health Check**
   ```bash
   redtriage health --verbose
   ```
   - Verify detailed output
   - Check all health categories
   - Validate timing information

3. **Specific Health Checks**
   ```bash
   redtriage health --run build-system,test-suites
   ```
   - Verify specific checks run
   - Check output accuracy
   - Validate skip behavior

#### Collection Command Testing
1. **Basic Collection**
   ```bash
   redtriage collect --output ./test-output
   ```
   - Verify output directory creation
   - Check artifact collection
   - Validate report generation

2. **Extended Collection**
   ```bash
   redtriage collect --extended --output ./test-output
   ```
   - Verify extended artifacts
   - Check performance impact
   - Validate output completeness

3. **Selective Collection**
   ```bash
   redtriage collect --include processes,network --output ./test-output
   ```
   - Verify artifact filtering
   - Check excluded artifacts
   - Validate output structure

#### Cross-Platform Testing
1. **Windows Testing**
   - Test in CMD, PowerShell, and Git Bash
   - Verify registry collection
   - Check Windows-specific artifacts

2. **Linux Testing**
   - Test in various Linux distributions
   - Verify file system operations
   - Check system call compatibility

3. **macOS Testing**
   - Test in Terminal and iTerm2
   - Verify property list handling
   - Check macOS-specific features

### Test Data Validation

#### Output Validation
- **File Structure**: Verify correct directory structure
- **Content Accuracy**: Validate collected data accuracy
- **Format Compliance**: Check output format specifications
- **Metadata**: Verify timestamps and checksums

#### Performance Validation
- **Collection Time**: Measure artifact collection duration
- **Resource Usage**: Monitor CPU and memory usage
- **Scalability**: Test with varying system sizes
- **Concurrency**: Validate parallel processing

## Build System

### Multi-Platform Build

#### Windows Build Process
```batch
@echo off
echo Building RedTriage for Windows...

set GOOS=windows
set GOARCH=amd64

echo Building main executable...
go build -ldflags="-s -w" -o build/redtriage.exe ./cmd/redtriage

echo Building CMD version...
go build -ldflags="-s -w" -o build/redtriage-cmd.exe ./cmd/redtriage-cmd

echo Building PowerShell version...
go build -ldflags="-s -w" -o build/redtriage-pwsh.exe ./cmd/redtriage-pwsh

echo Building generic CLI version...
go build -ldflags="-s -w" -o build/redtriage-cli.exe ./cmd/redtriage-cli

echo Windows build completed successfully!
```

#### Linux/macOS Build Process
```bash
#!/bin/bash

echo "Building RedTriage for multiple platforms..."

# Build for multiple architectures
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    export GOOS GOARCH
    
    echo "Building for $GOOS/$GOARCH..."
    
    # Build main executable
    go build -ldflags="-s -w" -o "build/redtriage-$GOOS-$GOARCH" ./cmd/redtriage
    
    # Build terminal-specific versions
    go build -ldflags="-s -w" -o "build/redtriage-bash-$GOOS-$GOARCH" ./cmd/redtriage-bash
    go build -ldflags="-s -w" -o "build/redtriage-cli-$GOOS-$GOARCH" ./cmd/redtriage-cli
done

echo "Multi-platform build completed successfully!"
```

#### PowerShell Build Process
```powershell
# PowerShell build script
Write-Host "Building RedTriage for multiple platforms..."

$platforms = @("windows/amd64", "linux/amd64", "darwin/amd64")

foreach ($platform in $platforms) {
    $parts = $platform -split "/"
    $env:GOOS = $parts[0]
    $env:GOARCH = $parts[1]
    
    Write-Host "Building for $env:GOOS/$env:GOARCH..."
    
    # Build main executable
    go build -ldflags="-s -w" -o "build/redtriage-$env:GOOS-$env:GOARCH.exe" ./cmd/redtriage
    
    # Build terminal-specific versions
    if ($env:GOOS -eq "windows") {
        go build -ldflags="-s -w" -o "build/redtriage-cmd-$env:GOOS-$env:GOARCH.exe" ./cmd/redtriage-cmd
        go build -ldflags="-s -w" -o "build/redtriage-pwsh-$env:GOOS-$env:GOARCH.exe" ./cmd/redtriage-pwsh
    } else {
        go build -ldflags="-s -w" -o "build/redtriage-bash-$env:GOOS-$env:GOARCH" ./cmd/redtriage-bash
    }
    
    go build -ldflags="-s -w" -o "build/redtriage-cli-$env:GOOS-$env:GOARCH" ./cmd/redtriage-cli
}

Write-Host "Multi-platform build completed successfully!"
```

### Build Validation

#### Binary Validation
- **File Size**: Verify reasonable binary sizes
- **Dependencies**: Check for missing dependencies
- **Permissions**: Validate executable permissions
- **Compatibility**: Test on target platforms

#### Build Artifacts
- **Checksums**: Generate and verify checksums
- **Signatures**: Digital signature validation
- **Packaging**: Archive creation and validation
- **Distribution**: Package distribution testing

## Deployment

### Deployment Strategy

#### Staging Deployment
1. **Environment Setup**
   - Configure staging environment
   - Install dependencies
   - Set up monitoring

2. **Deployment Testing**
   - Deploy to staging
   - Run integration tests
   - Validate functionality

3. **Performance Testing**
   - Load testing
   - Stress testing
   - Resource monitoring

#### Production Deployment
1. **Pre-Deployment**
   - Final testing validation
   - Backup procedures
   - Rollback planning

2. **Deployment Process**
   - Zero-downtime deployment
   - Health check validation
   - Performance monitoring

3. **Post-Deployment**
   - Functionality validation
   - Performance baseline
   - User feedback collection

### Deployment Automation

#### CI/CD Pipeline
```yaml
# .github/workflows/deploy.yml
name: Deploy RedTriage

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - name: Test
      run: go test -v ./...
    - name: Build
      run: |
        go build -o redtriage ./cmd/redtriage
        go build -o redtriage-cmd ./cmd/redtriage-cmd
        go build -o redtriage-pwsh ./cmd/redtriage-pwsh
        go build -o redtriage-bash ./cmd/redtriage-bash
        go build -o redtriage-cli ./cmd/redtriage-cli

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to production
      run: |
        echo "Deploying to production..."
        # Add deployment commands here
```

#### Automated Testing in CI/CD
- **Pre-deployment Tests**: Run all tests before deployment
- **Integration Tests**: Validate component interactions
- **Performance Tests**: Ensure performance requirements
- **Security Tests**: Validate security requirements

## Troubleshooting

### Common Issues

#### Build Issues
1. **Dependency Problems**
   ```bash
   go mod download
   go mod tidy
   go mod verify
   ```

2. **Cross-Compilation Issues**
   ```bash
   # Set environment variables
   export GOOS=windows
   export GOARCH=amd64
   export CGO_ENABLED=0
   ```

3. **Permission Issues**
   ```bash
   chmod +x build/*.exe
   chmod +x build/*
   ```

#### Runtime Issues
1. **File Permission Errors**
   - Check file permissions
   - Verify user privileges
   - Use elevated permissions if needed

2. **Path Issues**
   - Verify path separators
   - Check working directory
   - Validate file existence

3. **Memory Issues**
   - Monitor memory usage
   - Adjust collection limits
   - Use streaming for large files

### Debug Procedures

#### Verbose Logging
```bash
# Enable verbose output
redtriage --verbose health
redtriage --debug collect
```

#### Log Analysis
```bash
# Check log files
tail -f redtriage.log
grep ERROR redtriage.log
grep WARNING redtriage.log
```

#### Performance Profiling
```bash
# Enable profiling
redtriage --profile collect
redtriage --trace health
```

## Performance Testing

### Performance Metrics

#### Collection Performance
- **Artifact Count**: Number of artifacts collected
- **Collection Time**: Time to complete collection
- **Throughput**: Artifacts per second
- **Resource Usage**: CPU and memory consumption

#### System Performance
- **Response Time**: Command response time
- **Concurrency**: Parallel operation capability
- **Scalability**: Performance with system size
- **Efficiency**: Resource utilization

### Performance Testing Tools

#### Built-in Profiling
```go
import "runtime/pprof"

func enableProfiling() {
    cpuFile, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(cpuFile)
    defer pprof.StopCPUProfile()
    
    memFile, _ := os.Create("memory.prof")
    defer pprof.WriteHeapProfile(memFile)
}
```

#### External Benchmarking
```bash
# Time command execution
time redtriage health
time redtriage collect --output ./test

# Monitor resource usage
top -p $(pgrep redtriage)
htop -p $(pgrep redtriage)
```

## Security Testing

### Security Validation

#### Input Validation
- **Command Injection**: Test command injection prevention
- **Path Traversal**: Validate path traversal protection
- **Buffer Overflow**: Test buffer overflow protection
- **Format String**: Validate format string protection

#### Output Security
- **Data Leakage**: Check for sensitive data exposure
- **Access Control**: Validate file permission handling
- **Encryption**: Verify data encryption when enabled
- **Redaction**: Test sensitive data redaction

#### Runtime Security
- **Privilege Escalation**: Test privilege escalation prevention
- **Resource Exhaustion**: Validate resource limit enforcement
- **Network Security**: Test network access restrictions
- **File Security**: Verify secure file operations

### Security Testing Tools

#### Static Analysis
```bash
# Go security scanning
gosec ./...
golangci-lint run

# Dependency scanning
nancy sleuth
```

#### Dynamic Testing
```bash
# Fuzzing tests
go test -fuzz=Fuzz ./...
go test -fuzztime=30s ./...
```

## Continuous Integration

### CI/CD Pipeline

#### Automated Testing
- **Unit Tests**: Run on every commit
- **Integration Tests**: Run on pull requests
- **System Tests**: Run before deployment
- **Performance Tests**: Run on schedule

#### Quality Gates
- **Test Coverage**: Minimum 80% coverage required
- **Performance**: Performance regression detection
- **Security**: Security scan requirements
- **Documentation**: Documentation completeness

#### Deployment Automation
- **Staging**: Automatic deployment to staging
- **Production**: Manual approval for production
- **Rollback**: Automatic rollback on failure
- **Monitoring**: Post-deployment monitoring

### CI/CD Configuration

#### GitHub Actions
```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go: [1.21, 1.22]
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go }}
    
    - name: Test
      run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.txt
```

#### GitLab CI
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.21
  script:
    - go test -v ./...
    - go test -race ./...
    - go test -cover ./...

build:
  stage: build
  image: golang:1.21
  script:
    - go build -o redtriage ./cmd/redtriage
    - go build -o redtriage-cmd ./cmd/redtriage-cmd
    - go build -o redtriage-pwsh ./cmd/redtriage-pwsh
    - go build -o redtriage-bash ./cmd/redtriage-bash
    - go build -o redtriage-cli ./cmd/redtriage-cli
  artifacts:
    paths:
      - redtriage*
    expire_in: 1 week

deploy:
  stage: deploy
  script:
    - echo "Deploying to production..."
  only:
    - main
```

---

*This document provides comprehensive guidance for testing and deploying RedTriage across all platforms and environments.*
