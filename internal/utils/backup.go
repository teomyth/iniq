// Package utils provides utility functions for INIQ
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupFile creates a backup of a file with a timestamp
// If the file doesn't exist, it returns nil without creating a backup
// Returns the path to the backup file if successful
func BackupFile(filePath string, backupEnabled bool) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil // File doesn't exist, nothing to backup
	}

	// Read the original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup: %w", err)
	}

	var backupPath string
	if backupEnabled {
		// Create a timestamped backup file
		timestamp := time.Now().Format("20060102150405")
		backupPath = filePath + ".bak." + timestamp
	} else {
		// Create a simple backup file
		backupPath = filePath + ".bak"
	}

	// Create backup directory if it doesn't exist
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Write the backup file
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}
