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
// After the user makes a choice, it displays what was selected.
func PromptYesNo(question string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)

	// Determine the prompt format based on default value
	var prompt string
	if defaultValue {
		prompt = fmt.Sprintf("%s [Y/n]: ", question)
	} else {
		prompt = fmt.Sprintf("%s [y/N]: ", question)
	}

	// Print the prompt
	fmt.Print(prompt)

	// Read the input
	input, err := reader.ReadString('\n')
	if err != nil {
		// On error, return the default value
		return defaultValue
	}

	// Trim whitespace and convert to lowercase
	input = strings.TrimSpace(strings.ToLower(input))

	// Determine the result based on input
	var result bool
	if input == "" {
		// User pressed Enter, use default value
		result = defaultValue
	} else if input == "y" || input == "yes" {
		result = true
	} else if input == "n" || input == "no" {
		result = false
	} else {
		// Invalid input, use default value
		fmt.Printf("%sInvalid input, using default: %s%s\n",
			ColorYellow,
			getYesNoText(defaultValue),
			ColorReset)
		return defaultValue
	}

	// Show what was selected
	fmt.Printf("%sSelected: %s%s\n",
		ColorGray,
		getYesNoText(result),
		ColorReset)

	return result
}

// getYesNoText returns "Yes" or "No" based on the boolean value
func getYesNoText(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
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
		// Show what was selected
		if defaultValue != "" {
			fmt.Printf("%sUsing default: %s%s\n",
				ColorGray,
				defaultValue,
				ColorReset)
		}
	}

	return input
}
