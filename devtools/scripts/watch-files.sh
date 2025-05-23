#!/bin/bash
# watch-files.sh - Script to watch for file changes and automatically rebuild

# Source common paths
source "$(dirname "$0")/common-paths.sh"

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
GRAY='\033[0;90m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to get current timestamp
get_timestamp() {
    date "+%Y-%m-%d %H:%M:%S"
}

# Function to log with timestamp
log_with_timestamp() {
    local level="$1"
    local message="$2"
    local timestamp=$(get_timestamp)
    echo -e "${GRAY}[${timestamp}]${NC} $level $message"
}

# Ensure directories exist
ensure_dirs

# Get status directory
STATUS_DIR=$(get_status_dir)

# Set initial status
echo "starting" > "$STATUS_DIR/watcher.status"

# Function to update watcher status
update_status() {
    echo "$1" > "$STATUS_DIR/watcher.status"
    if [ "$1" = "error" ]; then
        echo "$2" > "$STATUS_DIR/watcher.error"
    fi
}

# Start the watcher with status updates
start_watcher() {
    update_status "running"
    log_with_timestamp "${GREEN}[WATCH]${NC}" "File watcher is running"

    if command -v watchexec &> /dev/null; then
        # Use watchexec for file watching
        log_with_timestamp "${BLUE}[WATCH]${NC}" "Starting file watcher with watchexec"

        # Make sure the binary directory exists
        mkdir -p bin

        # Start watchexec for install.sh in root directory
        (
            log_with_timestamp "${BLUE}[WATCH]${NC}" "Watching install.sh for changes"
            watchexec -w install.sh --verbose --print-events "echo '[SCRIPTS] Rebuilding scripts...' && task build:scripts && echo '[SCRIPTS] Scripts rebuilt successfully'" 2>&1 |
            while IFS= read -r line; do
                # Check for errors
                if [[ "$line" == *"error"* || "$line" == *"Error"* || "$line" == *"ERROR"* ]]; then
                    log_with_timestamp "${RED}[WATCH:SCRIPTS]${NC}" "$line"
                elif [[ "$line" == *"Rebuilding scripts"* ]]; then
                    log_with_timestamp "${YELLOW}[WATCH:SCRIPTS]${NC}" "Rebuilding scripts..."
                elif [[ "$line" == *"Scripts rebuilt successfully"* ]]; then
                    log_with_timestamp "${GREEN}[WATCH:SCRIPTS]${NC}" "Scripts rebuilt successfully"
                elif [[ "$line" == *"event"* && "$line" == *"path"* ]]; then
                    # Extract file path from event line
                    FILE_PATH=$(echo "$line" | grep -o 'path="[^"]*"' | sed 's/path="//;s/"$//')
                    EVENT_TYPE=$(echo "$line" | grep -o 'event=[^ ]*' | sed 's/event=//')
                    if [[ -n "$FILE_PATH" && -n "$EVENT_TYPE" ]]; then
                        log_with_timestamp "${BLUE}[WATCH:SCRIPTS]${NC}" "Detected ${EVENT_TYPE} on file: ${FILE_PATH}"
                    else
                        log_with_timestamp "${BLUE}[WATCH:SCRIPTS]${NC}" "$line"
                    fi
                else
                    log_with_timestamp "${BLUE}[WATCH:SCRIPTS]${NC}" "$line"
                fi
            done
        ) &
        SCRIPTS_WATCHER_PID=$!
        echo "$SCRIPTS_WATCHER_PID" > "$STATUS_DIR/scripts_watcher.pid"

        # Start watchexec for Go source files
        (
            log_with_timestamp "${BLUE}[WATCH]${NC}" "Watching cmd/ and internal/ directories for changes"
            watchexec -w cmd/ -w internal/ -w pkg/ -e go --verbose --print-events "echo '[GO] Rebuilding Go binaries...' && task goreleaser:build && echo '[GO] Copying binaries to bin directory...' && devtools/scripts/copy-binaries.sh && echo '[GO] Generating hashes...' && task build:hashes && echo '[GO] Go binaries rebuilt successfully'" 2>&1 |
            while IFS= read -r line; do
                # Check for errors
                if [[ "$line" == *"error"* || "$line" == *"Error"* || "$line" == *"ERROR"* ]]; then
                    log_with_timestamp "${RED}[WATCH:GO]${NC}" "$line"
                elif [[ "$line" == *"Rebuilding Go binaries"* ]]; then
                    log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Rebuilding Go binaries..."
                elif [[ "$line" == *"Copying binaries to bin directory"* ]]; then
                    log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Copying binaries to bin directory..."
                elif [[ "$line" == *"Generating hashes"* ]]; then
                    log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Generating hashes..."
                elif [[ "$line" == *"Go binaries rebuilt successfully"* ]]; then
                    log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Go binaries rebuilt successfully"
                elif [[ "$line" == *"event"* && "$line" == *"path"* ]]; then
                    # Extract file path from event line
                    FILE_PATH=$(echo "$line" | grep -o 'path="[^"]*"' | sed 's/path="//;s/"$//')
                    EVENT_TYPE=$(echo "$line" | grep -o 'event=[^ ]*' | sed 's/event=//')
                    if [[ -n "$FILE_PATH" && -n "$EVENT_TYPE" ]]; then
                        log_with_timestamp "${BLUE}[WATCH:GO]${NC}" "Detected ${EVENT_TYPE} on file: ${FILE_PATH}"
                    else
                        log_with_timestamp "${BLUE}[WATCH:GO]${NC}" "$line"
                    fi

                    # Verify the binaries were actually built
                    # Note: GoReleaser only builds Linux binaries, macOS binaries are for development only
                    LINUX_PLATFORMS=("linux-amd64" "linux-arm64")
                    MACOS_PLATFORMS=("darwin-amd64" "darwin-arm64")
                    ALL_BUILT=true
                    MISSING_PLATFORMS=""

                    # Check Linux binaries (should always be built by GoReleaser)
                    for platform in "${LINUX_PLATFORMS[@]}"; do
                        if [ -f "bin/iniq-${platform}" ]; then
                            BINARY_TIME=$(stat -c %y "bin/iniq-${platform}" 2>/dev/null || stat -f "%m" "bin/iniq-${platform}" 2>/dev/null)
                            log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Binary for ${platform} successfully built (${BINARY_TIME})"
                        else
                            ALL_BUILT=false
                            MISSING_PLATFORMS="${MISSING_PLATFORMS} ${platform}"
                        fi
                    done

                    # Check macOS binaries (development only - may not exist from GoReleaser)
                    for platform in "${MACOS_PLATFORMS[@]}"; do
                        if [ -f "bin/iniq-${platform}" ]; then
                            BINARY_TIME=$(stat -c %y "bin/iniq-${platform}" 2>/dev/null || stat -f "%m" "bin/iniq-${platform}" 2>/dev/null)
                            log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Binary for ${platform} available (development only) (${BINARY_TIME})"
                        else
                            log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Binary for ${platform} not available (development only - run 'task build:${platform}' to build)"
                        fi
                    done

                    # Also check the current platform binary
                    if [ -f bin/iniq ]; then
                        BINARY_TIME=$(stat -c %y bin/iniq 2>/dev/null || stat -f "%m" bin/iniq 2>/dev/null)
                        log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Current platform binary successfully built at bin/iniq (${BINARY_TIME})"
                    else
                        ALL_BUILT=false
                        log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Warning: Current platform binary not found at bin/iniq"
                    fi

                    if [ "$ALL_BUILT" = false ]; then
                        log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Warning: Some platform binaries are missing:${MISSING_PLATFORMS}"
                    else
                        log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "All platform binaries successfully built"
                    fi
                else
                    log_with_timestamp "${BLUE}[WATCH:GO]${NC}" "$line"
                fi
            done
        ) &
        GO_WATCHER_PID=$!
        echo "$GO_WATCHER_PID" > "$STATUS_DIR/go_watcher.pid"

        # Start watchexec for devtools
        (
            log_with_timestamp "${BLUE}[WATCH]${NC}" "Watching devtools/ directory for changes"
            watchexec -w devtools/ -e go --verbose --print-events "echo '[DEVTOOLS] Rebuilding development tools...' && task build:devserver && echo '[DEVTOOLS] Development tools rebuilt successfully'" 2>&1 |
            while IFS= read -r line; do
                # Check for errors
                if [[ "$line" == *"error"* || "$line" == *"Error"* || "$line" == *"ERROR"* ]]; then
                    log_with_timestamp "${RED}[WATCH:DEVTOOLS]${NC}" "$line"
                elif [[ "$line" == *"Rebuilding development tools"* ]]; then
                    log_with_timestamp "${YELLOW}[WATCH:DEVTOOLS]${NC}" "Rebuilding development tools..."
                elif [[ "$line" == *"Development tools rebuilt successfully"* ]]; then
                    log_with_timestamp "${GREEN}[WATCH:DEVTOOLS]${NC}" "Development tools rebuilt successfully"
                elif [[ "$line" == *"event"* && "$line" == *"path"* ]]; then
                    # Extract file path from event line
                    FILE_PATH=$(echo "$line" | grep -o 'path="[^"]*"' | sed 's/path="//;s/"$//')
                    EVENT_TYPE=$(echo "$line" | grep -o 'event=[^ ]*' | sed 's/event=//')
                    if [[ -n "$FILE_PATH" && -n "$EVENT_TYPE" ]]; then
                        log_with_timestamp "${BLUE}[WATCH:DEVTOOLS]${NC}" "Detected ${EVENT_TYPE} on file: ${FILE_PATH}"
                    else
                        log_with_timestamp "${BLUE}[WATCH:DEVTOOLS]${NC}" "$line"
                    fi
                else
                    log_with_timestamp "${BLUE}[WATCH:DEVTOOLS]${NC}" "$line"
                fi
            done
        ) &
        DEVTOOLS_WATCHER_PID=$!
        echo "$DEVTOOLS_WATCHER_PID" > "$STATUS_DIR/devtools_watcher.pid"

        # Keep this process running (use a large number instead of infinity for macOS compatibility)
        sleep 2147483647
    else
        # Fallback to inotifywait if watchexec is not available
        log_with_timestamp "${YELLOW}[WATCH]${NC}" "watchexec not found. Falling back to inotifywait."

        if command -v inotifywait &> /dev/null; then
            log_with_timestamp "${BLUE}[WATCH]${NC}" "Setting up file watcher with inotifywait"
            (
                while true; do
                    inotifywait -q -e modify -e create -e delete -r --exclude '(\.git|bin|tmp)' .
                    log_with_timestamp "${YELLOW}[WATCH]${NC}" "File change detected, rebuilding..."

                    # Check which files changed and rebuild accordingly
                    if [ install.sh -nt "$STATUS_DIR/last_build" 2>/dev/null ]; then
                        log_with_timestamp "${YELLOW}[WATCH:SCRIPTS]${NC}" "Script changes detected, rebuilding scripts..."
                        task build:scripts
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${GREEN}[WATCH:SCRIPTS]${NC}" "Scripts rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:SCRIPTS]${NC}" "Scripts rebuild failed"
                        fi
                    fi

                    if [ -n "$(find cmd/ internal/ pkg/ -name "*.go" -newer "$STATUS_DIR/last_build" 2>/dev/null)" ]; then
                        log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Go source changes detected, rebuilding binaries..."
                        task goreleaser:build
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Copying binaries to bin directory..."
                            devtools/scripts/copy-binaries.sh
                            log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Generating hashes..."
                            task build:hashes
                            log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Go binaries rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:GO]${NC}" "Go binaries rebuild failed"
                        fi
                    fi

                    if [ -n "$(find devtools/ -name "*.go" -newer "$STATUS_DIR/last_build" 2>/dev/null)" ]; then
                        log_with_timestamp "${YELLOW}[WATCH:DEVTOOLS]${NC}" "Devtools changes detected, rebuilding development tools..."
                        task build:devserver
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${GREEN}[WATCH:DEVTOOLS]${NC}" "Development tools rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:DEVTOOLS]${NC}" "Development tools rebuild failed"
                        fi
                    fi

                    # Update last build timestamp
                    touch "$STATUS_DIR/last_build"

                    # Small delay to avoid multiple rebuilds
                    sleep 1
                done
            ) &
            INOTIFY_PID=$!
            echo "$INOTIFY_PID" > "$STATUS_DIR/inotify.pid"
        else
            log_with_timestamp "${YELLOW}[WATCH]${NC}" "inotifywait not found. Using fallback polling method."
            (
                while true; do
                    log_with_timestamp "${BLUE}[WATCH]${NC}" "Checking for file changes..."

                    # Check which files changed and rebuild accordingly
                    if [ install.sh -nt "$STATUS_DIR/last_build" 2>/dev/null ]; then
                        log_with_timestamp "${YELLOW}[WATCH:SCRIPTS]${NC}" "Script changes detected, rebuilding scripts..."
                        task build:scripts
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${GREEN}[WATCH:SCRIPTS]${NC}" "Scripts rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:SCRIPTS]${NC}" "Scripts rebuild failed"
                        fi
                    fi

                    if [ -n "$(find cmd/ internal/ pkg/ -name "*.go" -newer "$STATUS_DIR/last_build" 2>/dev/null)" ]; then
                        log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Go source changes detected, rebuilding binaries..."
                        task goreleaser:build
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Copying binaries to bin directory..."
                            devtools/scripts/copy-binaries.sh
                            log_with_timestamp "${YELLOW}[WATCH:GO]${NC}" "Generating hashes..."
                            task build:hashes
                            log_with_timestamp "${GREEN}[WATCH:GO]${NC}" "Go binaries rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:GO]${NC}" "Go binaries rebuild failed"
                        fi
                    fi

                    if [ -n "$(find devtools/ -name "*.go" -newer "$STATUS_DIR/last_build" 2>/dev/null)" ]; then
                        log_with_timestamp "${YELLOW}[WATCH:DEVTOOLS]${NC}" "Devtools changes detected, rebuilding development tools..."
                        task build:devserver
                        if [ $? -eq 0 ]; then
                            log_with_timestamp "${GREEN}[WATCH:DEVTOOLS]${NC}" "Development tools rebuilt successfully"
                        else
                            log_with_timestamp "${RED}[WATCH:DEVTOOLS]${NC}" "Development tools rebuild failed"
                        fi
                    fi

                    # Update last build timestamp
                    touch "$STATUS_DIR/last_build"

                    # Sleep for a while before checking again
                    sleep 10
                done
            ) &
            POLL_PID=$!
            echo "$POLL_PID" > "$STATUS_DIR/poll.pid"
        fi

        # Keep this process running (use a large number instead of infinity for macOS compatibility)
        sleep 2147483647
    fi
}

# Set trap to update status on exit
trap 'update_status "stopped"; exit' EXIT INT TERM

# Start the watcher
start_watcher
