# Scripts Directory

## Purpose

The `scripts` directory contains installation and utility scripts for INIQ. These scripts are used for installation, development setup, and one-line initialization of systems.

## Structure

- `install.sh`: Installation script for INIQ
  - Downloads and installs the INIQ binary
  - Handles platform detection and appropriate binary selection
  - Installs to system path with proper permissions

- `iniq.sh`: One-line initialization script
  - Downloads and runs INIQ in interactive mode or with specified parameters
  - Properly handles command-line arguments with the `--` separator
  - Can be run using either curl or wget:
    ```bash
    # Interactive mode
    bash <(curl -L https://example.com/iniq.sh)
    bash <(wget -qO- https://example.com/iniq.sh)

    # With parameters
    bash <(curl -L https://example.com/iniq.sh) -- --user myuser --key github:username
    bash <(wget -qO- https://example.com/iniq.sh) -- --user myuser --key github:username
    ```

- `common.sh`: Common functions used by other scripts
  - Provides shared functionality for installation scripts
  - Handles platform detection, URL generation, and binary installation
  - Manages temporary files and cleanup

- `dev-setup.sh`: Development environment setup script
  - Installs required development tools
  - Sets up the project directory structure
  - Initializes Go modules and dependencies

## Usage

These scripts are designed to be:

1. Downloaded and executed directly (e.g., via curl)
2. Included in the project repository for development
3. Served dynamically by the development server with URL replacement

## Development Guidelines

- Ensure scripts work on both Linux and macOS
- Handle errors gracefully with informative messages
- Use common functions from `common.sh` to avoid duplication
- Test scripts in different environments before release
- Keep scripts compatible with common shell variants (bash, sh)
- Include clear documentation and usage examples

## Important Note

The installation scripts are a critical part of the user experience, as they are often the first interaction users have with INIQ. Ensure they are robust, user-friendly, and provide clear feedback.
