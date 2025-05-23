#!/bin/bash
# common.sh - Common functions for INIQ installation scripts

# Set strict error handling
set -e

# Default values - will be replaced by the development server in dev mode
# In production, this is set to the GitHub release download URL
DEFAULT_DOWNLOAD_BASE_URL="https://github.com/teomyth/iniq/releases/download/latest"

# Detect if we're in a development environment
is_development() {
    # Extract domain from DEFAULT_DOWNLOAD_BASE_URL
    local domain=$(echo "$DEFAULT_DOWNLOAD_BASE_URL" | sed -E 's|^https?://([^/]+).*|\1|')

    # Check if domain contains development indicators
    [[ "$domain" == *"localhost"* ]] ||
    [[ "$domain" == *"127.0.0.1"* ]] ||
    [[ "$domain" == *"trycloudflare.com"* ]] ||
    [[ "$domain" == *"dev"* ]]
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

    # Always use the full base URL including protocol
    # For GitHub releases, the URL format is:
    # https://github.com/user/repo/releases/download/tag/filename
    echo "${base_url}/iniq-${OS}-${ARCH}.tar.gz"
}

# This function has been removed as we don't want to assume GitHub hosting

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

# Download file using available tool
download_file() {
    local url="$1"
    local output_file="$2"
    local download_tool=$(check_download_tools)

    case "$download_tool" in
        curl)
            curl -L "$url" -o "$output_file"
            ;;
        wget)
            wget -O "$output_file" "$url"
            ;;
        *)
            return 1
            ;;
    esac

    return $?
}

# Calculate SHA-256 hash of a file
calculate_hash() {
    local file="$1"

    if command -v sha256sum &>/dev/null; then
        sha256sum "$file" | cut -d ' ' -f 1
    elif command -v shasum &>/dev/null; then
        shasum -a 256 "$file" | cut -d ' ' -f 1
    else
        echo "Error: Neither sha256sum nor shasum found. Cannot verify binary integrity." >&2
        return 1
    fi
}

# Download hash file for binary
download_hash() {
    local download_url="$1"
    local output_file="${2:-iniq.sha256}"

    # Construct hash URL by adding .sha256 extension
    local hash_url="${download_url}.sha256"

    # Define colors if not already defined
    GREEN=${GREEN:-'\033[0;32m'}
    BLUE=${BLUE:-'\033[0;34m'}
    YELLOW=${YELLOW:-'\033[0;33m'}
    RED=${RED:-'\033[0;31m'}
    NC=${NC:-'\033[0m'}

    echo -e "${BLUE}→${NC} Downloading hash file..."
    if download_file "$hash_url" "$output_file"; then
        echo -e "${GREEN}✓${NC} Hash file downloaded successfully"
        return 0
    else
        echo -e "${YELLOW}!${NC} Failed to download hash file. Will skip hash verification." >&2
        return 1
    fi
}

# Verify binary against hash file
verify_binary() {
    local binary_file="$1"
    local hash_file="$2"

    # Define colors if not already defined
    GREEN=${GREEN:-'\033[0;32m'}
    BLUE=${BLUE:-'\033[0;34m'}
    YELLOW=${YELLOW:-'\033[0;33m'}
    RED=${RED:-'\033[0;31m'}
    NC=${NC:-'\033[0m'}

    if [[ ! -f "$binary_file" ]]; then
        echo -e "${RED}×${NC} Binary file not found: $binary_file" >&2
        return 1
    fi

    if [[ ! -f "$hash_file" ]]; then
        echo -e "${RED}×${NC} Hash file not found: $hash_file" >&2
        return 1
    fi

    echo -e "${BLUE}→${NC} Verifying binary integrity..."

    # Extract expected hash from hash file (first field)
    local expected_hash=$(cat "$hash_file" | awk '{print $1}')

    # Calculate actual hash of binary
    local actual_hash=$(calculate_hash "$binary_file")

    if [[ "$expected_hash" == "$actual_hash" ]]; then
        echo -e "${GREEN}✓${NC} Binary integrity verified successfully"
        return 0
    else
        echo -e "${RED}×${NC} Binary integrity verification failed!" >&2
        echo -e "  Expected: $expected_hash" >&2
        echo -e "  Actual:   $actual_hash" >&2
        return 1
    fi
}

# Download binary
download_binary() {
    local download_url="$1"
    local output_file="${2:-iniq}"
    local force_download="${3:-false}"
    local install_path="${4:-/usr/local/bin/iniq}"

    # Define colors if not already defined
    GREEN=${GREEN:-'\033[0;32m'}
    BLUE=${BLUE:-'\033[0;34m'}
    YELLOW=${YELLOW:-'\033[0;33m'}
    RED=${RED:-'\033[0;31m'}
    BOLD=${BOLD:-'\033[1m'}
    NC=${NC:-'\033[0m'}

    # Hash file path
    local hash_file="${output_file}.sha256"
    local hash_downloaded=false
    local hash_verified=false

    # First, download the hash file once to use for all verifications
    if download_hash "$download_url" "$hash_file"; then
        hash_downloaded=true
    else
        echo -e "${YELLOW}!${NC} Could not download hash file. Verification will be skipped." >&2
    fi

    # Check if binary already exists in the final install location
    if [[ -f "$install_path" && "$force_download" != "true" ]]; then
        echo -e "${BLUE}→${NC} Binary already exists at ${BOLD}$install_path${NC}"

        # Verify existing binary if we have the hash file
        if [[ "$hash_downloaded" == "true" ]]; then
            # Copy existing binary to temp location for verification
            cp "$install_path" "$output_file.existing"

            # Verify existing binary
            if verify_binary "$output_file.existing" "$hash_file"; then
                echo -e "${GREEN}✓${NC} Existing binary is up-to-date"
                # Copy the existing binary to the output location
                cp "$install_path" "$output_file"
                rm -f "$output_file.existing"
                return 0
            else
                echo -e "${YELLOW}!${NC} Existing binary is outdated or corrupted"
                rm -f "$output_file.existing"
            fi
        else
            echo -e "${YELLOW}!${NC} Cannot verify existing binary without hash file"
        fi
    fi

    # Create a temporary file for download
    local temp_output_file="${output_file}.tmp.tar.gz"
    local extract_dir="${output_file}.extract"

    echo -e "${BLUE}→${NC} Downloading INIQ binary..."
    if download_file "$download_url" "$temp_output_file"; then
        # Create extraction directory
        mkdir -p "$extract_dir"
        
        # Extract tar.gz file
        echo -e "${BLUE}→${NC} Extracting archive..."
        if tar -xzf "$temp_output_file" -C "$extract_dir"; then
            # Find the binary in the extracted directory
            local binary_path=$(find "$extract_dir" -name "iniq" -type f)
            
            if [[ -n "$binary_path" ]]; then
                echo -e "${GREEN}✓${NC} Archive extracted successfully"
                # Copy the binary to the temporary output location
                cp "$binary_path" "${output_file}.tmp"
                chmod +x "${output_file}.tmp"
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

        # Clean up extraction directory
        rm -rf "$extract_dir" "$temp_output_file"

        # Verify downloaded binary if we have the hash file
        if [[ "$hash_downloaded" == "true" ]]; then
            # Verify downloaded binary
            if verify_binary "$temp_output_file" "$hash_file"; then
                hash_verified=true
            else
                echo -e "${RED}×${NC} Binary verification failed. The binary might be corrupted." >&2
                echo -e "${YELLOW}!${NC} Do you want to continue anyway? (y/N): " >&2
                read -r response
                if [[ ! "$response" =~ ^[Yy]$ ]]; then
                    echo -e "${RED}×${NC} Installation aborted." >&2
                    rm -f "$temp_output_file"
                    return 1
                fi
            fi
        else
            echo -e "${YELLOW}!${NC} Skipping verification due to missing hash file." >&2
        fi

        # Move binary to final location (atomic operation)
        mv "$temp_output_file" "$output_file"

        if [[ "$hash_verified" == "true" ]]; then
            echo -e "${GREEN}✓${NC} Download completed and binary integrity verified"
        else
            echo -e "${YELLOW}!${NC} Download completed but binary integrity could not be verified"
        fi

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

    # Define colors if not already defined
    GREEN=${GREEN:-'\033[0;32m'}
    BLUE=${BLUE:-'\033[0;34m'}
    YELLOW=${YELLOW:-'\033[0;33m'}
    RED=${RED:-'\033[0;31m'}
    BOLD=${BOLD:-'\033[1m'}
    NC=${NC:-'\033[0m'}

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

# Set up environment variables for development mode
setup_dev_environment() {
    # No environment variables needed as we use the full URL directly
    :
}

# Initialize script
init_script() {
    # Detect platform
    detect_platform

    # Set up development environment if needed
    setup_dev_environment
}
