// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ZSHManager manages ZSH plugins and configurations.
type ZSHManager struct {
	executor  CommandExecutor
	pluginDir string
}

// NewZSHManager creates a new ZSH manager.
func NewZSHManager(executor CommandExecutor) *ZSHManager {
	pluginDir := getZSHPluginDir()
	return &ZSHManager{
		executor:  executor,
		pluginDir: pluginDir,
	}
}

// getZSHPluginDir returns the ZSH plugin directory location.
func getZSHPluginDir() string {
	if dir := os.Getenv("ZPLUGINDIR"); dir != "" {
		return dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".config/zsh/plugins" // fallback.
	}

	return filepath.Join(homeDir, ".config", "zsh", "plugins")
}

// IsAvailable checks if ZSH is available.
func (z *ZSHManager) IsAvailable() bool {
	cmd := z.executor.Execute("zsh", "--version")
	err := cmd.Run()
	return err == nil
}

// Install installs a ZSH plugin by cloning from GitHub.
func (z *ZSHManager) Install(pluginRepo string) error {
	pluginName := getPluginName(pluginRepo)
	pluginPath := filepath.Join(z.pluginDir, pluginName)

	// Check if plugin already exists
	if _, err := os.Stat(pluginPath); err == nil {
		return nil // Already installed.
	}

	// Create plugin directory if it doesn't exist.
	if err := os.MkdirAll(z.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Clone the plugin.
	githubURL := fmt.Sprintf("https://github.com/%s", pluginRepo)
	cmd := z.executor.Execute("git", "clone", "-q", "--depth", "1", "--recursive", "--shallow-submodules", githubURL, pluginPath)
	return cmd.Run()
}

// ListInstalled lists all installed ZSH plugins.
func (z *ZSHManager) ListInstalled() ([]string, error) {
	if _, err := os.Stat(z.pluginDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(z.pluginDir)
	if err != nil {
		return nil, err
	}

	plugins := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}

// Update updates a specific plugin or all plugins.
func (z *ZSHManager) Update(pluginName string) error {
	if pluginName == "" {
		return z.updateAllPlugins()
	}

	pluginPath := filepath.Join(z.pluginDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	cmd := z.executor.Execute("git", "-C", pluginPath, "pull", "--ff", "--recurse-submodules", "--depth", "1", "--rebase", "--autostash")
	return cmd.Run()
}

// updateAllPlugins updates all installed plugins.
func (z *ZSHManager) updateAllPlugins() error {
	plugins, err := z.ListInstalled()
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		if err := z.Update(plugin); err != nil {
			return fmt.Errorf("failed to update plugin %s: %w", plugin, err)
		}
	}

	return nil
}

// IsInstalled checks if a plugin is installed.
func (z *ZSHManager) IsInstalled(pluginRepo string) bool {
	pluginName := getPluginName(pluginRepo)
	pluginPath := filepath.Join(z.pluginDir, pluginName)
	_, err := os.Stat(pluginPath)
	return err == nil
}

// Remove removes a ZSH plugin.
func (z *ZSHManager) Remove(pluginName string) error {
	pluginPath := filepath.Join(z.pluginDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	return os.RemoveAll(pluginPath)
}

// getPluginName extracts the plugin name from a GitHub repository path.
func getPluginName(pluginRepo string) string {
	parts := strings.Split(pluginRepo, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (repository name).
	}
	return pluginRepo
}
