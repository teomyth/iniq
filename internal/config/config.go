package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// User management
	User     string `mapstructure:"user"`
	Password bool   `mapstructure:"password"`

	// SSH key management
	Keys []string `mapstructure:"keys"`

	// Sudo configuration
	SudoNoPasswd bool `mapstructure:"sudo-nopasswd"`
	SkipSudo     bool `mapstructure:"skip-sudo"`

	// SSH security
	SSHNoRoot     bool `mapstructure:"ssh-no-root"`
	SSHNoPassword bool `mapstructure:"ssh-no-password"`
	All           bool `mapstructure:"all"`
	Backup        bool `mapstructure:"backup"`

	// General options
	Verbose bool `mapstructure:"verbose"`
	Quiet   bool `mapstructure:"quiet"`
	Yes     bool `mapstructure:"yes"`
	DryRun  bool `mapstructure:"dry-run"`
	Status  bool `mapstructure:"status"`
}

// InitConfig initializes the configuration system
func InitConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}

		// Search config in home directory with name ".iniq" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".iniq")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("INIQ")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return nil
}

// LoadConfig loads the configuration from viper into a Config struct
func LoadConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}
	return &config, nil
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, filePath string) error {
	// If no path is provided, use default
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}
		filePath = filepath.Join(home, ".iniq.yaml")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	// Set config values in viper
	viper.Set("user", config.User)
	viper.Set("password", config.Password)
	viper.Set("keys", config.Keys)
	viper.Set("sudo-nopasswd", config.SudoNoPasswd)
	viper.Set("skip-sudo", config.SkipSudo)
	viper.Set("ssh-no-root", config.SSHNoRoot)
	viper.Set("ssh-no-password", config.SSHNoPassword)
	viper.Set("all", config.All)
	viper.Set("backup", config.Backup)
	viper.Set("verbose", config.Verbose)
	viper.Set("quiet", config.Quiet)
	viper.Set("yes", config.Yes)
	viper.Set("dry-run", config.DryRun)
	viper.Set("status", config.Status)

	// Write config to file
	if err := viper.WriteConfigAs(filePath); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns a configuration with default values
func GetDefaultConfig() *Config {
	return &Config{
		User:          "",
		Password:      false,
		Keys:          []string{},
		SudoNoPasswd:  true,
		SkipSudo:      false,
		SSHNoRoot:     true,
		SSHNoPassword: true,
		All:           false,
		Backup:        false,
		Verbose:       false,
		Quiet:         false,
		Yes:           false,
		DryRun:        false,
		Status:        false,
	}
}

// ShowConfig displays the current configuration
func ShowConfig() {
	fmt.Println("INIQ Configuration")
	fmt.Println("------------------")

	// Show configuration file path
	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("No config file in use")
		fmt.Printf("Default locations: $HOME/.iniq.yaml, $HOME/.iniq/config.yaml\n\n")
	}

	// Show all settings
	settings := viper.AllSettings()
	if len(settings) == 0 {
		fmt.Println("No configuration settings found")
		return
	}

	fmt.Println("Current settings:")
	for k, v := range settings {
		fmt.Printf("  %s: %v\n", k, v)
	}
}
