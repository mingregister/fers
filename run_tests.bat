@echo off
echo Running unit tests for FERS project...
echo.

echo Testing pkg/crypto...
go test ./pkg/crypto -v
if %errorlevel% neq 0 (
    echo FAILED: pkg/crypto tests
    exit /b 1
)
echo.

echo Testing pkg/storage...
go test ./pkg/storage -v
if %errorlevel% neq 0 (
    echo FAILED: pkg/storage tests
    exit /b 1
)
echo.

echo Testing pkg/dir...
go test ./pkg/dir -v
if %errorlevel% neq 0 (
    echo FAILED: pkg/dir tests
    exit /b 1
)
echo.

echo Testing pkg/config...
go test ./pkg/config -v
if %errorlevel% neq 0 (
    echo FAILED: pkg/config tests
    exit /b 1
)
echo.

echo Running all tests with coverage...
go test ./pkg/... -cover
if %errorlevel% neq 0 (
    echo FAILED: Coverage tests
    exit /b 1
)
echo.

echo All tests passed successfully!
