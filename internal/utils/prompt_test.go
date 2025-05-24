package utils

import (
	"testing"
)

func TestParseBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		wantErr  bool
	}{
		// True values
		{"yes", "yes", true, false},
		{"YES", "YES", true, false},
		{"enable", "enable", true, false},
		{"ENABLE", "ENABLE", true, false},
		{"true", "true", true, false},
		{"TRUE", "TRUE", true, false},
		{"1", "1", true, false},
		{"y", "y", true, false},
		{"Y", "Y", true, false},
		{"t", "t", true, false},
		{"T", "T", true, false},
		{"on", "on", true, false},
		{"ON", "ON", true, false},
		
		// False values
		{"no", "no", false, false},
		{"NO", "NO", false, false},
		{"disable", "disable", false, false},
		{"DISABLE", "DISABLE", false, false},
		{"false", "false", false, false},
		{"FALSE", "FALSE", false, false},
		{"0", "0", false, false},
		{"n", "n", false, false},
		{"N", "N", false, false},
		{"f", "f", false, false},
		{"F", "F", false, false},
		{"off", "off", false, false},
		{"OFF", "OFF", false, false},
		
		// Whitespace handling
		{"  yes  ", "  yes  ", true, false},
		{"  no  ", "  no  ", false, false},
		
		// Invalid values
		{"invalid", "invalid", false, true},
		{"maybe", "maybe", false, true},
		{"", "", false, true},
		{"2", "2", false, true},
		{"yep", "yep", false, true},
		{"nope", "nope", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBoolValue(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseBoolValue(%q) expected error, got nil", tt.input)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseBoolValue(%q) unexpected error: %v", tt.input, err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("ParseBoolValue(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
