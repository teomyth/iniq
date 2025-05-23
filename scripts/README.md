# Scripts Directory

## Purpose

The `scripts` directory contains installation scripts for INIQ.

## Structure

- `install.sh`: Standalone installation script for INIQ
  - Downloads and installs the INIQ binary
  - Handles platform detection and appropriate binary selection
  - Installs to system path with proper permissions
  - Includes all necessary functions for a complete installation experience

## Usage

The installation script is designed to be:

1. Downloaded and executed directly via curl or wget
2. Included in GitHub releases for easy access
3. Self-contained with all necessary functions

Example usage:
```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/install.sh | sudo bash
```

## Development Guidelines

- Ensure scripts work on Linux systems
- Handle errors gracefully with informative messages
- Keep the script self-contained and standalone
- Test scripts in different environments before release
- Keep scripts compatible with common shell variants (bash, sh)
- Include clear documentation and usage examples

## Important Note

The installation script is a critical part of the user experience, as it is often the first interaction users have with INIQ. Ensure it is robust, user-friendly, and provides clear feedback.
