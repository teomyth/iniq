#!/bin/bash

# Copy GoReleaser binaries to bin directory for HTTP server
# Note: GoReleaser only builds Linux binaries for production releases
# macOS binaries are built separately for development purposes
mkdir -p bin

# Detect current platform and copy the appropriate binary
if [[ "$OSTYPE" == "darwin"* ]]; then
  # For macOS development, try to copy from GoReleaser first, then fallback to development build
  if [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_darwin_arm64_v8.0/iniq bin/iniq 2>/dev/null || \
    cp bin/iniq-darwin-arm64 bin/iniq 2>/dev/null || \
    echo "Warning: No Darwin ARM64 binary found. Run 'task build:darwin-arm64' to build for development."
  else
    cp dist/iniq_darwin_amd64_v1/iniq bin/iniq 2>/dev/null || \
    cp bin/iniq-darwin-amd64 bin/iniq 2>/dev/null || \
    echo "Warning: No Darwin AMD64 binary found. Run 'task build:darwin-amd64' to build for development."
  fi
else
  if [[ $(uname -m) == "aarch64" ]] || [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"
  else
    cp dist/iniq_linux_amd64_v1/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
  fi
fi

# Copy all platform binaries for HTTP server downloads
# Linux binaries (production supported)
cp dist/iniq_linux_amd64_v1/iniq bin/iniq-linux-amd64 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq-linux-arm64 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"

# macOS binaries (development only - these won't exist from GoReleaser)
# Try to copy from development builds if they exist
cp bin/iniq-darwin-amd64 bin/iniq-darwin-amd64 2>/dev/null || echo "Info: Darwin AMD64 binary not available (development only)"
cp bin/iniq-darwin-arm64 bin/iniq-darwin-arm64 2>/dev/null || echo "Info: Darwin ARM64 binary not available (development only)"
