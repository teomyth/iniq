package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestConfigFileLoading tests the loading of configuration from files
func TestConfigFileLoading(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "iniq-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name           string
		configFile     string
		configContent  string
		expectedValues map[string]interface{}
	}{
		{
			name:          "Basic YAML config",
			configFile:    "config.yaml",
			configContent: "user: testuser\nssh-no-root: true\nssh-no-password: true\n",
			expectedValues: map[string]interface{}{
				"user":            "testuser",
				"ssh-no-root":     true,
				"ssh-no-password": true,
			},
		},
		{
			name:          "Config with array values",
			configFile:    "config-arrays.yaml",
			configContent: "user: testuser\nkeys:\n  - github:testuser\n  - gitlab:testuser\n",
			expectedValues: map[string]interface{}{
				"user": "testuser",
				"keys": []interface{}{"github:testuser", "gitlab:testuser"},
			},
		},
		{
			name:          "Config with string list format",
			configFile:    "config-string-list.yaml",
			configContent: "user: testuser\nkey: \"github:testuser;gitlab:testuser\"\n",
			expectedValues: map[string]interface{}{
				"user": "testuser",
				"key":  "github:testuser;gitlab:testuser",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			// Create config file
			configPath := filepath.Join(tempDir, tc.configFile)
			err := os.WriteFile(configPath, []byte(tc.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Initialize config
			err = InitConfig(configPath)
			assert.NoError(t, err, "InitConfig should not return an error")

			// Verify that values were loaded correctly
			for key, expectedValue := range tc.expectedValues {
				actualValue := viper.Get(key)
				assert.Equal(t, expectedValue, actualValue, "Config value for %s does not match", key)
			}
		})
	}
}

// TestConfigPrecedence tests that command line flags take precedence over config file values
func TestConfigPrecedence(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "iniq-config-precedence-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config file with some values
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
user: config-user
ssh-no-root: false
ssh-no-password: false
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a new viper instance for this test
	v := viper.New()

	// Set config file
	v.SetConfigFile(configPath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Verify config file values were loaded
	assert.Equal(t, "config-user", v.GetString("user"), "Config file value should be loaded")
	assert.Equal(t, false, v.GetBool("ssh-no-root"), "Config file value should be loaded")
	assert.Equal(t, false, v.GetBool("ssh-no-password"), "Config file value should be loaded")

	// Now set values that would come from command line flags
	v.Set("user", "cli-user")
	v.Set("ssh-no-root", true)
	// Don't set ssh-no-password to test that config file value is used

	// Verify that the new values take precedence
	assert.Equal(t, "cli-user", v.GetString("user"), "CLI flag should override config file")
	assert.Equal(t, true, v.GetBool("ssh-no-root"), "CLI flag should override config file")
	assert.Equal(t, false, v.GetBool("ssh-no-password"), "Config file value should be used when no CLI flag")
}

// TestConfigFileNotFound tests behavior when config file is not found
func TestConfigFileNotFound(t *testing.T) {
	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "iniq-notfound-test-home")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Reset viper
	viper.Reset()

	// Initialize config with non-existent file - this should not panic
	nonExistentPath := "/path/to/nonexistent/config.yaml"
	err = InitConfig(nonExistentPath)
	assert.NoError(t, err, "InitConfig should not return an error for non-existent file")

	// Verify that default values are used
	assert.Equal(t, "", viper.GetString("user"), "Default value should be used when config file not found")
}

// TestEnvironmentVariables tests loading configuration from environment variables
func TestEnvironmentVariables(t *testing.T) {
	// Save original environment and restore after test
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	// Create a new viper instance for this test
	v := viper.New()

	// Set environment prefix and enable automatic environment variable binding
	v.SetEnvPrefix("INIQ")
	v.AutomaticEnv()

	// Viper needs special handling for environment variables with dashes
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set environment variables
	os.Setenv("INIQ_USER", "env-user")
	os.Setenv("INIQ_SSH_NO_ROOT", "true")
	os.Setenv("INIQ_SSH_NO_PASSWORD", "true")

	// Verify that environment variables were loaded
	assert.Equal(t, "env-user", v.GetString("user"), "Environment variable should be loaded")
	assert.Equal(t, true, v.GetBool("ssh-no-root"), "Environment variable should be loaded")
	assert.Equal(t, true, v.GetBool("ssh-no-password"), "Environment variable should be loaded")
}

// TestLoadConfig tests loading configuration into a struct
func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "iniq-load-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
user: testuser
keys:
  - github:testuser
  - gitlab:testuser
sudo-nopasswd: true
skip-sudo: false
ssh-no-root: true
ssh-no-password: true
verbose: true
quiet: false
yes: true
dry-run: false
status: true
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Reset viper
	viper.Reset()

	// Initialize config
	err = InitConfig(configPath)
	assert.NoError(t, err, "InitConfig should not return an error")

	// Load config
	config, err := LoadConfig()
	assert.NoError(t, err, "LoadConfig should not return an error")
	assert.NotNil(t, config, "Config should not be nil")

	// Verify config values
	assert.Equal(t, "testuser", config.User, "User should match")
	assert.Equal(t, []string{"github:testuser", "gitlab:testuser"}, config.Keys, "Keys should match")
	assert.Equal(t, true, config.SudoNoPasswd, "SudoNoPasswd should match")
	assert.Equal(t, false, config.SkipSudo, "SkipSudo should match")
	assert.Equal(t, true, config.SSHNoRoot, "SSHNoRoot should match")
	assert.Equal(t, true, config.SSHNoPassword, "SSHNoPassword should match")
	assert.Equal(t, true, config.Verbose, "Verbose should match")
	assert.Equal(t, false, config.Quiet, "Quiet should match")
	assert.Equal(t, true, config.Yes, "Yes should match")
	assert.Equal(t, false, config.DryRun, "DryRun should match")
	assert.Equal(t, true, config.Status, "Status should match")
}
