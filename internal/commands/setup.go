// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/setup"
	"github.com/spf13/cobra"
)

var (
	setupYes bool
)

var setupCmd = &cobra.Command{
	Use:   "setup [git-repo]",
	Short: "DEPRECATED: Use 'plonk init' or 'plonk clone' instead",
	Long: `DEPRECATED: The 'plonk setup' command is deprecated.

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
	// Show deprecation warning
	fmt.Println("⚠️  WARNING: The 'plonk setup' command is deprecated and will be removed in a future version.")
	fmt.Println()

	if len(args) == 0 {
		fmt.Println("Please use 'plonk init' instead to initialize a new plonk configuration.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  plonk init                        # Initialize with all package managers")
		fmt.Println("  plonk init --no-cargo --no-gem   # Skip specific managers")
		fmt.Println()
		fmt.Println("Run 'plonk init --help' for more information.")
	} else {
		fmt.Println("Please use 'plonk clone' instead to clone a dotfiles repository.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Printf("  plonk clone %s\n", args[0])
		fmt.Println()
		fmt.Println("Run 'plonk clone --help' for more information.")
	}

	fmt.Println()
	fmt.Println("Continuing with deprecated command for backward compatibility...")
	fmt.Println()

	// Continue with old behavior for now
	ctx := context.Background()

	setupConfig := setup.Config{
		Interactive: !setupYes,
		Verbose:     false,
	}

	if len(args) == 0 {
		// Setup without repository - use InitializeNew
		return setup.InitializeNew(ctx, setupConfig)
	} else {
		// Setup with repository - use CloneAndSetup
		gitRepo := args[0]
		return setup.CloneAndSetup(ctx, gitRepo, setupConfig)
	}
}
