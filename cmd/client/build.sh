#!/bin/bash

# Define build variables
BUILD_VERSION="v1.0.0"
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_COMMIT=$(git rev-parse HEAD)
LDFLAGS="-X main.buildVersion=$BUILD_VERSION -X main.buildDate=$BUILD_DATE -X main.buildCommit=$BUILD_COMMIT -s -w"

# Build for Windows
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ./builds/client_amd64_windows.exe main.go
echo "build client_amd64_windows.exe"

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ./builds/client_amd64_macos main.go
echo "build client_amd64_macos"

GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o ./builds/client_arm64_macos main.go
echo "build client_arm64_macos"

# Build for Linux
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ./builds/client_amd64_linux main.go
echo "build client_amd64_linux"

