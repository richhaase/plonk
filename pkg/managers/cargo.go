package managers

import (
	"strings"
)

// CargoManager manages Rust Cargo packages
type CargoManager struct {
	runner *CommandRunner
}

// NewCargoManager creates a new Cargo manager
func NewCargoManager(executor CommandExecutor) *CargoManager {
	return &CargoManager{
		runner: NewCommandRunner(executor, "cargo"),
	}
}

// IsAvailable checks if Cargo is installed
func (c *CargoManager) IsAvailable() bool {
	err := c.runner.RunCommand("--version")
	return err == nil
}

// Install installs a package via Cargo
func (c *CargoManager) Install(packageName string) error {
	return c.runner.RunCommand("install", packageName)
}

// Update updates a specific package via Cargo
func (c *CargoManager) Update(packageName string) error {
	return c.runner.RunCommand("install", packageName, "--force")
}

// UpdateAll updates all installed packages via Cargo
func (c *CargoManager) UpdateAll() error {
	// Get list of installed packages first
	packages, err := c.ListInstalled()
	if err != nil {
		return err
	}
	
	// Update each package individually
	for _, pkg := range packages {
		err := c.Update(pkg)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// IsInstalled checks if a package is installed via Cargo
func (c *CargoManager) IsInstalled(packageName string) bool {
	packages, err := c.ListInstalled()
	if err != nil {
		return false
	}
	
	for _, pkg := range packages {
		if pkg == packageName {
			return true
		}
	}
	
	return false
}

// ListInstalled lists all installed Cargo packages
func (c *CargoManager) ListInstalled() ([]string, error) {
	output, err := c.runner.RunCommandWithOutput("install", "--list")
	if err != nil {
		return nil, err
	}
	
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}
	
	lines := strings.Split(output, "\n")
	
	// Parse cargo install --list output format:
	// package-name v1.2.3:
	//     binary1
	//     binary2
	// next-package v2.0.0:
	//     binary3
	result := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Lines ending with ':' contain package names and versions
		if strings.HasSuffix(line, ":") {
			// Extract package name (everything before the first space)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				packageName := parts[0]
				if packageName != "" {
					result = append(result, packageName)
				}
			}
		}
	}
	
	return result, nil
}