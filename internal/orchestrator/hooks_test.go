// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/config"
)

func TestNewHookRunner(t *testing.T) {
	runner := NewHookRunner()

	if runner == nil {
		t.Fatal("Expected hook runner to be created")
	}

	expectedTimeout := 10 * time.Minute
	if runner.defaultTimeout != expectedTimeout {
		t.Errorf("Expected default timeout %v, got %v", expectedTimeout, runner.defaultTimeout)
	}
}

func TestHookRunner_ParseTimeout(t *testing.T) {
	runner := NewHookRunner()

	testCases := []struct {
		name            string
		timeoutStr      string
		expectedDefault bool
	}{
		{"Valid duration", "30s", false},
		{"Invalid duration", "invalid", true},
		{"Empty duration", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			timeout := runner.defaultTimeout
			if tc.timeoutStr != "" {
				if d, err := time.ParseDuration(tc.timeoutStr); err == nil {
					timeout = d
				}
			}

			if tc.expectedDefault {
				if timeout != runner.defaultTimeout {
					t.Errorf("Expected default timeout %v, got %v", runner.defaultTimeout, timeout)
				}
			} else if tc.timeoutStr == "30s" {
				expected := 30 * time.Second
				if timeout != expected {
					t.Errorf("Expected parsed timeout %v, got %v", expected, timeout)
				}
			}
		})
	}
}

func TestHookRunner_TimeoutParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"30s", 30 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			duration, err := time.ParseDuration(tc.input)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', got %v", tc.input, err)
				}
				if duration != tc.expected {
					t.Errorf("Expected duration %v for input '%s', got %v", tc.expected, tc.input, duration)
				}
			}
		})
	}
}

func TestHookConfiguration(t *testing.T) {
	testCases := []struct {
		name  string
		hook  config.Hook
		valid bool
	}{
		{
			name: "Valid hook with command only",
			hook: config.Hook{
				Command: "echo 'test'",
			},
			valid: true,
		},
		{
			name: "Valid hook with timeout",
			hook: config.Hook{
				Command: "echo 'test'",
				Timeout: "30s",
			},
			valid: true,
		},
		{
			name: "Valid hook with continue on error",
			hook: config.Hook{
				Command:         "exit 1",
				ContinueOnError: true,
			},
			valid: true,
		},
		{
			name: "Invalid hook with empty command",
			hook: config.Hook{
				Command: "",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.hook.Command != ""

			if isValid != tc.valid {
				t.Errorf("Expected validity %v for hook %+v, got %v", tc.valid, tc.hook, isValid)
			}
		})
	}
}
