package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test file for backup"
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		filePath      string
		backupEnabled bool
		expectError   bool
		expectBackup  bool
	}{
		{
			name:          "Backup with timestamp",
			filePath:      testFilePath,
			backupEnabled: true,
			expectError:   false,
			expectBackup:  true,
		},
		{
			name:          "Simple backup",
			filePath:      testFilePath,
			backupEnabled: false,
			expectError:   false,
			expectBackup:  true,
		},
		{
			name:          "Non-existent file",
			filePath:      filepath.Join(tempDir, "nonexistent.txt"),
			backupEnabled: true,
			expectError:   false,
			expectBackup:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run the backup function
			backupPath, err := BackupFile(tc.filePath, tc.backupEnabled)

			// Check error
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check backup file
			if tc.expectBackup {
				if backupPath == "" {
					t.Errorf("Expected backup path but got empty string")
				}

				// Check if backup file exists
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("Backup file does not exist: %s", backupPath)
				}

				// Check backup file content
				content, err := os.ReadFile(backupPath)
				if err != nil {
					t.Errorf("Failed to read backup file: %v", err)
				}
				if string(content) != testContent {
					t.Errorf("Backup file content does not match original. Got: %s, Want: %s", string(content), testContent)
				}

				// Check if backup file has the correct extension
				if tc.backupEnabled {
					if !strings.HasPrefix(filepath.Base(backupPath), "test.txt.bak.") {
						t.Errorf("Backup file does not have timestamp: %s", backupPath)
					}
				} else {
					if filepath.Base(backupPath) != "test.txt.bak" {
						t.Errorf("Backup file does not have correct name: %s", backupPath)
					}
				}
			} else {
				if backupPath != "" {
					t.Errorf("Expected empty backup path but got: %s", backupPath)
				}
			}
		})
	}
}
