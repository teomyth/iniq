#!/bin/bash
# serve-http.sh - Script to start the development HTTP server

# Debug mode can be enabled by uncommenting the following line
# set -x

# Source common paths
source "$(dirname "$0")/common-paths.sh"

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
GRAY='\033[0;90m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Ensure directories exist
ensure_dirs

# Get status directory
STATUS_DIR=$(get_status_dir)

# Set initial status
echo "starting" > "$STATUS_DIR/http-server.status"

# Make sure the find-available-port script is executable
chmod +x devtools/scripts/find-available-port.sh

# Get port from environment, or use default
if [ -n "$1" ]; then
    # Use specified port if provided
    PORT=$1
    echo -e "${BLUE}[SERVER]${NC} Using specified port: $PORT"
else
    # Use find-available-port.sh to find an available port starting from 12345
    PORT=$(./devtools/scripts/find-available-port.sh 12345)
    if [ $? -ne 0 ]; then
        # If find-available-port.sh fails, fall back to a random port
        PORT=$((20000 + RANDOM % 10000))
        echo -e "${YELLOW}[SERVER]${NC} Failed to find available port, using random port: $PORT"
    else
        echo -e "${BLUE}[SERVER]${NC} Using available port: $PORT"
    fi
fi

echo -e "${BLUE}[SERVER]${NC} Starting server on port: $PORT"

# Create a port file for other scripts to use
echo $PORT > "$DEV_SERVER_PORT_FILE"

# Function to update server status
update_status() {
    echo "$1" > "$STATUS_DIR/http-server.status"
}

# Start server with custom prefix for its output and monitor its status
(
    # Set trap to update status on exit
    trap 'update_status "stopped"; exit' EXIT INT TERM

    # Start the server
    update_status "starting"

    # Initialize counter for auto-detection
    i=0

    # Run the server and capture its output
    (
        # Add debug output
        echo -e "${YELLOW}[SERVER DEBUG]${NC} Starting HTTP server with command: ./bin/devserver --port $PORT --bin bin"
        echo -e "${YELLOW}[SERVER DEBUG]${NC} Current directory: $(pwd)"
        echo -e "${YELLOW}[SERVER DEBUG]${NC} Binary exists: $([ -f ./bin/devserver ] && echo "Yes" || echo "No")"

        # Use the pre-compiled binary with 0.0.0.0 binding for WSL mirror mode
        # This ensures we listen on all interfaces and can detect all available IPs
        ./bin/devserver --port $PORT --bin bin 2>&1 |
        while IFS= read -r line; do
            # Check if the line indicates server is ready
            if [[ "$line" == *"Starting development server"* || "$line" == *"Server started"* || "$line" == *"Listening"* || "$line" == *"Available on"* ]]; then
                update_status "running"
                echo -e "${GREEN}[SERVER]${NC} HTTP server is ready and listening on port $PORT"
            fi

            # Also detect the "Available on" section which indicates server is running
            if [[ "$line" == *"Available on"* ]]; then
                update_status "running"
                echo -e "${GREEN}[SERVER]${NC} HTTP server detected available network interfaces"
            fi

            # Check for errors
            if [[ "$line" == *"error"* || "$line" == *"Error"* || "$line" == *"ERROR"* || "$line" == *"bind: address already in use"* ]]; then
                update_status "error"
                echo -e "${RED}[SERVER]${NC} $line"
            else
                # Only show important server messages, filter out empty lines
                if [[ "$line" != "" ]]; then
                    echo -e "${BLUE}[SERVER]${NC} $line"
                fi
            fi

            # Increment counter
            i=$((i+1))

            # Force status to running after a short delay if not already set
            if [ $i -eq 5 ]; then
                if [ "$(cat "$STATUS_DIR/http-server.status" 2>/dev/null)" = "starting" ]; then
                    update_status "running"
                    echo -e "${YELLOW}[SERVER]${NC} Assuming HTTP server is running (no explicit ready message detected)"
                fi
            fi
        done
    ) &
    SERVER_PID=$!

    # Save PID for cleanup
    echo "$SERVER_PID" > "$STATUS_DIR/http-server.pid"

    # Wait a moment for the server to start
    sleep 1

    # Check initial status
    for i in {1..10}; do
        if [ -f "$STATUS_DIR/http-server.status" ]; then
            STATUS=$(cat "$STATUS_DIR/http-server.status")
            if [ "$STATUS" = "running" ]; then
                echo -e "${GREEN}[SERVER]${NC} HTTP server is running successfully on port $PORT"
                # Create a flag file to indicate server is fully initialized
                touch "$STATUS_DIR/http-server.initialized"
                break
            elif [ "$STATUS" = "error" ]; then
                echo -e "${RED}[SERVER]${NC} HTTP server encountered an error"
                break
            fi
        fi
        sleep 1
    done
)
