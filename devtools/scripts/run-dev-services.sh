#!/bin/bash
# run-dev-services.sh - Script to run development services in a coordinated way

# Source common paths
echo "Sourcing common paths..."
source "$(dirname "$0")/common-paths.sh"
echo "Common paths sourced."

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

# Print header
echo -e "${BLUE}=== INIQ Development Environment ===${NC}"

# Stop any existing development services
echo -e "${YELLOW}→ Cleaning up existing services...${NC}"
# Skip task clean:dev for now
# task clean:dev > /dev/null
echo -e "${YELLOW}→ Skipping task clean:dev for debugging${NC}"

# Clean up status files
rm -f "$STATUS_DIR"/*.status "$STATUS_DIR"/*.error "$STATUS_DIR"/*.pid "$STATUS_DIR"/*.url "$STATUS_DIR"/*.log

# Build all binaries and scripts first
echo -e "${YELLOW}→ Building all binaries and scripts...${NC}"
# Make sure bin/scripts directory exists
mkdir -p bin/scripts

# Build scripts first
echo -e "${YELLOW}  Building scripts...${NC}"
task build:scripts

# Then build binaries
echo -e "${YELLOW}  Building binaries...${NC}"
task goreleaser:build

# Copy GoReleaser binaries to bin directory for HTTP server
echo -e "${YELLOW}  Copying binaries to bin directory...${NC}"
mkdir -p bin
# Detect current platform and copy the appropriate binary
if [[ "$OSTYPE" == "darwin"* ]]; then
  if [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_darwin_arm64_v8.0/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Darwin ARM64 binary"
  else
    cp dist/iniq_darwin_amd64_v1/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Darwin AMD64 binary"
  fi
else
  if [[ $(uname -m) == "aarch64" ]] || [[ $(uname -m) == "arm64" ]]; then
    cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"
  else
    cp dist/iniq_linux_amd64_v1/iniq bin/iniq 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
  fi
fi

# Copy all platform binaries for HTTP server downloads
cp dist/iniq_linux_amd64_v1/iniq bin/iniq-linux-amd64 2>/dev/null || echo "Warning: Could not copy Linux AMD64 binary"
cp dist/iniq_linux_arm64_v8.0/iniq bin/iniq-linux-arm64 2>/dev/null || echo "Warning: Could not copy Linux ARM64 binary"
cp dist/iniq_darwin_amd64_v1/iniq bin/iniq-darwin-amd64 2>/dev/null || echo "Warning: Could not copy Darwin AMD64 binary"
cp dist/iniq_darwin_arm64_v8.0/iniq bin/iniq-darwin-arm64 2>/dev/null || echo "Warning: Could not copy Darwin ARM64 binary"

# Build devserver
echo -e "${YELLOW}  Building development server...${NC}"
task build:devserver

# Check if build was successful
if [ ! -f bin/iniq ]; then
  echo -e "${RED}  Error: Failed to build binary. Check for build errors.${NC}"
  # Continue anyway, as the watcher might fix it
fi

# Check if devserver was built
if [ ! -f bin/devserver ]; then
  echo -e "${RED}  Error: Failed to build devserver. Check for build errors.${NC}"
  # Try to build it again
  task build:devserver
fi

# Check if scripts were built
if [ ! -f bin/scripts/install.sh ] || [ ! -f bin/scripts/iniq.sh ]; then
  echo -e "${RED}  Error: Failed to build scripts. Check for build errors.${NC}"
  # Try to build scripts again
  task build:scripts
fi

# Function to check if a process is running
is_process_running() {
  local pid_file="$STATUS_DIR/$1.pid"
  if [ -f "$pid_file" ]; then
    local pid=$(cat "$pid_file")
    if kill -0 "$pid" 2>/dev/null; then
      return 0  # Process is running
    fi
  fi
  return 1  # Process is not running
}

# Function to wait for a status file
wait_for_status() {
  local service=$1
  local status=$2
  local timeout=$3
  local status_file="$STATUS_DIR/$service.status"

  echo -e "${GRAY}  Waiting for $service to be $status...${NC}"

  for i in $(seq 1 $timeout); do
    if [ -f "$status_file" ] && [ "$(cat $status_file)" = "$status" ]; then
      return 0  # Status matched
    fi
    sleep 1
    echo -n "."
  done

  echo ""
  echo -e "${RED}  Timed out waiting for $service to be $status${NC}"
  return 1  # Timeout
}

# Start file watcher in background with output redirected to log file and terminal
echo -e "${YELLOW}→ Starting file watcher...${NC}"
# Use tee to send output to both terminal and log file
./devtools/scripts/watch-files.sh 2>&1 | tee "$STATUS_DIR/watcher.log" &
WATCH_PID=$!
echo "$WATCH_PID" > "$STATUS_DIR/watcher.pid"

# Wait for file watcher to be ready
if wait_for_status "watcher" "running" 10; then
  echo -e "${GREEN}  File watcher is running${NC}"
else
  echo -e "${YELLOW}  File watcher may not be fully ready${NC}"
  # Continue anyway, not critical
fi

# Start HTTP server in background with output redirected to log file
echo -e "${YELLOW}→ Starting HTTP server...${NC}"
./devtools/scripts/serve-http.sh > "$STATUS_DIR/http-server.log" 2>&1 &
HTTP_PID=$!
echo "$HTTP_PID" > "$STATUS_DIR/http-server.pid"

# Wait for HTTP server to be ready (increased timeout for WSL)
if wait_for_status "http-server" "running" 60; then
  echo -e "${GREEN}  HTTP server is running${NC}"
else
  echo -e "${YELLOW}  HTTP server status check timed out, but we'll continue anyway${NC}"
  echo -e "${YELLOW}  This may happen in WSL mirror mode due to network interface detection${NC}"

  # Check if the process is actually running despite the status timeout
  if is_process_running http-server; then
    echo -e "${GREEN}  HTTP server process is running, continuing...${NC}"
    # Force the status to running
    echo "running" > "$STATUS_DIR/http-server.status"
  else
    echo -e "${RED}  HTTP server process is not running, exiting${NC}"
    exit 1
  fi
fi

# Cloudflared tunnel has been completely removed from the project

# Wait for HTTP server to be fully initialized
for i in {1..10}; do
  if [ -f "$STATUS_DIR/http-server.initialized" ]; then
    break
  fi
  sleep 0.5
done

# Display summary
echo -e "${GREEN}=== Development Environment Ready ===${NC}"

# 1. Build result
echo -e "${BLUE}→ Build:${NC}"
if [ -f bin/iniq ]; then
  LAST_MODIFIED=$(stat -c %y bin/iniq 2>/dev/null || stat -f "%m" bin/iniq 2>/dev/null)
  if [ -n "$LAST_MODIFIED" ]; then
    echo -e "  • Binary compiled: ${GREEN}Yes${NC} (Last modified: $LAST_MODIFIED)"
  else
    echo -e "  • Binary compiled: ${GREEN}Yes${NC}"
  fi
else
  echo -e "  • Binary compiled: ${YELLOW}No${NC} (Binary not found)"
fi

if [ -f bin/scripts/install.sh ] && [ -f bin/scripts/iniq.sh ]; then
  echo -e "  • Scripts compiled: ${GREEN}Yes${NC}"
else
  echo -e "  • Scripts compiled: ${YELLOW}No${NC} (Scripts not found)"
fi

# 2. Services status
echo -e "${BLUE}→ Services:${NC}"
echo -e "  • File Watcher: $(is_process_running watcher && echo "${GREEN}Running${NC}" || echo "${RED}Not running${NC}")"
echo -e "  • HTTP Server: $(is_process_running http-server && echo "${GREEN}Running${NC}" || echo "${RED}Not running${NC}")"

# 3. URLs
echo -e "${BLUE}→ Access URLs:${NC}"
URL_AVAILABLE=false

# Get server port file path
SERVER_PORT_FILE=$(get_server_port_file)

if [ -f "$SERVER_PORT_FILE" ]; then
  SERVER_PORT=$(cat "$SERVER_PORT_FILE")
  URL_AVAILABLE=true

  # Show access URLs
  echo -e "    ‣ ${BLUE}http://127.0.0.1:${GREEN}${SERVER_PORT}${NC}"

  # Get all non-loopback IPv4 addresses (cross-platform)
  if command -v ip >/dev/null 2>&1; then
    # Linux: use ip command
    ip -4 addr show | grep -v "127.0.0.1" | grep "inet " | awk '{print $2}' | cut -d/ -f1 | while read -r ip; do
      echo -e "    ‣ ${BLUE}http://${ip}:${GREEN}${SERVER_PORT}${NC}"
    done
  elif command -v ifconfig >/dev/null 2>&1; then
    # macOS/BSD: use ifconfig command
    ifconfig | grep "inet " | grep -v "127.0.0.1" | awk '{print $2}' | while read -r ip; do
      echo -e "    ‣ ${BLUE}http://${ip}:${GREEN}${SERVER_PORT}${NC}"
    done
  fi
fi

# Cloudflared tunnel URLs have been removed

# If no URLs are available, show a message
if [ "$URL_AVAILABLE" = "false" ]; then
  echo -e "  ${YELLOW}• No URLs available. HTTP server is not running.${NC}"
fi

# We don't need to show server status here, as it's already in the logs

# Set up trap to kill background processes on exit
trap 'echo -e "${YELLOW}→ Stopping all services...${NC}";
      # Remove heartbeat flag file to signal graceful shutdown
      rm -f "$STATUS_DIR/heartbeat.running"

      # Kill heartbeat process explicitly before running clean:dev
      if [ -f "$STATUS_DIR/heartbeat.pid" ]; then
        HEARTBEAT_PID=$(cat "$STATUS_DIR/heartbeat.pid")
        if [ -n "$HEARTBEAT_PID" ]; then
          echo -e "${GRAY}  Stopping heartbeat process (PID: $HEARTBEAT_PID)${NC}"
          # Send SIGTERM first for graceful shutdown
          kill -15 $HEARTBEAT_PID 2>/dev/null || true
          # Wait a moment for graceful shutdown
          sleep 1
          # Force kill if still running
          if kill -0 $HEARTBEAT_PID 2>/dev/null; then
            kill -9 $HEARTBEAT_PID 2>/dev/null || true
          fi
        fi
      fi

      task clean:dev > /dev/null;
      echo -e "${GREEN}✓ All services stopped${NC}"' EXIT INT TERM

# Wait for Ctrl+C
echo -e "${GRAY}Press Ctrl+C to stop all components${NC}"

# Create a log file for non-essential output
touch "$STATUS_DIR/dev.log"

# Inform user about log file
echo -e "${GRAY}Detailed logs are available at $STATUS_DIR/dev.log${NC}"

# Setup a heartbeat to keep connections alive
echo -e "${GRAY}Setting up heartbeat to prevent timeouts...${NC}"

# Function to send heartbeat with safe exit
send_heartbeat() {
  # Create a flag file to indicate heartbeat is running
  touch "$STATUS_DIR/heartbeat.running"

  # Log heartbeat start
  current_time=$(date "+%Y-%m-%d %H:%M:%S")
  echo -e "${GRAY}[${current_time}] Heartbeat: Started${NC}" >> "$STATUS_DIR/dev.log"

  # Simple heartbeat loop
  while [ -f "$STATUS_DIR/heartbeat.running" ]; do
    # Log heartbeat
    current_time=$(date "+%Y-%m-%d %H:%M:%S")
    echo -e "${GRAY}[${current_time}] Heartbeat: Keeping services alive${NC}" >> "$STATUS_DIR/dev.log"

    # If HTTP server is running, send a request to it
    if is_process_running http-server && [ -f "$SERVER_PORT_FILE" ]; then
      SERVER_PORT=$(cat "$SERVER_PORT_FILE")
      echo -e "${GRAY}[${current_time}] Heartbeat: Pinging local server${NC}" >> "$STATUS_DIR/dev.log"
      curl -s "http://127.0.0.1:${SERVER_PORT}/health" > /dev/null 2>&1 || true
    fi

    # Cloudflared tunnel heartbeat has been removed

    # Sleep for 2 minutes, but check every 15 seconds if we should exit
    # Reduced interval to ensure more frequent activity on the tunnel
    for ((i=0; i<8; i++)); do
      if [ ! -f "$STATUS_DIR/heartbeat.running" ]; then
        break  # Exit the sleep loop if flag file is gone
      fi
      sleep 15
    done
  done

  # Log heartbeat termination
  current_time=$(date "+%Y-%m-%d %H:%M:%S")
  echo -e "${GRAY}[${current_time}] Heartbeat: Terminated${NC}" >> "$STATUS_DIR/dev.log"
}

# Start heartbeat in background
send_heartbeat &
HEARTBEAT_PID=$!
echo "$HEARTBEAT_PID" > "$STATUS_DIR/heartbeat.pid"

# Sleep indefinitely until interrupted
while true; do
  # Check if all services are still running
  all_running=true

  # Check file watcher
  if ! is_process_running watcher; then
    echo -e "${YELLOW}→ File watcher is not running, restarting...${NC}"
    ./devtools/scripts/watch-files.sh > "$STATUS_DIR/watcher.log" 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > "$STATUS_DIR/watcher.pid"
    all_running=false
  fi

  # Check HTTP server
  if ! is_process_running http-server; then
    echo -e "${YELLOW}→ HTTP server is not running, restarting...${NC}"
    # Remove initialization flag if it exists
    rm -f "$STATUS_DIR/http-server.initialized"
    ./devtools/scripts/serve-http.sh > "$STATUS_DIR/http-server.log" 2>&1 &
    HTTP_PID=$!
    echo "$HTTP_PID" > "$STATUS_DIR/http-server.pid"
    all_running=false
  fi

  # Cloudflared tunnel has been completely removed from the project

  # If any service was restarted, show a message
  if [ "$all_running" = false ]; then
    echo -e "${GREEN}→ Services have been restarted${NC}"
  fi

  # Sleep for 1 minute before checking again
  sleep 60
done
