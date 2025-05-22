#!/bin/bash
# install-watchexec.sh - Script to install watchexec based on OS

# Define colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_os() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  # Map architecture names
  case "$ARCH" in
    x86_64)
      ARCH="x86_64"
      ;;
    aarch64|arm64)
      ARCH="aarch64"
      ;;
  esac

  echo -e "${BLUE}[WATCHEXEC]${NC} Detected OS: $OS, Architecture: $ARCH"
}

# Install watchexec on macOS
install_macos() {
  echo -e "${BLUE}[WATCHEXEC]${NC} Installing watchexec via Homebrew..."
  if ! command -v brew &> /dev/null; then
    echo -e "${YELLOW}[WATCHEXEC]${NC} Homebrew not found. Please install Homebrew first:"
    echo -e "${YELLOW}[WATCHEXEC]${NC} /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    return 1
  fi

  brew install watchexec
  return $?
}

# Install watchexec on Linux via direct download
install_linux() {
  echo -e "${BLUE}[WATCHEXEC]${NC} Installing watchexec via direct download..."

  # Create temporary directory
  TEMP_DIR=$(mktemp -d)
  cd $TEMP_DIR

  # Download latest release
  echo -e "${BLUE}[WATCHEXEC]${NC} Downloading watchexec..."

  # Use a direct download URL for the latest version
  VERSION="1.25.1"  # Hardcoded latest stable version
  DOWNLOAD_URL="https://github.com/watchexec/watchexec/releases/download/v${VERSION}/watchexec-${VERSION}-${ARCH}-unknown-linux-gnu.tar.xz"

  echo -e "${BLUE}[WATCHEXEC]${NC} Using download URL: $DOWNLOAD_URL"

  # Download the file
  if ! curl -L -o watchexec.tar.xz "$DOWNLOAD_URL"; then
    echo -e "${RED}[WATCHEXEC]${NC} Failed to download watchexec from $DOWNLOAD_URL"
    cd /
    rm -rf $TEMP_DIR
    return 1
  fi

  # Extract archive
  echo -e "${BLUE}[WATCHEXEC]${NC} Extracting archive..."
  tar -xf watchexec.tar.xz

  # Find the extracted directory
  EXTRACTED_DIR=$(find . -type d -name "watchexec-*" | head -n 1)

  if [ -z "$EXTRACTED_DIR" ]; then
    echo -e "${RED}[WATCHEXEC]${NC} Failed to find extracted directory"
    cd /
    rm -rf $TEMP_DIR
    return 1
  fi

  # Install binary
  echo -e "${BLUE}[WATCHEXEC]${NC} Installing binary..."
  cd "$EXTRACTED_DIR"
  sudo install -m 755 watchexec /usr/local/bin/

  # Clean up
  cd /
  rm -rf $TEMP_DIR

  return 0
}

# Main installation function
install_watchexec() {
  # Check if watchexec is already installed
  if command -v watchexec &> /dev/null; then
    echo -e "${GREEN}[WATCHEXEC]${NC} watchexec is already installed."
    return 0
  fi

  # Detect OS
  detect_os

  # Install based on OS
  case "$OS" in
    darwin)
      install_macos
      ;;
    linux)
      install_linux
      ;;
    *)
      echo -e "${RED}[WATCHEXEC]${NC} Automatic installation not available for your operating system: $OS"
      echo -e "${YELLOW}[WATCHEXEC]${NC} Please install watchexec manually from: https://github.com/watchexec/watchexec"
      return 1
      ;;
  esac

  # Verify installation
  if command -v watchexec &> /dev/null; then
    echo -e "${GREEN}[WATCHEXEC]${NC} watchexec installed successfully."
    return 0
  else
    echo -e "${RED}[WATCHEXEC]${NC} Failed to install watchexec."
    return 1
  fi
}

# Run the installation if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  install_watchexec
fi
