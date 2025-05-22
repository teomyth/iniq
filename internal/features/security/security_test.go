package security

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

	if feature.Name() != "security" {
		t.Errorf("Expected feature name 'security', got %q", feature.Name())
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
			name: "With ssh-no-root true",
			options: map[string]any{
				"ssh-no-root": true,
			},
			expected: true,
		},
		{
			name: "With ssh-no-password true",
			options: map[string]any{
				"ssh-no-password": true,
			},
			expected: true,
		},
		{
			name: "With both options true",
			options: map[string]any{
				"ssh-no-root":     true,
				"ssh-no-password": true,
			},
			expected: true,
		},
		{
			name: "With both options false",
			options: map[string]any{
				"ssh-no-root":     false,
				"ssh-no-password": false,
			},
			expected: false,
		},
		{
			name: "With skip-sudo true",
			options: map[string]any{
				"ssh-no-root": true,
				"skip-sudo":   true,
			},
			expected: false,
		},
		{
			name: "With all option true",
			options: map[string]any{
				"all": true,
			},
			expected: true,
		},
		{
			name: "With all option true but skip-sudo true",
			options: map[string]any{
				"all":       true,
				"skip-sudo": true,
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

	// Test cases - security feature doesn't have complex validation
	tests := []struct {
		name        string
		options     map[string]any
		expectError bool
	}{
		{
			name:        "Empty options",
			options:     map[string]any{},
			expectError: false,
		},
		{
			name: "With ssh-no-root true",
			options: map[string]any{
				"ssh-no-root": true,
			},
			expectError: false,
		},
		{
			name: "With ssh-no-password true",
			options: map[string]any{
				"ssh-no-password": true,
			},
			expectError: false,
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

	// Test cases for different option combinations
	testCases := []struct {
		name    string
		options map[string]any
	}{
		{
			name: "Basic security options",
			options: map[string]any{
				"ssh-no-root":     true,
				"ssh-no-password": true,
			},
		},
		{
			name: "With backup option",
			options: map[string]any{
				"ssh-no-root":     true,
				"ssh-no-password": true,
				"backup":          true,
			},
		},
		{
			name: "With all option",
			options: map[string]any{
				"all": true,
			},
		},
		{
			name: "With all and backup options",
			options: map[string]any{
				"all":    true,
				"backup": true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a context with the test options
			ctx := &features.ExecutionContext{
				Options:     tc.options,
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
					// but it shouldn't panic
					t.Logf("Execute returned error (expected in dry run): %v", err)
				}
			}()
		})
	}
}

func TestDisableRootLogin(t *testing.T) {
	// Test the disableRootLogin function
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	// Test cases
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: "# Added by INIQ (Previous setting: none)\nPermitRootLogin no",
		},
		{
			name:     "Content without PermitRootLogin",
			content:  "# SSH config\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Added by INIQ (Previous setting: none)\nPermitRootLogin no",
		},
		{
			name:     "Content with PermitRootLogin yes",
			content:  "# SSH config\nPermitRootLogin yes\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Modified by INIQ (Previous setting: PermitRootLogin yes)\nPermitRootLogin no",
		},
		{
			name:     "Content with commented PermitRootLogin",
			content:  "# SSH config\n#PermitRootLogin yes\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Modified by INIQ (Previous setting: #PermitRootLogin yes)\nPermitRootLogin no",
		},
		{
			name:     "Content with existing INIQ comment",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PermitRootLogin yes)\nPermitRootLogin no\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Modified by INIQ (Previous setting: PermitRootLogin no)\nPermitRootLogin no",
		},
		{
			name:     "Content with multiple existing INIQ comments",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PermitRootLogin prohibit-password)\n# Modified by INIQ (Previous setting: PermitRootLogin no)\nPermitRootLogin no\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Modified by INIQ (Previous setting: PermitRootLogin no)\nPermitRootLogin no",
		},
		{
			name:     "Content with duplicate INIQ comments",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PermitRootLogin prohibit-password)\n# Modified by INIQ (Previous setting: PermitRootLogin prohibit-password)\n# Modified by INIQ (Previous setting: PermitRootLogin prohibit-password)\nPermitRootLogin no\nPasswordAuthentication yes\n",
			expected: "# SSH config\nPasswordAuthentication yes\n# Modified by INIQ (Previous setting: PermitRootLogin no)\nPermitRootLogin no",
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := feature.disableRootLogin(tc.content)
			if result != tc.expected {
				t.Errorf("disableRootLogin() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestDisablePasswordAuth(t *testing.T) {
	// Test the disablePasswordAuth function
	osInfo, _ := osdetect.Detect()
	feature := New(osInfo)

	// Test cases
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: "# Added by INIQ (Previous setting: none)\nPasswordAuthentication no",
		},
		{
			name:     "Content without PasswordAuthentication",
			content:  "# SSH config\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Added by INIQ (Previous setting: none)\nPasswordAuthentication no",
		},
		{
			name:     "Content with PasswordAuthentication yes",
			content:  "# SSH config\nPasswordAuthentication yes\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\nPasswordAuthentication no",
		},
		{
			name:     "Content with commented PasswordAuthentication",
			content:  "# SSH config\n#PasswordAuthentication yes\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Modified by INIQ (Previous setting: #PasswordAuthentication yes)\nPasswordAuthentication no",
		},
		{
			name:     "Content with existing INIQ comment",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\nPasswordAuthentication no\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Modified by INIQ (Previous setting: PasswordAuthentication no)\nPasswordAuthentication no",
		},
		{
			name:     "Content with multiple existing INIQ comments",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\n# Modified by INIQ (Previous setting: PasswordAuthentication no)\nPasswordAuthentication no\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Modified by INIQ (Previous setting: PasswordAuthentication no)\nPasswordAuthentication no",
		},
		{
			name:     "Content with duplicate INIQ comments",
			content:  "# SSH config\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\n# Modified by INIQ (Previous setting: PasswordAuthentication yes)\nPasswordAuthentication no\nPermitRootLogin no\n",
			expected: "# SSH config\nPermitRootLogin no\n# Modified by INIQ (Previous setting: PasswordAuthentication no)\nPasswordAuthentication no",
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := feature.disablePasswordAuth(tc.content)
			if result != tc.expected {
				t.Errorf("disablePasswordAuth() = %q, expected %q", result, tc.expected)
			}
		})
	}
}
