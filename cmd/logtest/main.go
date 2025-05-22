package main

import (
	"fmt"
	"os"

	"github.com/teomyth/iniq/internal/logger"
)

func main() {
	// Create a new logger
	log := logger.New(true, false)

	// Print header
	log.PrintHeader("INIQ v1.0.0", "Cross-platform system initialization tool")
	fmt.Println("Starting system initialization...")
	fmt.Printf("System detected: %s\n", "Ubuntu 22.04 LTS")

	// Define operations
	operations := []string{
		"Create user 'devops'",
		"Configure SSH keys",
		"Set up sudo permissions",
		"Configure SSH security settings",
	}

	// Print operations list
	log.PrintOperationList(operations)

	// User creation operation
	log.StartOperation("Creating user 'devops'")
	log.StepWithIndent("Checking if user already exists...")
	log.IncreaseIndent()
	log.TextWithIndent("User does not exist, proceeding with creation")
	log.DecreaseIndent()
	log.StepWithIndent("Creating user 'devops'...")
	log.IncreaseIndent()
	log.SuccessWithIndent("User created successfully")
	log.DecreaseIndent()
	log.StepWithIndent("Setting up home directory...")
	log.IncreaseIndent()
	log.SuccessWithIndent("Home directory created at /home/devops")
	log.DecreaseIndent()
	log.StepWithIndent("Setting shell to /bin/bash...")
	log.IncreaseIndent()
	log.SuccessWithIndent("Shell configured")
	log.DecreaseIndent()
	log.SuccessWithIndent("User 'devops' created successfully")

	// SSH keys operation
	log.StartOperation("Configuring SSH keys")
	log.StepWithIndent("Creating .ssh directory...")
	log.IncreaseIndent()
	log.SuccessWithIndent("Directory created at /home/devops/.ssh")
	log.DecreaseIndent()
	log.StepWithIndent("Adding SSH keys...")
	log.IncreaseIndent()
	log.ListItemWithIndent("ssh-rsa AAAAB3Nz... (RSA)")
	log.ListItemWithIndent("ssh-ed25519 AAAAC3... (ED25519)")
	log.SuccessWithIndent("Keys added to authorized_keys")
	log.DecreaseIndent()
	log.StepWithIndent("Setting proper permissions...")
	log.IncreaseIndent()
	log.SuccessWithIndent("Permissions set to 700 for .ssh directory")
	log.SuccessWithIndent("Permissions set to 600 for authorized_keys file")
	log.DecreaseIndent()
	log.SuccessWithIndent("SSH keys configured successfully")

	// Sudo permissions operation
	log.StartOperation("Setting up sudo permissions")
	log.StepWithIndent("Checking sudo configuration...")
	log.IncreaseIndent()
	log.TextWithIndent("Previous setting: none")
	log.DecreaseIndent()
	log.StepWithIndent("Adding sudo configuration for user 'devops'...")
	log.IncreaseIndent()
	log.TextWithIndent("Adding: devops ALL=(ALL) NOPASSWD: ALL")
	log.SuccessWithIndent("Configuration added to /etc/sudoers.d/devops")
	log.DecreaseIndent()
	log.StepWithIndent("Setting proper permissions...")
	log.IncreaseIndent()
	log.SuccessWithIndent("Permissions set to 440 for sudo configuration")
	log.DecreaseIndent()
	log.SuccessWithIndent("Sudo permissions configured successfully")

	// SSH security settings operation
	log.StartOperation("Configuring SSH security settings")
	log.StepWithIndent("Modifying SSH configuration...")
	log.IncreaseIndent()
	log.ListItemWithIndent("PasswordAuthentication: yes → no")
	log.ListItemWithIndent("PermitRootLogin: yes → no")
	log.SuccessWithIndent("Settings updated")
	log.DecreaseIndent()
	log.StepWithIndent("Restarting SSH service...")
	log.IncreaseIndent()
	log.SuccessWithIndent("SSH service restarted")
	log.DecreaseIndent()
	log.SuccessWithIndent("SSH security settings configured successfully")

	// Print operation summary
	results := map[string]bool{
		"User 'devops' created successfully":            true,
		"SSH keys configured successfully":              true,
		"Sudo permissions configured successfully":      true,
		"SSH security settings configured successfully": true,
	}
	log.PrintOperationSummary(results)

	// Print system configuration
	configs := map[string]string{
		"User":                "devops (with sudo access)",
		"SSH Password Auth":   "disabled",
		"SSH Root Login":      "disabled",
		"SSH Authorized Keys": "2 keys configured",
	}
	log.PrintSystemConfig(configs)

	os.Exit(0)
}
