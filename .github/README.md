# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automating various tasks in the INIQ project.

## Available Workflows

### CI Workflow (`ci.yml`)

Runs on every push to the main branch and on pull requests:

- Lints the code
- Runs tests
- Builds the project for all platforms
- Uploads build artifacts

### Version Management and Release Workflow (`version.yml`)

Manually triggered to bump the version and create a release:

- Calculates the next version based on the selected type (patch, minor, major)
- Generates a changelog based on commit history
- Creates and pushes a new Git tag
- Builds the project for all platforms
- Generates checksums for all binaries
- Creates a GitHub release with all artifacts and changelog
- Supports dry run mode to preview the next version without creating a tag or release
- Allows adding custom release notes

## Usage

### Continuous Integration

The CI workflow runs automatically on pushes and pull requests. No manual action is required.

### Creating a Release

There are two ways to create a release:

#### Method 1: Using the Version Management and Release Workflow (Recommended)

1. Go to the "Actions" tab in the GitHub repository
2. Select the "Version Management and Release" workflow
3. Click "Run workflow"
4. Select the version type to bump (patch, minor, major)
5. (Optional) Add custom release notes
6. Click "Run workflow"

This will:
- Calculate the next version
- Generate a changelog
- Create and push a new tag
- Build the project for all platforms
- Create a GitHub release with the changelog and installation instructions

#### Method 2: Manually Creating a Tag

1. Create and push a new tag locally:
   ```bash
   task version:patch  # or version:minor or version:major
   git push origin <new-tag>
   ```

2. The release part of the workflow will not run automatically with this method. You'll need to create a GitHub release manually.

## Best Practices

1. Always use the Version Management and Release workflow or the `task version:*` commands to create new versions
2. Write clear commit messages following the conventional commits format:
   - `feat: add new feature`
   - `fix: fix bug`
   - `docs: update documentation`
   - `chore: update dependencies`
   - `refactor: refactor code`
   - `test: add tests`
   - `perf: improve performance`
3. Use the dry run option to preview the next version before creating it
4. Add custom release notes for important changes or migration instructions
5. Verify the release artifacts after a release is created
