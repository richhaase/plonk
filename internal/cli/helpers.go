// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package cli contains helpers specific to CLI interaction,
// such as argument parsing, confirmation prompts, and command helpers.
package cli

import (
	"strings"

	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/spf13/cobra"
)

// GetMetadataString safely extracts string metadata from operation results
func GetMetadataString(result operations.OperationResult, key string) string {
	if result.Metadata == nil {
		return ""
	}
	if value, ok := result.Metadata[key].(string); ok {
		return value
	}
	return ""
}

// SimpleFlags represents basic command flags without detection logic
type SimpleFlags struct {
	Manager string
	DryRun  bool
	Force   bool
	Verbose bool
	Output  string
}

// ParseSimpleFlags parses basic flags for package commands
func ParseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	flags := &SimpleFlags{}

	// Parse manager flags with precedence
	if brew, _ := cmd.Flags().GetBool("brew"); brew {
		flags.Manager = "homebrew"
	} else if npm, _ := cmd.Flags().GetBool("npm"); npm {
		flags.Manager = "npm"
	} else if cargo, _ := cmd.Flags().GetBool("cargo"); cargo {
		flags.Manager = "cargo"
	} else if pip, _ := cmd.Flags().GetBool("pip"); pip {
		flags.Manager = "pip"
	} else if gem, _ := cmd.Flags().GetBool("gem"); gem {
		flags.Manager = "gem"
	} else if goFlag, _ := cmd.Flags().GetBool("go"); goFlag {
		flags.Manager = "go"
	}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")
	flags.Output, _ = cmd.Flags().GetString("output")

	return flags, nil
}

// CompleteDotfilePaths provides file path completion for dotfiles
func CompleteDotfilePaths(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get home directory from shared context (no error handling needed)
	sharedCtx := runtime.GetSharedContext()
	_ = sharedCtx.HomeDir()

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
