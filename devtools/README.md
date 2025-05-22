# Development Tools Directory

## Purpose

The `devtools` directory contains tools and utilities specifically for development purposes. These tools are not part of the main application but help streamline the development process.

## Structure

- `tunnelserver/`: Development HTTP server
  - Serves installation scripts and binaries during development
  - Dynamically replaces download URLs in scripts
  - Provides endpoints for testing installation flows

- `tunnelrunner/`: Cloudflared tunnel integration
  - Manages the creation and configuration of Cloudflared tunnels
  - Exposes the local development server to the internet
  - Provides public URLs for testing installation scripts

- `watcher/`: File watching and hot reload functionality
  - Integrates with Air for automatic rebuilding
  - Monitors file changes and triggers rebuilds
  - Improves development workflow efficiency

- `scripts/`: Script content handling
  - Loads and processes script files
  - Replaces placeholder URLs with actual development URLs
  - Manages script content for the development server

## Usage

These tools are primarily used during development to:

1. Test installation scripts without deploying to production
2. Enable rapid development with hot reloading
3. Simulate production environments locally
4. Test cross-platform functionality

## Development Guidelines

- Keep development tools separate from application code
- Document tool usage and configuration
- Ensure tools work consistently across development environments
- Focus on improving developer productivity
- Test tools thoroughly to avoid development workflow issues

## Important Note

The code in this directory is not included in the production build of INIQ. It is solely for development purposes and should not contain any code that is required for the actual functionality of INIQ.
