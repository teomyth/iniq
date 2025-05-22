// Package features defines the core feature interface and common functionality
package features

import (
	"github.com/teomyth/iniq/internal/logger"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// Feature registration functions
var (
	RegisterUserFeature     func(*Registry, *osdetect.Info)
	RegisterSSHFeature      func(*Registry, *osdetect.Info)
	RegisterSudoFeature     func(*Registry, *osdetect.Info)
	RegisterSecurityFeature func(*Registry, *osdetect.Info)
)

// Flag represents a command-line flag for a feature
type Flag struct {
	// Name is the long name of the flag (e.g., "user")
	Name string

	// Shorthand is the short name of the flag (e.g., "u")
	Shorthand string

	// Usage is the help text for the flag
	Usage string

	// Default is the default value for the flag
	Default any

	// Required indicates if the flag is required
	Required bool
}

// ExecutionContext provides context for feature execution
type ExecutionContext struct {
	// Options contains the parsed command-line options
	Options map[string]any

	// Logger is the logger instance
	Logger *logger.Logger

	// DryRun indicates if the feature should only simulate execution
	DryRun bool

	// Interactive indicates if the feature should prompt for user input
	Interactive bool

	// Verbose indicates if the feature should output verbose information
	Verbose bool
}

// Feature defines the interface that all features must implement
type Feature interface {
	// Name returns the feature name
	Name() string

	// Description returns the feature description
	Description() string

	// Flags returns the command-line flags for the feature
	Flags() []Flag

	// ShouldActivate determines if the feature should be activated based on options
	ShouldActivate(options map[string]any) bool

	// ValidateOptions validates the feature options
	ValidateOptions(options map[string]any) error

	// Execute executes the feature functionality
	Execute(ctx *ExecutionContext) error

	// Priority returns the feature execution priority (lower numbers run first)
	Priority() int

	// DetectCurrentState detects and returns the current state of the feature
	// The returned map contains state information that can be used for display and decision making
	DetectCurrentState(ctx *ExecutionContext) (map[string]any, error)

	// DisplayCurrentState displays the current state of the feature to the user
	// This is used in interactive mode to show the user the current state before prompting
	DisplayCurrentState(ctx *ExecutionContext, state map[string]any)

	// ShouldPromptUser determines if the user should be prompted for input based on the current state
	// This allows features to skip prompting if the current state already matches the desired state
	ShouldPromptUser(ctx *ExecutionContext, state map[string]any) bool
}

// Registry manages the registration and retrieval of features
type Registry struct {
	features []Feature
}

// NewRegistry creates a new feature registry
func NewRegistry() *Registry {
	return &Registry{
		features: make([]Feature, 0),
	}
}

// Register registers a feature with the registry
func (r *Registry) Register(feature Feature) {
	r.features = append(r.features, feature)
}

// GetFeatures returns all registered features
func (r *Registry) GetFeatures() []Feature {
	return r.features
}

// GetActiveFeatures returns features that should be activated based on options
func (r *Registry) GetActiveFeatures(options map[string]any) []Feature {
	var activeFeatures []Feature

	for _, feature := range r.features {
		if feature.ShouldActivate(options) {
			activeFeatures = append(activeFeatures, feature)
		}
	}

	return activeFeatures
}

// SortFeaturesByPriority sorts features by priority (lower numbers first)
func SortFeaturesByPriority(features []Feature) []Feature {
	// Create a copy to avoid modifying the original slice
	sortedFeatures := make([]Feature, len(features))
	copy(sortedFeatures, features)

	// Simple insertion sort by priority
	for i := 1; i < len(sortedFeatures); i++ {
		j := i
		for j > 0 && sortedFeatures[j-1].Priority() > sortedFeatures[j].Priority() {
			sortedFeatures[j-1], sortedFeatures[j] = sortedFeatures[j], sortedFeatures[j-1]
			j--
		}
	}

	return sortedFeatures
}
