// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

// getOSPackageManagerSupport returns which package managers are supported on the current OS
func getOSPackageManagerSupport() map[string]bool {
	switch runtime.GOOS {
	case "darwin":
		return map[string]bool{
			"homebrew": true,
			"npm":      true,
			"cargo":    true,
			"gem":      true,
			"go":       true,
			"pip":      true,
		}
	case "linux":
		return map[string]bool{
			"homebrew": true, // Supported on Linux
			"npm":      true,
			"cargo":    true,
			"gem":      true,
			"go":       true,
			"pip":      true,
			// apt would go here when implemented
		}
	default:
		// Unsupported OS - return empty map
		// Windows and other OSes are not currently supported by plonk
		return map[string]bool{}
	}
}

// getManagerInstallSuggestion returns installation instructions for different package managers
func getManagerInstallSuggestion(manager string) string {
	osSpecific := getOSSpecificInstallCommand(manager, runtime.GOOS)
	if osSpecific != "" {
		return osSpecific
	}
	return "Install the required package manager or change the default manager with 'plonk config edit'"
}

// getOSSpecificInstallCommand returns OS-specific installation instructions for package managers
func getOSSpecificInstallCommand(manager, os string) string {
	// First check if the manager is supported on this OS
	support := getOSPackageManagerSupport()
	if supported, exists := support[manager]; exists && !supported {
		return fmt.Sprintf("%s is not supported on %s", manager, os)
	}

	switch manager {
	case "homebrew":
		switch os {
		case "darwin":
			return "Install Homebrew: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
		case "linux":
			return "Install Homebrew on Linux: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
		}
	case "npm":
		switch os {
		case "darwin":
			return "Install Node.js and npm: brew install node OR download from https://nodejs.org"
		case "linux":
			return "Install Node.js and npm: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"
		}
	case "cargo":
		return "Install Rust and Cargo: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
	case "gem":
		switch os {
		case "darwin":
			return "Ruby comes pre-installed on macOS. For a newer version: brew install ruby"
		case "linux":
			return "Install Ruby: sudo apt-get install ruby-full OR use rbenv/rvm"
		}
	case "go":
		switch os {
		case "darwin":
			return "Install Go: brew install go OR download from https://go.dev/dl/"
		case "linux":
			return "Install Go: Download from https://go.dev/dl/ OR sudo snap install go --classic"
		}
	case "pip":
		switch os {
		case "darwin":
			return "Install Python and pip: brew install python3 OR download from https://python.org"
		case "linux":
			return "Install Python and pip: sudo apt-get install python3-pip"
		}
	}

	return ""
}

// GetMetadataString safely extracts string metadata from operation results
func GetMetadataString(result state.OperationResult, key string) string {
	if result.Metadata == nil {
		return ""
	}
	if value, ok := result.Metadata[key].(string); ok {
		return value
	}
	return ""
}

// CreateOperationContext creates a context with timeout for batch operations
func CreateOperationContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CheckCancellation checks if the context has been canceled and returns appropriate error
func CheckCancellation(ctx context.Context, domain errors.Domain, operation string) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), errors.ErrInternal, domain, operation,
			"operation canceled or timed out")
	}
	return nil
}

// DetermineExitCode determines the appropriate exit code based on operation results
func DetermineExitCode(results []state.OperationResult, domain errors.Domain, operation string) error {
	if len(results) == 0 {
		return nil
	}

	summary := state.CalculateSummary(results)

	// Success if any items were added or updated
	if summary.Added > 0 || summary.Updated > 0 {
		return nil
	}

	// Failure only if all items failed
	if summary.Failed > 0 && summary.Added == 0 && summary.Updated == 0 && summary.Skipped == 0 {
		return errors.NewError(
			getErrorCodeForDomain(domain),
			domain,
			operation,
			fmt.Sprintf("failed to process %d item(s)", summary.Failed),
		)
	}

	return nil
}

// getErrorCodeForDomain returns the appropriate error code for a domain
func getErrorCodeForDomain(domain errors.Domain) errors.ErrorCode {
	switch domain {
	case errors.DomainPackages:
		return errors.ErrPackageInstall
	case errors.DomainDotfiles:
		return errors.ErrFileIO
	default:
		return errors.ErrInternal
	}
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
	// Get home directory (no error handling needed)
	_ = orchestrator.GetHomeDir()

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
