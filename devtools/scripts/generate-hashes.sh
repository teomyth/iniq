#!/bin/bash
# generate-hashes.sh - Generate SHA-256 hash files for binaries

# Define colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Log function
log() {
    local level="$1"
    local message="$2"
    local color="${NC}"

    case "$level" in
        "info") color="${BLUE}" ;;
        "success") color="${GREEN}" ;;
        "warning") color="${YELLOW}" ;;
        "error") color="${RED}" ;;
    esac

    echo -e "${color}[HASH]${NC} $message"
}

# Generate hash for a single file
generate_hash() {
    local file="$1"
    local output_file="${file}.sha256"

    if [[ ! -f "$file" ]]; then
        log "error" "File not found: $file"
        return 1
    fi

    log "info" "Generating hash for: $file"

    if command -v sha256sum &>/dev/null; then
        sha256sum "$file" > "$output_file"
    elif command -v shasum &>/dev/null; then
        shasum -a 256 "$file" > "$output_file"
    else
        log "error" "Neither sha256sum nor shasum found. Cannot generate hash."
        return 1
    fi

    log "success" "Hash saved to: $output_file"
    return 0
}

# Generate hashes for all binaries in a directory
generate_hashes_for_dir() {
    local dir="$1"
    local pattern="${2:-*}"
    local exclude_pattern="${3:-*.sha256}"

    if [[ ! -d "$dir" ]]; then
        log "error" "Directory not found: $dir"
        return 1
    fi

    log "info" "Generating hashes for binaries in: $dir"

    local count=0
    for file in "$dir"/$pattern; do
        # Skip hash files and non-files
        if [[ -f "$file" && ! "$file" =~ \.sha256$ ]]; then
            generate_hash "$file"
            ((count++))
        fi
    done

    if [[ $count -eq 0 ]]; then
        log "warning" "No matching files found in: $dir"
        return 0
    fi

    log "success" "Generated hashes for $count files"
    return 0
}

# Verify hash for a single file
verify_hash() {
    local file="$1"
    local hash_file="${file}.sha256"

    if [[ ! -f "$file" ]]; then
        log "error" "File not found: $file"
        return 1
    fi

    if [[ ! -f "$hash_file" ]]; then
        log "error" "Hash file not found: $hash_file"
        return 1
    fi

    log "info" "Verifying hash for: $file"

    local result=0
    if command -v sha256sum &>/dev/null; then
        sha256sum -c "$hash_file" &>/dev/null
        result=$?
    elif command -v shasum &>/dev/null; then
        # Extract hash and filename from hash file
        local hash=$(cat "$hash_file" | cut -d ' ' -f 1)
        local calculated_hash=$(shasum -a 256 "$file" | cut -d ' ' -f 1)

        if [[ "$hash" == "$calculated_hash" ]]; then
            result=0
        else
            result=1
        fi
    else
        log "error" "Neither sha256sum nor shasum found. Cannot verify hash."
        return 1
    fi

    if [[ $result -eq 0 ]]; then
        log "success" "Hash verification passed for: $file"
    else
        log "error" "Hash verification failed for: $file"
    fi

    return $result
}

# Main function
main() {
    local command="${1:-generate}"
    local target="${2:-bin}"
    local pattern="${3:-*}"

    case "$command" in
        "generate")
            generate_hashes_for_dir "$target" "$pattern"
            ;;
        "verify")
            # Verify all files in directory
            local count=0
            local failed=0

            for file in "$target"/$pattern; do
                # Skip hash files and non-files
                if [[ -f "$file" && ! "$file" =~ \.sha256$ ]]; then
                    verify_hash "$file" || ((failed++))
                    ((count++))
                fi
            done

            if [[ $count -eq 0 ]]; then
                log "warning" "No matching files found in: $target"
                return 0
            fi

            if [[ $failed -eq 0 ]]; then
                log "success" "All $count files verified successfully"
                return 0
            else
                log "error" "$failed of $count files failed verification"
                return 1
            fi
            ;;
        *)
            log "error" "Unknown command: $command"
            echo "Usage: $0 [generate|verify] [directory] [pattern]"
            return 1
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
