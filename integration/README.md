# INIQ Integration Tests

This directory contains integration tests for the INIQ CLI tool. These tests use the `testscript` package to run the INIQ binary in a controlled environment and verify its behavior.

## Running the Tests

To run the integration tests:

```bash
go test -tags=integration ./integration/...
```

## Test Structure

The integration tests are organized as follows:

- `iniq_test.go`: Main test file that sets up the test environment and runs the tests
- `testdata/`: Directory containing test scripts
  - `version.txt`: Tests for version information display
  - `sudo_detection.txt`: Tests for sudo detection functionality
  - `binary_cache.txt`: Tests for binary caching functionality

## Writing Test Scripts

Test scripts use the `testscript` format, which is similar to shell scripts but with special commands for testing. Here's an example:

```
# Test that INIQ displays version information with -v flag
exec iniq -v
stdout 'INIQ v'
! stderr .
```

Common commands:

- `exec`: Execute a command
- `stdout`: Check that stdout contains a pattern
- `stderr`: Check that stderr contains a pattern
- `!`: Negate the following check
- `mkdir`: Create a directory
- `cp`: Copy a file
- `env`: Set an environment variable

For more information, see the [testscript documentation](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript).

## Mocking

Some tests require mocking system functionality. Custom commands are provided for this purpose:

- `mock-sudo`: Mock sudo functionality

## Adding New Tests

To add a new test:

1. Create a new `.txt` file in the `testdata/` directory
2. Write test scripts using the `testscript` format
3. Run the tests to verify your new test works correctly
