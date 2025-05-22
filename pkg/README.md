# Package Directory

## Purpose

The `pkg` directory contains library code that can be used by external applications. Unlike the `internal` directory, code in this directory can be imported and used by other Go projects.

## Structure

- `osdetect/`: Operating system detection utilities
  - Provides functions to detect the operating system type and version
  - Identifies Linux distributions and package managers
  - Abstracts OS-specific details

- `sshkeys/`: SSH key handling utilities
  - Manages SSH key parsing and validation
  - Provides functions to fetch keys from various sources (GitHub, GitLab, URLs)
  - Handles key formatting and storage

## Usage

Code in this directory should:

1. Provide reusable functionality that may be useful to other projects
2. Have stable, well-documented APIs
3. Be well-tested and robust
4. Have minimal dependencies on other packages

## Development Guidelines

- Design for reuse by external applications
- Provide comprehensive documentation
- Maintain backward compatibility when possible
- Follow semantic versioning for breaking changes
- Keep dependencies minimal and explicit
- Write thorough tests for all functionality

## Important Note

Unlike code in the `internal` directory, packages in the `pkg` directory can be imported by external projects. This means that changes to these packages may affect other projects that depend on them. Be careful when making changes to ensure backward compatibility.
