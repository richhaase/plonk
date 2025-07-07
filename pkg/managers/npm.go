// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"path/filepath"
	"strings"
	
	"plonk/pkg/config"
)

// NpmManager manages NPM global packages.
type NpmManager struct {
	runner   *CommandRunner
	plonkDir string
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

// Search searches for packages via NPM.
func (n *NpmManager) Search(query string) ([]string, error) {
	output, err := n.runner.RunCommandWithOutput("search", query, "--parseable")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// NPM search output format: "name\tdescription\tauthor\tdate\tversion\tkeywords"
			parts := strings.Split(line, "\t")
			if len(parts) > 0 {
				result = append(result, parts[0])
			}
		}
	}

	return result, nil
}

// Info gets information about a package via NPM.
func (n *NpmManager) Info(packageName string) (string, error) {
	output, err := n.runner.RunCommandWithOutput("info", packageName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ListInstalledPackages returns detailed information about installed packages.
func (n *NpmManager) ListInstalledPackages() ([]PackageInfo, error) {
	// Get basic package list
	packages, err := n.ListInstalled()
	if err != nil {
		return nil, err
	}

	result := make([]PackageInfo, 0, len(packages))
	for _, pkg := range packages {
		info := PackageInfo{
			Name:    pkg,
			Status:  PackageInstalled,
			Manager: "npm",
		}

		// Try to get version information
		if versionInfo, err := n.getPackageVersion(pkg); err == nil {
			info.Version = versionInfo
		}

		result = append(result, info)
	}

	return result, nil
}

// getPackageVersion attempts to get version information for a package
func (n *NpmManager) getPackageVersion(packageName string) (string, error) {
	output, err := n.runner.RunCommandWithOutput("list", "-g", "--depth=0", packageName)
	if err != nil {
		return "", err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return "", nil
	}

	// Parse output like "/usr/local/lib/node_modules/typescript@4.8.4"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, packageName+"@") {
			parts := strings.Split(line, "@")
			if len(parts) >= 2 {
				return parts[len(parts)-1], nil
			}
		}
	}

	return "", nil
}

// SetConfigDir sets the plonk config directory
func (n *NpmManager) SetConfigDir(plonkDir string) {
	n.plonkDir = plonkDir
}

// getManagedPackages returns NPM packages listed in plonk.yaml
func (n *NpmManager) getManagedPackages() ([]string, error) {
	if n.plonkDir == "" {
		return []string{}, nil
	}
	
	cfg, err := config.LoadConfig(n.plonkDir)
	if err != nil {
		// If no config exists, return empty list
		return []string{}, nil
	}
	
	var packages []string
	// Extract package names from NPM packages
	for _, pkg := range cfg.NPM {
		if pkg.Package != "" {
			packages = append(packages, pkg.Package)
		} else {
			packages = append(packages, pkg.Name)
		}
	}
	return packages, nil
}

// ListManagedPackages returns packages that are managed by plonk
func (n *NpmManager) ListManagedPackages() ([]PackageInfo, error) {
	managedNames, err := n.getManagedPackages()
	if err != nil {
		return nil, err
	}
	
	var managed []PackageInfo
	
	for _, name := range managedNames {
		info := PackageInfo{
			Name:    name,
			Manager: "npm",
		}
		
		// Check if package is actually installed
		if n.IsInstalled(name) {
			info.Status = PackageInstalled
			// Try to get version
			if version, err := n.getPackageVersion(name); err == nil {
				info.Version = version
			}
		} else {
			info.Status = PackageAvailable
		}
		
		managed = append(managed, info)
	}
	
	return managed, nil
}

// ListUntrackedPackages returns installed packages not managed by plonk
func (n *NpmManager) ListUntrackedPackages() ([]PackageInfo, error) {
	allInstalled, err := n.ListInstalledPackages()
	if err != nil {
		return nil, err
	}
	
	managedNames, err := n.getManagedPackages()
	if err != nil {
		return nil, err
	}
	
	// Create a map of managed names for quick lookup
	managedMap := make(map[string]bool)
	for _, name := range managedNames {
		managedMap[name] = true
	}
	
	var untracked []PackageInfo
	for _, pkg := range allInstalled {
		if !managedMap[pkg.Name] {
			pkg.Status = PackageInstalled
			untracked = append(untracked, pkg)
		}
	}
	
	return untracked, nil
}

// ListMissingPackages returns packages in plonk.yaml that aren't installed
func (n *NpmManager) ListMissingPackages() ([]PackageInfo, error) {
	managedNames, err := n.getManagedPackages()
	if err != nil {
		return nil, err
	}
	
	var missing []PackageInfo
	
	for _, name := range managedNames {
		if !n.IsInstalled(name) {
			info := PackageInfo{
				Name:    name,
				Status:  PackageAvailable,
				Manager: "npm",
			}
			missing = append(missing, info)
		}
	}
	
	return missing, nil
}
