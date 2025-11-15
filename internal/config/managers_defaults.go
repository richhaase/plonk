// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// GetDefaultManagers returns built-in package manager configurations
func GetDefaultManagers() map[string]ManagerConfig {
	return map[string]ManagerConfig{
		"pipx": {
			Binary: "pipx",
			List: ListConfig{
				Command: []string{"pipx", "list", "--short"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"pipx", "install", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"pipx", "upgrade", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"pipx", "upgrade-all"},
			},
			Uninstall: CommandConfig{
				Command:          []string{"pipx", "uninstall", "{{.Package}}"},
				IdempotentErrors: []string{"not installed"},
			},
		},
		"cargo": {
			Binary:      "cargo",
			Description: "Cargo (Rust package manager)",
			InstallHint: "Install Rust from https://rustup.rs/",
			HelpURL:     "https://www.rust-lang.org/tools/install",
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
			Binary:      "npm",
			Description: "npm (Node.js package manager)",
			InstallHint: "Install Node.js from https://nodejs.org/ or use brew install node",
			HelpURL:     "https://nodejs.org/",
			List: ListConfig{
				Command:   []string{"npm", "list", "-g", "--depth=0", "--json"},
				Parse:     "json-map",
				JSONField: "dependencies",
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
			List: ListConfig{
				Command:   []string{"pnpm", "list", "-g", "--depth=0", "--json"},
				Parse:     "json",
				JSONField: "name",
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
		"conda": {
			Binary:      "conda",
			Description: "conda (Python/conda package manager)",
			InstallHint: "Install conda via Anaconda or Miniconda distributions",
			HelpURL:     "https://docs.conda.io/",
			List: ListConfig{
				Command:   []string{"conda", "list", "--json"},
				Parse:     "json",
				JSONField: "name",
			},
			Install: CommandConfig{
				Command:          []string{"conda", "install", "-y", "{{.Package}}"},
				IdempotentErrors: []string{"already installed"},
			},
			Upgrade: CommandConfig{
				Command:          []string{"conda", "update", "-y", "{{.Package}}"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			UpgradeAll: CommandConfig{
				Command:          []string{"conda", "update", "-y", "--all"},
				IdempotentErrors: []string{"already up-to-date", "up to date"},
			},
			Uninstall: CommandConfig{
				Command: []string{"conda", "remove", "-y", "{{.Package}}"},
			},
		},
		"uv": {
			Binary:      "uv",
			Description: "uv (Python package manager)",
			InstallHint: "Install UV from https://docs.astral.sh/uv/ or use brew install uv",
			HelpURL:     "https://docs.astral.sh/uv/",
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
