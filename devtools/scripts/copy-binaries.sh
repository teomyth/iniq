#!/bin/bash

# Copy GoReleaser binaries to bin directory for HTTP server
mkdir -p bin

# Detect current platform and copy the appropriate binary
if [[ "$OSTYPE" == "darwin"* ]]; then
  if [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_darwin_arm64_v8.0/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Darwin ARM64 binary"
  else
    cp dist/iniq_darwin_amd64_v1/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Darwin AMD64 binary"
  fi
else
  if [[ $(uname -m) == "aarch64" ]] || [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"
  else
    cp dist/iniq_linux_amd64_v1/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
  fi
fi

# Copy all platform binaries for HTTP server downloads
cp dist/iniq_linux_amd64_v1/iniq bin/iniq-linux-amd64 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq-linux-arm64 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"
cp dist/iniq_darwin_amd64_v1/iniq bin/iniq-darwin-amd64 2>/dev/null || echo "Warning: Could not copy Darwin AMD64 binary"
cp dist/iniq_darwin_arm64_v8.0/iniq bin/iniq-darwin-arm64 2>/dev/null || echo "Warning: Could not copy Darwin ARM64 binary"
