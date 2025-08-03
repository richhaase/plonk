// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

// ParsePackageSpec splits "manager:package" into (manager, package)
// Returns ("", package) if no prefix is found
func ParsePackageSpec(spec string) (manager, packageName string) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", spec
}

// IsValidManager checks if the given manager name is supported
func IsValidManager(manager string) bool {
	registry := packages.NewManagerRegistry()
	validManagers := registry.GetAllManagerNames()
	for _, valid := range validManagers {
		if manager == valid {
			return true
		}
	}
	return false
}

// GetValidManagers returns a list of all valid manager names
func GetValidManagers() []string {
	registry := packages.NewManagerRegistry()
	return registry.GetAllManagerNames()
}

// GetMetadataString safely extracts string metadata from operation results
func GetMetadataString(result resources.OperationResult, key string) string {
	if result.Metadata == nil {
		return ""
	}
	if value, ok := result.Metadata[key].(string); ok {
		return value
	}
	return ""
}

// SimpleFlags represents basic command flags
type SimpleFlags struct {
	DryRun  bool
	Force   bool
	Verbose bool
	Output  string
}

// ParseSimpleFlags parses basic flags for commands
func ParseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	flags := &SimpleFlags{}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")
	flags.Output, _ = cmd.Flags().GetString("output")

	return flags, nil
}

// validatePackageSpec validates a package specification and returns an error if invalid
func validatePackageSpec(spec, manager, packageName string) error {
	// Check for empty package name
	if packageName == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	// Check for empty manager prefix when colon is present
	if manager == "" && spec != packageName {
		// This means there was a colon but empty prefix (e.g., ":package")
		return fmt.Errorf("manager prefix cannot be empty")
	}

	return nil
}

// resolvePackageManager determines the package manager to use based on the spec and config
func resolvePackageManager(manager string, cfg *config.Config) string {
	if manager != "" {
		return manager
	}

	if cfg != nil && cfg.DefaultManager != "" {
		return cfg.DefaultManager
	}

	return packages.DefaultManager
}

// ValidatedPackageSpec represents a validated package specification
type ValidatedPackageSpec struct {
	Manager      string
	PackageName  string
	OriginalSpec string
}

// parseAndValidatePackageSpecs processes package specifications and returns validated specs or errors
func parseAndValidatePackageSpecs(args []string, cfg *config.Config) ([]ValidatedPackageSpec, []resources.OperationResult) {
	var specs []ValidatedPackageSpec
	var errors []resources.OperationResult

	for _, packageSpec := range args {
		manager, packageName := ParsePackageSpec(packageSpec)

		// Validate the package specification
		if err := validatePackageSpec(packageSpec, manager, packageName); err != nil {
			errors = append(errors, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: %w", packageSpec, err),
			})
			continue
		}

		// Resolve the package manager
		manager = resolvePackageManager(manager, cfg)

		// Validate the manager
		if !IsValidManager(manager) {
			errors = append(errors, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   fmt.Errorf("unknown package manager %q", manager),
			})
			continue
		}

		specs = append(specs, ValidatedPackageSpec{
			Manager:      manager,
			PackageName:  packageName,
			OriginalSpec: packageSpec,
		})
	}

	return specs, errors
}

// parseAndValidateUninstallSpecs processes package specifications for uninstall command
// Unlike install, this doesn't resolve default managers - it lets the uninstall command determine which manager to use
func parseAndValidateUninstallSpecs(args []string) ([]ValidatedPackageSpec, []resources.OperationResult) {
	var specs []ValidatedPackageSpec
	var errors []resources.OperationResult

	for _, packageSpec := range args {
		manager, packageName := ParsePackageSpec(packageSpec)

		// Validate the package specification
		if err := validatePackageSpec(packageSpec, manager, packageName); err != nil {
			errors = append(errors, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: %w", packageSpec, err),
			})
			continue
		}

		// Only validate manager if explicitly specified
		if manager != "" && !IsValidManager(manager) {
			errors = append(errors, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   fmt.Errorf("unknown package manager %q", manager),
			})
			continue
		}

		specs = append(specs, ValidatedPackageSpec{
			Manager:      manager, // Can be empty - uninstall will determine
			PackageName:  packageName,
			OriginalSpec: packageSpec,
		})
	}

	return specs, errors
}

// validateStatusFlags checks for incompatible flag combinations
func validateStatusFlags(showUnmanaged, showMissing bool) error {
	if showUnmanaged && showMissing {
		return fmt.Errorf("--unmanaged and --missing are mutually exclusive: items cannot be both untracked and missing")
	}
	return nil
}

// normalizeDisplayFlags sets defaults when no flags specified
func normalizeDisplayFlags(showPackages, showDotfiles bool) (packages, dotfiles bool) {
	// If neither flag is set, show both
	if !showPackages && !showDotfiles {
		return true, true
	}
	return showPackages, showDotfiles
}

// validateSearchSpec validates a search specification
func validateSearchSpec(manager, packageName string) error {
	// Package name must not be empty
	if packageName == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	// If manager is specified, it must be valid
	if manager != "" && !IsValidManager(manager) {
		return fmt.Errorf("unknown package manager %q", manager)
	}

	return nil
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
