package managers

import (
	"strings"
)

// AsdfManager manages ASDF tools and versions
type AsdfManager struct {
	runner *CommandRunner
}

// NewAsdfManager creates a new ASDF manager
func NewAsdfManager(executor CommandExecutor) *AsdfManager {
	return &AsdfManager{
		runner: NewCommandRunner(executor, "asdf"),
	}
}

// IsAvailable checks if ASDF is installed
func (a *AsdfManager) IsAvailable() bool {
	err := a.runner.RunCommand("version")
	return err == nil
}

// Install installs a tool/version via ASDF
// packageName should be in format "tool version" like "nodejs 20.0.0"
func (a *AsdfManager) Install(packageName string) error {
	parts := strings.Fields(packageName)
	if len(parts) < 2 {
		return a.runner.RunCommand("install", packageName)
	}
	
	// asdf install <tool> <version>
	args := append([]string{"install"}, parts...)
	return a.runner.RunCommand(args...)
}

// ListInstalled lists all installed ASDF plugins
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
	
	// Clean up any empty strings
	result := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		if trimmed := strings.TrimSpace(plugin); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result, nil
}

// Update updates a tool to the latest version via ASDF
func (a *AsdfManager) Update(toolName string) error {
	// Get the latest version available
	output, err := a.runner.RunCommandWithOutput("latest", toolName)
	if err != nil {
		return err
	}
	
	latestVersion := strings.TrimSpace(output)
	if latestVersion == "" {
		return nil // No version available
	}
	
	// Install the latest version
	return a.runner.RunCommand("install", toolName, latestVersion)
}

// IsInstalled checks if a tool is installed via ASDF (has any versions)
func (a *AsdfManager) IsInstalled(toolName string) bool {
	err := a.runner.RunCommand("list", toolName)
	return err == nil
}

// GetInstalledVersions returns all installed versions for a tool
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
	// The * indicates the current version
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Remove the * marker for current version
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		}
		
		version := strings.TrimSpace(line)
		if version != "" {
			result = append(result, version)
		}
	}
	
	return result, nil
}

// IsVersionInstalled checks if a specific version of a tool is installed
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

// InstallVersion installs a specific version of a tool
func (a *AsdfManager) InstallVersion(toolName, version string) error {
	return a.runner.RunCommand("install", toolName, version)
}