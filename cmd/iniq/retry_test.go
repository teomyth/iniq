package main

import (
	"errors"
	"testing"
)

func TestIsNonRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "conflicting password options",
			err:      errors.New("cannot specify both --password and --no-pass options"),
			expected: true,
		},
		{
			name:     "password input required",
			err:      errors.New("creating user 'test' requires password input"),
			expected: true,
		},
		{
			name:     "invalid flag",
			err:      errors.New("unknown flag: --invalid"),
			expected: true,
		},
		{
			name:     "unsupported OS",
			err:      errors.New("unsupported OS: windows"),
			expected: true,
		},
		{
			name:     "validation failed",
			err:      errors.New("validation failed: invalid username"),
			expected: true,
		},
		{
			name:     "network error should be retryable",
			err:      errors.New("network connection failed"),
			expected: false,
		},
		{
			name:     "timeout error should be retryable",
			err:      errors.New("operation timeout"),
			expected: false,
		},
		{
			name:     "permission error should be retryable",
			err:      errors.New("permission denied"),
			expected: false,
		},
		{
			name:     "generic system error should be retryable",
			err:      errors.New("failed to execute command"),
			expected: false,
		},
		{
			name:     "parse error should not be retryable",
			err:      errors.New("parse error: invalid syntax"),
			expected: true,
		},
		{
			name:     "syntax error should not be retryable",
			err:      errors.New("syntax error in configuration"),
			expected: true,
		},
		{
			name:     "stdin already set error should not be retryable",
			err:      errors.New("failed to create stdin pipe: exec: Stdin already set"),
			expected: true,
		},
		{
			name:     "process already started error should not be retryable",
			err:      errors.New("exec: process already started"),
			expected: true,
		},
		{
			name:     "command not found error should not be retryable",
			err:      errors.New("exec: command not found"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNonRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isNonRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
