// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"strings"

	"plonk/pkg/config"
)

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct {
	runner   *CommandRunner
	plonkDir string
}

// NewHomebrewManager creates a new Homebrew manager.
func NewHomebrewManager(executor CommandExecutor) *HomebrewManager {
	return &HomebrewManager{
		runner: NewCommandRunner(executor, "brew"),
	}
}

// IsAvailable checks if Homebrew is installed.
func (h *HomebrewManager) IsAvailable() bool {
	err := h.runner.RunCommand("--version")
	return err == nil
}

// Install installs a package via Homebrew.
func (h *HomebrewManager) Install(packageName string) error {
	return h.runner.RunCommand("install", packageName)
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled() ([]string, error) {
	output, err := h.runner.RunCommandWithOutput("list")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	packages := strings.Split(output, "\n")

	// Clean up any empty strings.
	result := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// Update updates a specific package via Homebrew.
func (h *HomebrewManager) Update(packageName string) error {
	return h.runner.RunCommand("upgrade", packageName)
}

// UpdateAll updates all packages via Homebrew.
func (h *HomebrewManager) UpdateAll() error {
	return h.runner.RunCommand("upgrade")
}

// IsInstalled checks if a specific package is installed via Homebrew.
func (h *HomebrewManager) IsInstalled(packageName string) bool {
	err := h.runner.RunCommand("list", packageName)
	return err == nil
}

// InstallCask installs a cask via Homebrew.
func (h *HomebrewManager) InstallCask(caskName string) error {
	return h.runner.RunCommand("install", "--cask", caskName)
}

// Search searches for packages via Homebrew.
func (h *HomebrewManager) Search(query string) ([]string, error) {
	output, err := h.runner.RunCommandWithOutput("search", query)
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
		if line != "" && !strings.HasPrefix(line, "==>") {
			result = append(result, line)
		}
	}

	return result, nil
}

// Info gets information about a package via Homebrew.
func (h *HomebrewManager) Info(packageName string) (string, error) {
	output, err := h.runner.RunCommandWithOutput("info", packageName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ListInstalledPackages returns detailed information about installed packages.
func (h *HomebrewManager) ListInstalledPackages() ([]PackageInfo, error) {
	// Get basic package list
	packages, err := h.ListInstalled()
	if err != nil {
		return nil, err
	}

	result := make([]PackageInfo, 0, len(packages))
	for _, pkg := range packages {
		info := PackageInfo{
			Name:    pkg,
			Status:  PackageInstalled,
			Manager: "homebrew",
		}

		// Try to get version information
		if versionInfo, err := h.getPackageVersion(pkg); err == nil {
			info.Version = versionInfo
		}

		result = append(result, info)
	}

	return result, nil
}

// getPackageVersion attempts to get version information for a package
func (h *HomebrewManager) getPackageVersion(packageName string) (string, error) {
	output, err := h.runner.RunCommandWithOutput("list", "--versions", packageName)
	if err != nil {
		return "", err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return "", nil
	}

	// Parse output like "git 2.42.0" to extract version
	parts := strings.Fields(output)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return "", nil
}

// SetConfigDir sets the plonk config directory
func (h *HomebrewManager) SetConfigDir(plonkDir string) {
	h.plonkDir = plonkDir
}

// getManagedPackages returns homebrew packages listed in plonk.yaml
func (h *HomebrewManager) getManagedPackages() ([]string, error) {
	if h.plonkDir == "" {
		return []string{}, nil
	}

	cfg, err := config.LoadConfig(h.plonkDir)
	if err != nil {
		// If no config exists, return empty list
		return []string{}, nil
	}

	var packages []string
	// Extract package names from both brews and casks
	for _, brew := range cfg.Homebrew.Brews {
		packages = append(packages, brew.Name)
	}
	for _, cask := range cfg.Homebrew.Casks {
		packages = append(packages, cask.Name)
	}
	return packages, nil
}

// ListManagedPackages returns packages that are managed by plonk
func (h *HomebrewManager) ListManagedPackages() ([]PackageInfo, error) {
	managedNames, err := h.getManagedPackages()
	if err != nil {
		return nil, err
	}

	var managed []PackageInfo

	for _, name := range managedNames {
		info := PackageInfo{
			Name:    name,
			Manager: "homebrew",
		}

		// Check if package is actually installed
		if h.IsInstalled(name) {
			info.Status = PackageInstalled
			// Try to get version
			if version, err := h.getPackageVersion(name); err == nil {
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
func (h *HomebrewManager) ListUntrackedPackages() ([]PackageInfo, error) {
	allInstalled, err := h.ListInstalledPackages()
	if err != nil {
		return nil, err
	}

	managedNames, err := h.getManagedPackages()
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
			pkg.Status = PackageInstalled // Should already be set, but ensure consistency
			untracked = append(untracked, pkg)
		}
	}

	return untracked, nil
}

// ListMissingPackages returns packages in plonk.yaml that aren't installed
func (h *HomebrewManager) ListMissingPackages() ([]PackageInfo, error) {
	managedNames, err := h.getManagedPackages()
	if err != nil {
		return nil, err
	}

	var missing []PackageInfo

	for _, name := range managedNames {
		if !h.IsInstalled(name) {
			info := PackageInfo{
				Name:    name,
				Status:  PackageAvailable,
				Manager: "homebrew",
			}
			missing = append(missing, info)
		}
	}

	return missing, nil
}
