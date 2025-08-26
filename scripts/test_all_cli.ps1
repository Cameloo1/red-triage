# RedTriage Multi-CLI Comprehensive Test
# Tests all CLI versions to ensure 100% functionality

param(
    [switch]$QuickTest,
    [string]$OutputDir = "./redtriage-reports/multi-cli-tests"
)

Write-Host "RedTriage Multi-CLI Comprehensive Test" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host ""

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    Write-Host "Created output directory: $OutputDir" -ForegroundColor Green
}

# Define CLI versions to test
$cliVersions = @(
    @{Name = "redtriage"; Description = "Main CLI"},
    @{Name = "redtriage-cmd"; Description = "Windows CMD"},
    @{Name = "redtriage-pwsh"; Description = "PowerShell"},
    @{Name = "redtriage-bash"; Description = "Bash"},
    @{Name = "redtriage-cli"; Description = "Generic CLI"}
)

# Define test commands
$testCommands = @(
    @{Command = "--version"; Description = "Version display"},
    @{Command = "--help"; Description = "Help display"},
    @{Command = "health"; Description = "Health command"},
    @{Command = "health --help"; Description = "Health help"},
    @{Command = "collect --help"; Description = "Collect help"},
    @{Command = "check --help"; Description = "Check help"},
    @{Command = "profile --help"; Description = "Profile help"},
    @{Command = "report --help"; Description = "Report help"}
)

# Initialize results
$allResults = @()
$startTime = Get-Date

# Test function
function Test-CLIVersion {
    param($cliName, $cliDescription)
    
    Write-Host "Testing $cliDescription ($cliName.exe)..." -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Cyan
    
    $cliResults = @()
    
    foreach ($test in $testCommands) {
        $testNumber = $cliResults.Count + 1
        Write-Host "  Test $testNumber`: $($test.Command)" -ForegroundColor Yellow
        
        try {
            $processStartInfo = New-Object System.Diagnostics.ProcessStartInfo
            $processStartInfo.FileName = "./$cliName.exe"
            $processStartInfo.Arguments = $test.Command
            $processStartInfo.UseShellExecute = $false
            $processStartInfo.RedirectStandardOutput = $true
            $processStartInfo.RedirectStandardError = $true
            $processStartInfo.CreateNoWindow = $true
            
            $process = New-Object System.Diagnostics.Process
            $process.StartInfo = $processStartInfo
            
            Write-Host "    Starting..." -NoNewline -ForegroundColor Gray
            
            if ($process.Start()) {
                if ($process.WaitForExit(15000)) { # 15 second timeout
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
                    
                } else {
                    $process.Kill()
                    Write-Host " TIMEOUT" -ForegroundColor Red
                    $status = "TIMEOUT"
                    $exitCode = -2
                    $output = "Process timed out after 15 seconds"
                }
            } else {
                Write-Host " FAILED TO START" -ForegroundColor Red
                $status = "ERROR"
                $exitCode = -3
                $output = "Failed to start process"
            }
            
            $process.Dispose()
            
        } catch {
            Write-Host " EXCEPTION" -ForegroundColor Red
            $status = "EXCEPTION"
            $exitCode = -4
            $output = $_.Exception.Message
        }
        
        $result = [PSCustomObject]@{
            CLIName = $cliName
            CLIDescription = $cliDescription
            TestNumber = $testNumber
            Command = $test.Command
            Description = $test.Description
            Status = $status
            ExitCode = $exitCode
            Output = $output
            Timestamp = Get-Date
        }
        
        $cliResults += $result
        $script:allResults += $result
    }
    
    # Display CLI summary
    $cliPassed = ($cliResults | Where-Object { $_.Status -eq "PASS" }).Count
    $cliTotal = $cliResults.Count
    $cliSuccessRate = [math]::Round(($cliPassed / $cliTotal) * 100, 1)
    
    Write-Host "  $cliDescription Results: $cliPassed/$cliTotal PASS ($cliSuccessRate%)" -ForegroundColor $(if ($cliSuccessRate -eq 100) { "Green" } else { "Yellow" })
    Write-Host ""
    
    return $cliResults
}

# Test all CLI versions
foreach ($cli in $cliVersions) {
    if (Test-Path "$($cli.Name).exe") {
        Test-CLIVersion -cliName $cli.Name -cliDescription $cli.Description
    } else {
        Write-Host "WARNING: $($cli.Name).exe not found, skipping..." -ForegroundColor Yellow
    }
}

# Generate comprehensive summary
$endTime = Get-Date
$duration = $endTime - $startTime

$totalTests = $allResults.Count
$totalPassed = ($allResults | Where-Object { $_.Status -eq "PASS" }).Count
$totalFailed = ($allResults | Where-Object { $_.Status -ne "PASS" }).Count
$overallSuccessRate = [math]::Round(($totalPassed / $totalTests) * 100, 1)

# CLI-specific summaries
$cliSummaries = @()
foreach ($cli in $cliVersions) {
    $cliResults = $allResults | Where-Object { $_.CLIName -eq $cli.Name }
    if ($cliResults.Count -gt 0) {
        $cliPassed = ($cliResults | Where-Object { $_.Status -eq "PASS" }).Count
        $cliTotal = $cliResults.Count
        $cliSuccessRate = [math]::Round(($cliPassed / $cliTotal) * 100, 1)
        $cliSummaries += "$($cli.Description): $cliPassed/$cliTotal ($cliSuccessRate%)"
    }
}

$summary = @"
RedTriage Multi-CLI Comprehensive Test Summary
==============================================
Test Run: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
Duration: $($duration.TotalSeconds.ToString('F2')) seconds
Total Tests: $totalTests

Overall Results:
PASS: $totalPassed
FAIL: $totalFailed
Success Rate: $overallSuccessRate%

CLI Version Results:
$(($cliSummaries | ForEach-Object { "  $_" }) -join "`n")

Detailed Results:
$(($allResults | ForEach-Object { "$($_.CLIName): $($_.Command) - $($_.Status)" }) -join "`n")
"@

# Save results
$summary | Out-File -FilePath "$OutputDir/multi-cli-test-summary.txt" -Encoding UTF8
$allResults | ConvertTo-Json -Depth 3 | Out-File -FilePath "$OutputDir/multi-cli-test-results.json" -Encoding UTF8

# Display summary
Write-Host "`n$summary" -ForegroundColor White

Write-Host "`nTest results saved to:" -ForegroundColor Green
Write-Host "  Summary: $OutputDir/multi-cli-test-summary.txt" -ForegroundColor Cyan
Write-Host "  Details: $OutputDir/multi-cli-test-results.json" -ForegroundColor Cyan

# Final assessment
if ($overallSuccessRate -eq 100) {
    Write-Host "`n🎉 EXCELLENT: 100% Success Rate Achieved!" -ForegroundColor Green
    Write-Host "All CLI versions are working perfectly!" -ForegroundColor Green
    exit 0
} elseif ($overallSuccessRate -ge 95) {
    Write-Host "`n✅ VERY GOOD: $overallSuccessRate% Success Rate" -ForegroundColor Green
    Write-Host "Minor issues detected but overall functionality is excellent." -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n⚠️  ATTENTION NEEDED: $overallSuccessRate% Success Rate" -ForegroundColor Yellow
    Write-Host "Some issues detected that need attention." -ForegroundColor Yellow
    exit 1
}
