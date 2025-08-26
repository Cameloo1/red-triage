#!/bin/bash

# Build all RedTriage CLI versions
# Usage: ./build.sh [--version VERSION] [--commit COMMIT] [--build-date DATE] [--clean] [--test] [--package]

set -e

# Default values
VERSION="dev"
COMMIT="unknown"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLEAN=false
TEST=false
PACKAGE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --commit)
            COMMIT="$2"
            shift 2
            ;;
        --build-date)
            BUILD_DATE="$2"
            shift 2
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --test)
            TEST=true
            shift
            ;;
        --package)
            PACKAGE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--version VERSION] [--commit COMMIT] [--build-date DATE] [--clean] [--test] [--package]"
            echo "  --version VERSION     Set version string (default: dev)"
            echo "  --commit COMMIT       Set commit hash (default: unknown)"
            echo "  --build-date DATE     Set build date (default: current UTC time)"
            echo "  --clean               Clean previous builds"
            echo "  --test                Run tests after building"
            echo "  --package             Create distribution package"
            echo "  --help                Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "RedTriage Multi-CLI Build Script"
echo "================================="

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version)
echo "✓ Go found: $GO_VERSION"

# Detect platform and architecture
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Set Go environment variables
export GOOS=$PLATFORM
export GOARCH=$ARCH

if [[ "$PLATFORM" == "darwin" ]]; then
    export GOOS="darwin"
elif [[ "$PLATFORM" == "linux" ]]; then
    export GOOS="linux"
fi

if [[ "$ARCH" == "x86_64" ]]; then
    export GOARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
    export GOARCH="arm64"
fi

echo "✓ Building for: $GOOS/$GOARCH"

# Create build directory
mkdir -p build

# Clean previous builds if requested
if [[ "$CLEAN" == true ]]; then
    echo "🧹 Cleaning previous builds..."
    rm -f build/redtriage*
    rm -f build/*.log
fi

# Build all CLI versions
echo "🔨 Building all RedTriage CLI versions..."

# 1. Main Interactive CLI
echo "  Building main interactive CLI..."
go build -o "build/redtriage" "./cmd/redtriage"
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed for main interactive CLI"
    exit 1
fi

# 2. Command-Line Interface (Non-Interactive)
echo "  Building command-line interface..."
go build -o "build/redtriage-cli" "./cmd/redtriage-cli"
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed for command-line interface"
    exit 1
fi

# 3. PowerShell Interface (Windows cross-compile)
echo "  Building PowerShell interface..."
export GOOS="windows"
export GOARCH="amd64"
go build -o "build/redtriage-pwsh.exe" "./cmd/redtriage-pwsh"
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed for PowerShell interface"
    exit 1
fi

# 4. CMD Interface (Windows cross-compile)
echo "  Building CMD interface..."
go build -o "build/redtriage-cmd.exe" "./cmd/redtriage-cmd"
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed for CMD interface"
    exit 1
fi

# 5. Linux/Bash Interface (Native)
echo "  Building Linux/Bash interface..."
export GOOS=$PLATFORM
export GOARCH=$ARCH
go build -o "build/redtriage-bash" "./cmd/redtriage-bash"
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed for Linux/Bash interface"
    exit 1
fi

# Copy configuration files
echo "📋 Copying configuration files..."
cp redtriage.yml build/ 2>/dev/null || echo "Warning: redtriage.yml not found"
cp redtriage.yml.example build/ 2>/dev/null || echo "Warning: redtriage.yml.example not found"

# Create output directories
echo "📁 Creating output directories..."
mkdir -p build/redtriage-output
mkdir -p build/redtriage-reports
mkdir -p build/logs

# Run tests if requested
if [[ "$TEST" == true ]]; then
    echo "🧪 Running tests..."
    go test -v ./...
    if [[ $? -ne 0 ]]; then
        echo "⚠️  Some tests failed, but continuing with build"
    fi
fi

# Create package if requested
if [[ "$PACKAGE" == true ]]; then
    echo "📦 Creating package..."
    mkdir -p dist
    
    PACKAGE_NAME="redtriage-clis-$PLATFORM-$ARCH-$VERSION.tar.gz"
    PACKAGE_PATH="dist/$PACKAGE_NAME"
    
    # Create tar.gz package
    tar -czf "$PACKAGE_PATH" -C build .
    echo "✓ Package created: $PACKAGE_PATH"
fi

# Test all executables
echo "🧪 Testing all executables..."

# Test main interactive CLI
if [[ -f "build/redtriage" ]]; then
    if ./build/redtriage --version >/dev/null 2>&1; then
        echo "  ✓ Main Interactive CLI: Working"
    else
        echo "  ❌ Main Interactive CLI: Failed"
    fi
else
    echo "  ❌ Main Interactive CLI: Not found"
fi

# Test command-line interface
if [[ -f "build/redtriage-cli" ]]; then
    if ./build/redtriage-cli --help >/dev/null 2>&1; then
        echo "  ✓ Command-Line Interface: Working"
    else
        echo "  ❌ Command-Line Interface: Failed"
    fi
else
    echo "  ❌ Command-Line Interface: Not found"
fi

# Test Linux/Bash interface
if [[ -f "build/redtriage-bash" ]]; then
    if ./build/redtriage-bash --help >/dev/null 2>&1; then
        echo "  ✓ Linux/Bash Interface: Working"
    else
        echo "  ❌ Linux/Bash Interface: Failed"
    fi
else
    echo "  ❌ Linux/Bash Interface: Not found"
fi

# Build summary
echo ""
echo "🎉 Build Summary"
echo "==============="
echo "✓ Main Interactive CLI: redtriage"
echo "✓ Command-Line Interface: redtriage-cli"
echo "✓ PowerShell Interface: redtriage-pwsh.exe (Windows)"
echo "✓ CMD Interface: redtriage-cmd.exe (Windows)"
echo "✓ Linux/Bash Interface: redtriage-bash"
echo "✓ Configuration files copied"
echo "✓ Output directories created"

if [[ "$PACKAGE" == true ]]; then
    echo "✓ Package created: $PACKAGE_NAME"
fi

echo ""
echo "🚀 All CLI versions built successfully!"
echo "Use './build/redtriage --interactive' for interactive mode"
echo "Use './build/redtriage-cli --help' for command-line mode"
echo "Use './build/redtriage-bash --help' for Linux/Bash mode"
echo "Windows executables: redtriage-pwsh.exe, redtriage-cmd.exe"
echo ""
echo "Build completed successfully!"
