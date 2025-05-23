// Package ssh implements the SSH key management feature
package ssh

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/internal/utils"
	"github.com/teomyth/iniq/pkg/osdetect"
	"github.com/teomyth/iniq/pkg/sshkeys"
)

// Feature implements the SSH key management feature
type Feature struct {
	osInfo *osdetect.Info
}

// New creates a new SSH key management feature
func New(osInfo *osdetect.Info) *Feature {
	return &Feature{
		osInfo: osInfo,
	}
}

// Name returns the feature name
func (f *Feature) Name() string {
	return "ssh"
}

// Description returns the feature description
func (f *Feature) Description() string {
	return "Import SSH keys from various sources"
}

// Flags returns the command-line flags for the feature
func (f *Feature) Flags() []features.Flag {
	return []features.Flag{
		{
			Name:      "key",
			Shorthand: "k",
			Usage:     "SSH key sources (github:user, gitlab:user, url:URL, file:path)",
			Default:   []string{},
			Required:  false,
		},
	}
}

// ShouldActivate determines if the feature should be activated
func (f *Feature) ShouldActivate(options map[string]any) bool {
	// In interactive mode, always activate the ssh function
	interactive, hasInteractive := options["interactive"].(bool)
	if hasInteractive && interactive {
		return true
	}

	// Otherwise, it will only be activated if the SSH key is provided
	keys, ok := options["keys"].([]string)
	return ok && len(keys) > 0
}

// ValidateOptions validates the feature options
func (f *Feature) ValidateOptions(options map[string]any) error {
	// Check if keys are valid
	keys, ok := options["keys"].([]string)
	if !ok {
		return nil // No keys specified, nothing to validate
	}

	// Validate each key source
	for _, key := range keys {
		if err := validateKeySource(key); err != nil {
			return err
		}
	}

	return nil
}

// Execute executes the feature functionality
func (f *Feature) Execute(ctx *features.ExecutionContext) error {
	keys, ok := ctx.Options["keys"].([]string)
	if !ok || len(keys) == 0 {
		// If in interactive mode, prompt the user to enter the SSH key
		if ctx.Interactive {
			ctx.Logger.Info("SSH Key Management")
			fmt.Println("Enter SSH key sources:")
			fmt.Println("  • github:username     - Import keys from GitHub user")
			fmt.Println("  • gitlab:username     - Import keys from GitLab user")
			fmt.Println("  • url:https://...     - Import keys from URL")
			fmt.Println("  • file:/path/to/key   - Import keys from local file")
			fmt.Println("\nMultiple sources can be separated by semicolons (;)")
			fmt.Print("Or leave empty to skip: ")
			var input string
			_, _ = fmt.Scanln(&input) // Ignore error for user input
			if input != "" {
				// parsed input SSH key
				keyList := strings.Split(input, ";")
				for _, key := range keyList {
					keys = append(keys, strings.TrimSpace(key))
				}
				ctx.Options["keys"] = keys
			} else {
				ctx.Logger.Info("No SSH keys specified, skipping")
				return nil
			}
		} else {
			ctx.Logger.Info("No SSH keys specified, skipping")
			return nil
		}
	}

	// Get username
	username, ok := ctx.Options["user"].(string)
	if !ok || username == "" {
		// Use real user if not specified (considering sudo environment)
		currentUser, err := getRealUser()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		username = currentUser.Username
	}

	// Get user home directory
	homeDir := osdetect.GetUserHomeDir(username, f.osInfo)

	// Create .ssh directory if it doesn't exist
	sshDir := filepath.Join(homeDir, ".ssh")
	if !ctx.DryRun {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return fmt.Errorf("failed to create .ssh directory: %w", err)
		}
	}

	// Authorized keys file
	authKeysFile := filepath.Join(sshDir, "authorized_keys")

	// Process each key source
	var allKeys []*sshkeys.Key
	for _, keySource := range keys {
		sourceKeys, err := f.processKeySource(ctx, keySource)
		if err != nil {
			ctx.Logger.Warning("Failed to process key source %s: %v", keySource, err)
			continue
		}
		allKeys = append(allKeys, sourceKeys...)
	}

	// Skip if no keys found
	if len(allKeys) == 0 {
		ctx.Logger.Warning("No valid SSH keys found")
		return nil
	}

	// Format key information for display
	var keyLines []string
	for _, key := range allKeys {
		// Format key info with aligned type, partial content, and comment
		parts := strings.Fields(key.Content)
		var keyContent string
		if len(parts) > 1 {
			// Extract just the key content (second part), not the type
			keyContent = parts[1]
			if len(keyContent) > 30 {
				// Show first 15 and last 15 characters with ... in the middle
				keyContent = keyContent[:15] + "..." + keyContent[len(keyContent)-15:]
			}
		}

		// Format source description
		var sourceDesc string
		switch key.Source {
		case sshkeys.GitHub:
			sourceDesc = fmt.Sprintf("From GitHub (%s)", key.SourceValue)
		case sshkeys.GitLab:
			sourceDesc = fmt.Sprintf("From GitLab (%s)", key.SourceValue)
		case sshkeys.URL:
			sourceDesc = fmt.Sprintf("From URL (%s)", key.SourceValue)
		case sshkeys.File:
			sourceDesc = fmt.Sprintf("From local file (%s)", key.SourceValue)
		default:
			sourceDesc = fmt.Sprintf("From %s (%s)", key.Source, key.SourceValue)
		}

		// Format key type with padding for alignment
		keyType := fmt.Sprintf("%-8s", key.Type)

		// Add comment if available
		commentStr := ""
		if key.Comment != "" {
			commentStr = "    " + key.Comment
		}

		keyInfo := fmt.Sprintf("- %s %s%s", keyType, keyContent, commentStr)
		keyLines = append(keyLines, sourceDesc)
		keyLines = append(keyLines, keyInfo)
		keyLines = append(keyLines, "") // Add empty line between keys for better readability
	}

	// Remove the last empty line if there is one
	if len(keyLines) > 0 && keyLines[len(keyLines)-1] == "" {
		keyLines = keyLines[:len(keyLines)-1]
	}

	// Display found keys using multi-line format
	ctx.Logger.MultiLine("info", fmt.Sprintf("Found %d SSH keys:", len(allKeys)), keyLines)

	// Skip if dry run
	if ctx.DryRun {
		ctx.Logger.Info("Would add %d SSH keys to %s", len(allKeys), authKeysFile)
		return nil
	}

	// Check if backup option is enabled
	backupEnabled, hasBackup := ctx.Options["backup"].(bool)

	// Backup existing authorized_keys file if it exists
	if _, err := os.Stat(authKeysFile); err == nil {
		backupPath, err := utils.BackupFile(authKeysFile, hasBackup && backupEnabled)
		if err != nil {
			return fmt.Errorf("failed to create backup of authorized_keys file: %w", err)
		}
		if backupPath != "" {
			ctx.Logger.Info("Created backup of authorized_keys file: %s", backupPath)
		}
	}

	// First, clean any duplicate keys that might already exist in the file
	ctx.Logger.Step("Cleaning any duplicate SSH keys...")
	if err := sshkeys.CleanDuplicateKeys(authKeysFile); err != nil {
		ctx.Logger.Warning("Failed to clean duplicate keys: %v", err)
		// Continue anyway, this is not a critical error
	}

	// Write keys to authorized_keys file
	ctx.Logger.Step("Installing keys for user '%s'...", username)
	if err := sshkeys.WriteToAuthorizedKeys(authKeysFile, allKeys, true); err != nil {
		return fmt.Errorf("failed to write keys to authorized_keys: %w", err)
	}

	// Set correct permissions
	if err := os.Chmod(authKeysFile, 0600); err != nil {
		return fmt.Errorf("failed to set permissions on authorized_keys: %w", err)
	}

	// Set correct ownership
	if os.Geteuid() == 0 {
		// Get user UID and GID
		u, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("failed to lookup user: %w", err)
		}

		// Parse UID and GID
		uid, _ := fmt.Sscanf(u.Uid, "%d", new(int))
		gid, _ := fmt.Sscanf(u.Gid, "%d", new(int))

		// Set ownership
		if err := os.Chown(sshDir, uid, gid); err != nil {
			return fmt.Errorf("failed to set ownership of .ssh directory: %w", err)
		}
		if err := os.Chown(authKeysFile, uid, gid); err != nil {
			return fmt.Errorf("failed to set ownership of authorized_keys: %w", err)
		}
	}

	ctx.Logger.Success("Keys installed successfully")

	// Display final SSH key summary
	displaySSHKeySummary(ctx, allKeys, username, authKeysFile)

	return nil
}

// Priority returns the feature execution priority
func (f *Feature) Priority() int {
	return 20 // SSH key import should run after user creation
}

// getRealUser returns the real user, even when running with sudo
func getRealUser() (*user.User, error) {
	// Check if we're running with sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" && os.Geteuid() == 0 {
		// We're running with sudo, get the original user
		return user.Lookup(sudoUser)
	}

	// Not running with sudo or couldn't get SUDO_USER, return current user
	return user.Current()
}

// DetectCurrentState detects and returns the current state of the SSH feature
func (f *Feature) DetectCurrentState(ctx *features.ExecutionContext) (map[string]any, error) {
	state := make(map[string]any)

	// Get username from options or use current user
	username, ok := ctx.Options["user"].(string)
	if !ok || username == "" {
		// Use real user if not specified (considering sudo environment)
		currentUser, err := getRealUser()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		username = currentUser.Username
	}

	state["username"] = username

	// Get user home directory
	homeDir := osdetect.GetUserHomeDir(username, f.osInfo)
	state["home_dir"] = homeDir

	// Check if .ssh directory exists
	sshDir := filepath.Join(homeDir, ".ssh")
	sshDirExists := false
	if _, err := os.Stat(sshDir); err == nil {
		sshDirExists = true
	}
	state["ssh_dir_exists"] = sshDirExists

	// Check if authorized_keys file exists and read its content
	authKeysFile := filepath.Join(sshDir, "authorized_keys")
	authKeysExists := false
	var existingKeys []*sshkeys.Key

	if _, err := os.Stat(authKeysFile); err == nil {
		authKeysExists = true

		// Read authorized_keys file
		content, err := os.ReadFile(authKeysFile)
		if err == nil {
			// Parse keys
			existingKeys, _ = sshkeys.ParseKeysFromText(string(content))
		}
	}

	state["auth_keys_exists"] = authKeysExists
	state["existing_keys"] = existingKeys
	state["existing_key_count"] = len(existingKeys)

	return state, nil
}

// DisplayCurrentState displays the current state of the SSH feature
func (f *Feature) DisplayCurrentState(ctx *features.ExecutionContext, state map[string]any) {
	if !ctx.Interactive {
		return
	}

	sshDirExists, _ := state["ssh_dir_exists"].(bool)
	authKeysExists, _ := state["auth_keys_exists"].(bool)
	existingKeyCount, _ := state["existing_key_count"].(int)
	homeDir, _ := state["home_dir"].(string)

	// SSH directory status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "SSH Directory")
	if sshDirExists {
		fmt.Printf("\033[1;32m✓ Exists\033[0m")
		sshDir := filepath.Join(homeDir, ".ssh")
		fmt.Printf(" \033[90m(%s)\033[0m\n", sshDir)
	} else {
		fmt.Printf("\033[1;31m✗ Not found\033[0m\n")
		fmt.Printf("    \033[90mSSH directory will be created when you run INIQ with SSH key options\033[0m\n")
		return
	}

	// authorized_keys file status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Authorized Keys")
	if authKeysExists {
		fmt.Printf("\033[1;32m✓ Exists\033[0m\n")
	} else {
		fmt.Printf("\033[1;31m✗ Not found\033[0m\n")
		fmt.Printf("    \033[90mAuthorized keys file will be created when you run INIQ with SSH key options\033[0m\n")
		return
	}

	// SSH keys status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "SSH Keys")
	if existingKeyCount > 0 {
		fmt.Printf("\033[1;32m✓ %d key(s) found\033[0m\n", existingKeyCount)
	} else {
		fmt.Printf("\033[1;31m✗ No keys found\033[0m\n")
		fmt.Printf("    \033[90mSSH keys will be added when you run INIQ with SSH key options\033[0m\n")
		return
	}

	// Display existing keys
	existingKeys, ok := state["existing_keys"].([]*sshkeys.Key)
	if !ok || len(existingKeys) == 0 {
		return
	}

	// Format key information for display
	for i, key := range existingKeys {
		// Format key info with aligned type, partial content, and comment
		parts := strings.Fields(key.Content)
		var keyType, keyContent, comment string

		if len(parts) > 0 {
			keyType = parts[0]
		}

		if len(parts) > 1 {
			// Extract just the key content (second part), not the type
			keyContent = parts[1]
			if len(keyContent) > 45 {
				// Show first 20 and last 20 characters with ... in the middle
				keyContent = keyContent[:20] + "..." + keyContent[len(keyContent)-20:]
			}
		}

		if key.Comment != "" {
			comment = key.Comment
		} else if len(parts) > 2 {
			comment = strings.Join(parts[2:], " ")
		}

		// Use gray for key type, white for key content, blue for comments
		fmt.Printf("    %d. \033[90m%s\033[0m \033[0;37m%s\033[0m", i+1, keyType, keyContent)
		if comment != "" {
			fmt.Printf(" \033[0;36m%s\033[0m", comment)
		}
		fmt.Println()
	}
}

// ShouldPromptUser determines if the user should be prompted for input
func (f *Feature) ShouldPromptUser(ctx *features.ExecutionContext, state map[string]any) bool {
	if !ctx.Interactive {
		return false
	}

	// Always prompt for SSH keys, as users might want to add more keys
	// even if they already have some
	return true
}

// processKeySource processes a key source and returns the keys
func (f *Feature) processKeySource(ctx *features.ExecutionContext, keySource string) ([]*sshkeys.Key, error) {
	// Parse key source
	source, value, err := parseKeySource(keySource)
	if err != nil {
		return nil, err
	}

	ctx.Logger.Info("Processing SSH key source: %s:%s", source, value)

	// Get keys based on source
	switch source {
	case "github", "gh":
		return sshkeys.FetchFromGitHub(value)
	case "gitlab", "gl":
		return sshkeys.FetchFromGitLab(value)
	case "url":
		return sshkeys.FetchFromURL(value)
	case "file":
		return sshkeys.ReadFromFile(value)
	default:
		// Try as file path if no prefix
		return sshkeys.ReadFromFile(keySource)
	}
}

// validateKeySource validates a key source
func validateKeySource(keySource string) error {
	// Parse key source
	source, value, err := parseKeySource(keySource)
	if err != nil {
		return err
	}

	// Validate based on source
	switch source {
	case "github", "gh", "gitlab", "gl":
		if value == "" {
			return fmt.Errorf("username is required for %s", source)
		}
	case "url":
		if value == "" {
			return fmt.Errorf("URL is required")
		}
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("invalid URL: %s (must start with http:// or https://)", value)
		}
	case "file":
		if value == "" {
			return fmt.Errorf("file path is required")
		}
	default:
		// Try as file path if no prefix
		if _, err := os.Stat(keySource); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", keySource)
		}
	}

	return nil
}

// parseKeySource parses a key source string into source and value
func parseKeySource(keySource string) (string, string, error) {
	// Check for source prefix
	parts := strings.SplitN(keySource, ":", 2)
	if len(parts) == 2 {
		source := strings.ToLower(parts[0])
		value := parts[1]

		// Validate source
		switch source {
		case "github", "gh", "gitlab", "gl", "url", "file", "f":
			// Map short forms to full forms
			if source == "gh" {
				source = "github"
			} else if source == "gl" {
				source = "gitlab"
			} else if source == "f" {
				source = "file"
			}
			return source, value, nil
		default:
			return "", "", fmt.Errorf("invalid key source: %s", source)
		}
	}

	// No prefix, assume file path
	return "file", keySource, nil
}

// displaySSHKeySummary displays a summary of the imported SSH keys
func displaySSHKeySummary(ctx *features.ExecutionContext, keys []*sshkeys.Key, username, authKeysFile string) {
	if len(keys) == 0 {
		return
	}

	// Group keys by source
	keysBySource := make(map[string][]*sshkeys.Key)
	for _, key := range keys {
		sourceKey := fmt.Sprintf("%s:%s", key.Source, key.SourceValue)
		keysBySource[sourceKey] = append(keysBySource[sourceKey], key)
	}

	// Create summary lines
	var summaryLines []string
	summaryLines = append(summaryLines, "SSH Keys Summary")
	summaryLines = append(summaryLines, "---------------")
	summaryLines = append(summaryLines, fmt.Sprintf("%d keys imported to %s", len(keys), authKeysFile))
	summaryLines = append(summaryLines, "")

	// Add keys grouped by source
	for sourceKey, sourceKeys := range keysBySource {
		parts := strings.SplitN(sourceKey, ":", 2)
		source := parts[0]
		value := parts[1]

		// Format source description
		var sourceDesc string
		switch sshkeys.Source(source) {
		case sshkeys.GitHub:
			sourceDesc = fmt.Sprintf("From GitHub (%s):", value)
		case sshkeys.GitLab:
			sourceDesc = fmt.Sprintf("From GitLab (%s):", value)
		case sshkeys.URL:
			sourceDesc = fmt.Sprintf("From URL (%s):", value)
		case sshkeys.File:
			sourceDesc = fmt.Sprintf("From local file (%s):", value)
		default:
			sourceDesc = fmt.Sprintf("From %s (%s):", source, value)
		}

		summaryLines = append(summaryLines, sourceDesc)

		// Add keys for this source
		for _, key := range sourceKeys {
			// Format key info with aligned type, partial content, and comment
			parts := strings.Fields(key.Content)
			var keyContent string
			if len(parts) > 1 {
				// Extract just the key content (second part), not the type
				keyContent = parts[1]
				if len(keyContent) > 45 {
					// Show first 20 and last 20 characters with ... in the middle
					keyContent = keyContent[:20] + "..." + keyContent[len(keyContent)-20:]
				}
			}

			// Format key type with padding for alignment
			keyType := fmt.Sprintf("%-8s", key.Type)

			// Add comment if available
			commentStr := ""
			if key.Comment != "" {
				commentStr = "    " + key.Comment
			}

			keyInfo := fmt.Sprintf("- %s %s%s", keyType, keyContent, commentStr)
			summaryLines = append(summaryLines, keyInfo)
		}

		summaryLines = append(summaryLines, "")
	}

	// Remove the last empty line if there is one
	if len(summaryLines) > 0 && summaryLines[len(summaryLines)-1] == "" {
		summaryLines = summaryLines[:len(summaryLines)-1]
	}

	// Print summary
	fmt.Println()
	for _, line := range summaryLines {
		fmt.Println(line)
	}
	fmt.Println()
}
