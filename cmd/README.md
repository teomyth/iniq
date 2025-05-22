# Command Directory

## Purpose

The `cmd` directory contains the application's entry points. Each subdirectory represents a separate executable program.

## Structure

- `iniq/`: Main INIQ CLI application
  - `main.go`: Entry point for the INIQ command-line tool

## Usage

This directory follows the standard Go project layout. Each subdirectory should contain a `main.go` file that serves as the entry point for a specific executable.

The main application logic should not be implemented here. Instead, the code in this directory should:

1. Parse command-line flags and arguments
2. Set up configuration
3. Initialize and wire together components from the `internal` and `pkg` directories
4. Handle top-level error management
5. Start the application

## Development Guidelines

- Keep the code in this directory minimal
- Focus on initialization and coordination
- Import and use packages from `internal/` and `pkg/`
- Do not implement business logic here
