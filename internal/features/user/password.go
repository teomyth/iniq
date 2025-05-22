package user

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/teomyth/iniq/internal/features"
	"golang.org/x/term"
)

// promptForPassword prompts the user to enter a password
func promptForPassword(username string) (string, error) {
	fmt.Printf("Enter password for user '%s': ", username)
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	fmt.Print("Confirm password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read confirmation password: %w", err)
	}
	fmt.Println() // Add newline after password input

	if string(password) != string(confirmPassword) {
		return "", fmt.Errorf("passwords do not match")
	}

	return string(password), nil
}

// setUserPassword sets the password for a user
func (f *Feature) setUserPassword(ctx *features.ExecutionContext, username, password string) error {
	// Skip if dry run
	if ctx.DryRun {
		ctx.Logger.Info("Would set password for user %s", username)
		return nil
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("setting password requires root privileges")
	}

	// Set password based on OS
	switch f.osInfo.Type {
	case "linux":
		return f.setLinuxUserPassword(ctx, username, password)
	case "darwin":
		return f.setDarwinUserPassword(ctx, username, password)
	default:
		return fmt.Errorf("unsupported OS: %s", f.osInfo.Type)
	}
}

// setLinuxUserPassword sets the password for a user on Linux
func (f *Feature) setLinuxUserPassword(ctx *features.ExecutionContext, username, password string) error {
	ctx.Logger.Info("Setting password for user %s on Linux", username)

	// Use chpasswd to set the password
	cmd := exec.Command("chpasswd")
	// Note: Don't set cmd.Stdin = os.Stdin when using StdinPipe()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a pipe to write to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start chpasswd: %w", err)
	}

	// Write username:password to stdin
	_, err = fmt.Fprintf(stdin, "%s:%s\n", username, password)
	if err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	// Close stdin to signal end of input
	stdin.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	ctx.Logger.Success("Password set for user %s", username)
	return nil
}

// setDarwinUserPassword sets the password for a user on macOS
func (f *Feature) setDarwinUserPassword(ctx *features.ExecutionContext, username, password string) error {
	ctx.Logger.Info("Setting password for user %s on macOS", username)

	// Use dscl to set the password
	cmd := exec.Command("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", username), password)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	ctx.Logger.Success("Password set for user %s", username)
	return nil
}
