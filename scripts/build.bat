@echo off
echo Building RedTriage Multi-CLI Tool...
echo ===================================

REM Set Go environment variables
set GOOS=windows
set GOARCH=amd64

REM Check if Go is available
go version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Create build directory
if not exist build mkdir build

REM Clean previous builds
echo Cleaning previous builds...
if exist build\*.exe del /q build\*.exe
if exist build\*.log del /q build\*.log

REM Build all CLI versions
echo Building all RedTriage CLI versions...

REM 1. Main Interactive CLI
echo   Building main interactive CLI...
go build -o build\redtriage.exe ./cmd/redtriage
if %ERRORLEVEL% neq 0 (
    echo Build failed for main interactive CLI
    pause
    exit /b 1
)

REM 2. Command-Line Interface (Non-Interactive)
echo   Building command-line interface...
go build -o build\redtriage-cli.exe ./cmd/redtriage-cli
if %ERRORLEVEL% neq 0 (
    echo Build failed for command-line interface
    pause
    exit /b 1
)

REM 3. PowerShell Interface
echo   Building PowerShell interface...
go build -o build\redtriage-pwsh.exe ./cmd/redtriage-pwsh
if %ERRORLEVEL% neq 0 (
    echo Build failed for PowerShell interface
    pause
    exit /b 1
)

REM 4. CMD Interface
echo   Building CMD interface...
go build -o build\redtriage-cmd.exe ./cmd/redtriage-cmd
if %ERRORLEVEL% neq 0 (
    echo Build failed for CMD interface
    pause
    exit /b 1
)

REM 5. Linux/Bash Interface (Cross-compile for Linux)
echo   Building Linux/Bash interface...
set GOOS=linux
go build -o build\redtriage-bash ./cmd/redtriage-bash
if %ERRORLEVEL% neq 0 (
    echo Build failed for Linux/Bash interface
    pause
    exit /b 1
)

REM Reset to Windows
set GOOS=windows

REM Copy configuration files
echo Copying configuration files...
copy redtriage.yml build\ >nul 2>&1
copy redtriage.yml.example build\ >nul 2>&1

REM Create output directories
echo Creating output directories...
if not exist build\redtriage-output mkdir build\redtriage-output
if not exist build\redtriage-reports mkdir build\redtriage-reports
if not exist build\logs mkdir build\logs

REM Test all executables
echo Testing all executables...

REM Test main interactive CLI
if exist build\redtriage.exe (
    build\redtriage.exe --version >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo   ✓ Main Interactive CLI: Working
    ) else (
        echo   ❌ Main Interactive CLI: Failed
    )
) else (
    echo   ❌ Main Interactive CLI: Not found
)

REM Test command-line interface
if exist build\redtriage-cli.exe (
    build\redtriage-cli.exe --help >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo   ✓ Command-Line Interface: Working
    ) else (
        echo   ❌ Command-Line Interface: Failed
    )
) else (
    echo   ❌ Command-Line Interface: Not found
)

REM Test PowerShell interface
if exist build\redtriage-pwsh.exe (
    build\redtriage-pwsh.exe --help >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo   ✓ PowerShell Interface: Working
    ) else (
        echo   ❌ PowerShell Interface: Failed
    )
) else (
    echo   ❌ PowerShell Interface: Not found
)

REM Test CMD interface
if exist build\redtriage-cmd.exe (
    build\redtriage-cmd.exe --help >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo   ✓ CMD Interface: Working
    ) else (
        echo   ❌ CMD Interface: Failed
    )
) else (
    echo   ❌ CMD Interface: Not found
)

REM Build summary
echo.
echo 🎉 Build Summary
echo ===============
echo ✓ Main Interactive CLI: redtriage.exe
echo ✓ Command-Line Interface: redtriage-cli.exe
echo ✓ PowerShell Interface: redtriage-pwsh.exe
echo ✓ CMD Interface: redtriage-cmd.exe
echo ✓ Linux/Bash Interface: redtriage-bash
echo ✓ Configuration files copied
echo ✓ Output directories created
echo.
echo 🚀 All CLI versions built successfully!
echo Use 'build\redtriage.exe --interactive' for interactive mode
echo Use 'build\redtriage-cli.exe --help' for command-line mode
echo Use 'build\redtriage-pwsh.exe --help' for PowerShell mode
echo Use 'build\redtriage-cmd.exe --help' for CMD mode
echo.
pause
