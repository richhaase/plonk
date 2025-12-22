// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// GetDefaultManagers returns built-in package manager configurations.
// Note: This is used for metadata lookup (description, install hints) by the doctor command.
// The actual package management is handled by hardcoded managers in packages.Registry.
func GetDefaultManagers() map[string]ManagerConfig {
	return map[string]ManagerConfig{
		"cargo": {
			Binary:      "cargo",
			Description: "Cargo (Rust package manager)",
			InstallHint: "Install Rust from https://rustup.rs/",
			HelpURL:     "https://www.rust-lang.org/tools/install",
			Available: CommandConfig{
				Command: []string{"cargo", "--version"},
			},
			List: ListConfig{
				Command: []string{"cargo", "install", "--list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"cargo", "install", "{{.Package}}"},
				IdempotentErrors: []string{"already exists", "already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"cargo", "install", "--force", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{},
			Uninstall: CommandConfig{
				Command: []string{"cargo", "uninstall", "{{.Package}}"},
			},
		},
		"gem": {
			Binary:      "gem",
			Description: "gem (Ruby package manager)",
			InstallHint: "Install Ruby from https://ruby-lang.org/ or use brew install ruby",
			HelpURL:     "https://ruby-lang.org/",
			Available: CommandConfig{
				Command: []string{"gem", "--version"},
			},
			List: ListConfig{
				Command: []string{"gem", "list", "--local", "--no-versions"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"gem", "install", "{{.Package}}", "--user-install"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"gem", "update", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"gem", "update"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"gem", "uninstall", "{{.Package}}", "-x"},
			},
		},
		"brew": {
			Binary:      "brew",
			Description: "Homebrew (macOS/Linux package manager)",
			InstallHint: "Visit https://brew.sh for installation instructions (prerequisite)",
			HelpURL:     "https://brew.sh",
			Available: CommandConfig{
				Command: []string{"brew", "--version"},
			},
			List: ListConfig{
				Command: []string{"brew", "list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"brew", "install", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"brew", "upgrade", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"brew", "upgrade"},
				IdempotentErrors: []string{"already up-to-date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"brew", "uninstall", "{{.Package}}"},
			},
		},
		"npm": {
			Binary:        "npm",
			Description:   "npm (Node.js package manager)",
			InstallHint:   "Install Node.js from https://nodejs.org/ or use brew install node",
			HelpURL:       "https://nodejs.org/",
			UpgradeTarget: "full_name_preferred",
			Available: CommandConfig{
				Command: []string{"npm", "--version"},
			},
			List: ListConfig{
				Command:       []string{"npm", "list", "-g", "--depth=0", "--json"},
				Parse:         "jsonpath",
				ParseStrategy: "jsonpath",
				KeysFrom:      "$.dependencies",
			},
			Install: CommandConfig{
				Command:          []string{"npm", "install", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"npm", "update", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"npm", "update", "-g"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"npm", "uninstall", "-g", "{{.Package}}"},
			},
			MetadataExtractors: map[string]MetadataExtractorConfig{
				"scope": {
					Pattern: "^(@[^/]+)/",
					Group:   1,
					Source:  "name",
				},
				"full_name": {
					Source: "name",
				},
			},
		},
		"pnpm": {
			Binary:      "pnpm",
			Description: "pnpm (Node.js package manager)",
			InstallHint: "Install pnpm from https://pnpm.io/ or use brew install pnpm",
			HelpURL:     "https://pnpm.io/",
			Available: CommandConfig{
				Command: []string{"pnpm", "--version"},
			},
			List: ListConfig{
				Command:       []string{"pnpm", "list", "-g", "--depth=0", "--json"},
				Parse:         "jsonpath",
				ParseStrategy: "jsonpath",
				KeysFrom:      "$[*].dependencies",
			},
			Install: CommandConfig{
				Command:          []string{"pnpm", "add", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"pnpm", "update", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"pnpm", "update", "-g"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"pnpm", "remove", "-g", "{{.Package}}"},
			},
		},
		"go": {
			Binary:      "go",
			Description: "Go (Go package manager)",
			InstallHint: "Install Go from https://go.dev/dl/ or use brew install go",
			HelpURL:     "https://go.dev/",
			Available: CommandConfig{
				Command: []string{"go", "version"},
			},
			List: ListConfig{
				Command: []string{"ls", "$GOBIN"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"go", "install", "{{.Package}}@latest"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"go", "install", "{{.Package}}@latest"},
				IdempotentErrors: []string{"already up-to-date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"rm", "-f", "$GOBIN/{{.Package}}"},
			},
		},
		"bun": {
			Binary:      "bun",
			Description: "Bun (JavaScript runtime and package manager)",
			InstallHint: "Install Bun from https://bun.sh/ or use brew install bun",
			HelpURL:     "https://bun.sh/",
			Available: CommandConfig{
				Command: []string{"bun", "--version"},
			},
			List: ListConfig{
				Command: []string{"bun", "pm", "ls", "-g"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"bun", "add", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"bun", "update", "-g", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"bun", "remove", "-g", "{{.Package}}"},
			},
		},
		"uv": {
			Binary:      "uv",
			Description: "uv (Python package manager)",
			InstallHint: "Install UV from https://docs.astral.sh/uv/ or use brew install uv",
			HelpURL:     "https://docs.astral.sh/uv/",
			Available: CommandConfig{
				Command: []string{"uv", "--version"},
			},
			List: ListConfig{
				Command: []string{"uv", "tool", "list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"uv", "tool", "install", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"uv", "tool", "upgrade", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"uv", "tool", "upgrade", "--all"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"uv", "tool", "uninstall", "{{.Package}}"},
			},
		},
	}
}
