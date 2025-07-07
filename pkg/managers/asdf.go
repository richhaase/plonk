// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"plonk/pkg/config"
)

// AsdfManager manages ASDF tools and versions.
type AsdfManager struct {
	runner   *CommandRunner
	plonkDir string
}

// NewAsdfManager creates a new ASDF manager.
func NewAsdfManager(executor CommandExecutor) *AsdfManager {
	return &AsdfManager{
		runner: NewCommandRunner(executor, "asdf"),
	}
}

// IsAvailable checks if ASDF is installed.
func (a *AsdfManager) IsAvailable() bool {
	err := a.runner.RunCommand("version")
	return err == nil
}

// Install installs a tool/version via ASDF
// packageName should be in format "tool version" like "nodejs 20.0.0".
func (a *AsdfManager) Install(packageName string) error {
	parts := strings.Fields(packageName)
	if len(parts) < 2 {
		return a.runner.RunCommand("install", packageName)
	}

	// asdf install <tool> <version>.
	args := append([]string{"install"}, parts...)
	return a.runner.RunCommand(args...)
}

// ListInstalled lists all installed ASDF plugins.
func (a *AsdfManager) ListInstalled() ([]string, error) {
	output, err := a.runner.RunCommandWithOutput("plugin", "list")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	plugins := strings.Split(output, "\n")

	// Clean up any empty strings.
	result := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		if trimmed := strings.TrimSpace(plugin); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// ListGlobalTools returns a list of globally configured ASDF tools and versions.
// Reads from ~/.tool-versions file and returns tools in "tool version" format.
func (a *AsdfManager) ListGlobalTools() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	toolVersionsPath := filepath.Join(homeDir, ".tool-versions")
	file, err := os.Open(toolVersionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No .tool-versions file means no global tools
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Each line should be "tool version"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			toolAndVersion := parts[0] + " " + parts[1]
			result = append(result, toolAndVersion)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Update updates a tool to the latest version via ASDF.
func (a *AsdfManager) Update(toolName string) error {
	// Get the latest version available.
	output, err := a.runner.RunCommandWithOutput("latest", toolName)
	if err != nil {
		return err
	}

	latestVersion := strings.TrimSpace(output)
	if latestVersion == "" {
		return nil // No version available.
	}

	// Install the latest version.
	return a.runner.RunCommand("install", toolName, latestVersion)
}

// IsInstalled checks if a tool is installed via ASDF (has any versions).
func (a *AsdfManager) IsInstalled(toolName string) bool {
	err := a.runner.RunCommand("list", toolName)
	return err == nil
}

// GetInstalledVersions returns all installed versions for a tool.
func (a *AsdfManager) GetInstalledVersions(toolName string) ([]string, error) {
	output, err := a.runner.RunCommandWithOutput("list", toolName)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")

	// Parse ASDF output format: "  18.0.0\n* 20.0.0\n  21.0.0"
	// The * indicates the current version.
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove the * marker for current version (can be "* " or just "*").
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		} else if strings.HasPrefix(line, "*") {
			line = strings.TrimPrefix(line, "*")
		}

		version := strings.TrimSpace(line)
		if version != "" {
			result = append(result, version)
		}
	}

	return result, nil
}

// IsVersionInstalled checks if a specific version of a tool is installed.
func (a *AsdfManager) IsVersionInstalled(toolName, version string) bool {
	versions, err := a.GetInstalledVersions(toolName)
	if err != nil {
		return false
	}

	for _, installedVersion := range versions {
		if installedVersion == version {
			return true
		}
	}
	return false
}

// ListInstalledPackages returns detailed information about installed packages.
func (a *AsdfManager) ListInstalledPackages() ([]PackageInfo, error) {
	// Get installed plugins (tools)
	plugins, err := a.ListInstalled()
	if err != nil {
		return nil, err
	}

	result := make([]PackageInfo, 0)
	for _, plugin := range plugins {
		// Get installed versions for this plugin
		versions, err := a.GetInstalledVersions(plugin)
		if err != nil {
			// If we can't get versions, still include the plugin without version info
			info := PackageInfo{
				Name:    plugin,
				Status:  PackageInstalled,
				Manager: "asdf",
			}
			result = append(result, info)
			continue
		}

		// Create a PackageInfo for each installed version
		for _, version := range versions {
			info := PackageInfo{
				Name:    plugin + " " + version,
				Version: version,
				Status:  PackageInstalled,
				Manager: "asdf",
			}
			result = append(result, info)
		}
	}

	return result, nil
}

// InstallVersion installs a specific version of a tool.
func (a *AsdfManager) InstallVersion(toolName, version string) error {
	return a.runner.RunCommand("install", toolName, version)
}

// Search searches for available ASDF plugins.
func (a *AsdfManager) Search(query string) ([]string, error) {
	output, err := a.runner.RunCommandWithOutput("plugin", "list", "all")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	result := make([]string, 0)
	queryLower := strings.ToLower(query)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// ASDF plugin list format: "plugin-name    repository-url"
			parts := strings.Fields(line)
			if len(parts) > 0 {
				pluginName := parts[0]
				if strings.Contains(strings.ToLower(pluginName), queryLower) {
					result = append(result, pluginName)
				}
			}
		}
	}

	return result, nil
}

// Info gets information about a plugin via ASDF.
func (a *AsdfManager) Info(pluginName string) (string, error) {
	// For ASDF, we can show available versions for a plugin
	output, err := a.runner.RunCommandWithOutput("list", "all", pluginName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// UpdateAll updates all ASDF plugins.
func (a *AsdfManager) UpdateAll() error {
	return a.runner.RunCommand("plugin", "update", "--all")
}

// SetConfigDir sets the plonk config directory
func (a *AsdfManager) SetConfigDir(plonkDir string) {
	a.plonkDir = plonkDir
}

// getManagedPackages returns ASDF tools listed in plonk.yaml
func (a *AsdfManager) getManagedPackages() ([]string, error) {
	if a.plonkDir == "" {
		return []string{}, nil
	}

	cfg, err := config.LoadConfig(a.plonkDir)
	if err != nil {
		// If no config exists, return empty list
		return []string{}, nil
	}

	var packages []string
	// Extract tool and version from ASDF tools
	for _, tool := range cfg.ASDF {
		packages = append(packages, tool.Name+" "+tool.Version)
	}
	return packages, nil
}

// ListManagedPackages returns packages that are managed by plonk
func (a *AsdfManager) ListManagedPackages() ([]PackageInfo, error) {
	managedPackages, err := a.getManagedPackages()
	if err != nil {
		return nil, err
	}

	var managed []PackageInfo

	for _, pkg := range managedPackages {
		// Parse "tool version" format
		parts := strings.Fields(pkg)
		if len(parts) >= 2 {
			toolName := parts[0]
			version := parts[1]

			info := PackageInfo{
				Name:    pkg,
				Version: version,
				Manager: "asdf",
			}

			// Check if this version is actually installed
			if a.IsVersionInstalled(toolName, version) {
				info.Status = PackageInstalled
			} else {
				info.Status = PackageAvailable
			}

			managed = append(managed, info)
		}
	}

	return managed, nil
}

// ListUntrackedPackages returns installed packages not managed by plonk
func (a *AsdfManager) ListUntrackedPackages() ([]PackageInfo, error) {
	allInstalled, err := a.ListInstalledPackages()
	if err != nil {
		return nil, err
	}

	managedPackages, err := a.getManagedPackages()
	if err != nil {
		return nil, err
	}

	// Create a map of managed packages for quick lookup
	managedMap := make(map[string]bool)
	for _, pkg := range managedPackages {
		managedMap[pkg] = true
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
func (a *AsdfManager) ListMissingPackages() ([]PackageInfo, error) {
	managedPackages, err := a.getManagedPackages()
	if err != nil {
		return nil, err
	}

	var missing []PackageInfo

	for _, pkg := range managedPackages {
		// Parse "tool version" format
		parts := strings.Fields(pkg)
		if len(parts) >= 2 {
			toolName := parts[0]
			version := parts[1]

			if !a.IsVersionInstalled(toolName, version) {
				info := PackageInfo{
					Name:    pkg,
					Version: version,
					Status:  PackageAvailable,
					Manager: "asdf",
				}
				missing = append(missing, info)
			}
		}
	}

	return missing, nil
}
