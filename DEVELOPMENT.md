# INIQ Development Guide

This document provides detailed information for developers who want to contribute to the INIQ project.

## Development Setup

### Prerequisites

- Go 1.18 or later - [Download and Install Go](https://golang.org/doc/install)
- Task - Task runner for Go (installation instructions below)

### Installation Instructions

#### Step 1: Clone the Repository

```bash
git clone https://github.com/teomyth/iniq.git
cd iniq
```

#### Step 2: Install Dependencies

**Linux (Ubuntu/Debian)**
```bash
# 1. Install Go
## Option 1: Using apt (may not be the latest version)
sudo apt update
sudo apt install golang-go

## Option 2: Install from official Go download (recommended for latest version)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile

# 2. Install Task
## Option 1: Using the official installation script (recommended)
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
echo 'export PATH=$PATH:~/.local/bin' >> ~/.profile
source ~/.profile

## Option 2: Using npm (with latest Node.js)
# Install the latest version of Node.js and npm (using NodeSource repository)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
# This installation already includes npm, no need to install it separately
sudo npm install -g @go-task/cli

## Option 3: Using Go (if you installed Go above)
go install github.com/go-task/task/v3/cmd/task@latest

## Option 4: Using snap (if snap is available)
sudo snap install task --classic
```

**Linux (Fedora/RHEL/CentOS)**
```bash
# 1. Install Go
sudo dnf install golang

# 2. Install Task
## Option 1: Using dnf
sudo dnf install go-task

## Option 2: Using the official installation script
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
echo 'export PATH=$PATH:~/.local/bin' >> ~/.profile
source ~/.profile

## Option 3: Using Go
go install github.com/go-task/task/v3/cmd/task@latest
```

**Linux (Arch Linux)**
```bash
# 1. Install Go
sudo pacman -S go

# 2. Install Task
sudo pacman -S go-task
```

**macOS**
```bash
# 1. Install Go
## Option 1: Using Homebrew
brew install go

## Option 2: Download from go.dev
# Download from https://go.dev/dl/ and follow installation instructions

# 2. Install Task
## Using Homebrew
brew install go-task/tap/go-task
```

**Any platform with Go installed**
```bash
# Install Task directly using Go
go install github.com/go-task/task/v3/cmd/task@latest
```

#### Step 3: Setup Development Environment

Once Task is installed, you can set up the entire development environment with a single command:

```bash
# Setup development environment (installs all required tools)
task setup
```

This will:
- Install all required development tools (Air, linters, etc.)
- Initialize Go modules
- Build the project
- Configure development environment

### Development Commands

```bash
# Start complete development environment (hot reload and HTTP server)
task dev

# Build binaries
task build                # Build for current platform
task build:all            # Build for all supported platforms

# Testing
task test                 # Run tests
task test:coverage        # Run tests with coverage report
task test:race            # Run tests with race condition detection

# Code quality
task lint                 # Run static code analysis
task fmt                  # Format code

# Dependency management
task deps                 # Update dependencies

# Cleanup
task clean                # Clean build artifacts (bin/, dist/, tmp/)
task clean:all            # Clean all artifacts and stop development services
task clean:dev            # Stop development services only

# Version management
task version:current      # Display current version
task version:next         # Display next suggested version
task version:patch        # Increment patch version (Z in vX.Y.Z)
task version:minor        # Increment minor version (Y in vX.Y.Z)
task version:major        # Increment major version (X in vX.Y.Z)

# Release
task release              # Prepare for release (run tests and build for all platforms)
```

View all available commands:

```bash
task --list
```

## Project Structure

The INIQ project follows a standard Go project layout:

- `cmd/` - Main applications for this project
- `internal/` - Private application and library code
- `pkg/` - Library code that's safe to use by external applications
- `scripts/` - Scripts to perform various build, install, analysis, etc. operations
- `devtools/` - Development tools and configurations
- `integration/` - Integration tests

## Coding Standards

INIQ follows standard Go coding conventions:

1. Use `gofmt` to format your code
2. Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
3. Document all exported functions, types, and constants
4. Write tests for all new functionality

## Testing

INIQ uses Go's standard testing package for unit tests and a custom framework for integration tests.

### Running Tests

```bash
# Run all tests
task test

# Run tests with coverage report
task test:coverage

# Run tests with race condition detection
task test:race
```

## Continuous Integration

INIQ uses GitHub Actions for continuous integration. The CI pipeline runs on every pull request and includes:

1. Building the project
2. Running unit tests
3. Running integration tests
4. Linting the code
5. Checking code coverage

## Workflows and Automation

INIQ uses a combination of Task-based workflows and GitHub Actions to automate development, testing, and release processes.

### Task Workflows

The project uses [Task](https://taskfile.dev/) as a task runner to standardize and simplify common development operations. Task workflows are defined in `Taskfile.yaml` and provide a consistent interface for various development activities:

#### Development Workflow

The development workflow is designed to provide a smooth experience with hot reloading:

```bash
# Start the complete development environment
task dev
```

This command:
- Watches for file changes and automatically rebuilds the project
- Starts a development HTTP server for testing
- Provides immediate feedback during development

#### Testing Workflow

The testing workflow includes various levels of testing:

```bash
# Run unit tests
task test

# Run all tests (unit and integration)
task test:all

# Run tests with coverage report
task test:coverage

# Run tests with race condition detection
task test:race
```

#### Build Workflow

The build workflow supports multiple platforms:

```bash
# Build for current platform
task build

# Build for all supported platforms
task build:all
```

### Version Management

INIQ uses semantic versioning and provides tools to manage versions:

```bash
# Display current version
task version:current

# Display next suggested version
task version:next

# Increment patch version (Z in vX.Y.Z)
task version:patch

# Increment minor version (Y in vX.Y.Z)
task version:minor

# Increment major version (X in vX.Y.Z)
task version:major
```

The version management system:
- Uses Git tags for version tracking
- Automatically embeds version information in binaries
- Generates appropriate version numbers based on semantic versioning rules

### CI/CD Pipeline

INIQ uses GitHub Actions for continuous integration and deployment:

#### Continuous Integration

The CI workflow (`ci.yml`) runs on every push to the main branch and pull requests:

1. **Lint**: Runs static code analysis to ensure code quality
2. **Test**: Runs all tests to verify functionality
3. **Build**: Builds binaries for all supported platforms
4. **Artifacts**: Uploads build artifacts for inspection

The CI pipeline ensures that all code changes maintain quality standards and don't introduce regressions.

#### Version Management

The version management workflow (`version.yml`) is triggered manually through GitHub Actions:

1. **Version Bump**: Increments the version number (patch, minor, or major)
2. **Changelog Generation**: Automatically generates a changelog from commit messages
3. **Tag Creation**: Creates and pushes a Git tag for the new version

#### Release with GoReleaser

The GoReleaser workflow (`goreleaser.yml`) is triggered automatically when a new tag is pushed:

1. **Build**: Builds binaries for all supported platforms
2. **Package**: Creates archives for each platform
3. **Checksum**: Generates SHA256 checksums for all artifacts
4. **GitHub Release**: Creates a GitHub release with all artifacts and release notes

This workflow leverages GoReleaser to handle the entire release process in a standardized way, following industry best practices.

### Release Process

INIQ uses GoReleaser for automated releases. The release process combines manual and automated steps:

1. **Prepare for Release**:
   - Ensure all tests pass: `task test:all`
   - Review and update documentation if needed
   - Commit all changes to the main branch

2. **Create and Push a Tag**:
   - Use the version management commands to create a new tag:
     ```bash
     # For a patch release
     task version:patch

     # For a minor release
     task version:minor

     # For a major release
     task version:major
     ```
   - Push the tag to GitHub:
     ```bash
     git push origin <tag-name>
     ```

3. **Automated Release with GoReleaser**:
   - When a tag is pushed, the GoReleaser GitHub Action is automatically triggered
   - The workflow automatically:
     - Builds binaries for all platforms
     - Generates checksums for verification
     - Creates a GitHub release with all artifacts
     - Includes installation instructions in the release notes

4. **Verification**:
   - Download and test the released binaries
   - Verify the installation process works as expected

#### Local Testing with GoReleaser

You can test the GoReleaser configuration locally without creating a release:

```bash
# Install GoReleaser if not already installed
sudo snap install goreleaser --classic

# Test the configuration (dry-run)
goreleaser release --snapshot --clean --skip=publish
```

This will build the binaries and create archives in the `dist` directory without publishing a release.

#### GoReleaser Configuration

The GoReleaser configuration is stored in `.goreleaser.yaml` and defines:

- Build settings for different platforms
- Archive formats and naming
- Checksum generation
- Release notes format
- GitHub release settings

This automated release process ensures consistent, reliable releases with minimal manual intervention.

## Contributing

Contributions to INIQ are welcome! Please follow the coding standards and testing guidelines outlined in this document.
