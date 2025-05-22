package sshkeys

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test data
const (
	// These are valid SSH public keys for testing
	testRSAKey     = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCy9f0/nwkXESPwk2YFn7gUPP4k4bYkGpUUEyRsM6NQxgF4oGbIxdz6Y9MMj3O5ICwxT3dqXG7dkhFX0RQYJSRm9HSV/FCFRZsLLXxMVlvIZL9ZJpqCzYWwDL/xmR+cGN2VRCz4IuD5HGpHIY+EjUdX7/0LZ3MmGEPZNEAjZB5HvYqUlJj9T0/SqNLkGSk9DTN+0IY/cTQRfUBWiRr2YQGQhOzcLYuGQPLXVR+DS3XXHEbQQJYx1TFA+xS4c0TnBQ+Ou9lMPwdYh0G+fsf5OC+QQGlIL/WBDmCQpJBQpHRAQ9mwmyEfsMhTKY5wFNCvJ9ggGYSXHk6Xdj8Vy8yn4Qrj test@example.com`
	testED25519Key = `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKHEOLLbeuppGcBAwD1KQ31eOqmcBpH/B+7jUmNbw+7n test@example.com`
)

func TestParseKeyString(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		keyContent  string
		source      Source
		sourceValue string
		expectError bool
		keyType     string
		keyComment  string
	}{
		{
			name:        "Valid RSA key",
			keyContent:  testRSAKey,
			source:      Text,
			sourceValue: "",
			expectError: false,
			keyType:     "ssh-rsa",
			keyComment:  "test@example.com",
		},
		{
			name:        "Valid ED25519 key",
			keyContent:  testED25519Key,
			source:      Text,
			sourceValue: "",
			expectError: false,
			keyType:     "ssh-ed25519",
			keyComment:  "test@example.com",
		},
		{
			name:        "Invalid key",
			keyContent:  "not a valid ssh key",
			source:      Text,
			sourceValue: "",
			expectError: true,
		},
		{
			name:        "Empty key",
			keyContent:  "",
			source:      Text,
			sourceValue: "",
			expectError: true,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, err := ParseKeyString(tc.keyContent, tc.source, tc.sourceValue)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If we expect success, check the key details
			if !tc.expectError {
				if key.Type != tc.keyType {
					t.Errorf("Expected key type %q, got %q", tc.keyType, key.Type)
				}

				if key.Comment != tc.keyComment {
					t.Errorf("Expected key comment %q, got %q", tc.keyComment, key.Comment)
				}

				if key.Content == "" {
					t.Errorf("Key content is empty")
				}

				if key.Source != tc.source {
					t.Errorf("Expected source %q, got %q", tc.source, key.Source)
				}

				if key.SourceValue != tc.sourceValue {
					t.Errorf("Expected source value %q, got %q", tc.sourceValue, key.SourceValue)
				}
			}
		})
	}
}

func TestFetchFromURL(t *testing.T) {
	// Skip this test in automated environments
	// In a real implementation, you would use a mock HTTP server
	t.Skip("Skipping URL test as it requires network access")
}

func TestFetchFromGitHub(t *testing.T) {
	// Skip this test in automated environments
	// In a real implementation, you would use a mock HTTP client
	t.Skip("Skipping GitHub test as it requires network access")
}

func TestFetchFromGitLab(t *testing.T) {
	// Skip this test in automated environments
	// In a real implementation, you would use a mock HTTP client
	t.Skip("Skipping GitLab test as it requires network access")
}

func TestReadFromFile(t *testing.T) {
	// Create a temporary file with a test key
	tmpDir, err := os.MkdirTemp("", "sshkeys-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyFile := filepath.Join(tmpDir, "test_key.pub")
	err = os.WriteFile(keyFile, []byte(testRSAKey), 0644)
	if err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	// Test getting keys from file
	keys, err := ReadFromFile(keyFile)
	if err != nil {
		t.Errorf("Failed to read keys from file: %v", err)
		return
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
		return
	}

	// Verify key content
	key := keys[0]
	if key.Type != "ssh-rsa" {
		t.Errorf("Expected key type 'ssh-rsa', got %q", key.Type)
	}

	if key.Comment != "test@example.com" {
		t.Errorf("Expected key comment 'test@example.com', got %q", key.Comment)
	}

	if key.Source != File {
		t.Errorf("Expected source File, got %q", key.Source)
	}

	if key.SourceValue != keyFile {
		t.Errorf("Expected source value %q, got %q", keyFile, key.SourceValue)
	}

	// Test with non-existent file
	_, err = ReadFromFile("/non/existent/file.pub")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got none")
	}
}

func TestParseKeysFromText(t *testing.T) {
	// Test parsing keys from text
	text := testRSAKey + "\n\n" + testED25519Key

	keys, err := ParseKeysFromText(text)
	if err != nil {
		t.Errorf("Failed to parse keys from text: %v", err)
		return
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
		return
	}

	// Verify first key
	if keys[0].Type != "ssh-rsa" {
		t.Errorf("Expected first key type 'ssh-rsa', got %q", keys[0].Type)
	}

	// Verify second key
	if keys[1].Type != "ssh-ed25519" {
		t.Errorf("Expected second key type 'ssh-ed25519', got %q", keys[1].Type)
	}
}

func TestWriteToAuthorizedKeys(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "sshkeys-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test authorized_keys file
	authKeysFile := filepath.Join(tmpDir, "authorized_keys")

	// Parse test keys
	rsaKey, err := ParseKeyString(testRSAKey, Text, "")
	if err != nil {
		t.Fatalf("Failed to parse RSA key: %v", err)
	}

	ed25519Key, err := ParseKeyString(testED25519Key, Text, "")
	if err != nil {
		t.Fatalf("Failed to parse ED25519 key: %v", err)
	}

	// Test writing keys to a new file
	keys := []*Key{rsaKey, ed25519Key}
	err = WriteToAuthorizedKeys(authKeysFile, keys, false)
	if err != nil {
		t.Errorf("Failed to write to authorized_keys: %v", err)
		return
	}

	// Read the file and verify content
	content, err := os.ReadFile(authKeysFile)
	if err != nil {
		t.Errorf("Failed to read authorized_keys: %v", err)
		return
	}

	// Check that both keys are in the file
	contentStr := string(content)
	if !strings.Contains(contentStr, testRSAKey) {
		t.Errorf("RSA key not found in authorized_keys")
	}

	if !strings.Contains(contentStr, testED25519Key) {
		t.Errorf("ED25519 key not found in authorized_keys")
	}

	// Create a third key for appending test that has a different key value
	// We need to modify both the comment and the key value to avoid duplicate detection
	modifiedRSAKey := strings.Replace(testRSAKey, "AAAAB3NzaC1yc2EAAAADAQABAAABAQC", "AAAAB3NzaC1yc2EAAAADAQABAAABAQD", 1)
	thirdKey, err := ParseKeyString(modifiedRSAKey, Text, "third-key")
	if err != nil {
		t.Fatalf("Failed to parse third key: %v", err)
	}

	// Modify the comment to make it different
	thirdKey.Comment = "another@example.com"
	thirdKey.Content = strings.Replace(thirdKey.Content, "test@example.com", "another@example.com", 1)

	err = WriteToAuthorizedKeys(authKeysFile, []*Key{thirdKey}, true)
	if err != nil {
		t.Errorf("Failed to append to authorized_keys: %v", err)
		return
	}

	// Read the file again and verify all keys are present
	content, err = os.ReadFile(authKeysFile)
	if err != nil {
		t.Errorf("Failed to read authorized_keys after append: %v", err)
		return
	}

	contentStr = string(content)
	if !strings.Contains(contentStr, testRSAKey) {
		t.Errorf("Original RSA key not found after append")
	}

	if !strings.Contains(contentStr, testED25519Key) {
		t.Errorf("Original ED25519 key not found after append")
	}

	if !strings.Contains(contentStr, "another@example.com") {
		t.Errorf("Third key not found after append")
	}
}
