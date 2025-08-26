@echo off
echo  RedTriage Test Automation Runner
echo ===================================
echo.

echo  Running Comprehensive Test Runner...
go run run_comprehensive_tests.go
if %errorlevel% neq 0 (
    echo ❌ Comprehensive tests failed
    goto :error
)
echo ✅ Comprehensive tests passed
echo.

echo  Now running component tests...
echo.

echo  Testing Collector Component...
go run test_collector.go
if %errorlevel% neq 0 (
    echo ❌ Collector test failed
    goto :error
)
echo ✅ Collector test passed
echo.

echo  Testing Detector Component...
go run test_detector.go
if %errorlevel% neq 0 (
    echo ❌ Detector test failed
    goto :error
)
echo ✅ Detector test passed
echo.

echo  Testing Packager Component...
go run test_packager.go
if %errorlevel% neq 0 (
    echo ❌ Packager test failed
    goto :error
)
echo ✅ Packager test passed
echo.

echo  Testing Reporter Component...
go run test_reporter.go
if %errorlevel% neq 0 (
    echo ❌ Reporter test failed
    goto :error
)
echo ✅ Reporter test passed
echo.

echo  All component tests completed successfully!
echo.
echo  TEST SUMMARY
echo ===============
echo ✅ Comprehensive Tests: PASS
echo ✅ Collector Component: PASS
echo ✅ Detector Component: PASS
echo ✅ Packager Component: PASS
echo ✅ Reporter Component: PASS
echo.
echo  Overall Status: 🟢 EXCELLENT (100% Success Rate)
echo.
echo  RedTriage is ready for production use!
goto :end

:error
echo.
echo ❌ Test execution failed
echo Please check the error messages above
exit /b 1

:end
echo.
echo  Test automation completed successfully!
pause
