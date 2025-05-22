// Package integration provides integration tests for the INIQ CLI tool.
//
// These tests use the testscript package to run the INIQ binary in a controlled
// environment and verify its behavior.
//
// To run these tests:
//   go test -tags=integration ./integration/...
//
//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// Global variable to store the path to the temporary binary
var iniqBinaryPath string

// TestMain sets up the test environment.
func TestMain(m *testing.M) {
	// Create a temporary directory that will be cleaned up after tests
	tempDir, err := os.MkdirTemp("", "iniq-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}

	// Ensure the temporary directory is removed when tests are done
	defer func() {
		// Clean up the temporary directory and its contents
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove temp directory: %v\n", err)
		}
	}()

	// Get the project root directory using runtime.Caller
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")

	// Build the iniq binary in the temporary directory
	iniqBinaryPath = filepath.Join(tempDir, "iniq")

	// Use exec.Command to build the binary
	buildCmd := exec.Command("go", "build", "-o", iniqBinaryPath, filepath.Join(projectRoot, "cmd/iniq"))
	buildCmd.Dir = projectRoot // Set working directory to project root

	// Capture build output for debugging
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build iniq binary: %v\n%s\n", err, buildOutput)
		os.Exit(1)
	}

	// Verify the binary was created successfully
	if _, err := os.Stat(iniqBinaryPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Binary not found at %s after successful build\n", iniqBinaryPath)
		os.Exit(1)
	}

	// Run the tests with testscript
	exitCode := testscript.RunMain(m, map[string]func() int{
		"iniq": func() int {
			// Create a new command to run the binary with the provided arguments
			cmd := exec.Command(iniqBinaryPath, os.Args[1:]...)

			// Connect standard I/O
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			// Run the command and handle errors
			err := cmd.Run()
			if err != nil {
				// Extract exit code if available
				if exitErr, ok := err.(*exec.ExitError); ok {
					return exitErr.ExitCode()
				}
				// Default error exit code
				return 1
			}
			// Success
			return 0
		},
	})

	// Exit with the appropriate code
	os.Exit(exitCode)
}

// TestINIQ runs the testscript tests for the INIQ CLI.
func TestINIQ(t *testing.T) {
	// Skip tests on Windows as INIQ is designed for Linux/macOS
	if runtime.GOOS == "windows" {
		t.Skip("Skipping tests on Windows")
	}

	// Verify that the binary exists before running tests
	if iniqBinaryPath == "" {
		t.Fatal("INIQ binary path is not set. TestMain may not have run correctly.")
	}

	if _, err := os.Stat(iniqBinaryPath); os.IsNotExist(err) {
		t.Fatalf("INIQ binary not found at %s", iniqBinaryPath)
	}

	// Set up the test environment
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			// Set up any environment variables or files needed for tests
			env.Vars = append(env.Vars, "HOME="+env.WorkDir)

			// Add binary path to environment for debugging
			env.Vars = append(env.Vars, "INIQ_BINARY_PATH="+iniqBinaryPath)

			// Create a mock SSH key for testing
			sshDir := filepath.Join(env.WorkDir, ".ssh")
			if err := os.MkdirAll(sshDir, 0700); err != nil {
				return err
			}

			// Return nil to indicate success
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// Add custom commands for the test scripts if needed
			"mock-sudo": func(ts *testscript.TestScript, neg bool, args []string) {
				// Mock sudo command for testing
				if neg {
					ts.Fatalf("mock-sudo does not support negation")
				}
				if len(args) < 1 {
					ts.Fatalf("usage: mock-sudo <status>")
				}

				// Create a mock sudo environment
				ts.Setenv("MOCK_SUDO", args[0])
			},
		},
	})
}
