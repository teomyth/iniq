#!/bin/bash
# install.sh - INIQ installation script

# Note: common.sh is now directly included in this script during the build process
# No need to download it separately

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Print banner
print_banner() {
    echo -e "${BOLD}INIQ Installer${NC}"
    echo -e "────────────────────────────────────────────"
}

# Print step header
print_step() {
    local step_num=$1
    local total_steps=$2
    local step_name=$3
    echo -e "\n${BLUE}[$step_num/$total_steps]${NC} ${BOLD}$step_name${NC}"
    echo -e "────────────────────────────────────────"
}

# Main installation function
install_iniq() {
    # Initialize script
    init_script

    # Create temporary directory for downloads
    print_step 1 3 "Preparing Environment"
    TEMP_DIR=$(create_temp_dir)
    trap 'cleanup "${TEMP_DIR}"' EXIT
    echo -e "${BLUE}→${NC} Created temporary directory"

    # Get download URL
    DOWNLOAD_URL=$(get_download_url)
    echo -e "${BLUE}→${NC} Detected platform: ${BOLD}$OS-$ARCH${NC}"

    # Define install path
    INSTALL_PATH="/usr/local/bin/iniq"

    # Download binary (will use cache if available)
    print_step 2 3 "Installing Binary"
    download_binary "${DOWNLOAD_URL}" "${TEMP_DIR}/iniq" "false" "${INSTALL_PATH}"

    # Install binary
    install_binary "${TEMP_DIR}/iniq"

    # Print success message
    print_step 3 3 "Finalizing Installation"
    echo -e "${GREEN}✓${NC} INIQ installed successfully to ${BOLD}${INSTALL_PATH}${NC}"

    # Print completion message with clear separation
    echo -e "\n────────────────────────────────────────"
    echo -e "${GREEN}${BOLD}Installation Complete!${NC}"
    echo -e "Run '${BOLD}iniq --help${NC}' to see available options."
    echo -e "────────────────────────────────────────\n"
}

# Run the installation
print_banner
install_iniq
