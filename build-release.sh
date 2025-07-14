#!/bin/bash

# Build the Go application for release
echo "Building kanboard-mcp..."

# Build the application with optimization flags
go build -ldflags="-s -w" -o kanboard-mcp .

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful! Executable created: kanboard-mcp"
else
    echo "Error: Build failed!"
    exit 1
fi 
