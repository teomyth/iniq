// Package sshkeys provides SSH key handling functionality
package sshkeys

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Source represents the source of an SSH key
type Source string

const (
	// GitHub represents GitHub as a key source
	GitHub Source = "github"
	// GitLab represents GitLab as a key source
	GitLab Source = "gitlab"
	// URL represents a direct URL as a key source
	URL Source = "url"
	// File represents a local file as a key source
	File Source = "file"
	// Text represents a direct text input as a key source
	Text Source = "text"
)

// Key represents an SSH public key with metadata
type Key struct {
	// Content is the full content of the key
	Content string
	// Type is the key type (e.g., ssh-rsa, ssh-ed25519)
	Type string
	// Fingerprint is the key fingerprint
	Fingerprint string
	// Comment is the key comment (if any)
	Comment string
	// Source is where the key came from
	Source Source
	// SourceValue is the specific value for the source (e.g., GitHub username)
	SourceValue string
}

// ParseKeyString parses a key string and returns a Key
func ParseKeyString(keyString string, source Source, sourceValue string) (*Key, error) {
	keyString = strings.TrimSpace(keyString)
	if keyString == "" {
		return nil, fmt.Errorf("empty key string")
	}

	// Parse the key to validate and get type
	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(keyString))
	if err != nil {
		return nil, fmt.Errorf("invalid SSH key: %w", err)
	}

	// Get key type
	keyType := pubKey.Type()

	// Get fingerprint
	fingerprint := ssh.FingerprintSHA256(pubKey)

	return &Key{
		Content:     keyString,
		Type:        keyType,
		Fingerprint: fingerprint,
		Comment:     comment,
		Source:      source,
		SourceValue: sourceValue,
	}, nil
}

// FetchFromGitHub fetches SSH keys from GitHub
func FetchFromGitHub(username string) ([]*Key, error) {
	url := fmt.Sprintf("https://github.com/%s.keys", username)
	keys, err := fetchFromURL(url, GitHub, username)

	// Add GitHub username as comment if key doesn't have one
	for _, key := range keys {
		if key.Comment == "" {
			// Extract key parts
			parts := strings.Fields(key.Content)
			if len(parts) >= 2 {
				// Add GitHub username as comment in a more concise format
				comment := fmt.Sprintf("gh:%s", username)
				key.Comment = comment
				// Update content with comment
				key.Content = fmt.Sprintf("%s %s %s", parts[0], parts[1], comment)
			}
		}
	}

	return keys, err
}

// FetchFromGitLab fetches SSH keys from GitLab
func FetchFromGitLab(username string) ([]*Key, error) {
	url := fmt.Sprintf("https://gitlab.com/%s.keys", username)
	keys, err := fetchFromURL(url, GitLab, username)

	// Add GitLab username as comment if key doesn't have one
	for _, key := range keys {
		if key.Comment == "" {
			// Extract key parts
			parts := strings.Fields(key.Content)
			if len(parts) >= 2 {
				// Add GitLab username as comment in a more concise format
				comment := fmt.Sprintf("gl:%s", username)
				key.Comment = comment
				// Update content with comment
				key.Content = fmt.Sprintf("%s %s %s", parts[0], parts[1], comment)
			}
		}
	}

	return keys, err
}

// FetchFromURL fetches SSH keys from a URL
func FetchFromURL(url string) ([]*Key, error) {
	return fetchFromURL(url, URL, url)
}

// fetchFromURL is a helper function to fetch keys from a URL
func fetchFromURL(url string, source Source, sourceValue string) ([]*Key, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch keys from %s: status code %d", url, resp.StatusCode)
	}

	// Read response body
	return parseKeysFromReader(resp.Body, source, sourceValue)
}

// ReadFromFile reads SSH keys from a file
func ReadFromFile(filePath string) ([]*Key, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file %s: %w", filePath, err)
	}
	defer file.Close()

	// Parse keys
	return parseKeysFromReader(file, File, filePath)
}

// ParseKeysFromText parses SSH keys from a text string
func ParseKeysFromText(text string) ([]*Key, error) {
	return parseKeysFromReader(strings.NewReader(text), Text, "")
}

// parseKeysFromReader parses SSH keys from a reader
func parseKeysFromReader(reader io.Reader, source Source, sourceValue string) ([]*Key, error) {
	var keys []*Key

	// Create scanner to read line by line
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Parse key
		key, err := ParseKeyString(line, source, sourceValue)
		if err != nil {
			// Skip invalid keys but log the error
			fmt.Printf("Warning: Skipping invalid key: %v\n", err)
			continue
		}

		keys = append(keys, key)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no valid SSH keys found")
	}

	return keys, nil
}

// extractKeyTypeAndValue extracts just the type and value parts of an SSH key, ignoring comments
func extractKeyTypeAndValue(keyContent string) string {
	parts := strings.Fields(keyContent)
	if len(parts) >= 2 {
		// Return just the type and value (first two parts)
		return parts[0] + " " + parts[1]
	}
	// If the key format is unexpected, return the original content
	return keyContent
}

// CleanDuplicateKeys reads an authorized_keys file and removes duplicate keys
// keeping only the first occurrence of each unique key (based on type and value, not comments)
func CleanDuplicateKeys(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // Nothing to clean if file doesn't exist
	}

	// Read all keys from the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read authorized_keys file: %w", err)
	}

	// Parse keys
	keys, err := ParseKeysFromText(string(content))
	if err != nil {
		// If no valid keys found, just return
		if strings.Contains(err.Error(), "no valid SSH keys found") {
			return nil
		}
		return fmt.Errorf("failed to parse keys from authorized_keys: %w", err)
	}

	// Create a map to track unique keys
	uniqueKeys := make(map[string]*Key)
	var uniqueKeyOrder []string // To preserve original order

	// Process each key, keeping only the first occurrence of each unique key
	for _, key := range keys {
		keyTypeAndValue := extractKeyTypeAndValue(key.Content)
		if _, exists := uniqueKeys[keyTypeAndValue]; !exists {
			uniqueKeys[keyTypeAndValue] = key
			uniqueKeyOrder = append(uniqueKeyOrder, keyTypeAndValue)
		}
	}

	// If no duplicates were found, return early
	if len(uniqueKeys) == len(keys) {
		return nil
	}

	// Create a new slice with unique keys in original order
	var cleanedKeys []*Key
	for _, keyTypeAndValue := range uniqueKeyOrder {
		cleanedKeys = append(cleanedKeys, uniqueKeys[keyTypeAndValue])
	}

	// Write the cleaned keys back to the file
	return WriteToAuthorizedKeys(filePath, cleanedKeys, false)
}

// WriteToAuthorizedKeys writes keys to an authorized_keys file
func WriteToAuthorizedKeys(filePath string, keys []*Key, appendMode bool) error {
	// Determine file mode
	var file *os.File
	var err error

	if appendMode {
		// Open file in append mode, create if it doesn't exist
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	} else {
		// Create or truncate file
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	}

	if err != nil {
		return fmt.Errorf("failed to open authorized_keys file: %w", err)
	}
	defer file.Close()

	// If appending, check for duplicates
	var existingKeyValues map[string]bool
	if appendMode {
		existingKeyValues = make(map[string]bool)

		// Read existing keys
		existingFile, err := os.Open(filePath)
		if err == nil {
			defer existingFile.Close()

			scanner := bufio.NewScanner(existingFile)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					// Store just the type and value parts for comparison
					keyTypeAndValue := extractKeyTypeAndValue(line)
					existingKeyValues[keyTypeAndValue] = true
				}
			}
		}
	}

	// Write keys
	for _, key := range keys {
		// Skip duplicates if appending - compare only type and value, not comments
		keyTypeAndValue := extractKeyTypeAndValue(key.Content)
		if appendMode && existingKeyValues[keyTypeAndValue] {
			continue
		}

		// Write key with newline
		if _, err := fmt.Fprintln(file, key.Content); err != nil {
			return fmt.Errorf("failed to write key to file: %w", err)
		}
	}

	return nil
}
