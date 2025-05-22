#!/bin/bash
# clean-dev-env.sh - Script to stop all development services

# Source common paths
source "$(dirname "$0")/common-paths.sh"

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
GRAY='\033[0;90m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}[CLEANUP]${NC} Stopping development services..."

# Get server port file path
SERVER_PORT_FILE=$(get_server_port_file)

# Check for dev server port file
if [ -f "$SERVER_PORT_FILE" ]; then
  SERVER_PORT=$(cat "$SERVER_PORT_FILE")
  echo -e "${BLUE}[CLEANUP]${NC} Found server port: $SERVER_PORT"

  # Stop HTTP server on the specific port
  PID=$(lsof -t -i:$SERVER_PORT 2>/dev/null || echo "")
  if [ -n "$PID" ]; then
    echo -e "${BLUE}[CLEANUP]${NC} Stopping HTTP server on port $SERVER_PORT (PID: $PID)"
    kill -9 $PID 2>/dev/null || true
  fi

  # Remove port file
  rm -f "$SERVER_PORT_FILE"
else
  # Try common ports if no port file exists
  for port in 8080 $(seq 10000 10100); do
    PID=$(lsof -t -i:$port 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
      echo -e "${BLUE}[CLEANUP]${NC} Stopping HTTP server on port $port (PID: $PID)"
      kill -9 $PID 2>/dev/null || true
    fi
  done
fi

# Stop all Go processes related to our development
GO_PIDS=$(ps aux | grep "go run devtools/devserver" | grep -v grep | awk '{print $2}')
if [ -n "$GO_PIDS" ]; then
  echo -e "${BLUE}[CLEANUP]${NC} Stopping Go development server processes"
  for pid in $GO_PIDS; do
    echo -e "${GRAY}[CLEANUP]${NC} Killing PID: $pid"
    kill -9 $pid 2>/dev/null || true
  done
fi



# Stop any watchexec processes
WATCHEXEC_PIDS=$(pgrep watchexec 2>/dev/null || echo "")
if [ -n "$WATCHEXEC_PIDS" ]; then
  echo -e "${BLUE}[CLEANUP]${NC} Stopping watchexec processes"
  for pid in $WATCHEXEC_PIDS; do
    echo -e "${GRAY}[CLEANUP]${NC} Killing watchexec PID: $pid"
    kill -9 $pid 2>/dev/null || true
  done
fi

# Kill any heartbeat processes
HEARTBEAT_PID_FILE="$STATUS_DIR/heartbeat.pid"
# Remove heartbeat flag file to signal graceful shutdown
rm -f "$STATUS_DIR/heartbeat.running"

if [ -f "$HEARTBEAT_PID_FILE" ]; then
  HEARTBEAT_PID=$(cat "$HEARTBEAT_PID_FILE")
  if [ -n "$HEARTBEAT_PID" ]; then
    echo -e "${BLUE}[CLEANUP]${NC} Stopping heartbeat process (PID: $HEARTBEAT_PID)"
    # Send SIGTERM first for graceful shutdown
    kill -15 $HEARTBEAT_PID 2>/dev/null || true
    # Wait a moment for graceful shutdown
    sleep 1
    # Force kill if still running
    if kill -0 $HEARTBEAT_PID 2>/dev/null; then
      echo -e "${BLUE}[CLEANUP]${NC} Force killing heartbeat process (PID: $HEARTBEAT_PID)"
      kill -9 $HEARTBEAT_PID 2>/dev/null || true
    fi
  fi
  rm -f "$HEARTBEAT_PID_FILE"
fi

# Check if any processes are still running
REMAINING_PIDS=$(ps aux | grep -E "devserver|watchexec|run-dev-services" | grep -v grep | awk '{print $2}')
if [ -n "$REMAINING_PIDS" ]; then
  echo -e "${YELLOW}[CLEANUP]${NC} Some processes may still be running:"
  ps aux | grep -E "devserver|watchexec|run-dev-services" | grep -v grep
  echo -e "${YELLOW}[CLEANUP]${NC} Attempting to force kill all remaining processes..."
  for pid in $REMAINING_PIDS; do
    kill -9 $pid 2>/dev/null || true
  done
fi

# Clean up status files
STATUS_DIR=$(get_status_dir)
echo -e "${BLUE}[CLEANUP]${NC} Cleaning up status files in $STATUS_DIR"
rm -f "$STATUS_DIR"/*.status "$STATUS_DIR"/*.error "$STATUS_DIR"/*.pid "$STATUS_DIR"/*.url "$STATUS_DIR"/*.log

# Also clean legacy status files if they exist
if [ -d "$LEGACY_DEV_STATUS_DIR" ] && [ "$LEGACY_DEV_STATUS_DIR" != "$STATUS_DIR" ]; then
  echo -e "${BLUE}[CLEANUP]${NC} Cleaning up legacy status files in $LEGACY_DEV_STATUS_DIR"
  rm -f "$LEGACY_DEV_STATUS_DIR"/*.status "$LEGACY_DEV_STATUS_DIR"/*.error "$LEGACY_DEV_STATUS_DIR"/*.pid "$LEGACY_DEV_STATUS_DIR"/*.url "$LEGACY_DEV_STATUS_DIR"/*.log
fi

echo -e "${GREEN}[CLEANUP]${NC} All services stopped and temporary files cleaned"
