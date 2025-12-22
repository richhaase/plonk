// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

// managerInstallHints provides installation guidance for each supported manager
var managerInstallHints = map[string]string{
	"brew":  "Visit https://brew.sh for installation instructions",
	"cargo": "Install Rust from https://rustup.rs/",
	"go":    "Install Go from https://go.dev/dl/",
	"npm":   "Install Node.js from https://nodejs.org/",
	"pnpm":  "Install pnpm from https://pnpm.io/ or run: npm install -g pnpm",
	"bun":   "Install bun from https://bun.sh/",
	"uv":    "Install uv from https://docs.astral.sh/uv/ or run: brew install uv",
}

// managerInstallHint returns the install hint for a manager.
func managerInstallHint(manager string) string {
	if hint, ok := managerInstallHints[manager]; ok {
		return hint
	}
	return "check installation instructions for " + manager
}
