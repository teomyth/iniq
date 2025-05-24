// Package main is the entry point for the INIQ application
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/teomyth/iniq/internal/config"
	"github.com/teomyth/iniq/internal/features"
	_ "github.com/teomyth/iniq/internal/features/security" // Register security feature
	_ "github.com/teomyth/iniq/internal/features/ssh"      // Register SSH feature
	_ "github.com/teomyth/iniq/internal/features/sudo"     // Register sudo feature
	_ "github.com/teomyth/iniq/internal/features/user"     // Register user feature
	"github.com/teomyth/iniq/internal/logger"
	"github.com/teomyth/iniq/internal/utils"
	"github.com/teomyth/iniq/internal/version"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// These variables are kept for backward compatibility
// and are populated from the version package
var (
	Version   = version.ShortString()
	BuildDate = "unknown"
	Commit    = "unknown"
)

// Command line flags
var (
	cfgFile         string
	verbose         bool
	quiet           bool
	yes             bool
	dryRun          bool
	skipSudo        bool
	username        string
	keys            []string
	sshRootLogin    string
	sshPasswordAuth string
	sshNoRoot       bool
	sshNoPass       bool
	sudoNoPass      bool
	showStatus      bool
	backupFiles     bool
	allSecurity     bool
	setPassword     bool
	noPassword      bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"V"},
	Short:   "Display detailed version information",
	Long: `Display detailed version information about INIQ.
This includes the semantic version, build date, commit hash,
Go version, and platform information.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Display version information
		info := version.Get()

		// Use colors if terminal supports it
		useColors := utils.IsTerminal(os.Stdout)

		if useColors {
			fmt.Println(utils.ColorText("INIQ Version Information", utils.ColorHeaderCyan))
			fmt.Println(utils.ColorText("------------------------", utils.ColorGray))
			fmt.Printf("%s %s\n", utils.ColorText("Version:", utils.ColorBold), utils.ColorText(info.Version, utils.ColorSuccess))
			fmt.Printf("%s %s\n", utils.ColorText("Build Date:", utils.ColorBold), info.BuildDate)
			fmt.Printf("%s %s\n", utils.ColorText("Commit:", utils.ColorBold), info.Commit)
			fmt.Printf("%s %s\n", utils.ColorText("Go Version:", utils.ColorBold), info.GoVersion)
			fmt.Printf("%s %s\n", utils.ColorText("Platform:", utils.ColorBold), info.Platform)
		} else {
			fmt.Println("INIQ Version Information")
			fmt.Println("------------------------")
			fmt.Printf("Version:    %s\n", info.Version)
			fmt.Printf("Build Date: %s\n", info.BuildDate)
			fmt.Printf("Commit:     %s\n", info.Commit)
			fmt.Printf("Go Version: %s\n", info.GoVersion)
			fmt.Printf("Platform:   %s\n", info.Platform)
		}

		// Print additional build information if available
		if info.BuildInfo != nil && len(info.BuildInfo.Settings) > 0 {
			if useColors {
				fmt.Println("\n" + utils.ColorText("Build Settings:", utils.ColorBold))
			} else {
				fmt.Println("\nBuild Settings:")
			}

			// Collect and display relevant build settings
			buildSettings := make(map[string]string)
			for _, setting := range info.BuildInfo.Settings {
				buildSettings[setting.Key] = setting.Value
			}

			// Display CGO status
			if cgoEnabled, exists := buildSettings["CGO_ENABLED"]; exists {
				status := "Enabled"
				if cgoEnabled == "0" {
					status = "Disabled"
				}
				if useColors {
					fmt.Printf("  %s CGO: %s\n", utils.ColorText("•", utils.ColorInfo), status)
				} else {
					fmt.Printf("  • CGO: %s\n", status)
				}
			} else {
				// For local builds, CGO is typically enabled by default
				if useColors {
					fmt.Printf("  %s CGO: %s\n", utils.ColorText("•", utils.ColorInfo), "Enabled (default)")
				} else {
					fmt.Printf("  • CGO: %s\n", "Enabled (default)")
				}
			}

			// Display VCS information
			if vcsRevision, exists := buildSettings["vcs.revision"]; exists {
				if useColors {
					fmt.Printf("  %s VCS Revision: %s\n", utils.ColorText("•", utils.ColorInfo), vcsRevision[:min(len(vcsRevision), 12)])
				} else {
					fmt.Printf("  • VCS Revision: %s\n", vcsRevision[:min(len(vcsRevision), 12)])
				}
			}

			if vcsTime, exists := buildSettings["vcs.time"]; exists {
				if useColors {
					fmt.Printf("  %s VCS Time: %s\n", utils.ColorText("•", utils.ColorInfo), vcsTime)
				} else {
					fmt.Printf("  • VCS Time: %s\n", vcsTime)
				}
			}

			if vcsModified, exists := buildSettings["vcs.modified"]; exists && vcsModified == "true" {
				if useColors {
					fmt.Printf("  %s %s\n",
						utils.ColorText("•", utils.ColorWarning),
						"Built from modified source (dirty state)")
				} else {
					fmt.Println("  • Built from modified source (dirty state)")
				}
			}

			// Display compiler information
			if compiler, exists := buildSettings["GOARCH"]; exists {
				if useColors {
					fmt.Printf("  %s Target Architecture: %s\n", utils.ColorText("•", utils.ColorInfo), compiler)
				} else {
					fmt.Printf("  • Target Architecture: %s\n", compiler)
				}
			}

			if goos, exists := buildSettings["GOOS"]; exists {
				if useColors {
					fmt.Printf("  %s Target OS: %s\n", utils.ColorText("•", utils.ColorInfo), goos)
				} else {
					fmt.Printf("  • Target OS: %s\n", goos)
				}
			}

			// Display build mode if available
			if buildMode, exists := buildSettings["-buildmode"]; exists {
				if useColors {
					fmt.Printf("  %s Build Mode: %s\n", utils.ColorText("•", utils.ColorInfo), buildMode)
				} else {
					fmt.Printf("  • Build Mode: %s\n", buildMode)
				}
			}

			// Display if this is a stripped binary
			if ldflags, exists := buildSettings["-ldflags"]; exists {
				if strings.Contains(ldflags, "-s") && strings.Contains(ldflags, "-w") {
					if useColors {
						fmt.Printf("  %s %s\n", utils.ColorText("•", utils.ColorInfo), "Stripped binary (debug info removed)")
					} else {
						fmt.Println("  • Stripped binary (debug info removed)")
					}
				}
			}

			// Display build type based on version
			buildType := "Development"
			if !strings.Contains(info.Version, "dirty") && !strings.Contains(info.Version, "dev") {
				buildType = "Release"
			}
			if useColors {
				fmt.Printf("  %s Build Type: %s\n", utils.ColorText("•", utils.ColorInfo), buildType)
			} else {
				fmt.Printf("  • Build Type: %s\n", buildType)
			}
		}
	},
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "iniq",
	Short:   "INIQ - System Initialization Tool",
	Version: version.String(),
	Long: `INIQ is a cross-platform system initialization tool for Linux and macOS.
It streamlines the process of setting up new systems with proper user accounts,
SSH access, security configurations, password policy enforcement, and system hardening.`,

	// Only run the main function if no subcommand is specified
	Run: func(cmd *cobra.Command, args []string) {
		// Create logger
		log := logger.New(verbose, quiet)

		// Detect OS
		osInfo, err := osdetect.Detect()
		if err != nil {
			log.Fatal("Failed to detect operating system: %v", err)
		}

		// Check macOS support and show warnings if needed
		if err := checkMacOSSupport(osInfo, log, yes, quiet); err != nil {
			log.Info("Exiting...")
			return
		}

		// Check if running as root
		isRoot := os.Geteuid() == 0

		// Create options map from viper
		options := make(map[string]any)
		for _, key := range viper.AllKeys() {
			options[key] = viper.Get(key)
		}

		// Add command line flags that might not be in viper
		options["user"] = username
		options["keys"] = keys
		options["ssh-root-login"] = sshRootLogin
		options["ssh-password-auth"] = sshPasswordAuth
		options["ssh-no-root"] = sshNoRoot
		options["ssh-no-password"] = sshNoPass
		options["sudo-nopasswd"] = sudoNoPass
		options["skip-sudo"] = skipSudo
		options["yes"] = yes
		options["verbose"] = verbose
		options["quiet"] = quiet
		options["dry-run"] = dryRun
		options["status"] = showStatus
		options["backup"] = backupFiles
		options["all"] = allSecurity
		options["password"] = setPassword
		options["no-password"] = noPassword

		// Handle --status flag
		if showStatus {
			// Status command doesn't show top banner

			// Create feature registry
			registry := features.NewRegistry()

			// Register features
			features.RegisterFeatures(registry, osInfo)

			// Create execution context
			ctx := &features.ExecutionContext{
				Options:     options,
				Logger:      log,
				DryRun:      true, // Don't make any changes
				Interactive: true, // Enable interactive mode for status display
				Verbose:     verbose,
			}

			// Get all features
			allFeatures := registry.GetFeatures()

			// Sort features by priority
			sortedFeatures := features.SortFeaturesByPriority(allFeatures)

			// Display current state for each feature in a simplified format

			// Display title with color - no top margin
			fmt.Println("\033[1;36mINIQ SYSTEM STATUS\033[0m")

			// Check current privileges
			isRoot := os.Geteuid() == 0
			hasPrivileges := isRoot || userInSudoGroup()

			// 1. User Account (most important - identity information)
			for _, feature := range sortedFeatures {
				if feature.Name() == "user" {
					// Detect current state
					state, err := feature.DetectCurrentState(ctx)
					if err != nil {
						fmt.Println("\033[1;36m● User Account\033[0m")
						fmt.Printf("  \033[1;31m✗ Error: %v\033[0m\n", err)
						continue
					}

					// Get sudoers file for header
					sudoersFile, _ := state["sudoers_file"].(string)
					if sudoersFile != "" {
						fmt.Printf("\033[1;36m● User Account\033[0m \033[90m(%s)\033[0m\n", sudoersFile)
					} else {
						fmt.Println("\033[1;36m● User Account\033[0m")
					}

					// Display simplified user status
					displaySimplifiedUserStatus(state)
				}
			}
			fmt.Println()

			// 2. SSH Security (second most important - security configuration)
			for _, feature := range sortedFeatures {
				if feature.Name() == "security" {
					// Detect current state
					state, err := feature.DetectCurrentState(ctx)
					if err != nil {
						fmt.Println("\033[1;36m● SSH Security\033[0m")
						fmt.Printf("  \033[1;31m✗ Error: %v\033[0m\n", err)
						continue
					}

					// Get SSH config file for header
					sshConfigFile, _ := state["ssh_config_file"].(string)
					if sshConfigFile != "" {
						fmt.Printf("\033[1;36m● SSH Security\033[0m \033[90m(%s)\033[0m\n", sshConfigFile)
					} else {
						fmt.Println("\033[1;36m● SSH Security\033[0m")
					}

					// Display simplified security status
					displaySimplifiedSecurityStatus(state, hasPrivileges)
				}
			}
			fmt.Println()

			// 3. SSH Keys (third most important - key management)
			fmt.Println("\033[1;36m● SSH Keys\033[0m")
			for _, feature := range sortedFeatures {
				if feature.Name() == "ssh" {
					// Detect current state
					state, err := feature.DetectCurrentState(ctx)
					if err != nil {
						log.Warning("Failed to detect SSH keys state: %v", err)
						continue
					}

					// Display simplified SSH keys status
					displaySimplifiedSSHKeysStatus(state)
				}
			}
			fmt.Println()

			// 4. System Privileges (fourth most important - privilege status)
			fmt.Println("\033[1;36m● System Privileges\033[0m")
			displaySystemPrivileges(isRoot, hasPrivileges)

			return
		}

		// Check if user has sudo privileges
		hasSudoPrivileges := isRoot || userInSudoGroup()

		// If not running as root and not skipping sudo operations, handle permissions
		if !hasSudoPrivileges && !skipSudo && !yes {
			log.Warning("INIQ requires sudo privileges for some operations")

			// Get current username
			currentUser, err := user.Current()
			if err != nil {
				log.Error("Failed to get current user: %v", err)
				os.Exit(1)
			}
			currentUsername := currentUser.Username

			// Provide simplified options to the user
			fmt.Println("\nYou have the following options:")
			fmt.Println("• Add current user to sudo group (recommended)")
			fmt.Println("• Skip operations requiring sudo")

			// Use utility function, default value is true (Y)
			addToSudoGroup := utils.PromptYesNo("\nAdd current user to sudo group?", true)

			if addToSudoGroup {
				// Add user to sudo group
				log.Info("Adding user to sudo group requires root privileges")
				err := addUserToSudoGroup(log, currentUsername)
				if err != nil {
					log.Error("Failed to add user to sudo group: %v", err)
					log.Info("Please run INIQ with sudo privileges")
					fmt.Println("Run: sudo iniq [options]")
					os.Exit(1)
				}

				log.Success("User added to sudo group successfully")

				// Ask if user wants to try activating sudo group immediately
				activateNow := utils.PromptYesNo("\nDo you want to try activating sudo privileges immediately?", true)

				if activateNow {
					log.Info("Attempting to activate sudo group membership without logout...")
					err := activateSudoGroup(log)
					if err != nil {
						log.Warning("Could not activate sudo group immediately: %v", err)
						log.Info("Please log out and log back in for changes to take effect")
						fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
						fmt.Println("After logging back in, run INIQ again.")
						os.Exit(0)
					}

					// Verify if sudo access is now working
					if userInSudoGroup() {
						log.Success("Sudo group membership activated successfully!")
						log.Info("Continuing with INIQ execution...")
						// Don't exit, continue with execution
					} else {
						log.Warning("Sudo group membership could not be verified")
						log.Info("Please log out and log back in for changes to take effect")
						fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
						fmt.Println("After logging back in, run INIQ again.")
						os.Exit(0)
					}
				} else {
					log.Info("Please log out and log back in for changes to take effect")
					fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
					fmt.Println("After logging back in, run INIQ again.")
					os.Exit(0)
				}
			} else {
				// Skip sudo operations
				log.Info("Skipping operations requiring sudo privileges")
				log.Warning("Some features will not be available")
				skipSudo = true
				options["skip-sudo"] = true
			}
		}

		// In verbose mode, just show a simple version info
		if verbose {
			log.Info("INIQ %s", version.String())
			// Add a blank line after version info
			if !quiet {
				fmt.Println()
			}
		} else {
			// In normal mode, don't show any startup message
			// This keeps the output clean and focused on the actual operations
		}

		// Create feature registry
		registry := features.NewRegistry()

		// Register features
		features.RegisterFeatures(registry, osInfo)

		// Get active features
		activeFeatures := registry.GetActiveFeatures(options)

		// If no features are active and not in non-interactive mode (no -y/--yes flag), enter interactive mode
		if len(activeFeatures) == 0 && !yes {
			// Set interactive flag
			options["interactive"] = true

			// Display welcome message
			if !quiet {
				fmt.Println("\nINIQ " + version.ShortString() + " - Interactive Mode")
				fmt.Println("────────────────────────────────────────")
			}

			// Create execution context for state detection
			stateCtx := &features.ExecutionContext{
				Options:     options,
				Logger:      log,
				DryRun:      true, // Don't make any changes
				Interactive: true,
				Verbose:     verbose,
			}

			// Get all features
			allFeatures := registry.GetFeatures()

			// Sort features by priority
			sortedFeatures := features.SortFeaturesByPriority(allFeatures)

			// Prompt user to enter username
			if username == "" {
				// Get current user as default value
				currentUser, err := user.Current()
				if err == nil {
					defaultUsername := currentUser.Username

					// Prompt user to enter username, default is current user
					fmt.Printf("\n\033[0;34m[1/4]\033[0m \033[1mUser Management\033[0m\n")
					fmt.Printf("────────────────────────────────────────\n")

					// Detect and display current user state
					for _, feature := range sortedFeatures {
						if feature.Name() == "user" {
							// Detect current state
							state, err := feature.DetectCurrentState(stateCtx)
							if err == nil {
								// Display current state
								feature.DisplayCurrentState(stateCtx, state)

								// Check if we should prompt the user
								if !feature.ShouldPromptUser(stateCtx, state) {
									// Use current user
									username = defaultUsername
									options["user"] = username
									fmt.Printf("\033[90mUsing current user: %s\033[0m\n", username)
									break
								}
							}
						}
					}

					// Only prompt if we still need to
					if username == "" {
						// Use utility function to get username
						var prompt string
						if os.Geteuid() == 0 {
							// Get real user (considering sudo environment)
							realUser, err := getRealUser()
							if err == nil {
								defaultUsername = realUser.Username
							}

							// Check if running with sudo
							sudoUser := os.Getenv("SUDO_USER")
							if sudoUser != "" {
								fmt.Printf("You are running INIQ with sudo from user '%s'.\n", sudoUser)
							} else {
								fmt.Println("You are running INIQ as root.")
							}

							prompt = fmt.Sprintf("Enter username to configure (or press Enter to use '%s')", defaultUsername)
						} else {
							prompt = "Enter username to configure (or press Enter to use yourself)"
						}

						input := utils.PromptWithDefault(prompt, defaultUsername)
						username = input
						options["user"] = username
					}
				}
			} else {
				// If username is already specified, show information
				fmt.Printf("\n\033[0;34m[1/4]\033[0m \033[1mUser Management\033[0m\n")
				fmt.Printf("────────────────────────────────────────\n")

				// Detect and display current user state
				for _, feature := range sortedFeatures {
					if feature.Name() == "user" {
						// Detect current state
						state, err := feature.DetectCurrentState(stateCtx)
						if err == nil {
							// Display current state
							feature.DisplayCurrentState(stateCtx, state)
						}
					}
				}

				fmt.Printf("\033[90mUsing specified username: %s\033[0m\n", username)
			}

			// Prompt user to enter SSH keys
			if len(keys) == 0 {
				fmt.Printf("\n\033[0;34m[2/4]\033[0m \033[1mSSH Key Management\033[0m\n")
				fmt.Printf("────────────────────────────────────────\n")

				// Detect and display current SSH key state
				for _, feature := range sortedFeatures {
					if feature.Name() == "ssh" {
						// Detect current state
						state, err := feature.DetectCurrentState(stateCtx)
						if err == nil {
							// Display current state
							feature.DisplayCurrentState(stateCtx, state)

							// Check if we should prompt the user
							if !feature.ShouldPromptUser(stateCtx, state) {
								fmt.Printf("\033[90mSkipping SSH key configuration\033[0m\n")
								break
							}
						}
					}
				}

				// Only prompt if we still need to
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
					// Parse input SSH keys
					keyList := strings.Split(input, ";")
					for _, key := range keyList {
						keys = append(keys, strings.TrimSpace(key))
					}
					options["keys"] = keys
					fmt.Printf("\033[90mSelected: %d SSH key(s) to import\033[0m\n", len(keys))
				} else {
					fmt.Printf("\033[90mSelected: Skip SSH key configuration\033[0m\n")
				}
			}

			// Only prompt for sudo configuration if username is provided
			if username != "" {
				fmt.Printf("\n\033[0;34m[3/4]\033[0m \033[1mSudo Configuration\033[0m\n")
				fmt.Printf("────────────────────────────────────────\n")

				// Detect and display current sudo state
				var sudoState map[string]any
				for _, feature := range sortedFeatures {
					if feature.Name() == "sudo" {
						// Detect current state
						state, err := feature.DetectCurrentState(stateCtx)
						if err == nil {
							// Display current state
							feature.DisplayCurrentState(stateCtx, state)
							sudoState = state
						}
					}
				}

				// Get sudo state
				var userHasSudo bool
				var hasPasswordlessSudo bool

				if sudoState != nil {
					userHasSudo, _ = sudoState["has_sudo"].(bool)
					hasPasswordlessSudo, _ = sudoState["has_passwordless_sudo"].(bool)
				} else {
					// Fallback to old method if state detection failed
					userHasSudo = os.Geteuid() == 0 || userInSudoGroup()
				}

				// Check if current user has sudo privileges
				if !userHasSudo {
					// Use utility function, default value is true (Y)
					addToSudoGroup := utils.PromptYesNo("Add user to sudo group?", true)

					// Default to adding user to sudo group
					if addToSudoGroup {
						// Add user to sudo group
						if os.Geteuid() == 0 {
							// Already root, add directly
							err := addUserToSudoGroup(log, username)
							if err != nil {
								log.Error("Failed to add user to sudo group: %v", err)
								skipSudo = true
								options["skip-sudo"] = true
								return
							}
							log.Success("User added to sudo group successfully")
						} else {
							// Need to elevate privileges
							fmt.Println("Adding user to sudo group requires root privileges.")

							// Use utility function, default value is true (Y)
							proceed := utils.PromptYesNo("Proceed? (You'll be prompted for password)", true)

							if proceed {
								err := addUserToSudoGroup(log, username)
								if err != nil {
									log.Error("Failed to add user to sudo group: %v", err)
									skipSudo = true
									options["skip-sudo"] = true
									return
								}
								log.Success("User added to sudo group successfully")

								// Ask if user wants to try activating sudo group immediately
								// Use utility function, default value is true (Y)
								activateNow := utils.PromptYesNo("\nDo you want to try activating sudo privileges immediately?", true)

								if activateNow {
									log.Info("Attempting to activate sudo group membership without logout...")
									err := activateSudoGroup(log)
									if err != nil {
										log.Warning("Could not activate sudo group immediately: %v", err)
										fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
										// Continue with configuration but note that sudo might not work yet
									} else {
										// Verify if sudo access is now working
										if userInSudoGroup() {
											log.Success("Sudo group membership activated successfully!")
										} else {
											log.Warning("Sudo group membership could not be verified")
											fmt.Println("\nIMPORTANT: You may need to log out and log back in for the sudo group changes to take full effect.")
										}
									}
								} else {
									fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
								}
							} else {
								skipSudo = true
								options["skip-sudo"] = true
								return
							}
						}

						// Only prompt for passwordless sudo if not already configured
						if !hasPasswordlessSudo {
							// Continue with sudo configuration
							// Use utility function, default value is true (Y)
							sudoNoPass = utils.PromptYesNo("Enable passwordless sudo?", true)
							options["sudo-nopasswd"] = sudoNoPass
						} else {
							fmt.Printf("\033[90mKeeping existing passwordless sudo configuration\033[0m\n")
							// Mark sudo as already configured, no need to execute again
							options["sudo-already-configured"] = true
						}
					} else {
						// Skip sudo configuration
						skipSudo = true
						options["skip-sudo"] = true
					}
				} else {
					// User already has sudo privileges
					// Only prompt for passwordless sudo if not already configured
					if !hasPasswordlessSudo {
						// Directly prompt for passwordless sudo without asking if they want to configure sudo
						sudoNoPass = utils.PromptYesNo("Enable passwordless sudo?", true)
						options["sudo-nopasswd"] = sudoNoPass
					} else {
						fmt.Printf("\033[90mKeeping existing passwordless sudo configuration\033[0m\n")
						// Mark sudo as already configured, no need to execute again
						options["sudo-already-configured"] = true
					}
				}
			}

			// Only prompt for SSH security settings if sudo is not skipped
			if !skipSudo {
				fmt.Printf("\n\033[0;34m[4/4]\033[0m \033[1mSSH Security Configuration\033[0m\n")
				fmt.Printf("────────────────────────────────────────\n")

				// Detect and display current SSH security state
				var securityState map[string]any
				for _, feature := range sortedFeatures {
					if feature.Name() == "security" {
						// Detect current state
						state, err := feature.DetectCurrentState(stateCtx)
						if err == nil {
							// Display current state
							feature.DisplayCurrentState(stateCtx, state)
							securityState = state
						}
						break
					}
				}

				// Get security state
				var rootLoginDisabled bool
				var passwordAuthDisabled bool

				if securityState != nil {
					rootLoginDisabled, _ = securityState["root_login_disabled"].(bool)
					passwordAuthDisabled, _ = securityState["password_auth_disabled"].(bool)
				}

				// Use the new state toggle function for root login
				rootLoginResult := utils.PromptStateToggle(utils.StateToggleConfig{
					FeatureName:  "SSH root login",
					CurrentState: !rootLoginDisabled, // Invert because we track "disabled" but function expects "enabled"
				})

				// Use the new state toggle function for password authentication
				passwordAuthResult := utils.PromptStateToggle(utils.StateToggleConfig{
					FeatureName:  "SSH password authentication",
					CurrentState: !passwordAuthDisabled, // Invert because we track "disabled" but function expects "enabled"
				})

				// Convert actions to the expected format and track changes
				switch rootLoginResult.Action {
				case utils.StateToggleEnable:
					sshNoRoot = false
					options["ssh-root-login"] = "enable"
				case utils.StateToggleDisable:
					sshNoRoot = true
					options["ssh-root-login"] = "disable"
				case utils.StateToggleKeep:
					sshNoRoot = rootLoginDisabled // Keep current state
				}

				switch passwordAuthResult.Action {
				case utils.StateToggleEnable:
					sshNoPass = false
					options["ssh-password-auth"] = "enable"
				case utils.StateToggleDisable:
					sshNoPass = true
					options["ssh-password-auth"] = "disable"
				case utils.StateToggleKeep:
					sshNoPass = passwordAuthDisabled // Keep current state
				}

				// Set the old-style options for backward compatibility
				// Only include in active features if there are actual changes
				hasSSHChanges := rootLoginResult.HasChange || passwordAuthResult.HasChange
				if hasSSHChanges {
					options["ssh-no-root"] = sshNoRoot
					options["ssh-no-password"] = sshNoPass
					// Mark that we have SSH security changes
					options["ssh-security-has-changes"] = true
				} else {
					// Mark that we don't have SSH security changes
					options["ssh-security-has-changes"] = false
				}
			} else {
				fmt.Printf("\n\033[0;34m[4/4]\033[0m \033[1mSSH Security Configuration\033[0m\n")
				fmt.Printf("────────────────────────────────────────\n")
				fmt.Printf("\n\033[90mSkipping SSH security configuration (requires sudo)\033[0m\n")
			}

			// Get active features again
			activeFeatures = registry.GetActiveFeatures(options)

			// If still no active features, show warning and exit
			if len(activeFeatures) == 0 {
				log.Warning("No operations specified. Running in interactive mode...")
				return
			}

			// Calculate total number of operations first
			operationCount := 0

			// Check if user already exists
			userExists := false
			if username != "" {
				_, err := user.Lookup(username)
				if err == nil {
					userExists = true
					// Mark user as already existing, no need to create again
					options["user-already-exists"] = true
				}
			}

			// Only count user creation operation if user doesn't exist
			if username != "" && !userExists {
				operationCount++
			}
			if len(keys) > 0 {
				operationCount++
			}
			sudoAlreadyConfigured, _ := options["sudo-already-configured"].(bool)
			if !skipSudo && !sudoAlreadyConfigured {
				operationCount++
			}
			// Only count SSH security if there are actual changes
			sshSecurityHasChanges, _ := options["ssh-security-has-changes"].(bool)
			if sshSecurityHasChanges {
				operationCount++
			}

			// Check if we have any operations to perform
			if operationCount == 0 {
				// No operations needed
				fmt.Printf("\n✓ No changes needed - all settings already match your preferences\n")
				return
			}

			// Display operation summary header only if we have operations
			fmt.Printf("\nSummary of Actions\n")
			fmt.Printf("-----------------\n")

			// Display operation summary
			operationIndex := 1

			// Get flag indicating if user already exists
			userExists = options["user-already-exists"].(bool)

			// Only show user creation operation if user doesn't exist
			if username != "" && !userExists {
				fmt.Printf("[%d/%d] User Management\n", operationIndex, operationCount)
				fmt.Printf("  - Create user '%s' with shell '/bin/bash'\n", username)
				operationIndex++
				fmt.Println()
			}

			if len(keys) > 0 {
				fmt.Printf("[%d/%d] SSH Key Management\n", operationIndex, operationCount)
				fmt.Printf("  - Import SSH keys from:\n")
				for _, key := range keys {
					source, value, _ := parseKeySource(key)
					switch source {
					case "github":
						fmt.Printf("    * GitHub user: %s\n", value)
					case "gitlab":
						fmt.Printf("    * GitLab user: %s\n", value)
					case "url":
						fmt.Printf("    * URL: %s\n", value)
					case "file":
						fmt.Printf("    * Local file: %s\n", value)
					default:
						fmt.Printf("    * %s\n", key)
					}
				}
				operationIndex++
				fmt.Println()
			}

			sudoAlreadyConfigured, _ = options["sudo-already-configured"].(bool)
			if !skipSudo && !sudoAlreadyConfigured {
				fmt.Printf("[%d/%d] Sudo Configuration\n", operationIndex, operationCount)
				if sudoNoPass {
					fmt.Printf("  - Configure sudo access with passwordless sudo\n")
				} else {
					fmt.Printf("  - Configure sudo access with password required\n")
				}
				operationIndex++
				fmt.Println()
			}

			// Only show SSH security if there are actual changes
			sshSecurityHasChanges, _ = options["ssh-security-has-changes"].(bool)
			if sshSecurityHasChanges {
				fmt.Printf("[%d/%d] SSH Security Configuration\n", operationIndex, operationCount)

				// Show only the changes that will be made
				rootLoginResult, hasRootResult := options["ssh-root-login"].(string)
				passwordAuthResult, hasPasswordResult := options["ssh-password-auth"].(string)

				if hasRootResult {
					if rootLoginResult == "enable" {
						fmt.Printf("  - Enable root login via SSH\n")
					} else if rootLoginResult == "disable" {
						fmt.Printf("  - Disable root login via SSH\n")
					}
				}

				if hasPasswordResult {
					if passwordAuthResult == "enable" {
						fmt.Printf("  - Enable password authentication via SSH\n")
					} else if passwordAuthResult == "disable" {
						fmt.Printf("  - Disable password authentication via SSH\n")
					}
				}

				_ = operationIndex // Increment not needed here
				fmt.Println()
			}

			// Ask for confirmation (we know operationCount > 0 at this point)
			proceed := utils.PromptYesNo("Do you want to proceed with these operations?", true)
			if !proceed {
				fmt.Println("Operation cancelled by user.")
				return
			}
		}

		// If still no active features, show warning and exit
		if len(activeFeatures) == 0 {
			log.Warning("No operations specified. Please use --help to see available options.")
			return
		}

		// Sort features by priority
		sortedFeatures := features.SortFeaturesByPriority(activeFeatures)

		// Create execution context
		ctx := &features.ExecutionContext{
			Options:     options,
			Logger:      log,
			DryRun:      dryRun,
			Interactive: !yes,
			Verbose:     verbose,
		}

		// Prepare operation list
		var operationTitles []string

		// Track operation execution results
		operationResults := make(map[string]bool)

		// Check which operations actually need to be performed
		for _, feature := range sortedFeatures {
			// Get the title based on feature type and check if operation is needed
			var title string
			var shouldAdd bool = true

			switch feature.Name() {
			case "user":
				// Check if user already exists
				_, err := user.Lookup(username)
				if err == nil {
					// User already exists, skip this operation
					shouldAdd = false
				} else {
					title = "Create user '" + username + "'"
				}
			case "ssh":
				// Check if SSH keys are specified
				keys, ok := options["keys"].([]string)
				if !ok || len(keys) == 0 {
					// No SSH keys specified, skip this operation
					shouldAdd = false
				} else {
					title = "Configure SSH keys"
				}
			case "sudo":
				// Check if sudo configuration is needed
				skipSudo, _ := options["skip-sudo"].(bool)
				sudoAlreadyConfigured, _ := options["sudo-already-configured"].(bool)
				if skipSudo || sudoAlreadyConfigured {
					// Sudo operations are skipped or already configured
					shouldAdd = false
				} else {
					title = "Set up sudo permissions"
				}
			case "security":
				// Check if SSH security has actual changes
				sshSecurityHasChanges, _ := options["ssh-security-has-changes"].(bool)
				if !sshSecurityHasChanges {
					// No SSH security changes needed
					shouldAdd = false
				} else {
					title = "Configure SSH security settings"
				}
			default:
				title = feature.Description()
			}

			// Add operation to list if needed
			if shouldAdd {
				operationTitles = append(operationTitles, title)
			}
		}

		// Print operation list only if there are operations to perform
		if len(operationTitles) > 0 {
			log.PrintOperationList(operationTitles)
		}

		// Execute features
		for _, feature := range sortedFeatures {
			// Check if operation needs to be performed
			var shouldExecute bool = true
			var title string

			switch feature.Name() {
			case "user":
				// Check if user already exists
				_, err := user.Lookup(username)
				if err == nil {
					// User already exists, skip this operation
					shouldExecute = false
				} else {
					title = "Creating user '" + username + "'"
				}
			case "ssh":
				// Check if SSH keys are specified
				keys, ok := options["keys"].([]string)
				if !ok || len(keys) == 0 {
					// No SSH keys specified, skip this operation
					shouldExecute = false
				} else {
					title = "Configuring SSH keys"
				}
			case "sudo":
				// Check if sudo configuration is needed
				skipSudo, _ := options["skip-sudo"].(bool)
				sudoAlreadyConfigured, _ := options["sudo-already-configured"].(bool)
				if skipSudo || sudoAlreadyConfigured {
					// Sudo operations are skipped or already configured
					shouldExecute = false
				} else {
					title = "Setting up sudo permissions"
				}
			case "security":
				// Check if SSH security has actual changes
				sshSecurityHasChanges, _ := options["ssh-security-has-changes"].(bool)
				if !sshSecurityHasChanges {
					// No SSH security changes needed
					shouldExecute = false
				} else {
					title = "Configuring SSH security settings"
				}
			default:
				title = feature.Description()
			}

			// Skip if operation is not needed
			if !shouldExecute {
				continue
			}

			// Start operation
			log.StartOperation(title)

			// Validate options
			if err := feature.ValidateOptions(options); err != nil {
				log.Error("Failed to validate options for %s: %v", title, err)
				operationResults[title] = false
				continue
			}

			// Execute feature with retry mechanism
			var err error
			maxRetries := 3
			operationSuccess := false

			// Determine if this is a critical operation that should cause immediate exit on failure
			isCriticalOperation := feature.Name() == "user"

			for attempt := 1; attempt <= maxRetries; attempt++ {
				err = feature.Execute(ctx)
				if err == nil {
					operationSuccess = true
					break // Success, exit retry loop
				}

				// Check if this is a sudo group activation issue
				if strings.Contains(err.Error(), "sudo group membership not yet active") {
					log.Warning("You need to log out and log back in for sudo group membership to take effect")
					fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
					fmt.Println("After logging back in, run INIQ again.")
					os.Exit(0)
				}

				// Check if this is an error that should not be retried
				if isNonRetryableError(err) {
					log.Error("Failed to execute %s: %v", title, err)
					operationSuccess = false
					break // Don't retry for configuration/parameter errors
				}

				// Check error type
				if os.IsPermission(err) {
					// Permission error
					log.Error("Permission error: %v", err)
					if ctx.Interactive && os.Geteuid() != 0 {
						// Provide simplified options
						fmt.Printf("\n✗ Failed to execute %s: Permission denied\n\n", title)
						fmt.Println("You need sudo privileges to perform this operation.")

						// Ask if user wants to add to sudo group
						currentUser, err := user.Current()
						if err != nil {
							log.Error("Failed to get current user: %v", err)
							continue
						}

						fmt.Print("Add current user to sudo group? [Y/n]: ")
						var input string
						_, _ = fmt.Scanln(&input) // Ignore error for user input

						// Default to adding user to sudo group
						if input == "" || strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
							log.Info("Adding user '%s' to sudo group...", currentUser.Username)
							err = addUserToSudoGroup(log, currentUser.Username)
							if err != nil {
								log.Error("Failed to add user to sudo group: %v", err)
								log.Info("Please run INIQ with sudo privileges")
								fmt.Println("Run: sudo iniq [options]")
								os.Exit(1)
							}

							log.Success("User added to sudo group successfully")

							// Ask if user wants to try activating sudo group immediately
							// Use utility function, default value is true (Y)
							activateNow := utils.PromptYesNo("\nDo you want to try activating sudo privileges immediately?", true)

							if activateNow {
								log.Info("Attempting to activate sudo group membership without logout...")
								err := activateSudoGroup(log)
								if err != nil {
									log.Warning("Could not activate sudo group immediately: %v", err)
									log.Info("Please log out and log back in for changes to take effect")
									fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
									fmt.Println("After logging back in, run INIQ again.")
									os.Exit(0)
								}

								// Verify if sudo access is now working
								if userInSudoGroup() {
									log.Success("Sudo group membership activated successfully!")
									log.Info("Continuing with INIQ execution...")
									// Retry the current operation
									continue
								} else {
									log.Warning("Sudo group membership could not be verified")
									log.Info("Please log out and log back in for changes to take effect")
									fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
									fmt.Println("After logging back in, run INIQ again.")
									os.Exit(0)
								}
							} else {
								log.Info("Please log out and log back in for changes to take effect")
								fmt.Println("\nIMPORTANT: You must log out and log back in for the sudo group changes to take effect.")
								fmt.Println("After logging back in, run INIQ again.")
								os.Exit(0)
							}
						} else {
							// Ask if user wants to skip this operation
							// Use utility function, default value is true (Y)
							skipOperation := utils.PromptYesNo("Skip this operation?", true)

							if skipOperation {
								// Skip this operation
								log.Warning("Skipping %s", title)
								operationSuccess = true // Mark as successful to continue
								break                   // Exit retry loop
							} else {
								// Exit
								log.Error("Operation aborted by user")
								os.Exit(1)
							}
						}
					}
				} else if strings.Contains(err.Error(), "network") ||
					strings.Contains(err.Error(), "connection") ||
					strings.Contains(err.Error(), "timeout") {
					// Network error
					if attempt < maxRetries {
						log.Warning("Network error: %v. Retrying (%d/%d)...", err, attempt, maxRetries)
						time.Sleep(2 * time.Second) // Wait before retry
						continue
					}
				}

				// For other errors or if we've exhausted retries
				if attempt < maxRetries {
					log.Warning("Error: %v. Retrying (%d/%d)...", err, attempt, maxRetries)
					time.Sleep(1 * time.Second) // Wait before retry
				} else {
					log.Error("Failed to execute %s after %d attempts: %v", title, maxRetries, err)
					operationSuccess = false
				}
			}

			// Record operation result
			operationResults[title] = operationSuccess

			// For critical operations, exit immediately on failure
			if isCriticalOperation && !operationSuccess {
				log.Error("Critical operation '%s' failed. Exiting...", title)
				os.Exit(1)
			}
		}

		// Print operation summary
		if len(sortedFeatures) > 0 {
			// Prepare summaries
			results := make(map[string]bool)

			// Use the tracked operation results
			for title, success := range operationResults {
				// Map operation titles to summary text
				var summaryText string

				// Check if this is a security operation
				if strings.Contains(title, "SSH security") {
					// For SSH security, use more specific text
					var securityMeasures []string
					if sshNoRoot, ok := options["ssh-no-root"].(bool); ok && sshNoRoot {
						securityMeasures = append(securityMeasures, "no root login")
					}
					if sshNoPass, ok := options["ssh-no-password"].(bool); ok && sshNoPass {
						securityMeasures = append(securityMeasures, "no password auth")
					}
					summaryText = fmt.Sprintf("SSH security: %s", strings.Join(securityMeasures, ", "))
				} else if strings.Contains(title, "sudo") {
					// For sudo operations
					username, _ := options["user"].(string)
					nopasswd, _ := options["sudo-nopasswd"].(bool)
					if nopasswd {
						summaryText = fmt.Sprintf("Passwordless sudo for '%s'", username)
					} else {
						summaryText = fmt.Sprintf("Sudo access for '%s'", username)
					}
				} else if strings.Contains(title, "SSH key") {
					// For SSH key operations
					keys, ok := options["keys"].([]string)
					if ok && len(keys) > 0 {
						summaryText = fmt.Sprintf("SSH keys added: %d", len(keys))
					} else {
						summaryText = "SSH keys configured"
					}
				} else if strings.Contains(title, "user") {
					// For user operations
					username, _ := options["user"].(string)
					summaryText = fmt.Sprintf("User account '%s' configured", username)
				} else {
					// For other operations, use the title directly
					summaryText = title
				}

				// Add to results with actual success/failure status
				results[summaryText] = success
			}

			// Print operation summary only if there are results
			if len(results) > 0 {
				log.PrintOperationSummary(results)
			}
		}

		// Print system configuration
		if len(sortedFeatures) > 0 {
			// Prepare system configuration
			configs := make(map[string]string)

			// Check if all operations were successful
			allSuccess := true
			for _, success := range operationResults {
				if !success {
					allSuccess = false
					break
				}
			}

			// Add user configuration only if user was created or modified
			_, err := user.Lookup(username)
			if username != "" && err != nil {
				configs["User"] = username
				if !skipSudo {
					if sudoNoPass {
						configs["User"] += " (with passwordless sudo)"
					} else {
						configs["User"] += " (with sudo access)"
					}
				}
			}

			// Add SSH security configuration only if it was successfully configured
			sshSecurityTitle := "Configuring SSH security settings"
			sshSecuritySuccess, sshSecurityConfigured := operationResults[sshSecurityTitle]

			if sshSecurityConfigured && sshSecuritySuccess {
				// Only add these settings if the operation was successful
				sshNoRoot, hasNoRoot := options["ssh-no-root"].(bool)
				if hasNoRoot && sshNoRoot {
					configs["SSH Root Login"] = "disabled"
				}

				sshNoPass, hasNoPass := options["ssh-no-password"].(bool)
				if hasNoPass && sshNoPass {
					configs["SSH Password Auth"] = "disabled"
				}
			}

			// Add SSH keys configuration only if keys were successfully added
			sshKeysTitle := "Configure SSH keys"
			sshKeysSuccess, sshKeysConfigured := operationResults[sshKeysTitle]

			if sshKeysConfigured && sshKeysSuccess {
				keys, ok := options["keys"].([]string)
				if ok && len(keys) > 0 {
					configs["SSH Authorized Keys"] = fmt.Sprintf("%d keys configured", len(keys))
				}
			}

			// Print system configuration only if there are configs
			if len(configs) > 0 {
				log.PrintSystemConfig(configs)
			}

			// Print success message based on overall success
			if allSuccess {
				log.Success("INIQ completed successfully")
			} else {
				log.Warning("INIQ completed with errors")
				os.Exit(1) // Exit with error code when operations failed
			}
		} else {
			// No operations performed
			log.Success("INIQ completed successfully")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Custom help function to display help with colored groups
func coloredHelpFunction(cmd *cobra.Command, args []string) {
	// Display header
	fmt.Printf("\033[1;36mINIQ - System Initialization Tool\033[0m\n")
	fmt.Printf("\033[90mVersion: %s\033[0m\n\n", cmd.Version)

	// Display usage
	fmt.Printf("\033[1mUsage:\033[0m\n")
	fmt.Printf("  %s\n", cmd.UseLine())
	if cmd.HasAvailableSubCommands() {
		fmt.Printf("  %s [command]\n", cmd.CommandPath())
	}

	// Display available commands if any
	if cmd.HasAvailableSubCommands() {
		fmt.Printf("\n\033[1mAvailable Commands:\033[0m\n")
		for _, subcmd := range cmd.Commands() {
			if subcmd.IsAvailableCommand() || subcmd.Name() == "help" {
				fmt.Printf("  \033[36m%-15s\033[0m %s\n", subcmd.Name(), subcmd.Short)
			}
		}
	}

	// Display flag groups
	fmt.Printf("\n\033[1;36mCore Feature Flags:\033[0m\n")
	fmt.Printf("  -u, --user string           Username to create or configure\n")
	fmt.Printf("  -k, --key strings           SSH key sources (github:user, gitlab:user, url:URL, file:path)\n")
	fmt.Printf("  -p, --password              Set password for the user (interactive prompt)\n")
	fmt.Printf("  --no-pass                   Create user without password (skip password setup)\n")
	fmt.Printf("  --ssh-root-login string     Configure SSH root login (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)\n")
	fmt.Printf("  --ssh-password-auth string  Configure SSH password authentication (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)\n")
	fmt.Printf("  --ssh-no-root               Disable SSH root login (deprecated, use --ssh-root-login=disable)\n")
	fmt.Printf("  --ssh-no-password           Disable SSH password authentication (deprecated, use --ssh-password-auth=disable)\n")
	fmt.Printf("  --sudo-nopasswd             Configure sudo without password (default true)\n")

	fmt.Printf("\n\033[1;36mSecurity Enhancement Flags:\033[0m\n")
	fmt.Printf("  -a, --all                   Apply all security hardening options\n")

	fmt.Printf("\n\033[1;36mOperation Helper Flags:\033[0m\n")
	fmt.Printf("  -b, --backup                Backup original configuration files\n")
	fmt.Printf("  -y, --yes                   Answer yes to all prompts\n")
	fmt.Printf("  --dry-run                   Show what would be done without making changes\n")
	fmt.Printf("  -S, --skip-sudo             Skip operations requiring sudo\n")

	fmt.Printf("\n\033[1;36mOutput Control Flags:\033[0m\n")
	fmt.Printf("  -v, --verbose               Enable verbose output\n")
	fmt.Printf("  -q, --quiet                 Suppress all output except errors\n")
	fmt.Printf("  -s, --status                Show current system status and exit\n")

	fmt.Printf("\n\033[1;36mConfiguration Flags:\033[0m\n")
	fmt.Printf("  --config string             Config file (default is $HOME/.iniq.yaml)\n")

	// Display footer
	if cmd.HasAvailableSubCommands() {
		fmt.Printf("\nUse \"%s \033[36m[command]\033[0m --help\" for more information about a command.\n", cmd.CommandPath())
	}
}

// Custom version template with colors
const coloredVersionTemplate = `{{"\033[1;36m"}}INIQ Version Information{{"\033[0m"}}
{{"\033[90m"}}------------------------{{"\033[0m"}}
{{"\033[1m"}}Version:    {{"\033[0m"}}{{"\033[32m"}}{{.Version}}{{"\033[0m"}}
{{"\033[1m"}}Build Date: {{"\033[0m"}}{{.BuildDate}}
{{"\033[1m"}}Commit:     {{"\033[0m"}}{{.Commit}}
{{"\033[1m"}}Go Version: {{"\033[0m"}}{{.GoVersion}}
{{"\033[1m"}}Platform:   {{"\033[0m"}}{{.Platform}}
`

func init() {
	cobra.OnInitialize(initConfig)

	// Add version command
	rootCmd.AddCommand(versionCmd)

	// Set custom help function
	rootCmd.SetHelpFunc(coloredHelpFunction)

	// Set custom version template with more information
	info := version.Get()
	rootCmd.SetVersionTemplate(strings.Replace(
		coloredVersionTemplate,
		"{{.BuildDate}}",
		info.BuildDate,
		-1,
	))
	rootCmd.SetVersionTemplate(strings.Replace(
		rootCmd.VersionTemplate(),
		"{{.Commit}}",
		info.Commit,
		-1,
	))
	rootCmd.SetVersionTemplate(strings.Replace(
		rootCmd.VersionTemplate(),
		"{{.GoVersion}}",
		info.GoVersion,
		-1,
	))
	rootCmd.SetVersionTemplate(strings.Replace(
		rootCmd.VersionTemplate(),
		"{{.Platform}}",
		info.Platform,
		-1,
	))

	// Core Feature Flags - directly modify system functionality
	rootCmd.Flags().StringVarP(&username, "user", "u", "", "username to create or configure")
	rootCmd.Flags().StringSliceVarP(&keys, "key", "k", []string{}, "SSH key sources (github:user, gitlab:user, url:URL, file:path)")
	rootCmd.Flags().StringVar(&sshRootLogin, "ssh-root-login", "", "configure SSH root login (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)")
	rootCmd.Flags().StringVar(&sshPasswordAuth, "ssh-password-auth", "", "configure SSH password authentication (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)")
	rootCmd.Flags().BoolVar(&sshNoRoot, "ssh-no-root", false, "disable SSH root login (deprecated, use --ssh-root-login=disable)")
	rootCmd.Flags().BoolVar(&sshNoPass, "ssh-no-password", false, "disable SSH password authentication (deprecated, use --ssh-password-auth=disable)")
	rootCmd.Flags().BoolVar(&sudoNoPass, "sudo-nopasswd", true, "configure sudo without password")
	rootCmd.Flags().BoolVarP(&setPassword, "password", "p", false, "set password for the user (interactive prompt)")
	rootCmd.Flags().BoolVar(&noPassword, "no-pass", false, "create user without password (skip password setup)")

	// Security Enhancement Flags - security related combination options
	rootCmd.Flags().BoolVarP(&allSecurity, "all", "a", false, "apply all security hardening options")

	// Operation Helper Flags - affect operation behavior but don't directly modify system
	rootCmd.Flags().BoolVarP(&backupFiles, "backup", "b", false, "backup original configuration files")
	rootCmd.Flags().BoolVarP(&yes, "yes", "y", false, "answer yes to all prompts")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.Flags().BoolVarP(&skipSudo, "skip-sudo", "S", false, "skip operations requiring sudo")

	// Output Control Flags - control command output verbosity
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except errors")
	rootCmd.Flags().BoolVarP(&showStatus, "status", "s", false, "show current system status and exit")

	// Configuration Flags - related to configuration files
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iniq.yaml)")

	// Bind flags to viper
	_ = viper.BindPFlag("user", rootCmd.Flags().Lookup("user"))
	_ = viper.BindPFlag("keys", rootCmd.Flags().Lookup("key"))
	_ = viper.BindPFlag("ssh-root-login", rootCmd.Flags().Lookup("ssh-root-login"))
	_ = viper.BindPFlag("ssh-password-auth", rootCmd.Flags().Lookup("ssh-password-auth"))
	_ = viper.BindPFlag("ssh-no-root", rootCmd.Flags().Lookup("ssh-no-root"))
	_ = viper.BindPFlag("ssh-no-password", rootCmd.Flags().Lookup("ssh-no-password"))
	_ = viper.BindPFlag("sudo-nopasswd", rootCmd.Flags().Lookup("sudo-nopasswd"))
	_ = viper.BindPFlag("backup", rootCmd.Flags().Lookup("backup"))
	_ = viper.BindPFlag("all", rootCmd.Flags().Lookup("all"))
	_ = viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	_ = viper.BindPFlag("no-password", rootCmd.Flags().Lookup("no-pass"))
	_ = viper.BindPFlag("skip-sudo", rootCmd.PersistentFlags().Lookup("skip-sudo"))
	_ = viper.BindPFlag("yes", rootCmd.PersistentFlags().Lookup("yes"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("status", rootCmd.PersistentFlags().Lookup("status"))
}

// isNonRetryableError checks if an error should not be retried
// These are typically configuration errors, parameter errors, or user input errors
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()

	// Configuration and parameter errors that should not be retried
	nonRetryablePatterns := []string{
		"cannot specify both --password and --no-pass options",
		"requires password input",
		"conflicting password options",
		"invalid option",
		"invalid argument",
		"invalid flag",
		"unknown flag",
		"required flag",
		"unsupported OS",
		"user already exists",
		"invalid username",
		"invalid shell",
		"invalid home directory",
		"validation failed",
		"invalid configuration",
		"missing required",
		"duplicate option",
		"invalid format",
		"parse error",
		"syntax error",
		"Stdin already set",
		"Stdout already set",
		"Stderr already set",
		"process already started",
		"exec: already started",
		"exec: not started",
		"command not found",
		"no such file or directory",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// userInSudoGroup checks if the current user is in the sudo/wheel group
// and verifies if sudo privileges are actually working
func userInSudoGroup() bool {
	// Get current user, considering SUDO_USER environment variable
	currentUser, err := getRealUser()
	if err != nil {
		return false
	}

	// Get user's groups
	groupIds, err := currentUser.GroupIds()
	if err != nil {
		return false
	}

	// Check each group
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

	// If not in sudo group, return false immediately
	if !inSudoGroup {
		return false
	}

	// If we're already root, we don't need to check sudo
	if os.Geteuid() == 0 {
		return true
	}

	// Try to run a simple sudo command to verify if sudo works
	// Use -n flag to prevent sudo from asking for a password
	cmd := exec.Command("sudo", "-n", "true")
	err = cmd.Run()

	// If the command succeeds, sudo is working
	if err == nil {
		return true
	}

	// If the command fails with "sudo: no tty present and no askpass program specified",
	// it means the user is in sudo group but needs to enter a password (which is normal)
	if strings.Contains(err.Error(), "no tty present") || strings.Contains(err.Error(), "askpass") {
		return true
	}

	// If the command fails with "user is not in the sudoers file", it means the user is in sudo group
	// but the membership is not yet active (needs to log out and log back in)
	if strings.Contains(err.Error(), "not in the sudoers file") {
		return false
	}

	// For other errors, assume sudo is not working properly
	return false
}

// addUserToSudoGroup adds the specified user to the sudo group
func addUserToSudoGroup(log *logger.Logger, username string) error {
	// Check if we're running as root
	if os.Geteuid() == 0 {
		// We're root, so we can directly add the user to sudo group
		log.Info("Adding user '%s' to sudo group...", username)

		// Detect OS
		osInfo, err := osdetect.Detect()
		if err != nil {
			return fmt.Errorf("failed to detect operating system: %w", err)
		}

		var cmd *exec.Cmd
		switch osInfo.Type {
		case osdetect.Linux:
			// On Linux, use usermod to add user to sudo group
			cmd = exec.Command("usermod", "-aG", "sudo", username)
		case osdetect.Darwin:
			// On macOS, use dseditgroup to add user to admin group
			cmd = exec.Command("dseditgroup", "-o", "edit", "-a", username, "-t", "user", "admin")
		default:
			return fmt.Errorf("unsupported OS: %s", osInfo.Type)
		}

		// Execute the command
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add user to sudo group: %w", err)
		}

		// Verify the user was added to the sudo group
		if !verifyUserInSudoGroup(username) {
			return fmt.Errorf("failed to verify user was added to sudo group")
		}

		// Add a note about sudo group membership activation
		log.Info("Note: Sudo group membership requires logging out and back in to take full effect")

		return nil
	} else {
		// We're not root, so we need to use su or sudo to add the user
		log.Info("You need root privileges to add a user to the sudo group")
		log.Info("Please enter the root password when prompted")

		// Detect OS
		osInfo, err := osdetect.Detect()
		if err != nil {
			return fmt.Errorf("failed to detect operating system: %w", err)
		}

		var cmd *exec.Cmd
		switch osInfo.Type {
		case osdetect.Linux:
			// On Linux, always use su with full path to usermod
			// We can't use sudo here because the user doesn't have sudo privileges yet
			log.Info("Running: su -c \"/usr/sbin/usermod -aG sudo %s\"", username)
			cmd = exec.Command("su", "-c", fmt.Sprintf("/usr/sbin/usermod -aG sudo %s", username))
		case osdetect.Darwin:
			// On macOS, use sudo to run dseditgroup
			log.Info("Running: sudo dseditgroup -o edit -a %s -t user admin", username)
			cmd = exec.Command("sudo", "dseditgroup", "-o", "edit", "-a", username, "-t", "user", "admin")
		default:
			return fmt.Errorf("unsupported OS: %s", osInfo.Type)
		}

		// Connect command to terminal for password input
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Execute the command
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add user to sudo group: %w", err)
		}

		// Verify the user was added to the sudo group
		if !verifyUserInSudoGroup(username) {
			return fmt.Errorf("failed to verify user was added to sudo group")
		}

		// Add a note about sudo group membership activation
		log.Info("Note: Sudo group membership requires logging out and back in to take full effect")

		return nil
	}
}

// verifyUserInSudoGroup checks if the specified user is in the sudo/wheel/admin group
func verifyUserInSudoGroup(username string) bool {
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

	// Check each group
	for _, gid := range groupIds {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			continue
		}

		// Check if group name is sudo, wheel, or admin (common sudo group names)
		groupName := strings.ToLower(group.Name)
		if groupName == "sudo" || groupName == "wheel" || groupName == "admin" {
			return true
		}
	}

	return false
}

// activateSudoGroup attempts to activate sudo group membership without requiring logout
func activateSudoGroup(log *logger.Logger) error {
	// Detect OS
	osInfo, err := osdetect.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	// Different approaches based on OS
	switch osInfo.Type {
	case osdetect.Linux:
		log.Info("Attempting to activate sudo group membership...")

		// Try multiple methods to activate sudo group

		// Method 1: Use sg command to run a command as sudo group
		log.Info("Trying sg method...")
		sgCmd := exec.Command("sg", "sudo", "-c", "id")
		sgOutput, err := sgCmd.CombinedOutput()
		if err == nil {
			log.Info("sg command successful: %s", string(sgOutput))
		} else {
			log.Warning("sg command failed: %v", err)
		}

		// Method 2: Try to restart sudo service if available
		log.Info("Trying to restart sudo service...")
		sudoRestartCmd := exec.Command("sudo", "service", "sudo", "restart")
		if err := sudoRestartCmd.Run(); err != nil {
			log.Warning("Failed to restart sudo service: %v", err)
		}

		// Method 3: Try to use sudo once to "activate" it
		log.Info("Trying to use sudo command...")
		sudoTestCmd := exec.Command("sudo", "-v")
		sudoTestCmd.Stdin = os.Stdin
		sudoTestCmd.Stdout = os.Stdout
		sudoTestCmd.Stderr = os.Stderr
		if err := sudoTestCmd.Run(); err != nil {
			log.Warning("Failed to run sudo test: %v", err)
		}

		// We've tried our best, but warn that a logout might still be needed
		log.Warning("Attempted to activate sudo group membership using multiple methods")
		log.Info("If sudo commands still fail, you will need to log out and log back in")

		// Let the caller check if it worked
		return nil

	case osdetect.Darwin:
		// On macOS, group membership is usually applied immediately
		log.Info("Refreshing group membership on macOS...")

		// Try to use sudo once to "activate" it
		sudoTestCmd := exec.Command("sudo", "-v")
		sudoTestCmd.Stdin = os.Stdin
		sudoTestCmd.Stdout = os.Stdout
		sudoTestCmd.Stderr = os.Stderr
		if err := sudoTestCmd.Run(); err != nil {
			log.Warning("Failed to run sudo test: %v", err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported OS: %s", osInfo.Type)
	}
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

// checkMacOSSupport checks if running on macOS and shows appropriate warnings
func checkMacOSSupport(osInfo *osdetect.Info, log *logger.Logger, yes, quiet bool) error {
	if osInfo.Type != osdetect.Darwin {
		return nil
	}

	log.Warning("⚠️  macOS Support Notice:")
	log.Warning("   • INIQ is primarily designed for Linux servers")
	log.Warning("   • macOS support is experimental and for development/testing")
	log.Warning("   • Some features may not work as expected")
	log.Warning("   • Use with caution in production environments")

	if !yes && !quiet {
		fmt.Println()
		proceed := utils.PromptYesNo("Do you want to continue?", false)
		if !proceed {
			return fmt.Errorf("operation cancelled by user")
		}
		fmt.Println()
	}

	return nil
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

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Use the config package to initialize configuration
	if err := config.InitConfig(cfgFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// Set up logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Execute the root command
	Execute()
}

// displaySystemPrivileges shows current system privilege status
func displaySystemPrivileges(isRoot bool, hasPrivileges bool) {
	// Current privilege level
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Current Level")
	if isRoot {
		fmt.Printf("\033[1;32m✓ Root\033[0m")
		sudoUser := os.Getenv("SUDO_USER")
		if sudoUser != "" {
			fmt.Printf(" \033[90m(via sudo from %s)\033[0m", sudoUser)
		}
		fmt.Println()
	} else if hasPrivileges {
		fmt.Printf("\033[1;32m✓ Sudo Access\033[0m\n")
	} else {
		fmt.Printf("\033[1;33m⚠ Limited\033[0m\n")
		fmt.Printf("    \033[90mSome operations require elevated privileges\033[0m\n")
	}

	// Capability summary
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Capabilities")
	if hasPrivileges {
		fmt.Printf("\033[1;32m✓ Full system configuration\033[0m\n")
	} else {
		fmt.Printf("\033[1;33m⚠ User-level configuration only\033[0m\n")
	}
}

// displaySimplifiedSecurityStatus shows simplified SSH security status
func displaySimplifiedSecurityStatus(state map[string]any, _ bool) {
	sshConfigExists, _ := state["ssh_config_exists"].(bool)

	if !sshConfigExists {
		fmt.Printf("  \033[1;33m⚠ SSH not configured\033[0m\n")
		fmt.Printf("    \033[90mSSH daemon configuration not found\033[0m\n")
		return
	}

	rootLoginDisabled, _ := state["root_login_disabled"].(bool)
	passwordAuthDisabled, _ := state["password_auth_disabled"].(bool)
	permitRootLoginValue, _ := state["permit_root_login_value"].(string)
	passwordAuthValue, _ := state["password_auth_value"].(string)
	permitRootLoginExplicit, _ := state["permit_root_login_explicit"].(bool)
	passwordAuthExplicit, _ := state["password_auth_explicit"].(bool)

	// Root Login status with proper alignment
	fmt.Printf("  %-15s: ", "Root Login")
	if rootLoginDisabled {
		fmt.Printf("\033[1;32m✓ Disabled\033[0m")
		if permitRootLoginValue == "prohibit-password" {
			fmt.Printf(" \033[90m(key-only)\033[0m")
		}
	} else {
		fmt.Printf("\033[1;31m✗ Enabled\033[0m")
	}

	// Show if using defaults
	if permitRootLoginValue != "" && permitRootLoginValue != "unknown" && !permitRootLoginExplicit {
		fmt.Printf(" \033[90m(default)\033[0m")
	}
	fmt.Println()

	// Password Authentication status with proper alignment
	fmt.Printf("  %-15s: ", "Password Auth")
	if passwordAuthDisabled {
		fmt.Printf("\033[1;32m✓ Disabled\033[0m")
	} else {
		fmt.Printf("\033[1;31m✗ Enabled\033[0m")
	}

	// Show if using defaults
	if passwordAuthValue != "" && passwordAuthValue != "unknown" && !passwordAuthExplicit {
		fmt.Printf(" \033[90m(default)\033[0m")
	}
	fmt.Println()
}

// displaySimplifiedUserStatus shows simplified user account status
func displaySimplifiedUserStatus(state map[string]any) {
	username := state["username"].(string)
	userExists, _ := state["user_exists"].(bool)
	isCurrentUser, _ := state["is_current_user"].(bool)

	// User identity with proper alignment
	fmt.Printf("  %-15s: \033[0;37m%s\033[0m", "Username", username)
	if isCurrentUser {
		fmt.Printf(" \033[90m(current)\033[0m")
	}
	fmt.Println()

	// User status with proper alignment
	fmt.Printf("  %-15s: ", "Account Status")
	if userExists {
		fmt.Printf("\033[1;32m✓ Active\033[0m\n")

		// Sudo access with proper alignment (no file path since it's in header)
		hasSudo, _ := state["has_sudo"].(bool)
		hasPasswordlessSudo, _ := state["has_passwordless_sudo"].(bool)

		fmt.Printf("  %-15s: ", "Sudo Access")
		if hasSudo {
			if hasPasswordlessSudo {
				fmt.Printf("\033[1;32m✓ Passwordless\033[0m")
			} else {
				fmt.Printf("\033[1;32m✓ With Password\033[0m")
			}
			fmt.Println()
		} else {
			fmt.Printf("\033[1;31m✗ None\033[0m\n")
		}
	} else {
		fmt.Printf("\033[1;31m✗ Not Found\033[0m\n")
	}
}

// displaySimplifiedSSHKeysStatus shows simplified SSH keys status
func displaySimplifiedSSHKeysStatus(state map[string]any) {
	username := state["username"].(string)
	sshDirExists, _ := state["ssh_dir_exists"].(bool)
	authKeysExists, _ := state["auth_keys_exists"].(bool)
	existingKeyCount, _ := state["existing_key_count"].(int)

	// SSH setup status with proper alignment
	fmt.Printf("  %-15s: \033[0;37m%s\033[0m\n", "User", username)

	// Determine SSH status and show it directly
	fmt.Printf("  %-15s: ", "Authorized Keys")
	if !sshDirExists {
		fmt.Printf("\033[1;33m⚠ Not Configured\033[0m\n")
		fmt.Printf("    \033[90mSSH directory not found\033[0m\n")
		return
	}

	if !authKeysExists {
		fmt.Printf("\033[1;33m⚠ No Keys\033[0m\n")
		fmt.Printf("    \033[90mSSH directory exists but no authorized keys\033[0m\n")
		return
	}

	// Key count
	if existingKeyCount > 0 {
		fmt.Printf("\033[1;32m✓ %d key(s)\033[0m\n", existingKeyCount)
	} else {
		fmt.Printf("\033[1;31m✗ None\033[0m\n")
	}
}
