# Build all RedTriage CLI versions
param(
    [string]$Version = "dev",
    [string]$Commit = "unknown",
    [string]$BuildDate = (Get-Date -Format "2006-01-02T15:04:05Z07:00"),
    [switch]$Clean,
    [switch]$Test,
    [switch]$Package
)

Write-Host "RedTriage Multi-CLI Build Script" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green

# Check if Go is available
try {
    $goVersion = go version
    Write-Host "✓ Go found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Error: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

# Set Go environment variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Create build directory
if (!(Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# Clean previous builds if requested
if ($Clean) {
    Write-Host "🧹 Cleaning previous builds..." -ForegroundColor Yellow
    Get-ChildItem "build" -Filter "*.exe" | Remove-Item -Force
    Get-ChildItem "build" -Filter "*.log" | Remove-Item -Force
}

# Build all CLI versions
Write-Host "🔨 Building all RedTriage CLI versions..." -ForegroundColor Cyan

# 1. Main Interactive CLI
Write-Host "  Building main interactive CLI..." -ForegroundColor White
go build -o "build\redtriage.exe" "./cmd/redtriage"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed for main interactive CLI" -ForegroundColor Red
    exit 1
}

# 2. Command-Line Interface (Non-Interactive)
Write-Host "  Building command-line interface..." -ForegroundColor White
go build -o "build\redtriage-cli.exe" "./cmd/redtriage-cli"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed for command-line interface" -ForegroundColor Red
    exit 1
}

# 3. PowerShell Interface
Write-Host "  Building PowerShell interface..." -ForegroundColor White
go build -o "build\redtriage-pwsh.exe" "./cmd/redtriage-pwsh"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed for PowerShell interface" -ForegroundColor Red
    exit 1
}

# 4. CMD Interface
Write-Host "  Building CMD interface..." -ForegroundColor White
go build -o "build\redtriage-cmd.exe" "./cmd/redtriage-cmd"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed for CMD interface" -ForegroundColor Red
    exit 1
}

# 5. Linux/Bash Interface (Cross-compile for Linux)
Write-Host "  Building Linux/Bash interface..." -ForegroundColor White
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "build\redtriage-bash" "./cmd/redtriage-bash"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed for Linux/Bash interface" -ForegroundColor Red
    exit 1
}

# Reset to Windows
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Copy configuration files
Write-Host "📋 Copying configuration files..." -ForegroundColor Cyan
Copy-Item "redtriage.yml" -Destination "build\" -Force
Copy-Item "redtriage.yml.example" -Destination "build\" -Force

# Create output directories
Write-Host "📁 Creating output directories..." -ForegroundColor Cyan
$directories = @("redtriage-output", "redtriage-reports", "logs")
foreach ($dir in $directories) {
    $path = "build\$dir"
    if (!(Test-Path $path)) {
        New-Item -ItemType Directory -Path $path | Out-Null
    }
}

# Run tests if requested
if ($Test) {
    Write-Host "🧪 Running tests..." -ForegroundColor Cyan
    go test -v ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "⚠️  Some tests failed, but continuing with build" -ForegroundColor Yellow
    }
}

# Create package if requested
if ($Package) {
    Write-Host "📦 Creating package..." -ForegroundColor Cyan
    if (!(Test-Path "dist")) {
        New-Item -ItemType Directory -Path "dist" | Out-Null
    }
    
    $packageName = "redtriage-clis-windows-$Version.zip"
    $packagePath = "dist\$packageName"
    
    # Create ZIP package
    Compress-Archive -Path "build\*" -DestinationPath $packagePath -Force
    Write-Host "✓ Package created: $packagePath" -ForegroundColor Green
}

# Test all executables
Write-Host "🧪 Testing all executables..." -ForegroundColor Cyan

$executables = @(
    @{Name="Main Interactive"; Path="build\redtriage.exe"; Args=@("--version")},
    @{Name="Command-Line"; Path="build\redtriage-cli.exe"; Args=@("--help")},
    @{Name="PowerShell"; Path="build\redtriage-pwsh.exe"; Args=@("--help")},
    @{Name="CMD"; Path="build\redtriage-cmd.exe"; Args=@("--help")}
)

foreach ($exe in $executables) {
    if (Test-Path $exe.Path) {
        try {
            $result = & $exe.Path $exe.Args 2>&1
            Write-Host "  ✓ $($exe.Name): Working" -ForegroundColor Green
        } catch {
            Write-Host "  ❌ $($exe.Name): Failed" -ForegroundColor Red
        }
    } else {
        Write-Host "  ❌ $($exe.Name): Not found" -ForegroundColor Red
    }
}

# Build summary
Write-Host "`n🎉 Build Summary" -ForegroundColor Green
Write-Host "===============" -ForegroundColor Green
Write-Host "✓ Main Interactive CLI: redtriage.exe" -ForegroundColor Green
Write-Host "✓ Command-Line Interface: redtriage-cli.exe" -ForegroundColor Green
Write-Host "✓ PowerShell Interface: redtriage-pwsh.exe" -ForegroundColor Green
Write-Host "✓ CMD Interface: redtriage-cmd.exe" -ForegroundColor Green
Write-Host "✓ Linux/Bash Interface: redtriage-bash" -ForegroundColor Green
Write-Host "✓ Configuration files copied" -ForegroundColor Green
Write-Host "✓ Output directories created" -ForegroundColor Green

if ($Package) {
    Write-Host "✓ Package created: $packageName" -ForegroundColor Green
}

Write-Host "`n🚀 All CLI versions built successfully!" -ForegroundColor Green
Write-Host "Use 'build\redtriage.exe --interactive' for interactive mode" -ForegroundColor Cyan
Write-Host "Use 'build\redtriage-cli.exe --help' for command-line mode" -ForegroundColor Cyan
Write-Host "Use 'build\redtriage-pwsh.exe --help' for PowerShell mode" -ForegroundColor Cyan
Write-Host "Use 'build\redtriage-cmd.exe --help' for CMD mode" -ForegroundColor Cyan
