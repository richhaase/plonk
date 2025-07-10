// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/errors"

	"github.com/spf13/cobra"
)

var (
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize plonk configuration",
	Long: `Create a plonk configuration file with default settings.

This command creates a plonk.yaml file in the configuration directory with
sensible defaults. This is useful when you want to customize settings from
the built-in defaults.

Examples:
  plonk init                    # Create config with defaults
  plonk init --force           # Overwrite existing config file
`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing configuration file")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Check if config file already exists
	if !initForce {
		if _, err := os.Stat(configPath); err == nil {
			return errors.NewError(errors.ErrFileExists, errors.DomainConfig, "init",
				fmt.Sprintf("Configuration file already exists: %s\nUse --force to overwrite", configPath))
		}
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainConfig, "init",
			"failed to create config directory").WithItem(configDir)
	}

	// Create a configuration file with helpful comments and some common settings
	configContent := `# Plonk Configuration File
# This file contains your custom settings. All settings are optional.
# Remove or comment out any settings you want to use the defaults for.

settings:
  # Default package manager (homebrew, npm, cargo)
  default_manager: homebrew
  
  # Timeout settings (in seconds)
  # operation_timeout: 300   # 5 minutes - overall operation timeout
  # package_timeout: 180     # 3 minutes - individual package operations  
  # dotfile_timeout: 60      # 1 minute - dotfile operations
  
  # Directories to expand when listing dotfiles
  # expand_directories:
  #   - .config
  #   - .ssh
  #   - .aws
  #   - .kube
  #   - .docker
  #   - .gnupg
  #   - .local

# Files and patterns to ignore when discovering dotfiles
# ignore_patterns:
#   - .DS_Store
#   - .git
#   - "*.backup"
#   - "*.tmp"
#   - "*.swp"
`

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "init",
			"failed to write configuration file").WithItem(configPath)
	}

	// Success message
	fmt.Printf("✅ Created configuration file: %s\n", configPath)
	fmt.Printf("📝 Edit the file to customize your settings\n")
	fmt.Printf("💡 Run 'plonk config show' to see your current configuration\n")

	return nil
}
