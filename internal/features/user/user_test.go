package user

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

	if feature.Name() != "user" {
		t.Errorf("Expected feature name 'user', got %q", feature.Name())
	}

	if feature.Description() == "" {
		t.Error("Feature description is empty")
	}
}

func TestPasswordPolicyValidation(t *testing.T) {
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)
	log := logger.New(false, false)

	tests := []struct {
		name        string
		options     map[string]any
		interactive bool
		expectError bool
		errorMsg    string
	}{
		{
			name: "conflicting password options",
			options: map[string]any{
				"user":        "testuser",
				"password":    true,
				"no-password": true,
			},
			interactive: true,
			expectError: true,
			errorMsg:    "cannot specify both --password and --no-pass options",
		},
		{
			name: "no password in non-interactive mode",
			options: map[string]any{
				"user": "testuser",
			},
			interactive: false,
			expectError: true,
			errorMsg:    "creating user 'testuser' requires password input",
		},
		{
			name: "explicit no-password in non-interactive mode",
			options: map[string]any{
				"user":        "testuser",
				"no-password": true,
			},
			interactive: false,
			expectError: false,
		},
		{
			name: "password option in non-interactive mode",
			options: map[string]any{
				"user":     "testuser",
				"password": true,
			},
			interactive: false,
			expectError: true,
			errorMsg:    "creating user 'testuser' requires password input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &features.ExecutionContext{
				Logger:      log,
				Options:     tt.options,
				Interactive: tt.interactive,
				DryRun:      true, // Use dry run to avoid actual user creation
			}

			err := feature.Execute(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
			name: "With username",
			options: map[string]any{
				"user": "testuser",
			},
			expected: true,
		},
		{
			name: "With empty username",
			options: map[string]any{
				"user": "",
			},
			expected: false,
		},
		{
			name: "With non-string username",
			options: map[string]any{
				"user": 123,
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
			expectError: true,
		},
		{
			name: "With valid username",
			options: map[string]any{
				"user": "testuser",
			},
			expectError: false,
		},
		{
			name: "With empty username",
			options: map[string]any{
				"user": "",
			},
			expectError: true,
		},
		{
			name: "With invalid username (spaces)",
			options: map[string]any{
				"user": "test user",
			},
			expectError: true,
		},
		{
			name: "With invalid username (special chars)",
			options: map[string]any{
				"user": "test@user",
			},
			expectError: true,
		},
		{
			name: "With non-string username",
			options: map[string]any{
				"user": 123,
			},
			expectError: true,
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
