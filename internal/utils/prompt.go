package utils

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Colors for terminal output
const (
	// Basic colors
	ColorReset   = "\033[0m"
	ColorGray    = "\033[90m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorRed     = "\033[31m"
	ColorCyan    = "\033[36m"
	ColorMagenta = "\033[35m"

	// Bright colors
	ColorBrightGreen  = "\033[92m"
	ColorBrightYellow = "\033[93m"
	ColorBrightBlue   = "\033[94m"
	ColorBrightRed    = "\033[91m"
	ColorBrightCyan   = "\033[96m"

	// Text styles
	ColorBold      = "\033[1m"
	ColorUnderline = "\033[4m"

	// Combined styles
	ColorHeaderBlue = "\033[1;34m" // Bold Blue
	ColorHeaderCyan = "\033[1;36m" // Bold Cyan
	ColorSuccess    = "\033[32m"   // Green
	ColorWarning    = "\033[33m"   // Yellow
	ColorError      = "\033[31m"   // Red
	ColorInfo       = "\033[34m"   // Blue
	ColorDebug      = "\033[36m"   // Cyan
	ColorListItem   = "\033[90m"   // Gray
)

// PromptYesNo asks the user a yes/no question and returns true for yes, false for no.
// The defaultValue parameter determines the default answer if the user just presses Enter.
// Shows gray placeholder text for the default value that gets replaced when user types.
func PromptYesNo(question string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Determine the default character and prompt format
		var prompt string
		if defaultValue {
			prompt = fmt.Sprintf("%s [%sY%s/n]: ", question, ColorGray, ColorReset)
		} else {
			prompt = fmt.Sprintf("%s [y/%sN%s]: ", question, ColorGray, ColorReset)
		}

		// Print the prompt
		fmt.Print(prompt)

		// Read the input
		input, err := reader.ReadString('\n')
		if err != nil {
			// On error, return the default value
			return defaultValue
		}

		// Trim whitespace and convert to lowercase for comparison
		trimmedInput := strings.TrimSpace(strings.ToLower(input))

		// Determine the result based on input
		if trimmedInput == "" {
			// User pressed Enter, use default value
			// No feedback needed - just return the default
			return defaultValue
		} else if trimmedInput == "y" || trimmedInput == "yes" {
			return true
		} else if trimmedInput == "n" || trimmedInput == "no" {
			return false
		} else {
			// Invalid input, show error and retry
			fmt.Printf("Invalid input. Please enter 'y' for yes or 'n' for no.\n")
			continue
		}
	}
}

// PromptYesNoWithFeedback asks the user a yes/no question and provides feedback when using defaults
// featureName and currentState are used to provide meaningful feedback about the current setting
func PromptYesNoWithFeedback(question string, defaultValue bool, featureName string, currentState bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Determine the prompt format
		var prompt string
		if defaultValue {
			prompt = fmt.Sprintf("%s [%sY%s/n]: ", question, ColorGray, ColorReset)
		} else {
			prompt = fmt.Sprintf("%s [y/%sN%s]: ", question, ColorGray, ColorReset)
		}

		// Print the prompt
		fmt.Print(prompt)

		// Read the input
		input, err := reader.ReadString('\n')
		if err != nil {
			// On error, return the default value
			return defaultValue
		}

		// Trim whitespace and convert to lowercase for comparison
		trimmedInput := strings.TrimSpace(strings.ToLower(input))

		// Determine the result based on input
		if trimmedInput == "" {
			// User pressed Enter, use default value and show feedback
			showKeepCurrentFeedback(featureName, currentState, defaultValue)
			return defaultValue
		} else if trimmedInput == "y" || trimmedInput == "yes" {
			return true
		} else if trimmedInput == "n" || trimmedInput == "no" {
			return false
		} else {
			// Invalid input, show error and retry
			fmt.Printf("Invalid input. Please enter 'y' for yes or 'n' for no.\n")
			continue
		}
	}
}

// showKeepCurrentFeedback shows feedback when user keeps current setting
func showKeepCurrentFeedback(featureName string, currentState bool, defaultChoice bool) {
	// Determine the current setting description
	var currentDesc string
	if featureName == "SSH root login" {
		if currentState {
			currentDesc = "Root login enabled"
		} else {
			currentDesc = "Root login disabled"
		}
	} else if featureName == "SSH password authentication" {
		if currentState {
			currentDesc = "Password authentication enabled"
		} else {
			currentDesc = "Password authentication disabled"
		}
	} else {
		// Generic fallback
		if currentState {
			currentDesc = fmt.Sprintf("%s enabled", featureName)
		} else {
			currentDesc = fmt.Sprintf("%s disabled", featureName)
		}
	}

	// Show feedback with checkmark and gray color to indicate "keeping current"
	fmt.Printf("%s✓ Keeping current setting: %s%s\n", ColorGray, currentDesc, ColorReset)
}

// ParseBoolValue parses various boolean value formats
// Supports: yes|enable|true|1|y|t|on for true values
//
//	no|disable|false|0|n|f|off for false values
func ParseBoolValue(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "enable", "true", "1", "y", "t", "on":
		return true, nil
	case "no", "disable", "false", "0", "n", "f", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s (supported: yes|enable|true|1|y|t|on or no|disable|false|0|n|f|off)", value)
	}
}

// ColorText wraps text with the specified color code and resets the color afterward
func ColorText(text string, colorCode string) string {
	return colorCode + text + ColorReset
}

// BoldText returns the text in bold
func BoldText(text string) string {
	return "\033[1m" + text + ColorReset
}

// SuccessText returns the text in success color (green)
func SuccessText(text string) string {
	return ColorGreen + text + ColorReset
}

// ErrorText returns the text in error color (red)
func ErrorText(text string) string {
	return ColorRed + text + ColorReset
}

// WarningText returns the text in warning color (yellow)
func WarningText(text string) string {
	return ColorYellow + text + ColorReset
}

// InfoText returns the text in info color (blue)
func InfoText(text string) string {
	return ColorBlue + text + ColorReset
}

// StateToggleAction represents the action to take for a state toggle
type StateToggleAction string

const (
	StateToggleKeep    StateToggleAction = "keep"
	StateToggleEnable  StateToggleAction = "enable"
	StateToggleDisable StateToggleAction = "disable"
)

// StateToggleConfig holds configuration for a state toggle prompt
type StateToggleConfig struct {
	// Feature name (e.g., "SSH root login", "SSH password authentication")
	FeatureName string
	// Current state (true = enabled, false = disabled)
	CurrentState bool
	// Custom prompt (optional)
	AllowPrompt string // Default: "Allow {FeatureName}?"
}

// StateToggleResult holds the result of a state toggle interaction
type StateToggleResult struct {
	Action    StateToggleAction
	HasChange bool
}

// PromptStateToggle provides a user-friendly state toggle interaction using "Allow" syntax
// Returns the action to take and whether there's a change from current state
func PromptStateToggle(config StateToggleConfig) StateToggleResult {
	// Set default prompt if not provided, including current state information
	if config.AllowPrompt == "" {
		// Show current state in the prompt to make it clear
		currentStateDesc := "disabled"
		if config.CurrentState {
			currentStateDesc = "enabled"
		}
		config.AllowPrompt = fmt.Sprintf("Allow %s? (currently %s)", config.FeatureName, currentStateDesc)
	}

	// Use three-state prompt: y/n/keep-current (Enter)
	choice := PromptThreeState(config.AllowPrompt, config.FeatureName, config.CurrentState)

	// Determine the action and whether there's a change
	var action StateToggleAction
	var hasChange bool

	switch choice {
	case "enable":
		action = StateToggleEnable
		hasChange = !config.CurrentState // Change if currently disabled
	case "disable":
		action = StateToggleDisable
		hasChange = config.CurrentState // Change if currently enabled
	case "keep":
		action = StateToggleKeep
		hasChange = false // No change
	}

	return StateToggleResult{
		Action:    action,
		HasChange: hasChange,
	}
}

// PromptThreeState asks the user a three-state question: enable/disable/keep-current
// Returns "enable", "disable", or "keep"
func PromptThreeState(question string, featureName string, currentState bool) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Show prompt with [y/n] format (no default, Enter = keep current)
		prompt := fmt.Sprintf("%s [y/n]: ", question)
		fmt.Print(prompt)

		// Read the input
		input, err := reader.ReadString('\n')
		if err != nil {
			// On error, keep current state
			showKeepCurrentFeedback(featureName, currentState, true)
			return "keep"
		}

		// Trim whitespace and convert to lowercase for comparison
		trimmedInput := strings.TrimSpace(strings.ToLower(input))

		// Determine the result based on input
		if trimmedInput == "" {
			// User pressed Enter, keep current state
			showKeepCurrentFeedback(featureName, currentState, true)
			return "keep"
		} else if trimmedInput == "y" || trimmedInput == "yes" {
			return "enable"
		} else if trimmedInput == "n" || trimmedInput == "no" {
			return "disable"
		} else {
			// Invalid input, show error and retry
			fmt.Printf("Invalid input. Please enter 'y' for yes, 'n' for no, or press Enter to keep current setting.\n")
			continue
		}
	}
}

// IsTerminal checks if the given file is a terminal
func IsTerminal(f *os.File) bool {
	if runtime.GOOS == "windows" {
		return false // Simplified for Windows
	}
	stat, _ := f.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// PromptWithDefault asks the user for input with a default value
// and returns the user's input or the default if the user just presses Enter.
func PromptWithDefault(question string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	// Determine the prompt format based on whether there's a default value
	var prompt string
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s%s%s]: ",
			question,
			ColorGray,
			defaultValue,
			ColorReset)
	} else {
		prompt = fmt.Sprintf("%s: ", question)
	}

	// Print the prompt
	fmt.Print(prompt)

	// Read the input
	input, err := reader.ReadString('\n')
	if err != nil {
		// On error, return the default value
		return defaultValue
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// If input is empty, use default value
	if input == "" {
		input = defaultValue
		// No feedback needed - just return the default value
	}

	return input
}
