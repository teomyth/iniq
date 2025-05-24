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

# Constants
GITHUB_REPO="teomyth/iniq"
GITHUB_RELEASES_BASE_URL="https://github.com/${GITHUB_REPO}/releases"
GITHUB_API_BASE_URL="https://api.github.com/repos/${GITHUB_REPO}"
BINARY_NAME="iniq"
DEFAULT_INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"

# Default values
DEFAULT_DOWNLOAD_BASE_URL="${GITHUB_RELEASES_BASE_URL}/latest/download"
GITHUB_API_URL="${GITHUB_API_BASE_URL}/releases/latest"

# Helper function to get binary filename pattern
get_binary_filename() {
    echo "${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
}

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

    # Check if platform is supported
    if [[ "$OS" != "linux" ]]; then
        echo -e "${RED}×${NC} Unsupported operating system: ${BOLD}$OS${NC}" >&2
        echo -e "${YELLOW}!${NC} INIQ currently only supports Linux." >&2
        echo -e "${YELLOW}!${NC} Supported platforms: linux-amd64, linux-arm64" >&2
        return 1
    fi

    if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" ]]; then
        echo -e "${RED}×${NC} Unsupported architecture: ${BOLD}$ARCH${NC}" >&2
        echo -e "${YELLOW}!${NC} Supported architectures: amd64, arm64" >&2
        return 1
    fi

    # Export variables
    export OS
    export ARCH
}

# Get latest version from GitHub API
get_latest_version() {
    local download_tool=$(check_download_tools)
    local version=""

    case "$download_tool" in
        curl)
            version=$(curl -s "$GITHUB_API_URL" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
            ;;
        wget)
            version=$(wget -qO- "$GITHUB_API_URL" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
            ;;
        *)
            return 1
            ;;
    esac

    if [[ -n "$version" && "$version" != "null" ]]; then
        echo "$version"
        return 0
    else
        return 1
    fi
}

# Get download URL based on platform
get_download_url() {
    local base_url="${1:-$DEFAULT_DOWNLOAD_BASE_URL}"
    echo "${base_url}/$(get_binary_filename)"
}

# Check if we're in development environment
is_development_environment() {
    # Check if the base URL is not the default GitHub releases URL
    if [[ "$DEFAULT_DOWNLOAD_BASE_URL" != "${GITHUB_RELEASES_BASE_URL}/latest/download" ]]; then
        return 0  # true - we're in development
    fi
    return 1  # false - we're not in development
}

# Get download URL with version fallback
get_download_url_with_fallback() {
    local base_url="${1:-$DEFAULT_DOWNLOAD_BASE_URL}"

    # Check if we're in development environment
    if is_development_environment; then
        echo -e "${BLUE}→${NC} Development environment detected, using local server" >&2
        echo "${base_url}/$(get_binary_filename)"
        return 0
    fi

    # Production environment - try to get the latest version from API
    echo -e "${BLUE}→${NC} Checking latest version from GitHub API..." >&2
    local latest_version=$(get_latest_version)

    if [[ -n "$latest_version" ]]; then
        echo -e "${BLUE}→${NC} Found latest version: ${BOLD}$latest_version${NC}" >&2
        local versioned_url="${GITHUB_RELEASES_BASE_URL}/download/${latest_version}/$(get_binary_filename)"
        echo "$versioned_url"
    else
        echo -e "${YELLOW}→${NC} Could not get version from API, using latest URL" >&2
        echo "${base_url}/$(get_binary_filename)"
    fi
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

# Download binary with fallback
download_binary() {
    local download_url="$1"
    local output_file="${2:-${BINARY_NAME}}"

    # Create a temporary file for download
    local temp_output_file="${output_file}.tmp.tar.gz"
    local extract_dir="${output_file}.extract"

    echo -e "${BLUE}→${NC} Downloading INIQ binary..."

    # First attempt with provided URL
    if download_file "$download_url" "$temp_output_file"; then
        # Verify the downloaded file is a valid gzip archive
        if file "$temp_output_file" | grep -q "gzip compressed"; then
            echo -e "${GREEN}✓${NC} Downloaded valid archive"
        else
            echo -e "${YELLOW}!${NC} Downloaded file is not a valid gzip archive, trying fallback..."
            rm -f "$temp_output_file"

            # Try to get versioned URL as fallback
            local fallback_url=$(get_download_url_with_fallback)
            if [[ "$fallback_url" != "$download_url" ]]; then
                echo -e "${BLUE}→${NC} Trying fallback URL..."
                if download_file "$fallback_url" "$temp_output_file"; then
                    if ! file "$temp_output_file" | grep -q "gzip compressed"; then
                        echo -e "${RED}×${NC} Fallback download also failed - not a valid gzip archive" >&2
                        rm -f "$temp_output_file"
                        return 1
                    fi
                    echo -e "${GREEN}✓${NC} Fallback download successful"
                else
                    echo -e "${RED}×${NC} Fallback download failed" >&2
                    return 1
                fi
            else
                echo -e "${RED}×${NC} No fallback URL available" >&2
                return 1
            fi
        fi
    else
        echo -e "${YELLOW}!${NC} Initial download failed, trying fallback..."

        # Try to get versioned URL as fallback
        local fallback_url=$(get_download_url_with_fallback)
        if [[ "$fallback_url" != "$download_url" ]]; then
            echo -e "${BLUE}→${NC} Trying fallback URL..."
            if ! download_file "$fallback_url" "$temp_output_file"; then
                echo -e "${RED}×${NC} Fallback download also failed" >&2
                return 1
            fi
            echo -e "${GREEN}✓${NC} Fallback download successful"
        else
            echo -e "${RED}×${NC} Download failed and no fallback URL available" >&2
            return 1
        fi
    fi

    # Create extraction directory
    mkdir -p "$extract_dir"

    # Extract tar.gz file
    echo -e "${BLUE}→${NC} Extracting archive..."
    if tar -xzf "$temp_output_file" -C "$extract_dir"; then
        # Find the binary in the extracted directory
        local binary_path=$(find "$extract_dir" -name "${BINARY_NAME}" -type f)

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
}

# Install binary to system path
install_binary() {
    local binary_file="${1:-${BINARY_NAME}}"
    local install_path="${2:-${DEFAULT_INSTALL_PATH}}"

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
    if ! detect_platform; then
        exit 1
    fi
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

    # Get download URL with fallback support
    DOWNLOAD_URL=$(get_download_url_with_fallback)
    echo -e "${BLUE}→${NC} Detected platform: ${BOLD}$OS-$ARCH${NC}"

    # Define install path
    INSTALL_PATH="${DEFAULT_INSTALL_PATH}"

    # Download binary
    print_step 2 3 "Installing Binary"
    download_binary "${DOWNLOAD_URL}" "${TEMP_DIR}/${BINARY_NAME}"

    # Install binary
    install_binary "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_PATH}"

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
