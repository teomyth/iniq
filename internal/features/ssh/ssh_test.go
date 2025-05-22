package ssh

import (
	"testing"

	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/internal/logger"
	"github.com/teomyth/iniq/pkg/osdetect"
)

func TestFeatureInterface(t *testing.T) {
	// Verify that Feature implements the features.Feature interface
	var _ features.Feature = (*Feature)(nil)
}

func TestNew(t *testing.T) {
	// Test creating a new feature
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	if feature == nil {
		t.Fatal("New returned nil")
	}

	if feature.Name() != "ssh" {
		t.Errorf("Expected feature name 'ssh', got %q", feature.Name())
	}

	if feature.Description() == "" {
		t.Error("Feature description is empty")
	}
}

func TestShouldActivate(t *testing.T) {
	// Create a feature for testing
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	// Test cases
	tests := []struct {
		name     string
		options  map[string]any
		expected bool
	}{
		{
			name:     "Empty options",
			options:  map[string]any{},
			expected: false,
		},
		{
			name: "With keys",
			options: map[string]any{
				"keys": []string{"key1", "key2"},
			},
			expected: true,
		},
		{
			name: "With empty keys slice",
			options: map[string]any{
				"keys": []string{},
			},
			expected: false,
		},
		{
			name: "With non-slice keys",
			options: map[string]any{
				"keys": "not a slice",
			},
			expected: false,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := feature.ShouldActivate(tc.options)
			if result != tc.expected {
				t.Errorf("ShouldActivate() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	// Create a feature for testing
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	// Test cases
	tests := []struct {
		name        string
		options     map[string]any
		expectError bool
	}{
		{
			name:        "Empty options",
			options:     map[string]any{},
			expectError: false, // No keys is valid, just won't activate
		},
		{
			name: "With valid keys",
			options: map[string]any{
				"keys": []string{"github:username", "gitlab:username", "file:/path/to/key"},
			},
			expectError: false,
		},
		{
			name: "With invalid key format",
			options: map[string]any{
				"keys": []string{"invalid:format"},
			},
			expectError: true,
		},
		{
			name: "With non-slice keys",
			options: map[string]any{
				"keys": "not a slice",
			},
			expectError: false, // Will be handled by ShouldActivate
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := feature.ValidateOptions(tc.options)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// This is a complex function that interacts with the system
	// We'll skip the actual execution test and just verify it doesn't panic
	// In a real test, you would mock the system calls

	// Create a feature for testing
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	// Create a context with minimal options
	ctx := &features.ExecutionContext{
		Options: map[string]any{
			"user": "testuser",
			"keys": []string{"github:testuser"},
		},
		Logger:      logger.New(false, false),
		DryRun:      true, // Use dry run to avoid actual system changes
		Interactive: false,
		Verbose:     false,
	}

	// Execute should not panic in dry run mode
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Execute panicked: %v", r)
			}
		}()

		// Call Execute in dry run mode
		err := feature.Execute(ctx)
		if err != nil {
			// We expect an error in dry run mode on most systems
			// because the user doesn't exist, but it shouldn't panic
			t.Logf("Execute returned error (expected in dry run): %v", err)
		}
	}()
}
