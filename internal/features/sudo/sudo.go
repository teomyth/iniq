// Package sudo implements the sudo configuration feature
package sudo

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/internal/utils"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// Feature implements the sudo configuration feature
type Feature struct {
	osInfo *osdetect.Info
}

// New creates a new sudo configuration feature
func New(osInfo *osdetect.Info) *Feature {
	return &Feature{
		osInfo: osInfo,
	}
}

// Name returns the feature name
func (f *Feature) Name() string {
	return "sudo"
}

// Description returns the feature description
func (f *Feature) Description() string {
	return "Configure sudo privileges for users"
}

// Flags returns the command-line flags for the feature
func (f *Feature) Flags() []features.Flag {
	return []features.Flag{
		{
			Name:      "sudo-nopasswd",
			Shorthand: "",
			Usage:     "configure sudo without password",
			Default:   true,
			Required:  false,
		},
		{
			Name:      "skip-sudo",
			Shorthand: "S",
			Usage:     "skip operations requiring sudo",
			Default:   false,
			Required:  false,
		},
	}
}

// ShouldActivate determines if the feature should be activated
func (f *Feature) ShouldActivate(options map[string]any) bool {
	// In interactive mode, always activate the sudo function unless explicitly skipped
	interactive, hasInteractive := options["interactive"].(bool)
	skipSudo, hasSkipSudo := options["skip-sudo"].(bool)

	if hasInteractive && interactive {
		return !hasSkipSudo || !skipSudo
	}

	// Otherwise, it will only be activated if the username is provided and sudo is not skipped
	username, hasUser := options["user"].(string)
	return hasUser && username != "" && (!hasSkipSudo || !skipSudo)
}

// ValidateOptions validates the feature options
func (f *Feature) ValidateOptions(options map[string]any) error {
	// In interactive mode, there is no need to verify the user name, as the user will be prompted for input in the Execute method.
	interactive, hasInteractive := options["interactive"].(bool)
	if hasInteractive && interactive {
		return nil
	}

	// Check if the username is valid
	username, ok := options["user"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username is required")
	}

	return nil
}

// Execute executes the feature functionality
func (f *Feature) Execute(ctx *features.ExecutionContext) error {
	username := ctx.Options["user"].(string)
	nopasswd, hasNopasswd := ctx.Options["sudo-nopasswd"].(bool)

	// If the sudo-nopasswd option is not specified in interactive mode, prompt the user
	if ctx.Interactive && !hasNopasswd {
		ctx.Logger.Info("Sudo Configuration")

		// Use the utility function, default value is true (Y)
		nopasswd = utils.PromptYesNo("Enable passwordless sudo?", true)
		ctx.Options["sudo-nopasswd"] = nopasswd
	}

	// If the sudo-nopasswd option is not specified, use the default value (enabled)
	if !ctx.Interactive && !hasNopasswd {
		nopasswd = true
		ctx.Options["sudo-nopasswd"] = true
	}

	// Skip if dry run
	if ctx.DryRun {
		if nopasswd {
			ctx.Logger.Info("Would configure passwordless sudo for user %s", username)
		} else {
			ctx.Logger.Info("Would configure sudo with password for user %s", username)
		}
		return nil
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		// Not running as root, need to use sudo
		ctx.Logger.Warning("Configuring sudo requires root privileges")

		// Check if user was recently added to sudo group but hasn't logged out yet
		currentUser, err := user.Current()
		if err == nil && isUserInSudoGroupButNotActive(currentUser.Username) {
			ctx.Logger.Warning("You were recently added to the sudo group, but this change requires logging out and back in to take effect")
			ctx.Logger.Info("Please log out and log back in, then run INIQ again")
			return fmt.Errorf("sudo group membership not yet active; please log out and log back in")
		}

		ctx.Logger.Info("INIQ will now use sudo to configure sudo permissions")
		ctx.Logger.Info("You may be prompted for your password")

		// Determine the commands to run based on OS
		switch f.osInfo.Type {
		case osdetect.Linux:
			// Create sudoers.d directory and file
			sudoersDir := "/etc/sudoers.d"
			sudoersFile := filepath.Join(sudoersDir, username)
			var sudoersContent string
			if nopasswd {
				sudoersContent = fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL", username)
			} else {
				sudoersContent = fmt.Sprintf("%s ALL=(ALL) ALL", username)
			}

			// Show configuration details
			passwordRequired := "yes"
			if nopasswd {
				passwordRequired = "no"
			}
			configLines := []string{
				fmt.Sprintf("User: %s", username),
				"Sudo access: ALL commands",
				fmt.Sprintf("Password required: %s", passwordRequired),
				fmt.Sprintf("Configuration file: %s", sudoersFile),
			}
			ctx.Logger.MultiLine("info", "Configuring sudo with the following settings:", configLines)

			// Step 1: Create sudoers.d directory
			ctx.Logger.Step("Creating sudoers.d directory...")
			mkdirCmd := exec.Command("sudo", "mkdir", "-p", sudoersDir)
			mkdirCmd.Stdin = os.Stdin
			mkdirCmd.Stdout = os.Stdout
			mkdirCmd.Stderr = os.Stderr
			if err := mkdirCmd.Run(); err != nil {
				// Check if this is likely due to sudo group membership not being active yet
				if strings.Contains(err.Error(), "not in the sudoers file") ||
					strings.Contains(err.Error(), "not allowed to execute") {
					ctx.Logger.Warning("You appear to be in the sudo group, but the membership is not yet active")
					ctx.Logger.Info("This typically requires logging out and back in to take effect")
					ctx.Logger.Info("Please log out and log back in, then run INIQ again")
					return fmt.Errorf("sudo group membership not yet active; please log out and log back in")
				}
				return fmt.Errorf("failed to create sudoers.d directory: %w", err)
			}

			// Step 2: Create temporary sudoers file
			tempFile := filepath.Join(os.TempDir(), "iniq_sudoers_temp")
			if err := os.WriteFile(tempFile, []byte(sudoersContent), 0644); err != nil {
				return fmt.Errorf("failed to create temporary sudoers file: %w", err)
			}
			defer os.Remove(tempFile)

			// Step 3: Move file to sudoers.d with sudo
			ctx.Logger.Step("Creating sudoers file %s", sudoersFile)
			mvCmd := exec.Command("sudo", "cp", tempFile, sudoersFile)
			mvCmd.Stdin = os.Stdin
			mvCmd.Stdout = os.Stdout
			mvCmd.Stderr = os.Stderr
			if err := mvCmd.Run(); err != nil {
				return fmt.Errorf("failed to create sudoers file: %w", err)
			}

			// Step 4: Set correct permissions
			chmodCmd := exec.Command("sudo", "chmod", "0440", sudoersFile)
			chmodCmd.Stdin = os.Stdin
			chmodCmd.Stdout = os.Stdout
			chmodCmd.Stderr = os.Stderr
			if err := chmodCmd.Run(); err != nil {
				return fmt.Errorf("failed to set permissions on sudoers file: %w", err)
			}

			// Step 5: Validate sudoers file
			ctx.Logger.Step("Validating sudoers file...")
			visudoCmd := exec.Command("sudo", "visudo", "-c", "-f", sudoersFile)
			visudoCmd.Stdin = os.Stdin
			visudoCmd.Stdout = os.Stdout
			visudoCmd.Stderr = os.Stderr
			if err := visudoCmd.Run(); err != nil {
				// Remove invalid file
				rmCmd := exec.Command("sudo", "rm", sudoersFile)
				_ = rmCmd.Run() // Ignore errors here
				return fmt.Errorf("invalid sudoers file: %w", err)
			}

		case osdetect.Darwin:
			// On macOS, we need to add the user to the admin group
			ctx.Logger.Step("Adding user to admin group...")
			adminCmd := exec.Command("sudo", "dseditgroup", "-o", "edit", "-a", username, "-t", "user", "admin")
			adminCmd.Stdin = os.Stdin
			adminCmd.Stdout = os.Stdout
			adminCmd.Stderr = os.Stderr
			if err := adminCmd.Run(); err != nil {
				return fmt.Errorf("failed to add user to admin group: %w", err)
			}

			// For passwordless sudo, we need to modify the sudoers file
			if nopasswd {
				sudoersDir := "/etc/sudoers.d"
				sudoersFile := filepath.Join(sudoersDir, username)
				sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL", username)

				// Step 1: Create sudoers.d directory
				ctx.Logger.Step("Creating sudoers.d directory...")
				mkdirCmd := exec.Command("sudo", "mkdir", "-p", sudoersDir)
				mkdirCmd.Stdin = os.Stdin
				mkdirCmd.Stdout = os.Stdout
				mkdirCmd.Stderr = os.Stderr
				if err := mkdirCmd.Run(); err != nil {
					return fmt.Errorf("failed to create sudoers.d directory: %w", err)
				}

				// Step 2: Create temporary sudoers file
				tempFile := filepath.Join(os.TempDir(), "iniq_sudoers_temp")
				if err := os.WriteFile(tempFile, []byte(sudoersContent), 0644); err != nil {
					return fmt.Errorf("failed to create temporary sudoers file: %w", err)
				}
				defer os.Remove(tempFile)

				// Step 3: Move file to sudoers.d with sudo
				ctx.Logger.Step("Creating sudoers file %s", sudoersFile)
				mvCmd := exec.Command("sudo", "cp", tempFile, sudoersFile)
				mvCmd.Stdin = os.Stdin
				mvCmd.Stdout = os.Stdout
				mvCmd.Stderr = os.Stderr
				if err := mvCmd.Run(); err != nil {
					return fmt.Errorf("failed to create sudoers file: %w", err)
				}

				// Step 4: Set correct permissions
				chmodCmd := exec.Command("sudo", "chmod", "0440", sudoersFile)
				chmodCmd.Stdin = os.Stdin
				chmodCmd.Stdout = os.Stdout
				chmodCmd.Stderr = os.Stderr
				if err := chmodCmd.Run(); err != nil {
					return fmt.Errorf("failed to set permissions on sudoers file: %w", err)
				}

				// Step 5: Validate sudoers file
				ctx.Logger.Step("Validating sudoers file...")
				visudoCmd := exec.Command("sudo", "visudo", "-c", "-f", sudoersFile)
				visudoCmd.Stdin = os.Stdin
				visudoCmd.Stdout = os.Stdout
				visudoCmd.Stderr = os.Stderr
				if err := visudoCmd.Run(); err != nil {
					// Remove invalid file
					rmCmd := exec.Command("sudo", "rm", sudoersFile)
					_ = rmCmd.Run() // Ignore errors here
					return fmt.Errorf("invalid sudoers file: %w", err)
				}
			}
		}

		ctx.Logger.Success("Sudo configured successfully")
		return nil
	}

	// Configure sudo based on OS
	switch f.osInfo.Type {
	case osdetect.Linux:
		return f.configureLinuxSudo(ctx, username, nopasswd)
	case osdetect.Darwin:
		return f.configureDarwinSudo(ctx, username, nopasswd)
	default:
		return fmt.Errorf("unsupported OS: %s", f.osInfo.Type)
	}
}

// Priority returns the feature execution priority
func (f *Feature) Priority() int {
	return 30 // Sudo configuration should run after user creation and SSH key import
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

// DetectCurrentState detects and returns the current state of the sudo feature
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

	// Check if user exists
	_, err := user.Lookup(username)
	if err != nil {
		state["user_exists"] = false
		return state, nil
	}

	state["user_exists"] = true

	// Check if user has sudo privileges
	hasSudo, err := userHasSudo(username)
	if err != nil {
		ctx.Logger.Warning("Failed to check sudo privileges: %v", err)
	}
	state["has_sudo"] = hasSudo

	// Check if user is in sudo group
	inSudoGroup, err := isUserInSudoGroup(username)
	if err != nil {
		ctx.Logger.Warning("Failed to check sudo group membership: %v", err)
	}
	state["in_sudo_group"] = inSudoGroup

	// Check if sudo is passwordless
	hasPasswordlessSudo, err := hasPasswordlessSudo(username)
	if err != nil {
		ctx.Logger.Warning("Failed to check passwordless sudo: %v", err)
	}
	state["has_passwordless_sudo"] = hasPasswordlessSudo

	// Check if running as root
	state["is_root"] = os.Geteuid() == 0

	return state, nil
}

// DisplayCurrentState displays the current state of the sudo feature
func (f *Feature) DisplayCurrentState(ctx *features.ExecutionContext, state map[string]any) {
	if !ctx.Interactive {
		return
	}

	username := state["username"].(string)
	userExists, _ := state["user_exists"].(bool)

	// User status
	fmt.Printf("  \033[1;34m%s\033[0m: \033[0;37m%s\033[0m\n", "Username", username)

	if !userExists {
		fmt.Printf("  \033[1;34m%s\033[0m: \033[1;33m⚠ User does not exist\033[0m\n", "User Status")
		fmt.Printf("    \033[90mSudo configuration will be applied after user creation\033[0m\n")
		return
	}

	hasSudo, _ := state["has_sudo"].(bool)
	inSudoGroup, _ := state["in_sudo_group"].(bool)
	hasPasswordlessSudo, _ := state["has_passwordless_sudo"].(bool)
	isRoot, _ := state["is_root"].(bool)

	// Sudo access status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Sudo Access")
	if hasSudo {
		fmt.Printf("\033[1;32m✓ Enabled\033[0m\n")
	} else if inSudoGroup {
		fmt.Printf("\033[1;33m⚠ In sudo group but not active\033[0m\n")
		fmt.Printf("    \033[90mYou may need to log out and log back in for sudo privileges to take effect\033[0m\n")
	} else {
		fmt.Printf("\033[1;31m✗ Disabled\033[0m\n")
	}

	// Passwordless sudo status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Passwordless Sudo")
	if hasPasswordlessSudo {
		fmt.Printf("\033[1;32m✓ Enabled\033[0m\n")
	} else if hasSudo {
		fmt.Printf("\033[1;31m✗ Disabled\033[0m \033[90m(password required)\033[0m\n")
	} else {
		fmt.Printf("\033[1;31m✗ Not configured\033[0m\n")
	}

	// Sudo group membership status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Sudo Group")
	if inSudoGroup {
		fmt.Printf("\033[1;32m✓ Member\033[0m\n")
	} else {
		fmt.Printf("\033[1;31m✗ Not a member\033[0m\n")
	}

	// Root status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Root Privileges")
	if isRoot {
		fmt.Printf("\033[1;32m✓ Running as root\033[0m\n")
	} else {
		fmt.Printf("\033[1;31m✗ Not running as root\033[0m\n")
	}
}

// ShouldPromptUser determines if the user should be prompted for input
func (f *Feature) ShouldPromptUser(ctx *features.ExecutionContext, state map[string]any) bool {
	if !ctx.Interactive {
		return false
	}

	userExists, _ := state["user_exists"].(bool)
	hasSudo, _ := state["has_sudo"].(bool)

	// If user doesn't exist, we should prompt
	if !userExists {
		return true
	}

	// If user already has sudo, we should prompt to confirm if they want to change it
	return hasSudo
}

// userHasSudo checks if a user has sudo privileges
func userHasSudo(username string) (bool, error) {
	// Check if user is root (always has sudo)
	if username == "root" {
		return true, nil
	}

	// Check if current user is the target user
	currentUser, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("failed to get current user: %w", err)
	}

	// If we're checking the current user, try to run sudo -n true
	if currentUser.Username == username {
		cmd := exec.Command("sudo", "-n", "true")
		err := cmd.Run()

		// If the command succeeds, user has sudo
		if err == nil {
			return true, nil
		}

		// If the error contains "no tty present" or "askpass", it means
		// the user has sudo but needs to enter a password (which is normal)
		if strings.Contains(err.Error(), "no tty present") || strings.Contains(err.Error(), "askpass") {
			return true, nil
		}
	}

	// Check if user is in sudo group
	inSudoGroup, err := isUserInSudoGroup(username)
	if err != nil {
		return false, err
	}

	return inSudoGroup, nil
}

// isUserInSudoGroup checks if a user is in the sudo group
func isUserInSudoGroup(username string) (bool, error) {
	// Get groups for user
	cmd := exec.Command("groups", username)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get groups for user %s: %w", username, err)
	}

	// Parse output and check for sudo/admin/wheel group
	groups := strings.Fields(string(output))
	for _, group := range groups {
		if group == "sudo" || group == "admin" || group == "wheel" {
			return true, nil
		}
	}

	return false, nil
}

// hasPasswordlessSudo checks if a user has passwordless sudo
func hasPasswordlessSudo(username string) (bool, error) {
	// Check sudoers file
	sudoersFile := filepath.Join("/etc/sudoers.d", username)
	if _, err := os.Stat(sudoersFile); err == nil {
		// Read sudoers file
		content, err := os.ReadFile(sudoersFile)
		if err != nil {
			return false, fmt.Errorf("failed to read sudoers file: %w", err)
		}

		// Check if file contains NOPASSWD
		return strings.Contains(string(content), "NOPASSWD"), nil
	}

	// Check main sudoers file
	cmd := exec.Command("sudo", "-n", "grep", username, "/etc/sudoers")
	output, err := cmd.Output()
	if err == nil {
		// Check if output contains NOPASSWD
		return strings.Contains(string(output), "NOPASSWD"), nil
	}

	// If we can't check, assume no passwordless sudo
	return false, nil
}

// configureLinuxSudo configures sudo for a user on Linux
func (f *Feature) configureLinuxSudo(ctx *features.ExecutionContext, username string, nopasswd bool) error {
	ctx.Logger.Info("Configuring sudo for user %s on Linux", username)

	// Check if user exists
	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Create sudoers.d directory if it doesn't exist
	sudoersDir := "/etc/sudoers.d"
	if err := os.MkdirAll(sudoersDir, 0755); err != nil {
		return fmt.Errorf("failed to create sudoers.d directory: %w", err)
	}

	// Create sudoers file for user
	sudoersFile := filepath.Join(sudoersDir, username)
	var sudoersContent string
	if nopasswd {
		sudoersContent = fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL\n", username)
	} else {
		sudoersContent = fmt.Sprintf("%s ALL=(ALL) ALL\n", username)
	}

	// Show configuration details
	passwordRequired := "yes"
	if nopasswd {
		passwordRequired = "no"
	}
	configLines := []string{
		fmt.Sprintf("User: %s", username),
		"Sudo access: ALL commands",
		fmt.Sprintf("Password required: %s", passwordRequired),
		fmt.Sprintf("Configuration file: %s", sudoersFile),
	}
	ctx.Logger.MultiLine("info", "Configuring sudo with the following settings:", configLines)

	// Check if backup option is enabled
	backupEnabled, hasBackup := ctx.Options["backup"].(bool)

	// Backup existing sudoers file if it exists
	if _, err := os.Stat(sudoersFile); err == nil {
		backupPath, err := utils.BackupFile(sudoersFile, hasBackup && backupEnabled)
		if err != nil {
			return fmt.Errorf("failed to create backup of sudoers file: %w", err)
		}
		if backupPath != "" {
			ctx.Logger.Info("Created backup of sudoers file: %s", backupPath)
		}
	}

	// Write sudoers file
	ctx.Logger.Step("Creating sudoers file %s", sudoersFile)
	if err := os.WriteFile(sudoersFile, []byte(sudoersContent), 0440); err != nil {
		return fmt.Errorf("failed to write sudoers file: %w", err)
	}

	// Show file content
	contentLines := strings.Split(sudoersContent, "\n")
	var fileLines []string
	for _, line := range contentLines {
		if line != "" {
			fileLines = append(fileLines, line)
		}
	}
	ctx.Logger.MultiLine("info", "Sudoers file content:", fileLines)

	// Validate sudoers file
	ctx.Logger.Step("Validating sudoers file...")
	cmd := exec.Command("visudo", "-c", "-f", sudoersFile)
	if err := cmd.Run(); err != nil {
		// Remove invalid file
		os.Remove(sudoersFile)
		return fmt.Errorf("invalid sudoers file: %w", err)
	}

	ctx.Logger.Success("Sudo configured successfully")
	return nil
}

// isUserInSudoGroupButNotActive checks if a user is in the sudo group but the membership is not yet active
func isUserInSudoGroupButNotActive(username string) bool {
	// Get user information
	u, err := user.Lookup(username)
	if err != nil {
		return false
	}

	// Get user's groups
	groupIds, err := u.GroupIds()
	if err != nil {
		return false
	}

	// Check if user is in sudo group
	inSudoGroup := false
	for _, gid := range groupIds {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			continue
		}

		// Check if group name is sudo, wheel, or admin (common sudo group names)
		groupName := strings.ToLower(group.Name)
		if groupName == "sudo" || groupName == "wheel" || groupName == "admin" {
			inSudoGroup = true
			break
		}
	}

	// If not in sudo group, return false
	if !inSudoGroup {
		return false
	}

	// Try to run a simple sudo command to verify if sudo works
	// Use -n flag to prevent sudo from asking for a password
	cmd := exec.Command("sudo", "-n", "true")
	err = cmd.Run()

	// If the command fails and the user is in sudo group, it means the membership is not yet active
	return err != nil
}

// configureDarwinSudo configures sudo for a user on macOS
func (f *Feature) configureDarwinSudo(ctx *features.ExecutionContext, username string, nopasswd bool) error {
	ctx.Logger.Info("Configuring sudo for user %s on macOS", username)

	// Check if user exists
	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user %s does not exist", username)
	}

	// On macOS, we need to add the user to the admin group
	cmd := exec.Command("dseditgroup", "-o", "edit", "-a", username, "-t", "user", "admin")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add user to admin group: %w", err)
	}

	// For passwordless sudo, we need to modify the sudoers file
	if nopasswd {
		// Create sudoers.d directory if it doesn't exist
		sudoersDir := "/etc/sudoers.d"
		if err := os.MkdirAll(sudoersDir, 0755); err != nil {
			return fmt.Errorf("failed to create sudoers.d directory: %w", err)
		}

		// Create sudoers file for user
		sudoersFile := filepath.Join(sudoersDir, username)
		sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL\n", username)

		// Check if backup option is enabled
		backupEnabled, hasBackup := ctx.Options["backup"].(bool)

		// Backup existing sudoers file if it exists
		if _, err := os.Stat(sudoersFile); err == nil {
			backupPath, err := utils.BackupFile(sudoersFile, hasBackup && backupEnabled)
			if err != nil {
				return fmt.Errorf("failed to create backup of sudoers file: %w", err)
			}
			if backupPath != "" {
				ctx.Logger.Info("Created backup of sudoers file: %s", backupPath)
			}
		}

		// Write sudoers file
		if err := os.WriteFile(sudoersFile, []byte(sudoersContent), 0440); err != nil {
			return fmt.Errorf("failed to write sudoers file: %w", err)
		}

		// Validate sudoers file
		cmd := exec.Command("visudo", "-c", "-f", sudoersFile)
		if err := cmd.Run(); err != nil {
			// Remove invalid file
			os.Remove(sudoersFile)
			return fmt.Errorf("invalid sudoers file: %w", err)
		}
	}

	ctx.Logger.Success("Sudo configured for user %s", username)
	return nil
}
