# RedTriage Comprehensive Test Runner
# Consolidates all testing functionality into a single script

param(
    [switch]$SkipTimeoutTests,
    [switch]$SkipBuildTests,
    [switch]$QuickTest,
    [string]$OutputDir = "./redtriage-reports/tests",
    [string]$CLIVersion = "redtriage"
)

# Set execution policy for this session only
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

Write-Host "RedTriage Comprehensive Test Runner" -ForegroundColor Green
Write-Host "===================================" -ForegroundColor Green

if ($SkipTimeoutTests) {
    Write-Host "Skipping timeout-prone tests" -ForegroundColor Yellow
}
if ($SkipBuildTests) {
    Write-Host "Skipping build-related tests" -ForegroundColor Yellow
}
if ($QuickTest) {
    Write-Host "Running quick test suite" -ForegroundColor Yellow
}
Write-Host ""

# Create output directory structure
$reportsDir = "./redtriage-reports"
$testDir = "$reportsDir/tests"
$healthDir = "$reportsDir/health"
$systemDir = "$reportsDir/system-usage"

foreach ($dir in @($reportsDir, $testDir, $healthDir, $systemDir)) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "Created directory: $dir" -ForegroundColor Green
    }
}

# Initialize results tracking
$script:allResults = @()
$startTime = Get-Date

# Test function with improved timeout and error handling
function Test-Command {
    param($command, $description, $testNumber, $timeout = 30)
    
    Write-Host "Test $testNumber`: $command" -ForegroundColor Yellow
    
    try {
        $processStartInfo = New-Object System.Diagnostics.ProcessStartInfo
        $processStartInfo.FileName = "./$CLIVersion.exe"
        $processStartInfo.Arguments = $command
        $processStartInfo.UseShellExecute = $false
        $processStartInfo.RedirectStandardOutput = $true
        $processStartInfo.RedirectStandardError = $true
        $processStartInfo.CreateNoWindow = $true
        
        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $processStartInfo
        
        Write-Host "  Starting..." -NoNewline -ForegroundColor Gray
        
        if ($process.Start()) {
            if ($process.WaitForExit($timeout * 1000)) {
                $stdout = $process.StandardOutput.ReadToEnd()
                $stderr = $process.StandardError.ReadToEnd()
                $exitCode = $process.ExitCode
                $output = "$stdout`n$stderr"
                
                if ($exitCode -eq 0) {
                    Write-Host " PASS" -ForegroundColor Green
                    $status = "PASS"
                } else {
                    Write-Host " FAIL (Exit: $exitCode)" -ForegroundColor Red
                    $status = "FAIL"
                }
                
                $warning = ""
                if ($output -match "not yet implemented") {
                    $warning = "Feature not implemented"
                    Write-Host "  WARNING: $warning" -ForegroundColor Yellow
                }
                
            } else {
                $process.Kill()
                Write-Host " TIMEOUT" -ForegroundColor Red
                $status = "TIMEOUT"
                $exitCode = -2
                $output = "Process timed out after $timeout seconds"
                $warning = ""
            }
        } else {
            Write-Host " FAILED TO START" -ForegroundColor Red
            $status = "ERROR"
            $exitCode = -3
            $output = "Failed to start process"
            $warning = ""
        }
        
        $process.Dispose()
        
    } catch {
        Write-Host " EXCEPTION" -ForegroundColor Red
        $status = "EXCEPTION"
        $exitCode = -4
        $output = $_.Exception.Message
        $warning = ""
    }
    
    $result = [PSCustomObject]@{
        TestNumber = $testNumber
        Command = $command
        Description = $description
        Status = $status
        ExitCode = $exitCode
        Output = $output
        Warning = $warning
        Timestamp = Get-Date
    }
    
    $script:allResults += $result
    Write-Host "  Added result: $($result.Status)" -ForegroundColor Gray
    return $result
}

# Define test suites
$basicTests = @(
    @{Command = ""; Description = "Help/version display"},
    @{Command = "health"; Description = "Basic health check"},
    @{Command = "health --verbose"; Description = "Verbose health check"},
    @{Command = "health --help"; Description = "Health help"},
    @{Command = "collect --help"; Description = "Collect help"},
    @{Command = "check --help"; Description = "Check help"},
    @{Command = "profile --help"; Description = "Profile help"},
    @{Command = "report --help"; Description = "Report help"}
)

$healthTests = @(
    @{Command = "health --run system"; Description = "System health check"},
    @{Command = "health --run build-system"; Description = "Build system health"},
    @{Command = "health --run test-suites"; Description = "Test suites health"},
    @{Command = "health --skip memory-analysis"; Description = "Skip memory analysis"},
    @{Command = "health --output $healthDir/health-test.json"; Description = "Health output to file"}
)

$collectionTests = @(
    @{Command = "collect --output $systemDir/collection-test"; Description = "Basic collection"},
    @{Command = "collect --extended --output $systemDir/extended-test"; Description = "Extended collection"},
    @{Command = "collect --include processes,network --output $systemDir/selective-test"; Description = "Selective collection"},
    @{Command = "collect --exclude memory --output $systemDir/exclude-test"; Description = "Exclude memory"}
)

$checkTests = @(
    @{Command = "check --verbose"; Description = "Verbose check"},
    @{Command = "check --output $testDir/check-results.json"; Description = "Check with output"}
)

$profileTests = @(
    @{Command = "profile --output $systemDir/profile-test.json"; Description = "Basic profiling"},
    @{Command = "profile --verbose --output $systemDir/profile-verbose.json"; Description = "Verbose profiling"}
)

$reportTests = @(
    @{Command = "report --output $testDir/report-test.html"; Description = "HTML report"},
    @{Command = "report --format json --output $testDir/report-test.json"; Description = "JSON report"},
    @{Command = "report --format markdown --output $testDir/report-test.md"; Description = "Markdown report"}
)

# Run test suites
Write-Host "Running Basic Tests..." -ForegroundColor Cyan
$testNumber = 1
foreach ($test in $basicTests) {
    Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
    $testNumber++
}

if (!$QuickTest) {
    Write-Host "`nRunning Health Tests..." -ForegroundColor Cyan
    foreach ($test in $healthTests) {
        Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
        $testNumber++
    }
    
    Write-Host "`nRunning Collection Tests..." -ForegroundColor Cyan
    foreach ($test in $collectionTests) {
        Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
        $testNumber++
    }
    
    Write-Host "`nRunning Check Tests..." -ForegroundColor Cyan
    foreach ($test in $checkTests) {
        Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
        $testNumber++
    }
    
    Write-Host "`nRunning Profile Tests..." -ForegroundColor Cyan
    foreach ($test in $profileTests) {
        Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
        $testNumber++
    }
    
    Write-Host "`nRunning Report Tests..." -ForegroundColor Cyan
    foreach ($test in $reportTests) {
        Test-Command -command $test.Command -description $test.Description -testNumber $testNumber
        $testNumber++
    }
}

# Debug output
Write-Host "`nDebug: Results array contains $($script:allResults.Count) items" -ForegroundColor Yellow

# Generate test summary
$endTime = Get-Date
$duration = $endTime - $startTime

# Count test results by status
$passedTests = ($script:allResults | Where-Object { $_.Status -eq "PASS" }).Count
$failedTests = ($script:allResults | Where-Object { $_.Status -ne "PASS" }).Count
$totalTests = $script:allResults.Count

$summary = @"
RedTriage Test Summary
=====================
Test Run: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
Duration: $($duration.TotalSeconds.ToString('F2')) seconds
Total Tests: $totalTests

Results:
PASS: $passedTests
FAIL: $failedTests

Detailed Results:
$(($script:allResults | ForEach-Object { "$($_.TestNumber): $($_.Command) - $($_.Status)" }) -join "`n")
"@

# Save results
$summary | Out-File -FilePath "$testDir/test-summary.txt" -Encoding UTF8
$script:allResults | ConvertTo-Json -Depth 3 | Out-File -FilePath "$testDir/test-results.json" -Encoding UTF8

# Display summary
Write-Host "`n$summary" -ForegroundColor White

Write-Host "`nTest results saved to:" -ForegroundColor Green
Write-Host "  Summary: $testDir/test-summary.txt" -ForegroundColor Cyan
Write-Host "  Details: $testDir/test-results.json" -ForegroundColor Cyan
Write-Host "  Reports: $reportsDir" -ForegroundColor Cyan

# Return exit code based on test results
if ($failedTests -gt 0) {
    Write-Host "`nWARNING: $failedTests tests failed or had issues" -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "`nAll tests passed successfully!" -ForegroundColor Green
    exit 0
}
