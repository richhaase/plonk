// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathCheckResult represents the result of checking a directory's PATH status
type PathCheckResult struct {
	inPath      bool
	exists      bool
	suggestions []string
}

// shellInfo represents shell detection information
type shellInfo struct {
	name       string
	configFile string
	reload     string
}

// checkDirectoryInPath checks if a directory exists and is in PATH
func checkDirectoryInPath(directory string) PathCheckResult {
	result := PathCheckResult{}

	// Check if directory exists
	if _, err := os.Stat(directory); err == nil {
		result.exists = true
	}

	// Check if in PATH
	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, string(os.PathListSeparator))
	for _, pathDir := range pathDirs {
		if pathDir == directory {
			result.inPath = true
			break
		}
	}

	// Generate suggestions if not in PATH but exists
	if result.exists && !result.inPath {
		shellPath := os.Getenv("SHELL")
		shell := detectShell(shellPath)
		pathExport := generatePathExport([]string{directory})
		commands := generateShellCommands(shell, pathExport)

		result.suggestions = append(result.suggestions, fmt.Sprintf("Detected shell: %s", shell.name))
		result.suggestions = append(result.suggestions, "Add to PATH with:")
		for _, cmd := range commands {
			result.suggestions = append(result.suggestions, fmt.Sprintf("  %s", cmd))
		}
	}

	return result
}

// detectShell detects the user's shell from SHELL environment variable
func detectShell(shellPath string) shellInfo {
	// Default to bash if detection fails
	defaultShell := shellInfo{
		name:       "bash",
		configFile: "~/.bashrc",
		reload:     "source ~/.bashrc",
	}

	if shellPath == "" {
		return defaultShell
	}

	// Extract shell name from path
	shellName := filepath.Base(shellPath)

	switch shellName {
	case "zsh":
		return shellInfo{
			name:       "zsh",
			configFile: "~/.zshrc",
			reload:     "source ~/.zshrc",
		}
	case "bash":
		return defaultShell
	case "fish":
		return shellInfo{
			name:       "fish",
			configFile: "~/.config/fish/config.fish",
			reload:     "source ~/.config/fish/config.fish",
		}
	case "ksh":
		return shellInfo{
			name:       "ksh",
			configFile: "~/.kshrc",
			reload:     ". ~/.kshrc",
		}
	case "tcsh":
		return shellInfo{
			name:       "tcsh",
			configFile: "~/.tcshrc",
			reload:     "source ~/.tcshrc",
		}
	default:
		// Try to infer from common patterns
		if strings.Contains(shellPath, "zsh") {
			return shellInfo{
				name:       "zsh",
				configFile: "~/.zshrc",
				reload:     "source ~/.zshrc",
			}
		}
		return defaultShell
	}
}

// generatePathExport creates the PATH export line for missing paths
func generatePathExport(missingPaths []string) string {
	if len(missingPaths) == 0 {
		return ""
	}

	// Join all paths with colon
	pathString := strings.Join(missingPaths, ":")
	return fmt.Sprintf("export PATH=\"%s:$PATH\"", pathString)
}

// generateShellCommands generates shell-specific commands for PATH configuration
func generateShellCommands(shell shellInfo, pathExport string) []string {
	if shell.name == "fish" {
		// Fish shell has special syntax
		commands := []string{}
		// Extract paths from the export command
		// pathExport looks like: export PATH="/path1:/path2:$PATH"
		if strings.HasPrefix(pathExport, "export PATH=\"") && strings.HasSuffix(pathExport, ":$PATH\"") {
			pathString := strings.TrimPrefix(pathExport, "export PATH=\"")
			pathString = strings.TrimSuffix(pathString, ":$PATH\"")
			paths := strings.Split(pathString, ":")
			for _, path := range paths {
				if path != "" {
					commands = append(commands, fmt.Sprintf("fish_add_path %s", path))
				}
			}
		}
		return commands
	}

	// For most shells, we can append to config file
	commands := []string{
		fmt.Sprintf("echo '%s' >> %s", pathExport, shell.configFile),
		shell.reload,
	}

	return commands
}
