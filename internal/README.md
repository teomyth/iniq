# Internal Directory

## Purpose

The `internal` directory contains private application code that is specific to INIQ and not intended for use by external applications. Go enforces this by preventing imports of code from this directory by packages outside the parent module.

## Structure

- `config/`: Configuration handling and environment variables
  - Manages loading and parsing of configuration files
  - Handles environment variable integration
  - Provides configuration validation

- `features/`: Core feature implementations
  - Contains the implementation of all INIQ features
  - Each feature is implemented as a separate package
  - Features follow a common interface defined in `features/feature.go`

- `logger/`: Logging functionality
  - Provides structured logging with different levels
  - Supports different output formats (text, JSON)
  - Handles log rotation and filtering

- `utils/`: Utility functions for internal use
  - Common helper functions used across the application
  - Not intended for external use

## Usage

Code in this directory should:

1. Implement the core functionality of INIQ
2. Be organized into logical packages based on functionality
3. Follow the interfaces defined in the feature package
4. Be well-tested with unit tests

## Development Guidelines

- Keep packages focused on a single responsibility
- Use interfaces to define clear boundaries between components
- Write comprehensive unit tests for all functionality
- Do not expose implementation details unnecessarily
- Document all exported functions, types, and constants

## Important Note

Code in the `internal` directory cannot be imported by external projects. This is enforced by the Go compiler. If you need to create reusable components that can be used by other projects, place them in the `pkg` directory instead.
