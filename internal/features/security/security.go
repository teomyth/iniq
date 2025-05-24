// Package security implements the SSH security configuration feature
package security

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/internal/utils"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// Feature implements the SSH security configuration feature
type Feature struct {
	osInfo *osdetect.Info
}

// New creates a new SSH security configuration feature
func New(osInfo *osdetect.Info) *Feature {
	return &Feature{
		osInfo: osInfo,
	}
}

// Name returns the feature name
func (f *Feature) Name() string {
	return "security"
}

// Description returns the feature description
func (f *Feature) Description() string {
	return "Configure SSH security settings"
}

// Flags returns the command-line flags for the feature
func (f *Feature) Flags() []features.Flag {
	return []features.Flag{
		{
			Name:      "ssh-root-login",
			Shorthand: "",
			Usage:     "configure SSH root login (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)",
			Default:   "",
			Required:  false,
		},
		{
			Name:      "ssh-password-auth",
			Shorthand: "",
			Usage:     "configure SSH password authentication (yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)",
			Default:   "",
			Required:  false,
		},
		{
			Name:      "ssh-no-root",
			Shorthand: "",
			Usage:     "disable SSH root login (deprecated, use --ssh-root-login=disable)",
			Default:   false,
			Required:  false,
		},
		{
			Name:      "ssh-no-password",
			Shorthand: "",
			Usage:     "disable SSH password authentication (deprecated, use --ssh-password-auth=disable)",
			Default:   false,
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
	// In interactive mode, always activate security function
	interactive, hasInteractive := options["interactive"].(bool)
	skipSudo, hasSkipSudo := options["skip-sudo"].(bool)

	if hasInteractive && interactive {
		return !hasSkipSudo || !skipSudo
	}

	// Check if the --all option is provided
	allSecurity, hasAllSecurity := options["all"].(bool)
	if hasAllSecurity && allSecurity && (!hasSkipSudo || !skipSudo) {
		// If --all is specified, enable all security options
		options["ssh-no-root"] = true
		options["ssh-no-password"] = true
		return true
	}

	// Otherwise, it will only be activated if the SSH security option is provided and sudo is not skipped
	sshNoRoot, hasNoRoot := options["ssh-no-root"].(bool)
	sshNoPass, hasNoPass := options["ssh-no-password"].(bool)

	// Check new parameters
	sshRootLogin, hasRootLogin := options["ssh-root-login"].(string)
	sshPasswordAuth, hasPasswordAuth := options["ssh-password-auth"].(string)

	return (((hasNoRoot && sshNoRoot) || (hasNoPass && sshNoPass)) ||
		((hasRootLogin && sshRootLogin != "") || (hasPasswordAuth && sshPasswordAuth != ""))) &&
		(!hasSkipSudo || !skipSudo)
}

// ValidateOptions validates the feature options
func (f *Feature) ValidateOptions(options map[string]any) error {
	// Validate ssh-root-login parameter
	if rootLogin, ok := options["ssh-root-login"].(string); ok && rootLogin != "" {
		if _, err := utils.ParseBoolValue(rootLogin); err != nil {
			return fmt.Errorf("invalid value for --ssh-root-login: %w", err)
		}
	}

	// Validate ssh-password-auth parameter
	if passwordAuth, ok := options["ssh-password-auth"].(string); ok && passwordAuth != "" {
		if _, err := utils.ParseBoolValue(passwordAuth); err != nil {
			return fmt.Errorf("invalid value for --ssh-password-auth: %w", err)
		}
	}

	return nil
}

// SSHSecurityOptions represents the parsed SSH security configuration
type SSHSecurityOptions struct {
	RootLoginAction    string // "enable", "disable", or "keep"
	PasswordAuthAction string // "enable", "disable", or "keep"
}

// parseSSHSecurityOptions parses and resolves SSH security options from various sources
func parseSSHSecurityOptions(options map[string]any, _ map[string]any) (SSHSecurityOptions, error) {
	opts := SSHSecurityOptions{
		RootLoginAction:    "keep", // Default: keep current state
		PasswordAuthAction: "keep",
	}

	// Parse new parameters first (they take precedence)
	if rootLogin, ok := options["ssh-root-login"].(string); ok && rootLogin != "" {
		enable, err := utils.ParseBoolValue(rootLogin)
		if err != nil {
			return opts, err
		}
		if enable {
			opts.RootLoginAction = "enable"
		} else {
			opts.RootLoginAction = "disable"
		}
	}

	if passwordAuth, ok := options["ssh-password-auth"].(string); ok && passwordAuth != "" {
		enable, err := utils.ParseBoolValue(passwordAuth)
		if err != nil {
			return opts, err
		}
		if enable {
			opts.PasswordAuthAction = "enable"
		} else {
			opts.PasswordAuthAction = "disable"
		}
	}

	// Handle backward compatibility with old parameters
	if opts.RootLoginAction == "keep" {
		if noRoot, ok := options["ssh-no-root"].(bool); ok && noRoot {
			opts.RootLoginAction = "disable"
		}
	}

	if opts.PasswordAuthAction == "keep" {
		if noPassword, ok := options["ssh-no-password"].(bool); ok && noPassword {
			opts.PasswordAuthAction = "disable"
		}
	}

	return opts, nil
}

// Execute executes the feature functionality
func (f *Feature) Execute(ctx *features.ExecutionContext) error {
	// Get current state for intelligent prompting
	currentState, err := f.DetectCurrentState(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect current SSH security state: %w", err)
	}

	// Parse SSH security options
	opts, err := parseSSHSecurityOptions(ctx.Options, currentState)
	if err != nil {
		return err
	}

	// If in interactive mode and no security options are specified, prompt the user
	if ctx.Interactive && opts.RootLoginAction == "keep" && opts.PasswordAuthAction == "keep" {
		ctx.Logger.Info("SSH Security Configuration")

		// Get current state values
		rootLoginDisabled, _ := currentState["root_login_disabled"].(bool)
		passwordAuthDisabled, _ := currentState["password_auth_disabled"].(bool)

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

		// Convert actions to the expected format
		switch rootLoginResult.Action {
		case utils.StateToggleEnable:
			opts.RootLoginAction = "enable"
		case utils.StateToggleDisable:
			opts.RootLoginAction = "disable"
		case utils.StateToggleKeep:
			opts.RootLoginAction = "keep"
		}

		switch passwordAuthResult.Action {
		case utils.StateToggleEnable:
			opts.PasswordAuthAction = "enable"
		case utils.StateToggleDisable:
			opts.PasswordAuthAction = "disable"
		case utils.StateToggleKeep:
			opts.PasswordAuthAction = "keep"
		}
	}

	// Check if any changes are needed and provide appropriate feedback
	hasChanges := opts.RootLoginAction != "keep" || opts.PasswordAuthAction != "keep"

	if !hasChanges {
		// No changes needed - show confirmation message
		ctx.Logger.Info("")
		ctx.Logger.Info("SSH Security Configuration")
		ctx.Logger.Info("--------------------------")
		ctx.Logger.Success("✓ No changes needed - all settings already match your preferences")

		// Get current state values for display
		rootLoginDisabled, _ := currentState["root_login_disabled"].(bool)
		passwordAuthDisabled, _ := currentState["password_auth_disabled"].(bool)

		// Show current settings
		rootStatus := "enabled"
		if rootLoginDisabled {
			rootStatus = "disabled"
		}
		passwordStatus := "enabled"
		if passwordAuthDisabled {
			passwordStatus = "disabled"
		}

		ctx.Logger.Info("  - Root login: %s (unchanged)", rootStatus)
		ctx.Logger.Info("  - Password authentication: %s (unchanged)", passwordStatus)

		return nil
	}

	// Skip if dry run
	if ctx.DryRun {
		if opts.RootLoginAction == "enable" {
			ctx.Logger.Info("Would enable SSH root login")
		} else if opts.RootLoginAction == "disable" {
			ctx.Logger.Info("Would disable SSH root login")
		}
		if opts.PasswordAuthAction == "enable" {
			ctx.Logger.Info("Would enable SSH password authentication")
		} else if opts.PasswordAuthAction == "disable" {
			ctx.Logger.Info("Would disable SSH password authentication")
		}
		return nil
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("configuring SSH security requires root privileges")
	}

	// Get SSH config file path
	sshConfigFile := osdetect.GetSSHConfigPath(f.osInfo)

	// Read current config
	configContent, err := os.ReadFile(sshConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read SSH config file: %w", err)
	}

	// Check if backup option is enabled
	backupEnabled, hasBackup := ctx.Options["backup"].(bool)

	// Make a backup of the original file
	backupPath, err := utils.BackupFile(sshConfigFile, hasBackup && backupEnabled)
	if err != nil {
		return fmt.Errorf("failed to create backup of SSH config file: %w", err)
	}
	if backupPath != "" {
		ctx.Logger.Info("Created backup of SSH config file: %s", backupPath)
	}

	// Prepare configuration changes
	var configChanges []string
	newContent := string(configContent)

	// Apply root login configuration
	if opts.RootLoginAction == "enable" {
		ctx.Logger.Step("Enabling SSH root login...")
		newContent = f.enableRootLogin(newContent)
		configChanges = append(configChanges, "PermitRootLogin yes")
	} else if opts.RootLoginAction == "disable" {
		ctx.Logger.Step("Disabling SSH root login...")
		newContent = f.disableRootLogin(newContent)
		configChanges = append(configChanges, "PermitRootLogin no")
	}

	// Apply password authentication configuration
	if opts.PasswordAuthAction == "enable" {
		ctx.Logger.Step("Enabling SSH password authentication...")
		newContent = f.enablePasswordAuth(newContent)
		configChanges = append(configChanges, "PasswordAuthentication yes")
	} else if opts.PasswordAuthAction == "disable" {
		ctx.Logger.Step("Disabling SSH password authentication...")
		newContent = f.disablePasswordAuth(newContent)
		configChanges = append(configChanges, "PasswordAuthentication no")
	}

	// Show configuration changes
	ctx.Logger.MultiLine("info", "Applying the following SSH configuration changes:", configChanges)

	// Write updated config
	ctx.Logger.Step("Modifying SSH configuration file...")
	if err := os.WriteFile(sshConfigFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write SSH config file: %w", err)
	}

	// Restart SSH service
	if err := f.restartSSHService(ctx); err != nil {
		return fmt.Errorf("failed to restart SSH service: %w", err)
	}

	ctx.Logger.Success("SSH security settings configured")
	return nil
}

// Priority returns the feature execution priority
func (f *Feature) Priority() int {
	return 40 // Security configuration should run last
}

// DetectCurrentState detects and returns the current state of the security feature
func (f *Feature) DetectCurrentState(ctx *features.ExecutionContext) (map[string]any, error) {
	state := make(map[string]any)

	// Check if running with sufficient privileges
	if os.Geteuid() != 0 {
		return state, fmt.Errorf("SSH security configuration requires root privileges. Please run with sudo")
	}

	// Get SSH config file path
	sshConfigFile := osdetect.GetSSHConfigPath(f.osInfo)
	state["ssh_config_file"] = sshConfigFile

	// Check if SSH config file exists
	_, err := os.Stat(sshConfigFile)
	if err != nil {
		state["ssh_config_exists"] = false
		return state, nil
	}

	state["ssh_config_exists"] = true

	// Read current config - we have root privileges so this should work
	configContent, err := os.ReadFile(sshConfigFile)
	if err != nil {
		return state, fmt.Errorf("failed to read SSH config file %s: %w", sshConfigFile, err)
	}

	// Parse SSH configuration
	rootLoginResult := f.parseSSHSetting(string(configContent), "PermitRootLogin")
	passwordAuthResult := f.parseSSHSetting(string(configContent), "PasswordAuthentication")

	// Determine security status based on explicit settings and defaults
	state["root_login_disabled"] = f.isRootLoginSecure(rootLoginResult)
	state["password_auth_disabled"] = f.isPasswordAuthSecure(passwordAuthResult)
	state["permit_root_login_value"] = rootLoginResult.EffectiveValue
	state["password_auth_value"] = passwordAuthResult.EffectiveValue
	state["permit_root_login_explicit"] = rootLoginResult.IsExplicit
	state["password_auth_explicit"] = passwordAuthResult.IsExplicit
	state["permit_root_login_source"] = rootLoginResult.Source
	state["password_auth_source"] = passwordAuthResult.Source

	// Check if running as root
	state["is_root"] = os.Geteuid() == 0

	return state, nil
}

// SSHSettingResult represents the result of parsing an SSH setting
type SSHSettingResult struct {
	EffectiveValue string // The effective value (what SSH actually uses)
	IsExplicit     bool   // Whether the setting is explicitly configured
	Source         string // Source of the value: "explicit", "default", or "commented"
}

// parseSSHSetting parses a specific SSH setting from the configuration
func (f *Feature) parseSSHSetting(configContent, settingName string) SSHSettingResult {
	lines := strings.Split(configContent, "\n")

	// Look for explicit (uncommented) setting first
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, settingName+" ") && !strings.HasPrefix(trimmedLine, "#") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 1 {
				return SSHSettingResult{
					EffectiveValue: parts[1],
					IsExplicit:     true,
					Source:         "explicit",
				}
			}
		}
	}

	// Look for commented setting to understand what would be the default
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#"+settingName+" ") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 1 {
				// Remove the # prefix
				value := parts[1]
				return SSHSettingResult{
					EffectiveValue: value,
					IsExplicit:     false,
					Source:         "commented",
				}
			}
		}
	}

	// No setting found, use SSH defaults
	defaultValue := f.getSSHDefault(settingName)
	return SSHSettingResult{
		EffectiveValue: defaultValue,
		IsExplicit:     false,
		Source:         "default",
	}
}

// getSSHDefault returns the default value for an SSH setting
func (f *Feature) getSSHDefault(settingName string) string {
	switch settingName {
	case "PermitRootLogin":
		// Default varies by OS and SSH version, but commonly "prohibit-password" or "yes"
		// We'll be conservative and assume "yes" (less secure) unless we know better
		return "yes"
	case "PasswordAuthentication":
		// Default is typically "yes"
		return "yes"
	default:
		return "unknown"
	}
}

// isRootLoginSecure determines if the root login setting is secure
func (f *Feature) isRootLoginSecure(result SSHSettingResult) bool {
	// "no" is secure, "prohibit-password" is also secure
	return result.EffectiveValue == "no" || result.EffectiveValue == "prohibit-password"
}

// isPasswordAuthSecure determines if the password authentication setting is secure
func (f *Feature) isPasswordAuthSecure(result SSHSettingResult) bool {
	// Only "no" is secure for password authentication
	return result.EffectiveValue == "no"
}

// DisplayCurrentState displays the current state of the security feature
func (f *Feature) DisplayCurrentState(ctx *features.ExecutionContext, state map[string]any) {
	if !ctx.Interactive {
		return
	}

	sshConfigExists, _ := state["ssh_config_exists"].(bool)
	sshConfigFile, _ := state["ssh_config_file"].(string)

	// SSH config file status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "SSH Config File")
	if sshConfigExists {
		fmt.Printf("\033[1;32m✓ Found\033[0m")
		if sshConfigFile != "" {
			fmt.Printf(" \033[90m(%s)\033[0m", sshConfigFile)
		}
		fmt.Println()
	} else {
		fmt.Printf("\033[1;31m✗ Not found\033[0m\n")
		fmt.Printf("    \033[90mSSH configuration will be created when you run INIQ with security options\033[0m\n")
		return
	}

	rootLoginDisabled, _ := state["root_login_disabled"].(bool)
	passwordAuthDisabled, _ := state["password_auth_disabled"].(bool)
	permitRootLoginValue, _ := state["permit_root_login_value"].(string)
	passwordAuthValue, _ := state["password_auth_value"].(string)
	permitRootLoginExplicit, _ := state["permit_root_login_explicit"].(bool)
	passwordAuthExplicit, _ := state["password_auth_explicit"].(bool)
	permitRootLoginSource, _ := state["permit_root_login_source"].(string)
	passwordAuthSource, _ := state["password_auth_source"].(string)

	// Root login status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Root Login")
	if rootLoginDisabled {
		fmt.Printf("\033[1;32m✓ Disabled\033[0m")
	} else {
		fmt.Printf("\033[1;31m✗ Enabled\033[0m")
	}

	// Show the actual setting and source
	if permitRootLoginValue != "" && permitRootLoginValue != "unknown" {
		fmt.Printf(" \033[90m(PermitRootLogin %s", permitRootLoginValue)
		if !permitRootLoginExplicit {
			switch permitRootLoginSource {
			case "commented":
				fmt.Printf(" - from commented default")
			case "default":
				fmt.Printf(" - SSH default")
			}
		}
		fmt.Printf(")\033[0m")
	}
	fmt.Println()

	// Password authentication status
	fmt.Printf("  \033[1;34m%s\033[0m: ", "Password Authentication")
	if passwordAuthDisabled {
		fmt.Printf("\033[1;32m✓ Disabled\033[0m")
	} else {
		fmt.Printf("\033[1;31m✗ Enabled\033[0m")
	}

	// Show the actual setting and source
	if passwordAuthValue != "" && passwordAuthValue != "unknown" {
		fmt.Printf(" \033[90m(PasswordAuthentication %s", passwordAuthValue)
		if !passwordAuthExplicit {
			switch passwordAuthSource {
			case "commented":
				fmt.Printf(" - from commented default")
			case "default":
				fmt.Printf(" - SSH default")
			}
		}
		fmt.Printf(")\033[0m")
	}
	fmt.Println()

	// Security recommendations
	fmt.Printf("  \033[1;34m%s\033[0m:\n", "Security Recommendations")

	if !rootLoginDisabled {
		fmt.Printf("    \033[1;33m⚠\033[0m \033[0;37mDisable root login for better security\033[0m\n")
	}

	if !passwordAuthDisabled {
		fmt.Printf("    \033[1;33m⚠\033[0m \033[0;37mDisable password authentication and use SSH keys only\033[0m\n")
	}

	if rootLoginDisabled && passwordAuthDisabled {
		fmt.Printf("    \033[1;32m✓\033[0m \033[0;37mYour SSH configuration follows security best practices\033[0m\n")
	}
}

// ShouldPromptUser determines if the user should be prompted for input
func (f *Feature) ShouldPromptUser(ctx *features.ExecutionContext, state map[string]any) bool {
	if !ctx.Interactive {
		return false
	}

	sshConfigExists, _ := state["ssh_config_exists"].(bool)
	isRoot, _ := state["is_root"].(bool)

	// If SSH config doesn't exist or we're not root, we can't configure SSH security
	if !sshConfigExists || !isRoot {
		return false
	}

	// Always prompt for SSH security settings
	return true
}

// configureRootLogin configures root login in SSH config
func (f *Feature) configureRootLogin(config string, enable bool) string {
	// For empty content, just return the comment and setting
	if strings.TrimSpace(config) == "" {
		value := "no"
		if enable {
			value = "yes"
		}
		return fmt.Sprintf("# Added by INIQ (Previous setting: none)\nPermitRootLogin %s", value)
	}

	lines := strings.Split(config, "\n")

	// First, find if PermitRootLogin exists and its current value
	permitRootLoginFound := false
	permitRootLoginValue := ""
	permitRootLoginIndex := -1

	// Find the current PermitRootLogin setting
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "PermitRootLogin") && !strings.HasPrefix(trimmedLine, "#") {
			permitRootLoginFound = true
			permitRootLoginValue = trimmedLine
			permitRootLoginIndex = i
			break
		} else if strings.HasPrefix(trimmedLine, "#PermitRootLogin") {
			// If we find a commented setting but later find an active one, we'll prefer the active one
			if !permitRootLoginFound {
				permitRootLoginFound = true
				permitRootLoginValue = trimmedLine
				permitRootLoginIndex = i
			}
		}
	}

	// Remove all existing "Modified by INIQ" comments for PermitRootLogin
	// and the PermitRootLogin setting itself
	cleanedLines := make([]string, 0, len(lines))
	skipLine := false

	for i := 0; i < len(lines); i++ {
		if skipLine {
			skipLine = false
			continue
		}

		trimmedLine := strings.TrimSpace(lines[i])

		// Skip any line that is a comment about PermitRootLogin modification
		if strings.HasPrefix(trimmedLine, "# Modified by INIQ") &&
			strings.Contains(trimmedLine, "PermitRootLogin") {
			// If the next line is the PermitRootLogin setting, skip that too
			if i+1 < len(lines) &&
				strings.HasPrefix(strings.TrimSpace(lines[i+1]), "PermitRootLogin") {
				skipLine = true
			}
			continue
		}

		// Skip the PermitRootLogin line if we found it earlier
		if i == permitRootLoginIndex {
			continue
		}

		cleanedLines = append(cleanedLines, lines[i])
	}

	// Determine the new setting value
	value := "no"
	if enable {
		value = "yes"
	}

	// Add the new setting
	if permitRootLoginFound {
		commentLine := "# Modified by INIQ (Previous setting: " + permitRootLoginValue + ")"
		newSetting := "PermitRootLogin " + value

		// For test cases compatibility
		if len(cleanedLines) > 0 {
			result := ""
			for _, line := range cleanedLines {
				result += line + "\n"
			}
			return strings.TrimRight(result, "\n") + "\n" + commentLine + "\n" + newSetting
		} else {
			return commentLine + "\n" + newSetting
		}
	} else {
		// No existing setting found, add as new
		commentLine := "# Added by INIQ (Previous setting: none)"
		newSetting := "PermitRootLogin " + value

		if len(cleanedLines) > 0 {
			result := ""
			for _, line := range cleanedLines {
				result += line + "\n"
			}
			return strings.TrimRight(result, "\n") + "\n" + commentLine + "\n" + newSetting
		} else {
			return commentLine + "\n" + newSetting
		}
	}
}

// enableRootLogin enables root login in SSH config
func (f *Feature) enableRootLogin(config string) string {
	return f.configureRootLogin(config, true)
}

// disableRootLogin disables root login in SSH config
func (f *Feature) disableRootLogin(config string) string {
	return f.configureRootLogin(config, false)
}

// configurePasswordAuth configures password authentication in SSH config
func (f *Feature) configurePasswordAuth(config string, enable bool) string {
	// For empty content, just return the comment and setting
	if strings.TrimSpace(config) == "" {
		value := "no"
		if enable {
			value = "yes"
		}
		return fmt.Sprintf("# Added by INIQ (Previous setting: none)\nPasswordAuthentication %s", value)
	}

	lines := strings.Split(config, "\n")

	// First, find if PasswordAuthentication exists and its current value
	passwordAuthFound := false
	passwordAuthValue := ""
	passwordAuthIndex := -1

	// Find the current PasswordAuthentication setting
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "PasswordAuthentication") && !strings.HasPrefix(trimmedLine, "#") {
			passwordAuthFound = true
			passwordAuthValue = trimmedLine
			passwordAuthIndex = i
			break
		} else if strings.HasPrefix(trimmedLine, "#PasswordAuthentication") {
			// If we find a commented setting but later find an active one, we'll prefer the active one
			if !passwordAuthFound {
				passwordAuthFound = true
				passwordAuthValue = trimmedLine
				passwordAuthIndex = i
			}
		}
	}

	// Remove all existing "Modified by INIQ" comments for PasswordAuthentication
	// and the PasswordAuthentication setting itself
	cleanedLines := make([]string, 0, len(lines))
	skipLine := false

	for i := 0; i < len(lines); i++ {
		if skipLine {
			skipLine = false
			continue
		}

		trimmedLine := strings.TrimSpace(lines[i])

		// Skip any line that is a comment about PasswordAuthentication modification
		if strings.HasPrefix(trimmedLine, "# Modified by INIQ") &&
			strings.Contains(trimmedLine, "PasswordAuthentication") {
			// If the next line is the PasswordAuthentication setting, skip that too
			if i+1 < len(lines) &&
				strings.HasPrefix(strings.TrimSpace(lines[i+1]), "PasswordAuthentication") {
				skipLine = true
			}
			continue
		}

		// Skip the PasswordAuthentication line if we found it earlier
		if i == passwordAuthIndex {
			continue
		}

		cleanedLines = append(cleanedLines, lines[i])
	}

	// Determine the new setting value
	value := "no"
	if enable {
		value = "yes"
	}

	// Add the new setting
	if passwordAuthFound {
		commentLine := "# Modified by INIQ (Previous setting: " + passwordAuthValue + ")"
		newSetting := "PasswordAuthentication " + value

		// For test cases compatibility
		if len(cleanedLines) > 0 {
			result := ""
			for _, line := range cleanedLines {
				result += line + "\n"
			}
			return strings.TrimRight(result, "\n") + "\n" + commentLine + "\n" + newSetting
		} else {
			return commentLine + "\n" + newSetting
		}
	} else {
		// No existing setting found, add as new
		commentLine := "# Added by INIQ (Previous setting: none)"
		newSetting := "PasswordAuthentication " + value

		if len(cleanedLines) > 0 {
			result := ""
			for _, line := range cleanedLines {
				result += line + "\n"
			}
			return strings.TrimRight(result, "\n") + "\n" + commentLine + "\n" + newSetting
		} else {
			return commentLine + "\n" + newSetting
		}
	}
}

// enablePasswordAuth enables password authentication in SSH config
func (f *Feature) enablePasswordAuth(config string) string {
	return f.configurePasswordAuth(config, true)
}

// disablePasswordAuth disables password authentication in SSH config
func (f *Feature) disablePasswordAuth(config string) string {
	return f.configurePasswordAuth(config, false)
}

// restartSSHService restarts the SSH service
func (f *Feature) restartSSHService(ctx *features.ExecutionContext) error {
	ctx.Logger.Info("Restarting SSH service")

	// Get restart command based on OS
	restartCmd := osdetect.GetServiceRestartCommand("ssh", f.osInfo)
	if restartCmd == "" {
		return fmt.Errorf("failed to determine SSH service restart command")
	}

	// Execute restart command
	cmd := exec.Command("sh", "-c", restartCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart SSH service: %w", err)
	}

	ctx.Logger.Success("SSH service restarted")
	return nil
}
