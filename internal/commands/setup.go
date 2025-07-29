// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/setup"
	"github.com/spf13/cobra"
)

var (
	setupYes bool
)

var setupCmd = &cobra.Command{
	Use:   "setup [git-repo]",
	Short: "Set up plonk configuration and install required tools",
	Long: `Set up plonk for first-time use or clone an existing dotfiles repository.

Without arguments, this command:
- Creates the plonk configuration directory
- Creates default plonk.yaml and plonk.lock files
- Checks system requirements and offers to install missing package managers
- Installs bootstrap managers (Homebrew, Cargo) via official installers
- Installs language packages (Python, Ruby, Node.js, Go) via your default manager
- All installed language packages are tracked in plonk.lock

With a git repository argument, this command:
- Clones the repository into your plonk directory
- Sets up configuration files if they don't exist
- Installs missing package managers as described above
- Runs 'plonk apply' to configure your system

Supported package managers:
- Homebrew (macOS/Linux) - installed via official installer
- Cargo (Rust) - installed via rustup
- npm (Node.js), pip (Python), gem (Ruby), go - installed via default_manager

Git repository formats supported:
- GitHub shorthand: user/repo (defaults to HTTPS)
- HTTPS URL: https://github.com/user/repo.git
- SSH URL: git@github.com:user/repo.git
- Git protocol: git://github.com/user/repo.git

Examples:
  plonk setup                        # Initialize new plonk setup
  plonk setup user/dotfiles         # Clone GitHub repo (shorthand)
  plonk setup richhaase/dotfiles    # Clone specific user's dotfiles
  plonk setup https://github.com/user/repo.git  # Full HTTPS URL
  plonk setup --yes user/dotfiles   # Non-interactive setup`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().BoolVar(&setupYes, "yes", false, "Non-interactive mode - answer yes to all prompts")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	setupConfig := setup.Config{
		Interactive: !setupYes,
		Verbose:     false, // Could add --verbose flag later
	}

	if len(args) == 0 {
		// Setup without repository
		return setup.SetupWithoutRepo(ctx, setupConfig)
	} else {
		// Setup with repository
		gitRepo := args[0]
		return setup.SetupWithRepo(ctx, gitRepo, setupConfig)
	}
}
