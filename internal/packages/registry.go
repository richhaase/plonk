// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sort"

	"github.com/richhaase/plonk/internal/config"
)

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "brew"

// supportedManagers lists all available package managers
var supportedManagers = []string{"brew", "cargo", "gem", "go", "npm", "pnpm", "bun", "uv"}

func init() {
	// Register supported managers with the config validator
	config.SetValidManagers(supportedManagers)
}

// ManagerRegistry manages package manager creation
type ManagerRegistry struct{}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{}

// GetRegistry returns the shared manager registry instance
func GetRegistry() *ManagerRegistry {
	return defaultRegistry
}

// GetManager returns a package manager instance by name using the default executor
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	return r.GetManagerWithExecutor(name, nil)
}

// GetManagerWithExecutor returns a package manager instance with an injected executor.
func (r *ManagerRegistry) GetManagerWithExecutor(name string, exec CommandExecutor) (PackageManager, error) {
	if exec == nil {
		exec = defaultExecutor
	}

	switch name {
	case "brew":
		return NewBrewManager(exec), nil
	case "cargo":
		return NewCargoManager(exec), nil
	case "gem":
		return NewGemManager(exec), nil
	case "go":
		return NewGoManager(exec), nil
	case "npm":
		return NewNPMManager(exec), nil
	case "pnpm":
		return NewPNPMManager(exec), nil
	case "bun":
		return NewBunManager(exec), nil
	case "uv":
		return NewUVManager(exec), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}
}

// GetAllManagerNames returns all registered manager names sorted alphabetically
func (r *ManagerRegistry) GetAllManagerNames() []string {
	names := make([]string, len(supportedManagers))
	copy(names, supportedManagers)
	sort.Strings(names)
	return names
}

// HasManager checks if a manager is supported by the registry
func (r *ManagerRegistry) HasManager(name string) bool {
	for _, m := range supportedManagers {
		if m == name {
			return true
		}
	}
	return false
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}

// ManagerMetadata holds descriptive information about a package manager
type ManagerMetadata struct {
	Description string
	InstallHint string
	HelpURL     string
}

// managerMetadata contains metadata for all built-in managers
var managerMetadata = map[string]ManagerMetadata{
	"brew": {
		Description: "Homebrew (macOS/Linux package manager)",
		InstallHint: "Visit https://brew.sh for installation instructions (prerequisite)",
		HelpURL:     "https://brew.sh",
	},
	"cargo": {
		Description: "Cargo (Rust package manager)",
		InstallHint: "Install Rust from https://rustup.rs/",
		HelpURL:     "https://www.rust-lang.org/tools/install",
	},
	"gem": {
		Description: "gem (Ruby package manager)",
		InstallHint: "Install Ruby from https://ruby-lang.org/ or use brew install ruby",
		HelpURL:     "https://ruby-lang.org/",
	},
	"go": {
		Description: "Go (Go package manager)",
		InstallHint: "Install Go from https://go.dev/dl/ or use brew install go",
		HelpURL:     "https://go.dev/",
	},
	"npm": {
		Description: "npm (Node.js package manager)",
		InstallHint: "Install Node.js from https://nodejs.org/ or use brew install node",
		HelpURL:     "https://nodejs.org/",
	},
	"pnpm": {
		Description: "pnpm (Node.js package manager)",
		InstallHint: "Install pnpm from https://pnpm.io/ or use brew install pnpm",
		HelpURL:     "https://pnpm.io/",
	},
	"bun": {
		Description: "Bun (JavaScript runtime and package manager)",
		InstallHint: "Install Bun from https://bun.sh/ or use brew install bun",
		HelpURL:     "https://bun.sh/",
	},
	"uv": {
		Description: "uv (Python package manager)",
		InstallHint: "Install UV from https://docs.astral.sh/uv/ or use brew install uv",
		HelpURL:     "https://docs.astral.sh/uv/",
	},
}

// GetManagerMetadata returns metadata for a manager by name
func (r *ManagerRegistry) GetManagerMetadata(name string) (ManagerMetadata, bool) {
	meta, ok := managerMetadata[name]
	return meta, ok
}

// GetAllManagerMetadata returns metadata for all managers
func (r *ManagerRegistry) GetAllManagerMetadata() map[string]ManagerMetadata {
	result := make(map[string]ManagerMetadata, len(managerMetadata))
	for k, v := range managerMetadata {
		result[k] = v
	}
	return result
}
