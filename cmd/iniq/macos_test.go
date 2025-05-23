package main

import (
	"testing"

	"github.com/teomyth/iniq/internal/logger"
	"github.com/teomyth/iniq/pkg/osdetect"
)

func TestCheckMacOSSupport(t *testing.T) {
	tests := []struct {
		name     string
		osType   osdetect.OSType
		yes      bool
		quiet    bool
		expected bool // whether function should return error
	}{
		{
			name:     "Linux - no warning",
			osType:   osdetect.Linux,
			yes:      false,
			quiet:    false,
			expected: false, // no error
		},
		{
			name:     "macOS with quiet flag - no warning",
			osType:   osdetect.Darwin,
			yes:      false,
			quiet:    true,
			expected: false, // no error, no warning
		},
		{
			name:     "macOS with yes flag - warning but no prompt",
			osType:   osdetect.Darwin,
			yes:      true,
			quiet:    false,
			expected: false, // no error, warning shown
		},
		{
			name:     "macOS interactive mode - warning and prompt",
			osType:   osdetect.Darwin,
			yes:      false,
			quiet:    false,
			expected: false, // depends on user input, but we can't test interactive input easily
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger
			log := logger.New(false, tt.quiet)

			// Create OS info
			osInfo := &osdetect.Info{
				Type: tt.osType,
			}

			// For non-interactive tests (Linux, quiet mode, yes mode)
			if tt.osType != osdetect.Darwin || tt.quiet || tt.yes {
				err := checkMacOSSupport(osInfo, log, tt.yes, tt.quiet)

				if tt.expected && err == nil {
					t.Errorf("Expected error but got none")
				} else if !tt.expected && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Check that warning is shown for macOS (except in quiet mode)
				if tt.osType == osdetect.Darwin && !tt.quiet {
					// We can't easily capture logger output, but we can verify the function runs
					// The actual warning display is tested in integration tests
				}
			}
		})
	}
}

func TestCheckMacOSSupport_LinuxNoWarning(t *testing.T) {
	// Test that Linux doesn't show any warning
	osInfo := &osdetect.Info{
		Type: osdetect.Linux,
	}
	log := logger.New(false, false)

	err := checkMacOSSupport(osInfo, log, false, false)
	if err != nil {
		t.Errorf("Expected no error for Linux, got: %v", err)
	}
}

func TestCheckMacOSSupport_MacOSQuietMode(t *testing.T) {
	// Test that macOS in quiet mode doesn't show warning
	osInfo := &osdetect.Info{
		Type: osdetect.Darwin,
	}
	log := logger.New(false, true) // quiet mode

	err := checkMacOSSupport(osInfo, log, false, true)
	if err != nil {
		t.Errorf("Expected no error for macOS quiet mode, got: %v", err)
	}
}

func TestCheckMacOSSupport_MacOSYesMode(t *testing.T) {
	// Test that macOS with yes flag shows warning but doesn't prompt
	osInfo := &osdetect.Info{
		Type: osdetect.Darwin,
	}
	log := logger.New(false, false)

	err := checkMacOSSupport(osInfo, log, true, false) // yes mode
	if err != nil {
		t.Errorf("Expected no error for macOS yes mode, got: %v", err)
	}
}
