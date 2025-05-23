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
if go install github.com/caarlos0/svu@latest; then
    echo -e "${GREEN}✅ svu installed successfully.${NC}"

    # Check if GOPATH/bin is in PATH
    GOPATH=$(go env GOPATH)
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        echo -e "${YELLOW}GOPATH/bin is not in your PATH.${NC}"
        echo -e "To use svu globally, GOPATH/bin needs to be added to your PATH."
        echo ""

        # Detect shell and shell config file
        SHELL_NAME=$(basename "$SHELL")
        case "$SHELL_NAME" in
            bash)
                SHELL_CONFIG="$HOME/.bashrc"
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    SHELL_CONFIG="$HOME/.bash_profile"
                fi
                ;;
            zsh)
                SHELL_CONFIG="$HOME/.zshrc"
                ;;
            fish)
                SHELL_CONFIG="$HOME/.config/fish/config.fish"
                ;;
            *)
                SHELL_CONFIG="$HOME/.profile"
                ;;
        esac

        # Ask user if they want to add GOPATH/bin to PATH
        echo -e "${BLUE}Would you like to add GOPATH/bin to your PATH automatically? (y/N)${NC}"
        read -r response

        if [[ "$response" =~ ^[Yy]$ ]]; then
            # Add to shell config file
            if [[ "$SHELL_NAME" == "fish" ]]; then
                echo "set -gx PATH \$PATH $GOPATH/bin" >> "$SHELL_CONFIG"
            else
                echo "export PATH=\"\$PATH:$GOPATH/bin\"" >> "$SHELL_CONFIG"
            fi

            echo -e "${GREEN}✅ Added GOPATH/bin to $SHELL_CONFIG${NC}"
            echo -e "${YELLOW}Please restart your terminal or run:${NC}"
            if [[ "$SHELL_NAME" == "fish" ]]; then
                echo -e "  source $SHELL_CONFIG"
            else
                echo -e "  source $SHELL_CONFIG"
            fi
            echo -e "${YELLOW}to make svu available globally.${NC}"

            # Also export for current session
            export PATH="$PATH:$GOPATH/bin"
        else
            echo -e "${YELLOW}GOPATH/bin was not added to PATH.${NC}"
            echo -e "You can add it manually by adding this line to $SHELL_CONFIG:"
            if [[ "$SHELL_NAME" == "fish" ]]; then
                echo -e "  set -gx PATH \$PATH $GOPATH/bin"
            else
                echo -e "  export PATH=\"\$PATH:$GOPATH/bin\""
            fi
        fi
    fi

    # Verify installation
    if command -v svu &> /dev/null; then
        echo -e "${GREEN}svu is now available globally:${NC}"
        echo -e "  $(svu --version 2>&1)"
    else
        echo -e "${YELLOW}svu was installed but is not in your PATH.${NC}"
        echo -e "You can run it using: $GOPATH/bin/svu"
    fi
else
    echo -e "${RED}Failed to install svu.${NC}"
    exit 1
fi
