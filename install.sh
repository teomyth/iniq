#!/bin/bash
# install.sh - INIQ installation script
# This is a standalone installation script that includes all necessary functions

# Set strict error handling
set -e

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Default values
DEFAULT_DOWNLOAD_BASE_URL="https://github.com/teomyth/iniq/releases/latest/download"

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

# Detect OS and architecture
detect_platform() {
    # Detect OS
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')

    # Detect architecture
    ARCH=$(uname -m)

    # Map architecture to Go naming convention
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l)  ARCH="arm" ;;
        i386)    ARCH="386" ;;
    esac

    # Export variables
    export OS
    export ARCH
}

# Get download URL based on platform
get_download_url() {
    local base_url="${1:-$DEFAULT_DOWNLOAD_BASE_URL}"
    echo "${base_url}/iniq-${OS}-${ARCH}.tar.gz"
}

# Check for download tools
check_download_tools() {
    if command -v curl &>/dev/null; then
        echo "curl"
        return 0
    elif command -v wget &>/dev/null; then
        echo "wget"
        return 0
    else
        echo "Error: Neither curl nor wget found. Please install one of them." >&2
        return 1
    fi
}

# Download file using available tool with retry
download_file() {
    local url="$1"
    local output_file="$2"
    local download_tool=$(check_download_tools)
    local max_retries=3
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        case "$download_tool" in
            curl)
                if curl -L --fail "$url" -o "$output_file"; then
                    return 0
                fi
                ;;
            wget)
                if wget -O "$output_file" "$url"; then
                    return 0
                fi
                ;;
            *)
                return 1
                ;;
        esac

        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo -e "${YELLOW}→${NC} Download failed, retrying ($retry_count/$max_retries)..."
            sleep 2
        fi
    done

    return 1
}

# Download binary
download_binary() {
    local download_url="$1"
    local output_file="${2:-iniq}"

    # Create a temporary file for download
    local temp_output_file="${output_file}.tmp.tar.gz"
    local extract_dir="${output_file}.extract"

    echo -e "${BLUE}→${NC} Downloading INIQ binary..."
    if download_file "$download_url" "$temp_output_file"; then
        # Verify the downloaded file is a valid gzip archive
        if ! file "$temp_output_file" | grep -q "gzip compressed"; then
            echo -e "${RED}×${NC} Downloaded file is not a valid gzip archive" >&2
            echo -e "${YELLOW}!${NC} This might be a GitHub redirect issue. Try using a specific version URL." >&2
            rm -f "$temp_output_file"
            return 1
        fi

        # Create extraction directory
        mkdir -p "$extract_dir"

        # Extract tar.gz file
        echo -e "${BLUE}→${NC} Extracting archive..."
        if tar -xzf "$temp_output_file" -C "$extract_dir"; then
            # Find the binary in the extracted directory
            local binary_path=$(find "$extract_dir" -name "iniq" -type f)

            if [[ -n "$binary_path" ]]; then
                echo -e "${GREEN}✓${NC} Archive extracted successfully"
                # Copy the binary to the output location
                cp "$binary_path" "$output_file"
                chmod +x "$output_file"
            else
                echo -e "${RED}×${NC} Binary not found in extracted archive" >&2
                rm -rf "$extract_dir" "$temp_output_file"
                return 1
            fi
        else
            echo -e "${RED}×${NC} Failed to extract archive" >&2
            rm -f "$temp_output_file"
            return 1
        fi

        # Clean up extraction directory and temp file
        rm -rf "$extract_dir" "$temp_output_file"
        echo -e "${GREEN}✓${NC} Download completed successfully"
        return 0
    else
        echo -e "${RED}×${NC} Failed to download INIQ binary" >&2
        rm -f "$temp_output_file"
        return 1
    fi
}

# Install binary to system path
install_binary() {
    local binary_file="${1:-iniq}"
    local install_path="${2:-/usr/local/bin/iniq}"

    # Create a temporary file for installation
    local temp_install_file="${install_path}.tmp"

    echo -e "${BLUE}→${NC} Installing INIQ binary to: ${BOLD}${install_path}${NC}"

    # Copy to temporary location first
    cp "${binary_file}" "${temp_install_file}"
    chmod +x "${temp_install_file}"

    # Check if we need sudo for the move operation
    if [[ -w "$(dirname "${install_path}")" ]]; then
        # We have write permission, move directly
        mv "${temp_install_file}" "${install_path}"
        echo -e "${GREEN}✓${NC} Binary installed successfully"
    else
        # We need sudo
        echo -e "${YELLOW}!${NC} Elevated permissions required to install to ${BOLD}${install_path}${NC}"
        sudo mv "${temp_install_file}" "${install_path}"
        echo -e "${GREEN}✓${NC} Binary installed successfully with elevated permissions"
    fi
}

# Clean up temporary files
cleanup() {
    local temp_dir="$1"
    if [[ -n "${temp_dir}" && -d "${temp_dir}" ]]; then
        rm -rf "${temp_dir}"
    fi
}

# Create temporary directory
create_temp_dir() {
    mktemp -d 2>/dev/null || mktemp -d -t 'iniq-tmp'
}

# Initialize script
init_script() {
    detect_platform
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

    # Download binary
    print_step 2 3 "Installing Binary"
    download_binary "${DOWNLOAD_URL}" "${TEMP_DIR}/iniq"

    # Install binary
    install_binary "${TEMP_DIR}/iniq" "${INSTALL_PATH}"

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
