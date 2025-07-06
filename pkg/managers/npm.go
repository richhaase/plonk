package managers

import (
	"path/filepath"
	"strings"
)

// NpmManager manages NPM global packages.
type NpmManager struct {
	runner *CommandRunner
}

// NewNpmManager creates a new NPM manager.
func NewNpmManager(executor CommandExecutor) *NpmManager {
	return &NpmManager{
		runner: NewCommandRunner(executor, "npm"),
	}
}

// IsAvailable checks if NPM is installed.
func (n *NpmManager) IsAvailable() bool {
	err := n.runner.RunCommand("--version")
	return err == nil
}

// Install installs a package globally via NPM.
func (n *NpmManager) Install(packageName string) error {
	return n.runner.RunCommand("install", "-g", packageName)
}

// Update updates a specific package globally via NPM.
func (n *NpmManager) Update(packageName string) error {
	return n.runner.RunCommand("update", "-g", packageName)
}

// UpdateAll updates all global packages via NPM.
func (n *NpmManager) UpdateAll() error {
	return n.runner.RunCommand("update", "-g")
}

// IsInstalled checks if a package is installed globally via NPM.
func (n *NpmManager) IsInstalled(packageName string) bool {
	err := n.runner.RunCommand("list", "-g", "--depth=0", packageName)
	return err == nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled() ([]string, error) {
	output, err := n.runner.RunCommandWithOutput("list", "-g", "--depth=0", "--parseable")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")

	// Parse NPM output format: "/usr/local/lib/node_modules/package-name".
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract package name from path.
		// Handle scoped packages like /usr/local/lib/node_modules/@vue/cli.
		packageName := filepath.Base(line)

		// Check if this is a scoped package (starts with @).
		parentDir := filepath.Base(filepath.Dir(line))
		if strings.HasPrefix(parentDir, "@") {
			packageName = parentDir + "/" + packageName
		}

		// Skip npm itself and empty names.
		if packageName != "npm" && packageName != "" && packageName != "." {
			result = append(result, packageName)
		}
	}

	return result, nil
}
