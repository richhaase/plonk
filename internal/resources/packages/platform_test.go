// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"runtime"
	"testing"
)

func TestIsPackageManagerSupportedOnPlatform(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected bool
		reason   string
	}{
		// Platform-specific package managers
		{
			name:     "apt on current platform",
			manager:  "apt",
			expected: runtime.GOOS == "linux" && GetLinuxDistro() == DistroDebian,
			reason:   "APT should only be supported on Debian-based Linux",
		},
		{
			name:     "brew on current platform",
			manager:  "brew",
			expected: runtime.GOOS == "darwin" || runtime.GOOS == "linux",
			reason:   "Homebrew should be supported on macOS and Linux",
		},

		// Language-specific package managers (always supported)
		{
			name:     "cargo is cross-platform",
			manager:  "cargo",
			expected: true,
			reason:   "Cargo should be supported on all platforms",
		},
		{
			name:     "npm is cross-platform",
			manager:  "npm",
			expected: true,
			reason:   "NPM should be supported on all platforms",
		},
		{
			name:     "pip is cross-platform",
			manager:  "pip",
			expected: true,
			reason:   "Pip should be supported on all platforms",
		},
		{
			name:     "gem is cross-platform",
			manager:  "gem",
			expected: true,
			reason:   "Gem should be supported on all platforms",
		},
		{
			name:     "go is cross-platform",
			manager:  "go",
			expected: true,
			reason:   "Go should be supported on all platforms",
		},

		// Unknown manager
		{
			name:     "unknown manager",
			manager:  "unknown",
			expected: false,
			reason:   "Unknown managers should not be supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPackageManagerSupportedOnPlatform(tt.manager)
			if got != tt.expected {
				t.Errorf("IsPackageManagerSupportedOnPlatform(%q) = %v, want %v (%s)",
					tt.manager, got, tt.expected, tt.reason)
			}
		})
	}
}

func TestGetNativePackageManager(t *testing.T) {
	native := GetNativePackageManager()

	switch runtime.GOOS {
	case "darwin":
		if native != "brew" {
			t.Errorf("Expected native package manager 'brew' on macOS, got %q", native)
		}
	case "linux":
		// On Linux, it depends on the distribution
		distro := GetLinuxDistro()
		switch distro {
		case DistroDebian:
			if native != "apt" {
				t.Errorf("Expected native package manager 'apt' on Debian-based Linux, got %q", native)
			}
		case DistroUnknown:
			if native != "brew" {
				t.Errorf("Expected fallback to 'brew' on unknown Linux, got %q", native)
			}
		}
		// Note: Other distros would have their own native managers once implemented
	default:
		if native != "" {
			t.Errorf("Expected empty native package manager on unsupported platform, got %q", native)
		}
	}
}
