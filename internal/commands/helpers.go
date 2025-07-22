// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"runtime"
)

// getOSPackageManagerSupport returns which package managers are supported on the current OS
func getOSPackageManagerSupport() map[string]bool {
	switch runtime.GOOS {
	case "darwin":
		return map[string]bool{
			"homebrew": true,
			"npm":      true,
			"cargo":    true,
			"gem":      true,
			"go":       true,
			"pip":      true,
		}
	case "linux":
		return map[string]bool{
			"homebrew": true, // Supported on Linux
			"npm":      true,
			"cargo":    true,
			"gem":      true,
			"go":       true,
			"pip":      true,
			// apt would go here when implemented
		}
	default:
		// Unsupported OS - return empty map
		// Windows and other OSes are not currently supported by plonk
		return map[string]bool{}
	}
}

// getManagerInstallSuggestion returns installation instructions for different package managers
func getManagerInstallSuggestion(manager string) string {
	osSpecific := getOSSpecificInstallCommand(manager, runtime.GOOS)
	if osSpecific != "" {
		return osSpecific
	}
	return "Install the required package manager or change the default manager with 'plonk config edit'"
}

// getOSSpecificInstallCommand returns OS-specific installation instructions for package managers
func getOSSpecificInstallCommand(manager, os string) string {
	// First check if the manager is supported on this OS
	support := getOSPackageManagerSupport()
	if supported, exists := support[manager]; exists && !supported {
		return fmt.Sprintf("%s is not supported on %s", manager, os)
	}

	switch manager {
	case "homebrew":
		switch os {
		case "darwin":
			return "Install Homebrew: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
		case "linux":
			return "Install Homebrew on Linux: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
		}
	case "npm":
		switch os {
		case "darwin":
			return "Install Node.js and npm: brew install node OR download from https://nodejs.org"
		case "linux":
			return "Install Node.js and npm: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"
		}
	case "cargo":
		return "Install Rust and Cargo: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
	case "gem":
		switch os {
		case "darwin":
			return "Ruby comes pre-installed on macOS. For a newer version: brew install ruby"
		case "linux":
			return "Install Ruby: sudo apt-get install ruby-full OR use rbenv/rvm"
		}
	case "go":
		switch os {
		case "darwin":
			return "Install Go: brew install go OR download from https://go.dev/dl/"
		case "linux":
			return "Install Go: Download from https://go.dev/dl/ OR sudo snap install go --classic"
		}
	case "pip":
		switch os {
		case "darwin":
			return "Install Python and pip: brew install python3 OR download from https://python.org"
		case "linux":
			return "Install Python and pip: sudo apt-get install python3-pip"
		}
	}

	return ""
}
