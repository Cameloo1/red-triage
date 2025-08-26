# Contributing to RedTriage

Thank you for your interest in contributing to RedTriage! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for build automation)
- PowerShell (Windows) or Bash (Linux/macOS)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/redtriage.git
   cd redtriage
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/redtriage/redtriage.git
   ```

## Development Setup

### Install Dependencies

```bash
# Download Go modules
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Build the Project

```bash
# Build all CLI versions
./scripts/build.sh          # Linux/macOS
./scripts/build.bat         # Windows
./scripts/build.ps1         # PowerShell

# Build specific version
go build -o redtriage ./cmd/redtriage
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test ./internal/collector

# Run tests with coverage
go test -cover ./...

# Run comprehensive test suite
./scripts/run_tests.ps1
```

## Contributing Guidelines

### What We're Looking For

- **Bug fixes**: Fixes for existing issues
- **Feature enhancements**: New functionality that aligns with project goals
- **Documentation improvements**: Better docs, examples, and guides
- **Performance improvements**: Optimizations and efficiency gains
- **Testing**: Additional test coverage and test improvements

### What to Avoid

- **Breaking changes**: Unless absolutely necessary and well-justified
- **Large refactoring**: Without prior discussion and approval
- **Platform-specific code**: Without cross-platform considerations
- **Dependencies**: Adding new dependencies without justification

## Testing Guidelines

### Test Requirements

- **Unit tests**: Required for all new functionality
- **Integration tests**: Required for complex features
- **Cross-platform tests**: Ensure compatibility across platforms
- **Performance tests**: For performance-critical code

### Running Tests

```bash
# Quick test suite
./scripts/run_tests.ps1 -QuickTest

# Full test suite
./scripts/run_tests.ps1

# Multi-CLI tests
./scripts/test_all_cli.ps1

# Go tests
go test -v ./...
```

### Test Coverage

- Aim for at least 80% test coverage
- Focus on critical paths and edge cases
- Test error conditions and failure scenarios

## Pull Request Process

### Before Submitting

1. **Ensure tests pass**: All tests must pass locally
2. **Update documentation**: Update relevant documentation
3. **Check formatting**: Ensure code follows style guidelines
4. **Rebase on main**: Keep your branch up to date

### PR Guidelines

1. **Clear title**: Descriptive title that explains the change
2. **Detailed description**: Explain what and why, not how
3. **Related issues**: Link to any related issues
4. **Screenshots**: Include screenshots for UI changes
5. **Testing**: Describe how you tested the changes

### PR Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No breaking changes
```

## Code Style

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Use meaningful variable and function names
- Add comments for exported functions

### File Organization

- Keep files focused and single-purpose
- Use descriptive file names
- Group related functionality together
- Follow Go package conventions

### Error Handling

- Always check and handle errors
- Use meaningful error messages
- Log errors appropriately
- Return errors up the call stack

## Documentation

### Documentation Standards

- Use clear, concise language
- Include examples for complex features
- Keep documentation up to date
- Use consistent formatting

### Documentation Types

- **README.md**: Project overview and quick start
- **Technical Documentation**: Architecture and implementation details
- **Testing Guide**: Testing procedures and guidelines
- **API Reference**: Function and interface documentation

## Reporting Issues

### Issue Guidelines

- Use the issue template
- Provide detailed reproduction steps
- Include system information
- Attach relevant logs and screenshots

### Issue Types

- **Bug Report**: Something isn't working
- **Feature Request**: New functionality needed
- **Documentation**: Documentation improvements
- **Enhancement**: Improvement to existing features

## Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check existing documentation first
- **Code**: Review existing code for examples

## Recognition

Contributors will be recognized in:
- Project README
- Release notes
- Contributor statistics
- Special acknowledgments for significant contributions

Thank you for contributing to RedTriage!
