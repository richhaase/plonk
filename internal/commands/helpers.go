// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

// getManagerInstallSuggestion returns installation instructions for different package managers
func getManagerInstallSuggestion(manager string) string {
	switch manager {
	case "homebrew":
		return "Install Homebrew: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
	case "npm":
		return "Install Node.js and npm: visit https://nodejs.org or use 'brew install node'"
	case "cargo":
		return "Install Rust and Cargo: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
	default:
		return "Install the required package manager or change the default manager with 'plonk config edit'"
	}
}
