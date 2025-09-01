# RedTriage - Professional Incident Response Triage Tool

A comprehensive, static, cross-platform incident response triage CLI tool that collects volatile and persistent artifacts, runs local detections, and packages everything into a signed archive with a manifest and concise report.

## Overview

RedTriage is a professional-grade incident response tool that provides comprehensive system triage analysis, but focused on static endpoint forensics and incident response. JRedTriage captures and analyzes system artifacts, processes, and forensic data to provide complete visibility into system state during incident response scenarios.

## Key Features

### Core Capabilities
- **Artifact Collection**: Comprehensive collection of volatile and persistent system artifacts
- **Detection Engine**: Local threat detection using Sigma rules and custom detection logic
- **Forensic Packaging**: Secure packaging with checksums and manifest generation
- **Multi-Format Reporting**: HTML, Markdown, and JSON report generation
- **Cross-Platform Support**: Windows, Linux, and macOS compatibility

### Collection Artifacts
- **Process Information**: Running processes, services, and system state
- **Network Data**: Active connections, routing tables, and network configuration
- **File System**: File metadata, timestamps, and integrity checks
- **Registry/Configuration**: System configuration and registry data
- **Memory Analysis**: Volatile memory collection and analysis
- **Log Analysis**: System logs, security events, and application logs

### Detection & Analysis
- **Threat Detection**: Sigma rule-based detection engine
- **Anomaly Detection**: Behavioral analysis and pattern recognition
- **Forensic Timeline**: Event timeline reconstruction
- **Risk Assessment**: Automated risk scoring and prioritization

## Architecture

RedTriage follows a modular architecture designed for extensibility and reliability:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Collection    │    │   Detection     │    │   Packaging     │
│   Engine        │───▶│   Engine        │───▶│   & Reporting   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │  ▲                   │   ▲                     │
         ▼  │                   ▼   │                     ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Platform      │    │   Rule Engine   │    │   Output       │
│   Collectors    │    │   (Sigma)       │    │   Management   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites
- Go 1.21 or higher
- Windows, Linux, or macOS
- Sufficient disk space for artifact collection

### Quick Start
```bash
# Clone the repository
git clone https://github.com/red-triage/redtriage.git
cd redtriage

# Build all CLI versions
./scripts/build.sh          # Linux/macOS
./scripts/build.bat         # Windows
./scripts/build.ps1         # PowerShell

# Run health check
./build/redtriage-cmd health

# Start collection
./build/redtriage-cmd collect --output ./incident-response
```

## Usage

### Basic Commands
```bash
# System health check
redtriage health --output health-report.json

# Artifact collection
redtriage collect --extended --output ./triage-bundle

# Preflight checks
redtriage check --verbose

# Host profiling
redtriage profile --output host-profile.json

# Interactive mode
redtriage --interactive
```

### Advanced Collection
```bash
# Extended collection with specific artifacts
redtriage collect \
  --extended \
  --include processes,network,files \
  --exclude memory \
  --output ./comprehensive-triage

# Custom timeout and compression
redtriage collect \
  --timeout 600 \
  --compression tar.gz \
  --checksums \
  --output ./custom-triage
```

## Configuration

RedTriage uses a YAML configuration file (`redtriage.yml`) for customization:

```yaml
# Collection settings
detection_timeout: "5m"
min_severity: "medium"
compression_level: 6

# Security settings
checksum_algorithm: "sha256"
redaction_enabled: true
allow_network: false

# Artifact-specific settings
artifacts:
  processes:
    enabled: true
    timeout: "5m"
    max_size: "50MB"
  network:
    enabled: true
    timeout: "1m"
    max_size: "5MB"
```

## Output & Reports

### Report Formats
- **HTML Reports**: Comprehensive web-based reports with navigation
- **Markdown Reports**: Plain text reports for documentation
- **JSON Reports**: Machine-readable data for automation
- **Timeline Reports**: Chronological event reconstruction

### Output Structure
```
output-directory/
├── redtriage-RT-[timestamp]-[hash]/
│   ├── artifacts/           # Collected system artifacts
│   ├── findings/            # Detection results
│   ├── reports/             # Generated reports
│   ├── manifest.json        # Collection manifest
│   └── checksums.txt        # File integrity checksums
├── collection.log           # Collection process log
└── summary.json            # Collection summary
```

## Detection Rules

RedTriage supports Sigma rules for threat detection:

```yaml
title: Suspicious Process Creation
id: 12345678-1234-1234-1234-123456789012
status: test
description: Detects suspicious process creation patterns
author: RedTriage Team
date: 2025/08/26
logsource:
    category: process_creation
    product: windows
detection:
    selection:
        CommandLine|contains:
            - 'powershell.exe -enc'
            - 'cmd.exe /c powershell'
    condition: selection
level: high
```

## Testing & Validation

### Health Checks
```bash
# Comprehensive system health check
redtriage health --verbose --output health-report.json

# Specific health checks
redtriage health --run build-system,test-suites

# Skip specific checks
redtriage health --skip memory-analysis
```

### Test Suites
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Performance Tests**: Load and stress testing
- **Security Tests**: Vulnerability and security testing

### Comprehensive Feature Testing
```bash
# Run comprehensive CLI test (64 test combinations)
pwsh -File scripts/test_all_cli.ps1

# Run basic test suite
pwsh -File scripts/run_tests.ps1 -QuickTest

# Run full test suite
pwsh -File scripts/run_tests.ps1
```

The comprehensive test scripts validate every command and flag combination, generating detailed reports in `./redtriage-reports/`.

## Platform Support

### Windows
- Process and service enumeration
- Registry collection and analysis
- Event log analysis
- Windows-specific artifacts

### Linux
- Process and system call analysis
- File system forensics
- Kernel module analysis
- Systemd service analysis

### macOS
- Process and application analysis
- Property list collection
- System extension analysis
- macOS-specific artifacts

## Security Features

- **Checksum Verification**: SHA-256 integrity checking
- **Secure Packaging**: Encrypted archive support
- **Redaction**: Sensitive data masking
- **Audit Logging**: Complete operation logging
- **Access Control**: Role-based permissions

## Performance & Scalability

- **Parallel Collection**: Multi-threaded artifact collection
- **Incremental Updates**: Delta collection support
- **Resource Management**: Memory and CPU optimization
- **Distributed Collection**: Multi-host collection support

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build development version
go build -o redtriage-dev ./cmd/redtriage
```

## Documentation

- [Project Structure](PROJECT_STRUCTURE.md)
- [Technical Documentation](docs/TECHNICAL_DOCUMENTATION.md)
- [Testing and Deployment Guide](docs/TESTING_AND_DEPLOYMENT.md)
- [Comprehensive Test Report](docs/COMPREHENSIVE_TEST_REPORT.md)
- [Project Cleanup Summary](docs/CLEANUP_SUMMARY.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Cisco NetFlow's comprehensive flow analysis approach
- Built on industry-standard forensic tools and methodologies
- Community-driven development and testing

## Support

- **Issues**: [GitHub Issues](https://github.com/redtriage/redtriage/issues)
- **Discussions**: [GitHub Discussions](https://github.com/redtriage/redtriage/discussions)
- **Documentation**: [Wiki](https://github.com/redtriage/redtriage/wiki)

---

**RedTriage** - Professional Incident Response Triage Tool  
*Designed to work as NetFlow by Cisco*  
*Version: 1.0.0* | *Build: 2025-08-26*
