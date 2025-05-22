//go:build tools
// +build tools

package tools

import (
	// Development tools
	_ "github.com/go-task/task/v3/cmd/task"                 // Task runner
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
)

// This file is used to track development tool dependencies.
// It is not compiled into the final binary.
//
// These tools are automatically installed by running:
// task setup
//
// Or you can install them manually:
// go install github.com/go-task/task/v3/cmd/task@latest
// go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
