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
			Binary: "cargo",
			List: ListConfig{
				Command: []string{"cargo", "install", "--list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command:          []string{"cargo", "install", "{{.Package}}"},
				IdempotentErrors: []string{"already exists"},
			},
			Upgrade: CommandConfig{
				Command: []string{"cargo", "install", "--force", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{},
			Uninstall: CommandConfig{
				Command: []string{"cargo", "uninstall", "{{.Package}}"},
			},
		},
		"gem": {
			Binary: "gem",
			List: ListConfig{
				Command: []string{"gem", "list", "--local", "--no-versions"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command: []string{"gem", "install", "{{.Package}}", "--user-install"},
			},
			Upgrade: CommandConfig{
				Command: []string{"gem", "update", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"gem", "update"},
			},
			Uninstall: CommandConfig{
				Command: []string{"gem", "uninstall", "{{.Package}}", "-x"},
			},
		},
		"brew": {
			Binary: "brew",
			List: ListConfig{
				Command: []string{"brew", "list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command: []string{"brew", "install", "{{.Package}}"},
			},
			Upgrade: CommandConfig{
				Command: []string{"brew", "upgrade", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"brew", "upgrade"},
			},
			Uninstall: CommandConfig{
				Command: []string{"brew", "uninstall", "{{.Package}}"},
			},
		},
		"npm": {
			Binary: "npm",
			List: ListConfig{
				Command: []string{"npm", "list", "-g", "--depth=0", "--parseable"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command: []string{"npm", "install", "-g", "{{.Package}}"},
			},
			Upgrade: CommandConfig{
				Command: []string{"npm", "update", "-g", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"npm", "update", "-g"},
			},
			Uninstall: CommandConfig{
				Command: []string{"npm", "uninstall", "-g", "{{.Package}}"},
			},
		},
		"pnpm": {
			Binary: "pnpm",
			List: ListConfig{
				Command: []string{"pnpm", "list", "-g", "--depth=0", "--parseable"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command: []string{"pnpm", "add", "-g", "{{.Package}}"},
			},
			Upgrade: CommandConfig{
				Command: []string{"pnpm", "update", "-g", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"pnpm", "update", "-g"},
			},
			Uninstall: CommandConfig{
				Command: []string{"pnpm", "remove", "-g", "{{.Package}}"},
			},
		},
		"conda": {
			Binary: "conda",
			List: ListConfig{
				Command:   []string{"conda", "list", "--json"},
				Parse:     "json",
				JSONField: "name",
			},
			Install: CommandConfig{
				Command: []string{"conda", "install", "-y", "{{.Package}}"},
			},
			Upgrade: CommandConfig{
				Command: []string{"conda", "update", "-y", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"conda", "update", "-y", "--all"},
			},
			Uninstall: CommandConfig{
				Command: []string{"conda", "remove", "-y", "{{.Package}}"},
			},
		},
		"uv": {
			Binary: "uv",
			List: ListConfig{
				Command: []string{"uv", "tool", "list"},
				Parse:   "lines",
			},
			Install: CommandConfig{
				Command: []string{"uv", "tool", "install", "{{.Package}}"},
			},
			Upgrade: CommandConfig{
				Command: []string{"uv", "tool", "upgrade", "{{.Package}}"},
			},
			UpgradeAll: CommandConfig{
				Command: []string{"uv", "tool", "upgrade", "--all"},
			},
			Uninstall: CommandConfig{
				Command: []string{"uv", "tool", "uninstall", "{{.Package}}"},
			},
		},
	}
}
