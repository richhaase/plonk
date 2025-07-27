// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetOSPackageManagerSupport(t *testing.T) {
	support := getOSPackageManagerSupport()

	// Test based on current OS
	switch runtime.GOOS {
	case "darwin":
		// macOS should support all implemented managers
		expectedSupport := map[string]bool{
			"brew":  true,
			"npm":   true,
			"cargo": true,
			"gem":   true,
			"go":    true,
			"pip":   true,
		}
		for manager, expected := range expectedSupport {
			if support[manager] != expected {
				t.Errorf("macOS: expected %s support to be %v, got %v", manager, expected, support[manager])
			}
		}
	case "linux":
		// Linux should support all implemented managers
		for _, manager := range []string{"brew", "npm", "cargo", "gem", "go", "pip"} {
			if !support[manager] {
				t.Errorf("Linux: expected %s to be supported", manager)
			}
		}
	default:
		// Other OS should return empty map
		if len(support) != 0 {
			t.Errorf("Unsupported OS: expected empty map, got %v", support)
		}
	}
}

func TestGetManagerInstallSuggestion(t *testing.T) {
	// Test that all managers return useful suggestions
	managers := []string{"brew", "npm", "cargo", "gem", "go", "pip", "unknown"}

	for _, manager := range managers {
		t.Run(manager, func(t *testing.T) {
			suggestion := getManagerInstallSuggestion(manager)

			// Should always return something
			if suggestion == "" {
				t.Errorf("getManagerInstallSuggestion(%s) returned empty string", manager)
			}

			// Unknown managers should get the default message
			if manager == "unknown" {
				if !strings.Contains(suggestion, "plonk config edit") {
					t.Errorf("getManagerInstallSuggestion(unknown) should mention 'plonk config edit'")
				}
			}
		})
	}
}
