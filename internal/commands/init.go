// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/setup"
	"github.com/spf13/cobra"
)

var (
	initYes      bool
	skipHomebrew bool
	skipCargo    bool
	skipNPM      bool
	skipPip      bool
	skipGem      bool
	skipGo       bool
	installAll   bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new plonk configuration",
	Long: `Initialize plonk for first-time use with a fresh configuration.

This command:
- Creates the plonk configuration directory
- Creates default plonk.yaml and plonk.lock files
- Checks system requirements and offers to install missing package managers
- Installs bootstrap managers (Homebrew, Cargo) via official installers
- Installs language packages (Python, Ruby, Node.js, Go) via your default manager
- All installed language packages are tracked in plonk.lock

By default, all package managers are installed. Use --no-<manager> flags to skip
specific managers, or use --all to explicitly install all managers.

Supported package managers:
- Homebrew (macOS/Linux) - installed via official installer
- Cargo (Rust) - installed via rustup
- npm (Node.js), pip (Python), gem (Ruby), go - installed via default_manager

Examples:
  plonk init                        # Initialize with all package managers
  plonk init --no-cargo --no-gem   # Skip Rust and Ruby package managers
  plonk init --yes                  # Non-interactive mode`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initYes, "yes", false, "Non-interactive mode - answer yes to all prompts")
	initCmd.Flags().BoolVar(&skipHomebrew, "no-homebrew", false, "Skip Homebrew installation")
	initCmd.Flags().BoolVar(&skipCargo, "no-cargo", false, "Skip Cargo/Rust installation")
	initCmd.Flags().BoolVar(&skipNPM, "no-npm", false, "Skip npm/Node.js installation")
	initCmd.Flags().BoolVar(&skipPip, "no-pip", false, "Skip pip/Python installation")
	initCmd.Flags().BoolVar(&skipGem, "no-gem", false, "Skip gem/Ruby installation")
	initCmd.Flags().BoolVar(&skipGo, "no-go", false, "Skip Go installation")
	initCmd.Flags().BoolVar(&installAll, "all", false, "Install all package managers (default behavior)")

	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	skipManagers := setup.SkipManagers{
		Homebrew: skipHomebrew,
		Cargo:    skipCargo,
		NPM:      skipNPM,
		Pip:      skipPip,
		Gem:      skipGem,
		Go:       skipGo,
	}

	setupConfig := setup.Config{
		Interactive:  !initYes,
		Verbose:      false, // Could add --verbose flag later
		SkipManagers: skipManagers,
	}

	return setup.InitializeNew(ctx, setupConfig)
}
