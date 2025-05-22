// Package logger provides logging functionality for the application
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// Icons for different log levels
const (
	IconDebug    = "›"
	IconStep     = "→"
	IconWarning  = "!"
	IconError    = "×"
	IconFatal    = "×"
	IconSuccess  = "✓"
	IconListItem = "•"
)

// Logger wraps logrus.Logger to provide additional functionality
type Logger struct {
	*logrus.Logger
	verbose       bool
	quiet         bool
	fileLogger    *logrus.Logger
	currentOp     int
	totalOps      int
	indentLevel   int
	operationName string
}

// New creates a new logger instance
func New(verbose, quiet bool) *Logger {
	// Create main logger
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&CustomFormatter{
		UseColors: isTerminal(os.Stdout),
	})

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Create file logger if possible
	var fileLogger *logrus.Logger
	logPath := getLogFilePath()
	if logPath != "" {
		fileLogger = logrus.New()
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			fileLogger.SetOutput(logFile)
			fileLogger.SetFormatter(&logrus.JSONFormatter{})
			fileLogger.SetLevel(logrus.DebugLevel) // Always log everything to file
		}
	}

	return &Logger{
		Logger:        log,
		verbose:       verbose,
		quiet:         quiet,
		fileLogger:    fileLogger,
		currentOp:     0,
		totalOps:      0,
		indentLevel:   0,
		operationName: "",
	}
}

// getLogFilePath returns the path to the log file
func getLogFilePath() string {
	// If running as root, use system log directory
	if os.Geteuid() == 0 {
		return "/var/log/iniq.log"
	}

	// Otherwise, use user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Create .iniq/logs directory if it doesn't exist
	logDir := filepath.Join(homeDir, ".iniq", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return ""
	}

	return filepath.Join(logDir, "iniq.log")
}

// isTerminal checks if the given file is a terminal
func isTerminal(f *os.File) bool {
	if runtime.GOOS == "windows" {
		return false // Simplified for Windows
	}
	stat, _ := f.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// CustomFormatter is a custom logrus formatter
type CustomFormatter struct {
	UseColors bool
}

// Format formats the log entry
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Check if this is an indented line (part of a multi-line output)
	if _, ok := entry.Data["indent"]; ok {
		return []byte(fmt.Sprintf("  %s\n", entry.Message)), nil
	}

	// Get icon based on level or custom level
	var icon string
	if level, ok := entry.Data["level"]; ok {
		switch level {
		case "success":
			icon = IconSuccess
		case "step":
			icon = IconStep
		case "list-item":
			icon = IconListItem
		default:
			icon = IconStep // Default to step icon
		}
	} else {
		switch entry.Level {
		case logrus.DebugLevel:
			icon = IconDebug
		case logrus.InfoLevel:
			icon = IconStep // Info is now step
		case logrus.WarnLevel:
			icon = IconWarning
		case logrus.ErrorLevel:
			icon = IconError
		case logrus.FatalLevel, logrus.PanicLevel:
			icon = IconFatal
		default:
			icon = IconStep // Default to step icon
		}
	}

	// Format message with colors if enabled
	var output string
	if f.UseColors {
		var colorCode string
		if level, ok := entry.Data["level"]; ok {
			switch level {
			case "success":
				colorCode = "\033[32m" // Green
			case "step":
				colorCode = "\033[34m" // Blue
			case "list-item":
				colorCode = "\033[90m" // Gray
			default:
				colorCode = "\033[34m" // Blue
			}
		} else {
			switch entry.Level {
			case logrus.DebugLevel:
				colorCode = "\033[36m" // Cyan
			case logrus.InfoLevel:
				colorCode = "\033[34m" // Blue
			case logrus.WarnLevel:
				colorCode = "\033[33m" // Yellow
			case logrus.ErrorLevel:
				colorCode = "\033[31m" // Red
			case logrus.FatalLevel, logrus.PanicLevel:
				colorCode = "\033[31m" // Red
			default:
				colorCode = "\033[0m" // Reset
			}
		}

		resetCode := "\033[0m"
		output = fmt.Sprintf("%s%s%s %s", colorCode, icon, resetCode, entry.Message)
	} else {
		output = fmt.Sprintf("%s %s", icon, entry.Message)
	}

	// Add fields if any (but skip 'level' field as it's used internally)
	if len(entry.Data) > 0 {
		fields := make(map[string]any)
		for k, v := range entry.Data {
			// Skip the 'level' field as it's used internally for icon selection
			if k != "level" {
				fields[k] = v
			}
		}

		if len(fields) > 0 {
			output += " {"
			for k, v := range fields {
				output += fmt.Sprintf(" %s=%v", k, v)
			}
			output += " }"
		}
	}

	return []byte(output + "\n"), nil
}

// SetOutput sets the output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.Logger.SetOutput(w)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	if l.verbose && !l.quiet {
		l.Logger.Debugf(format, args...)
	}
	if l.fileLogger != nil {
		l.fileLogger.Debugf(format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	if !l.quiet {
		l.Logger.Infof(format, args...)
	}
	if l.fileLogger != nil {
		l.fileLogger.Infof(format, args...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...any) {
	if !l.quiet {
		l.Logger.Warnf(format, args...)
	}
	if l.fileLogger != nil {
		l.fileLogger.Warnf(format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.Logger.Errorf(format, args...)
	if l.fileLogger != nil {
		l.fileLogger.Errorf(format, args...)
	}
}

// Fatal logs a fatal error message and exits
func (l *Logger) Fatal(format string, args ...any) {
	l.Logger.Fatalf(format, args...)
	// No need to log to file as Fatalf will exit
}

// Success logs a success message
func (l *Logger) Success(format string, args ...any) {
	if !l.quiet {
		// Use info level with success icon
		entry := l.Logger.WithField("level", "success")
		entry.Logger.Formatter = &CustomFormatter{
			UseColors: isTerminal(os.Stdout),
		}
		entry.Info(fmt.Sprintf(format, args...))
	}
	if l.fileLogger != nil {
		l.fileLogger.WithField("level", "success").Info(fmt.Sprintf(format, args...))
	}
}

// MultiLine logs a multi-line message with the first line having an icon and subsequent lines indented
func (l *Logger) MultiLine(level string, heading string, lines []string) {
	if l.quiet && level != "error" && level != "warning" {
		return
	}

	// Log the heading with appropriate icon
	switch level {
	case "debug":
		l.Debug(heading)
	case "info":
		l.Info(heading)
	case "warning":
		l.Warning(heading)
	case "error":
		l.Error(heading)
	case "success":
		l.Success(heading)
	case "step":
		l.Step(heading)
	case "list-item":
		l.ListItem(heading)
	default:
		l.Step(heading) // Default to step
	}

	// Log each line with proper indentation
	for _, line := range lines {
		if !l.quiet {
			fmt.Fprintf(os.Stdout, "  %s\n", line)
		}
		if l.fileLogger != nil {
			l.fileLogger.WithField("indent", true).Info(line)
		}
	}
}

// Step logs a step message (replacing the old "working" concept)
func (l *Logger) Step(format string, args ...any) {
	if !l.quiet {
		// Use info level with step icon
		entry := l.Logger.WithField("level", "step")
		entry.Logger.Formatter = &CustomFormatter{
			UseColors: isTerminal(os.Stdout),
		}
		entry.Info(fmt.Sprintf(format, args...))
	}
	if l.fileLogger != nil {
		l.fileLogger.WithField("level", "step").Info(fmt.Sprintf(format, args...))
	}
}

// ListItem logs a list item
func (l *Logger) ListItem(format string, args ...any) {
	if !l.quiet {
		// Use info level with list item icon
		entry := l.Logger.WithField("level", "list-item")
		entry.Logger.Formatter = &CustomFormatter{
			UseColors: isTerminal(os.Stdout),
		}
		entry.Info(fmt.Sprintf(format, args...))
	}
	if l.fileLogger != nil {
		l.fileLogger.WithField("level", "list-item").Info(fmt.Sprintf(format, args...))
	}
}

// SetOperations sets the total number of operations
func (l *Logger) SetOperations(total int) {
	l.totalOps = total
}

// StartOperation starts a new operation with the given name
func (l *Logger) StartOperation(name string) {
	l.currentOp++
	l.operationName = name
	l.indentLevel = 0

	if !l.quiet {
		// Only add a newline if this is not the first operation
		if l.currentOp > 1 {
			fmt.Println()
		}

		// Use a more prominent format for operation title
		if isTerminal(os.Stdout) {
			// Format: [1/4] with blue color, then operation name with bold
			fmt.Printf("\033[0;34m[%d/%d]\033[0m \033[1m%s\033[0m\n", l.currentOp, l.totalOps, name)
			fmt.Printf("────────────────────────────────────────\n")
		} else {
			fmt.Printf("[%d/%d] %s\n", l.currentOp, l.totalOps, name)
			fmt.Printf("--------------------------------\n")
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"operation": l.currentOp,
			"total":     l.totalOps,
			"name":      name,
		}).Info("Starting operation")
	}
}

// IncreaseIndent increases the indent level
func (l *Logger) IncreaseIndent() {
	l.indentLevel++
}

// DecreaseIndent decreases the indent level
func (l *Logger) DecreaseIndent() {
	if l.indentLevel > 0 {
		l.indentLevel--
	}
}

// ResetIndent resets the indent level to 0
func (l *Logger) ResetIndent() {
	l.indentLevel = 0
}

// Indent returns the current indentation string
func (l *Logger) Indent() string {
	return strings.Repeat("  ", l.indentLevel)
}

// PrintIndent prints the current indentation
func (l *Logger) PrintIndent() {
	if !l.quiet {
		fmt.Print(l.Indent())
	}
}

// StepWithIndent logs a step message with the current indentation
func (l *Logger) StepWithIndent(format string, args ...any) {
	// Ensure we have at least one level of indentation for steps
	if l.indentLevel == 0 {
		l.indentLevel = 1
	}

	if !l.quiet {
		if isTerminal(os.Stdout) {
			// Blue color for step icon
			fmt.Printf("%s\033[34m→\033[0m %s\n", l.Indent(), fmt.Sprintf(format, args...))
		} else {
			fmt.Printf("%s→ %s\n", l.Indent(), fmt.Sprintf(format, args...))
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"operation": l.currentOp,
			"indent":    l.indentLevel,
			"level":     "step",
		}).Info(fmt.Sprintf(format, args...))
	}
}

// SuccessWithIndent logs a success message with the current indentation
func (l *Logger) SuccessWithIndent(format string, args ...any) {
	if !l.quiet {
		if isTerminal(os.Stdout) {
			// Green color for success icon
			fmt.Printf("%s\033[32m✓\033[0m %s\n", l.Indent(), fmt.Sprintf(format, args...))
		} else {
			fmt.Printf("%s✓ %s\n", l.Indent(), fmt.Sprintf(format, args...))
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"operation": l.currentOp,
			"indent":    l.indentLevel,
			"level":     "success",
		}).Info(fmt.Sprintf(format, args...))
	}
}

// ListItemWithIndent logs a list item with the current indentation
func (l *Logger) ListItemWithIndent(format string, args ...any) {
	if !l.quiet {
		if isTerminal(os.Stdout) {
			// Gray color for list item icon
			fmt.Printf("%s\033[90m•\033[0m %s\n", l.Indent(), fmt.Sprintf(format, args...))
		} else {
			fmt.Printf("%s• %s\n", l.Indent(), fmt.Sprintf(format, args...))
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"operation": l.currentOp,
			"indent":    l.indentLevel,
			"level":     "list-item",
		}).Info(fmt.Sprintf(format, args...))
	}
}

// TextWithIndent logs plain text with the current indentation
func (l *Logger) TextWithIndent(format string, args ...any) {
	if !l.quiet {
		fmt.Printf("%s%s\n", l.Indent(), fmt.Sprintf(format, args...))
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"operation": l.currentOp,
			"indent":    l.indentLevel,
		}).Info(fmt.Sprintf(format, args...))
	}
}

// PrintOperationSummary prints a summary of all operations
func (l *Logger) PrintOperationSummary(results map[string]bool) {
	if !l.quiet {
		fmt.Println("\n\033[1mOperation Summary\033[0m")
		fmt.Println("────────────────────────────────────────")

		useColors := isTerminal(os.Stdout)
		for name, success := range results {
			if success {
				if useColors {
					fmt.Printf("\033[32m✓\033[0m %s\n", name)
				} else {
					fmt.Printf("✓ %s\n", name)
				}
			} else {
				if useColors {
					fmt.Printf("\033[31m×\033[0m %s\n", name)
				} else {
					fmt.Printf("× %s\n", name)
				}
			}
		}

		// Check if all operations were successful
		allSuccess := true
		for _, success := range results {
			if !success {
				allSuccess = false
				break
			}
		}

		fmt.Println()
		if allSuccess {
			if useColors {
				fmt.Println("\033[32mSystem initialization completed successfully!\033[0m")
			} else {
				fmt.Println("System initialization completed successfully!")
			}
		} else {
			if useColors {
				fmt.Println("\033[31mSystem initialization completed with errors.\033[0m")
			} else {
				fmt.Println("System initialization completed with errors.")
			}
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithField("level", "summary").Info("Operation summary")
		for name, success := range results {
			l.fileLogger.WithFields(logrus.Fields{
				"operation": name,
				"success":   success,
				"level":     "summary",
			}).Info("Operation result")
		}
	}
}

// PrintSystemConfig prints the system configuration summary
func (l *Logger) PrintSystemConfig(configs map[string]string) {
	if !l.quiet {
		fmt.Println("\n\033[1mSystem Configuration\033[0m")
		fmt.Println("────────────────────────────────────────")

		useColors := isTerminal(os.Stdout)
		for name, value := range configs {
			if useColors {
				fmt.Printf("\033[0;34m•\033[0m %s: \033[32m%s\033[0m\n", name, value)
			} else {
				fmt.Printf("• %s: %s\n", name, value)
			}
		}
	}

	if l.fileLogger != nil {
		for name, value := range configs {
			l.fileLogger.WithFields(logrus.Fields{
				"config": name,
				"value":  value,
				"level":  "config",
			}).Info("System configuration")
		}
	}
}

// PrintOperationList prints the list of operations to be performed
func (l *Logger) PrintOperationList(operations []string) {
	l.SetOperations(len(operations))

	if !l.quiet {
		fmt.Println("\n\033[1mOperations to be performed:\033[0m")
		fmt.Println("────────────────────────────────────────")

		useColors := isTerminal(os.Stdout)
		for i, op := range operations {
			if useColors {
				fmt.Printf("\033[0;34m%d.\033[0m %s\n", i+1, op)
			} else {
				fmt.Printf("%d. %s\n", i+1, op)
			}
		}
		// No extra newline here, as StartOperation will add one for the first operation
	}

	if l.fileLogger != nil {
		l.fileLogger.WithField("total_operations", len(operations)).Info("Operations to be performed")
		for i, op := range operations {
			l.fileLogger.WithFields(logrus.Fields{
				"operation_index": i + 1,
				"operation_name":  op,
			}).Info("Planned operation")
		}
	}
}

// PrintHeader prints the application header
func (l *Logger) PrintHeader(version, description string) {
	if !l.quiet {
		if isTerminal(os.Stdout) {
			fmt.Printf("\033[1m%s\033[0m - %s\n", version, description)
		} else {
			fmt.Printf("%s - %s\n", version, description)
		}
	}

	if l.fileLogger != nil {
		l.fileLogger.WithFields(logrus.Fields{
			"version":     version,
			"description": description,
		}).Info("Application started")
	}
}
