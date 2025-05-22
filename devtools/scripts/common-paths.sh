#!/bin/bash
# common-paths.sh - Common path definitions for development scripts

# Define the base directory for temporary files
DEV_TMP_DIR="devtools/tmp"

# Define the status directory
DEV_STATUS_DIR="$DEV_TMP_DIR/status"

# Define the server port file
DEV_SERVER_PORT_FILE="$DEV_TMP_DIR/server-port"

# Function to ensure directories exist
ensure_dirs() {
    mkdir -p "$DEV_TMP_DIR"
    mkdir -p "$DEV_STATUS_DIR"
}

# Function to get the server port file path
get_server_port_file() {
    echo "$DEV_SERVER_PORT_FILE"
}

# Function to get the status directory
get_status_dir() {
    echo "$DEV_STATUS_DIR"
}

# Ensure directories exist when this script is sourced
ensure_dirs
