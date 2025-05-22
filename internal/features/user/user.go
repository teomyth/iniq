// Package user implements the user management feature
package user

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// Feature implements the user management feature
type Feature struct {
	osInfo *osdetect.Info
}

// New creates a new user management feature
func New(osInfo *osdetect.Info) *Feature {
	return &Feature{
		osInfo: osInfo,
	}
}

// Name returns the feature name
func (f *Feature) Name() string {
	return "user"
}

// Description returns the feature description
func (f *Feature) Description() string {
	return "Create and configure users"
}

// Flags returns the command-line flags for the feature
func (f *Feature) Flags() []features.Flag {
	return []features.Flag{
		{
			Name:      "user",
			Shorthand: "u",
			Usage:     "username to create or configure",
			Default:   "",
			Required:  false,
		},
		{
			Name:      "shell",
			Shorthand: "",
			Usage:     "shell for the user",
			Default:   "/bin/bash",
			Required:  false,
		},
	}
}

// ShouldActivate determines if the feature should be activated
func (f *Feature) ShouldActivate(options map[string]any) bool {
	// In interactive mode, always activate the user function
	interactive, hasInteractive := options["interactive"].(bool)
	if hasInteractive && interactive {
		return true
	}

	// Otherwise, it will only be activated if the username is provided
	username, ok := options["user"].(string)
	return ok && username != ""
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

	// Check whether the username is valid (excluding spaces, etc.)
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
			return fmt.Errorf("invalid username: %s (only letters, numbers, underscore, hyphen, and dot are allowed)", username)
		}
	}

	return nil
}

// Execute executes the feature functionality
func (f *Feature) Execute(ctx *features.ExecutionContext) error {
	username, ok := ctx.Options["user"].(string)
	if !ok || username == "" {
		// If in interactive mode, prompt the user to enter the user name
		if ctx.Interactive {
			// Get the real user (considering sudo environment)
			realUser, err := getRealUser()
			if err == nil {
				defaultUsername := realUser.Username
				// Prompt the user to enter the user name, the default is the real user
				ctx.Logger.Info("User Management")
				if os.Geteuid() == 0 {
					// Check if we're running with sudo
					sudoUser := os.Getenv("SUDO_USER")
					if sudoUser != "" {
						fmt.Printf("You are running INIQ with sudo from user '%s'.\n", sudoUser)
					} else {
						fmt.Println("You are running INIQ as root.")
					}
					fmt.Printf("Enter username to configure (or press Enter to use '%s'): ", defaultUsername)
				} else {
					fmt.Printf("Enter username to configure (or press Enter to use yourself): ")
				}
				var input string
				fmt.Scanln(&input)
				if input == "" {
					username = defaultUsername
				} else {
					username = input
				}
				ctx.Options["user"] = username
			}
		} else {
			// In non-interactive mode, if no username is provided, use the real user
			realUser, err := getRealUser()
			if err == nil {
				username = realUser.Username
				ctx.Options["user"] = username
				ctx.Logger.Info("No username specified, using detected user: %s", username)
			} else {
				ctx.Logger.Info("No username specified and couldn't detect user, skipping")
				return nil
			}
		}
	}

	shell, _ := ctx.Options["shell"].(string)
	if shell == "" {
		shell = "/bin/bash"
	}

	// Check if user exists
	_, err := user.Lookup(username)
	if err == nil {
		ctx.Logger.Info("User %s already exists", username)

		// Check if password option is enabled
		setPassword, hasPassword := ctx.Options["password"].(bool)
		if hasPassword && setPassword {
			// Prompt for password
			password, err := promptForPassword(username)
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			// Set password
			if err := f.setUserPassword(ctx, username, password); err != nil {
				return err
			}
		}

		return nil
	}

	// User does not exist, need to create
	// Check password policy for new user creation
	setPassword, hasSetPassword := ctx.Options["password"].(bool)
	noPassword, hasNoPassword := ctx.Options["no-password"].(bool)

	// Validate password options are not conflicting
	if hasSetPassword && setPassword && hasNoPassword && noPassword {
		return fmt.Errorf("cannot specify both --password and --no-pass options")
	}

	// Determine if we need to set a password
	needsPassword := true
	if hasNoPassword && noPassword {
		needsPassword = false
	} else if hasSetPassword && setPassword {
		needsPassword = true
	} else {
		// Default behavior: require password unless explicitly disabled
		needsPassword = true
	}

	// Check if we can handle password input in current environment
	if needsPassword && !ctx.Interactive {
		return fmt.Errorf("creating user '%s' requires password input.\nOptions:\n  1. Remove -y flag to enable interactive password input\n  2. Add --no-pass flag to create user without password", username)
	}

	// Skip if dry run
	if ctx.DryRun {
		ctx.Logger.Info("Would create user %s with shell %s", username, shell)

		if needsPassword {
			ctx.Logger.Info("Would set password for user %s", username)
		} else {
			ctx.Logger.Info("Would create user %s without password", username)
		}

		return nil
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("creating users requires root privileges")
	}

	// Create user based on OS
	switch f.osInfo.Type {
	case osdetect.Linux:
		return f.createLinuxUser(ctx, username, shell, needsPassword)
	case osdetect.Darwin:
		return f.createDarwinUser(ctx, username, shell, needsPassword)
	default:
		return fmt.Errorf("unsupported OS: %s", f.osInfo.Type)
	}
}

// Priority returns the feature execution priority
func (f *Feature) Priority() int {
	return 10 // User creation should run first
}

// getUserShell returns the shell for a user
func getUserShell(username string) string {
	// Try to get shell from /etc/passwd
	cmd := exec.Command("getent", "passwd", username)
	output, err := cmd.Output()
	if err == nil {
		// Parse output (format: username:x:uid:gid:gecos:home:shell)
		parts := strings.Split(string(output), ":")
		if len(parts) >= 7 {
			return strings.TrimSpace(parts[6])
		}
	}

	// Fallback to default shell
	return "/bin/bash"
}

// userHasSudo checks if a user has sudo privileges
func userHasSudo(username string) (bool, error) {
	// Check if user is root (always has sudo)
	if username == "root" {
		return true, nil
	}

	// Check if current user is the target user
	realUser, err := getRealUser()
	if err != nil {
		return false, fmt.Errorf("failed to get real user: %w", err)
	}

	// If we're checking the current user, try to run sudo -n true
	if realUser.Username == username {
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

// DetectCurrentState detects and returns the current state of the user feature
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
		state["is_current_user"] = true
	} else {
		state["is_current_user"] = false
	}

	state["username"] = username

	// Check if user exists
	u, err := user.Lookup(username)
	if err == nil {
		// User exists
		state["user_exists"] = true
		state["user_home"] = u.HomeDir

		// Get user shell
		shell := getUserShell(u.Username)
		state["user_shell"] = shell

		// Check if user has sudo privileges
		hasSudo, err := userHasSudo(u.Username)
		if err != nil {
			ctx.Logger.Warning("Failed to check sudo privileges: %v", err)
		}
		state["has_sudo"] = hasSudo

		// Check if user is in sudo group
		inSudoGroup, err := isUserInSudoGroup(u.Username)
		if err != nil {
			ctx.Logger.Warning("Failed to check sudo group membership: %v", err)
		}
		state["in_sudo_group"] = inSudoGroup
	} else {
		// User does not exist
		state["user_exists"] = false
	}

	// Check if running as root
	state["is_root"] = os.Geteuid() == 0

	return state, nil
}

// DisplayCurrentState displays the current state of the user feature
func (f *Feature) DisplayCurrentState(ctx *features.ExecutionContext, state map[string]any) {
	if !ctx.Interactive {
		return
	}

	username := state["username"].(string)
	userExists, _ := state["user_exists"].(bool)
	isCurrentUser, _ := state["is_current_user"].(bool)

	// Use colors and icons to display status
	// Green: Configured/Enabled
	// Yellow: Warning/Needs attention
	// Red: Not configured/Disabled
	// Blue: Title/Key
	// Gray: Additional information

	// User status
	if userExists {
		userShell, _ := state["user_shell"].(string)
		hasSudo, _ := state["has_sudo"].(bool)
		inSudoGroup, _ := state["in_sudo_group"].(bool)
		userHome, _ := state["user_home"].(string)

		// Username
		fmt.Printf("  \033[1;34m%s\033[0m: \033[1;32m%s\033[0m", "Username", username)
		if isCurrentUser {
			fmt.Printf(" \033[90m(current user)\033[0m")
		}
		fmt.Println()

		// User status
		fmt.Printf("  \033[1;34m%s\033[0m: \033[1;32m✓ Exists\033[0m\n", "User Status")

		// Home directory
		if userHome != "" {
			fmt.Printf("  \033[1;34m%s\033[0m: \033[0;37m%s\033[0m\n", "Home Directory", userHome)
		}

		// User shell
		if userShell != "" {
			fmt.Printf("  \033[1;34m%s\033[0m: \033[0;37m%s\033[0m\n", "Shell", userShell)
		}

		// Sudo privileges
		if hasSudo {
			fmt.Printf("  \033[1;34m%s\033[0m: \033[1;32m✓ Enabled\033[0m\n", "Sudo Access")
		} else if inSudoGroup {
			fmt.Printf("  \033[1;34m%s\033[0m: \033[1;33m⚠ In sudo group but not active\033[0m\n", "Sudo Access")
			fmt.Printf("    \033[90mYou may need to log out and log back in for sudo privileges to take effect\033[0m\n")
		} else {
			fmt.Printf("  \033[1;34m%s\033[0m: \033[1;31m✗ Disabled\033[0m\n", "Sudo Access")
		}
	} else {
		// Username
		fmt.Printf("  \033[1;34m%s\033[0m: \033[0;37m%s\033[0m\n", "Username", username)

		// User status
		fmt.Printf("  \033[1;34m%s\033[0m: \033[1;31m✗ Does not exist\033[0m\n", "User Status")
		fmt.Printf("    \033[90mUser will be created when you run INIQ with appropriate options\033[0m\n")
	}
}

// ShouldPromptUser determines if the user should be prompted for input
func (f *Feature) ShouldPromptUser(ctx *features.ExecutionContext, state map[string]any) bool {
	if !ctx.Interactive {
		return false
	}

	userExists, _ := state["user_exists"].(bool)
	isCurrentUser, _ := state["is_current_user"].(bool)
	isRoot, _ := state["is_root"].(bool)

	// If user doesn't exist, we should prompt
	if !userExists {
		return true
	}

	// If it's the current user and we're not root, no need to prompt
	if isCurrentUser && !isRoot {
		return false
	}

	// Otherwise, prompt to confirm using existing user
	return true
}

// createLinuxUser creates a new user on Linux
func (f *Feature) createLinuxUser(ctx *features.ExecutionContext, username, shell string, needsPassword bool) error {
	ctx.Logger.Info("Creating user %s on Linux", username)

	// Create user with home directory and specified shell
	cmd := exec.Command("useradd", "-m", "-s", shell, username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	ctx.Logger.Success("User %s created successfully", username)

	// Set password if needed
	if needsPassword {
		// Prompt for password
		password, err := promptForPassword(username)
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}

		// Set password
		if err := f.setUserPassword(ctx, username, password); err != nil {
			return err
		}
	} else {
		ctx.Logger.Info("User %s created without password. Use 'passwd %s' to set password later.", username, username)
	}

	return nil
}

// createDarwinUser creates a new user on macOS
func (f *Feature) createDarwinUser(ctx *features.ExecutionContext, username, shell string, needsPassword bool) error {
	ctx.Logger.Info("Creating user %s on macOS", username)

	// Generate a unique UID and GID (starting from 501, which is the first regular user on macOS)
	uid := 501
	gid := 20 // Default group is 'staff' (20) on macOS

	// Find next available UID
	for {
		_, err := user.LookupId(fmt.Sprintf("%d", uid))
		if err != nil {
			break // UID is available
		}
		uid++
	}

	// Create user home directory
	homeDir := filepath.Join("/Users", username)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}

	// Create user using dscl
	commands := [][]string{
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username)},
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username), "UserShell", shell},
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username), "RealName", username},
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username), "UniqueID", fmt.Sprintf("%d", uid)},
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username), "PrimaryGroupID", fmt.Sprintf("%d", gid)},
		{"dscl", ".", "-create", fmt.Sprintf("/Users/%s", username), "NFSHomeDirectory", homeDir},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute %v: %w", cmdArgs, err)
		}
	}

	// Set ownership of home directory
	if err := os.Chown(homeDir, uid, gid); err != nil {
		return fmt.Errorf("failed to set ownership of home directory: %w", err)
	}

	ctx.Logger.Success("User %s created successfully", username)

	// Set password if needed
	if needsPassword {
		// Prompt for password
		password, err := promptForPassword(username)
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}

		// Set password
		if err := f.setUserPassword(ctx, username, password); err != nil {
			return err
		}
	} else {
		ctx.Logger.Info("User %s created without password. Use 'passwd %s' to set password later.", username, username)
	}

	return nil
}
