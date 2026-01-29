// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

// buildInstallExamples generates CLI examples for the install command based on
// the available package managers.
func buildInstallExamples() string {
	var lines []string

	// Always include a simple, manager-agnostic example.
	lines = append(lines, "plonk install htop git neovim ripgrep")

	// Add a few manager-prefixed examples using supported managers.
	managerNames := packages.SupportedManagers

	const maxManagers = 4
	for i, name := range managerNames {
		if i >= maxManagers {
			break
		}
		lines = append(lines, fmt.Sprintf("plonk install %s:PACKAGE", name))
	}

	return strings.Join(lines, "\n")
}

// buildUninstallExamples generates CLI examples for the uninstall command
// based on the available package managers.
func buildUninstallExamples() string {
	var lines []string

	lines = append(lines, "plonk uninstall htop git")

	// Add a few manager-prefixed examples using supported managers.
	managerNames := packages.SupportedManagers

	const maxManagers = 2
	for i, name := range managerNames {
		if i >= maxManagers {
			break
		}
		lines = append(lines, fmt.Sprintf("plonk uninstall %s:PACKAGE", name))
	}

	return strings.Join(lines, "\n")
}

// buildUpgradeExamples generates CLI examples for the upgrade command using
// the available managers.
func buildUpgradeExamples() string {
	var lines []string

	// Generic examples that do not depend on specific manager names.
	lines = append(lines, "plonk upgrade")
	lines = append(lines, "plonk upgrade ripgrep")

	// Add manager-specific examples using supported managers.
	managerNames := packages.SupportedManagers

	if len(managerNames) > 0 {
		lines = append(lines, fmt.Sprintf("plonk upgrade %s", managerNames[0]))
	}
	if len(managerNames) > 1 {
		lines = append(lines, fmt.Sprintf("plonk upgrade %s %s", managerNames[0], managerNames[1]))
	}

	return strings.Join(lines, "\n")
}

// SimpleFlags represents basic command flags
type SimpleFlags struct {
	DryRun  bool
	Force   bool
	Verbose bool
}

// ParseSimpleFlags parses basic flags for commands
func ParseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	flags := &SimpleFlags{}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")

	return flags, nil
}

// normalizeDisplayFlags sets defaults when no flags specified
func normalizeDisplayFlags(showPackages, showDotfiles bool) (packages, dotfiles bool) {
	// If neither flag is set, show both
	if !showPackages && !showDotfiles {
		return true, true
	}
	return showPackages, showDotfiles
}

// parseSimpleFlags parses basic flags for commands
func parseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	return ParseSimpleFlags(cmd)
}

// CompleteDotfilePaths provides file path completion for dotfiles
func CompleteDotfilePaths(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get home directory (no error handling needed)
	_ = config.GetHomeDir()

	// Define common dotfile suggestions
	commonDotfiles := []string{
		"~/.zshrc", "~/.bashrc", "~/.bash_profile", "~/.profile",
		"~/.vimrc", "~/.vim/", "~/.nvim/",
		"~/.gitconfig", "~/.gitignore_global",
		"~/.tmux.conf", "~/.tmux/",
		"~/.ssh/config", "~/.ssh/",
		"~/.aws/config", "~/.aws/credentials",
		"~/.config/", "~/.config/nvim/", "~/.config/fish/", "~/.config/alacritty/",
		"~/.docker/config.json",
		"~/.zprofile", "~/.zshenv",
		"~/.inputrc", "~/.editorconfig",
	}

	// If no input yet, return all common suggestions
	if toComplete == "" {
		return commonDotfiles, cobra.ShellCompDirectiveNoSpace
	}

	// If starts with tilde, filter common dotfiles
	if strings.HasPrefix(toComplete, "~/") {
		var filtered []string
		for _, suggestion := range commonDotfiles {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}

		// Fall back to file completion for ~/.config/ style paths
		return nil, cobra.ShellCompDirectiveDefault
	}

	// For relative paths, try to suggest based on common dotfile names
	if !strings.HasPrefix(toComplete, "/") {
		relativeSuggestions := []string{
			".zshrc", ".bashrc", ".bash_profile", ".profile",
			".vimrc", ".gitconfig", ".tmux.conf", ".inputrc",
			".editorconfig", ".zprofile", ".zshenv",
		}

		var filtered []string
		for _, suggestion := range relativeSuggestions {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}
	}

	// Fall back to default file completion for absolute paths and other cases
	return nil, cobra.ShellCompDirectiveDefault
}

