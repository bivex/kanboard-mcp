@echo off
setlocal

rem Build the Go application for release
echo Building kanboard-mcp.exe...
go build -ldflags="-s -w" -o kanboard-mcp.exe .

if %errorlevel% neq 0 (
    echo Error: Build failed!
    exit /b %errorlevel%
)

echo Build successful! Executable created: kanboard-mcp.exe

endlocal 
