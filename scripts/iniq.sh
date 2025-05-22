#!/bin/bash
# iniq.sh - One-line initialization script for INIQ

# Helper function to download files
download_with_available_tool() {
    local url="$1"
    local output_file="$2"

    if command -v curl &>/dev/null; then
        curl -L "$url" -o "$output_file"
        return $?
    elif command -v wget &>/dev/null; then
        wget -O "$output_file" "$url"
        return $?
    else
        echo "Error: Neither curl nor wget found. Please install one of them." >&2
        return 1
    fi
}

# Check if the current user has sudo privileges
check_sudo() {
    # If we're already root, we don't need sudo
    if [ "$(id -u)" -eq 0 ]; then
        return 0
    fi

    # Try to run a simple sudo command to check if we have sudo privileges
    if sudo -n true 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Elevate privileges if needed
elevate_privileges() {
    local script_path="$1"
    shift
    local script_args=("$@")

    echo "INIQ requires sudo privileges for full functionality."
    echo "Requesting elevated privileges..."

    # Use sudo to run the script with preserved environment variables
    if sudo -E bash "$script_path" "${script_args[@]}"; then
        exit 0
    else
        echo "Failed to run with sudo. Continuing with limited functionality..."
        return 1
    fi
}

# Source common functions if running locally, otherwise use embedded functions
if [[ -f "$(dirname "$0")/common.sh" ]]; then
    source "$(dirname "$0")/common.sh"
else
    # We already have all common functions embedded in this script
    # No need to download common.sh

    # Create temporary directory for later use
    TEMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'iniq-tmp')

    # Set cleanup trap
    trap 'rm -rf "${TEMP_DIR}"' EXIT
fi

# Print banner
print_banner() {
    echo "=================================================="
    echo "  INIQ One-Line Initialization"
    echo "  A cross-platform system initialization tool"
    echo "=================================================="
    echo ""
}

# Parse command line arguments
parse_args() {
    # Process force download parameter
    FORCE_DOWNLOAD=false

    # Check if the first parameter is --force-download
    if [[ "$1" == "--force-download" ]]; then
        FORCE_DOWNLOAD=true
        shift
    fi

    # Process arguments to handle -- separator
    local args=()
    local skip_next=false
    local found_separator=false

    for arg in "$@"; do
        if [ "$skip_next" = true ]; then
            skip_next=false
            continue
        fi

        if [ "$arg" = "--" ]; then
            found_separator=true
            continue
        fi

        if [ "$found_separator" = true ]; then
            args+=("$arg")
        fi
    done

    # If we found the separator and have arguments after it, use those
    if [ "$found_separator" = true ] && [ ${#args[@]} -gt 0 ]; then
        INIQ_ARGS=("${args[@]}")
    else
        # Otherwise use all arguments
        INIQ_ARGS=("$@")
    fi

    # Uncomment for debugging
    # echo "Debug: Received arguments: ${INIQ_ARGS[*]}"
}

# Main function
run_iniq() {
    # Initialize script
    init_script

    # Create temporary directory for downloads
    TEMP_DIR=$(create_temp_dir)
    trap 'cleanup "${TEMP_DIR}"' EXIT

    # If force download is enabled, show message
    if [ "$FORCE_DOWNLOAD" = true ]; then
        echo "Forcing fresh download..."
    fi

    # Get download URL
    DOWNLOAD_URL=$(get_download_url)

    # Download binary (will use cache if available)
    download_binary "${DOWNLOAD_URL}" "${TEMP_DIR}/iniq"

    # Make the binary executable
    chmod +x "${TEMP_DIR}/iniq"

    # Create a temporary script that will run the binary
    # This avoids process substitution issues with sudo
    RUNNER_SCRIPT="${TEMP_DIR}/run_iniq.sh"
    cat > "$RUNNER_SCRIPT" << EOF
#!/bin/bash
# Temporary runner script for INIQ

# Run the binary with provided arguments
"${TEMP_DIR}/iniq" "\$@"
EOF
    chmod +x "$RUNNER_SCRIPT"

    # Check if we need sudo and don't already have it
    if ! check_sudo; then
        # We need sudo but don't have it, try to elevate privileges
        # Pass all arguments to the elevated script
        elevate_privileges "$RUNNER_SCRIPT" "${INIQ_ARGS[@]}"
        SUDO_RESULT=$?

        # If elevation failed, run with limited functionality
        if [ $SUDO_RESULT -ne 0 ]; then
            echo "Running with limited functionality (without sudo)..."
            # Add --skip-sudo flag to arguments if not already present
            if [[ ! " ${INIQ_ARGS[*]} " =~ " -S " ]] && [[ ! " ${INIQ_ARGS[*]} " =~ " --skip-sudo " ]]; then
                INIQ_ARGS+=("--skip-sudo")
            fi
        fi
    fi

    # If no arguments provided, use default recommended settings
    if [[ ${#INIQ_ARGS[@]} -eq 0 ]]; then
        # Default recommended settings with interactive mode
        "$RUNNER_SCRIPT"
    else
        # Run with provided arguments
        "$RUNNER_SCRIPT" "${INIQ_ARGS[@]}"
    fi
}

# Run the script
print_banner
parse_args "$@"
run_iniq
