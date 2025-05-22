#!/bin/bash
# Script to install svu (Semantic Version Util) tool

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    echo -e "Please install Go before continuing: https://golang.org/doc/install"
    exit 1
fi

# Check if svu is already installed
if command -v svu &> /dev/null; then
    echo -e "${GREEN}svu is already installed.${NC}"
    echo -e "Current version: $(svu --version 2>&1)"
    exit 0
fi

echo -e "${BLUE}Installing svu (Semantic Version Util)...${NC}"

# Install svu
if go install github.com/caarlos0/svu/v2/cmd/svu@latest; then
    echo -e "${GREEN}✅ svu installed successfully.${NC}"
    
    # Check if GOPATH/bin is in PATH
    GOPATH=$(go env GOPATH)
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        echo -e "${YELLOW}Warning: GOPATH/bin is not in your PATH.${NC}"
        echo -e "To use svu, you need to add GOPATH/bin to your PATH:"
        echo -e "  export PATH=\"\$GOPATH/bin:\$PATH\""
        echo -e "You can add this line to your ~/.bashrc or ~/.zshrc file."
    fi
    
    # Verify installation
    if command -v svu &> /dev/null; then
        echo -e "${GREEN}svu is now available:${NC}"
        echo -e "  $(svu --version 2>&1)"
    else
        echo -e "${YELLOW}svu was installed but is not in your PATH.${NC}"
        echo -e "You can run it using: $GOPATH/bin/svu"
    fi
else
    echo -e "${RED}Failed to install svu.${NC}"
    exit 1
fi
