#!/bin/bash
# find-available-port.sh - Find an available port starting from a given base port

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

# Default starting port
START_PORT=${1:-10000}
MAX_PORT=65535

# Function to check if a port is available
is_port_available() {
    local port=$1

    # Try to bind to the port using netcat if available
    if command -v nc &> /dev/null; then
        nc -z 127.0.0.1 $port &> /dev/null
        if [ $? -ne 0 ]; then
            return 0  # Port is available
        else
            return 1  # Port is in use
        fi
    fi

    # Alternative check using lsof if available
    if command -v lsof &> /dev/null; then
        lsof -i :$port &> /dev/null
        if [ $? -ne 0 ]; then
            return 0  # Port is available
        else
            return 1  # Port is in use
        fi
    fi

    # Alternative check using ss if available
    if command -v ss &> /dev/null; then
        ss -tuln | grep ":$port " &> /dev/null
        if [ $? -ne 0 ]; then
            return 0  # Port is available
        else
            return 1  # Port is in use
        fi
    fi

    # Last resort: try to bind to the port using bash
    (echo > /dev/tcp/127.0.0.1/$port) &> /dev/null
    if [ $? -ne 0 ]; then
        return 0  # Port is available
    else
        return 1  # Port is in use
    fi
}

# Find an available port
find_available_port() {
    local port=$START_PORT
    local max_search=$((START_PORT + 1000))  # Search up to 1000 ports to find an available one

    # Debug output
    echo -e "${BLUE}[PORT]${NC} Searching for available port starting from $START_PORT..." >&2

    while [ $port -le $max_search ]; do
        if is_port_available $port; then
            # Debug output
            echo -e "${GREEN}[PORT]${NC} Found available port: $port" >&2

            # Output just the port number to stdout for capture
            echo $port
            return 0
        fi

        # Increment port number
        port=$((port + 1))

        # Print a progress message every 100 ports
        if [ $((port % 100)) -eq 0 ]; then
            echo -e "${BLUE}[PORT]${NC} Still searching... checked up to port $port" >&2
        fi
    done

    # Error output (to stderr)
    echo -e "${YELLOW}[PORT]${NC} No available ports found between $START_PORT and $max_search" >&2
    return 1
}

# Run the function and output the result
find_available_port
