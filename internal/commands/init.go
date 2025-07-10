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

	// Get actual default values
	defaults := config.GetDefaults()

	// Create a configuration file with all default values shown
	configContent := fmt.Sprintf(`# Plonk Configuration File
# This file contains your custom settings. All settings are optional.
# Remove or comment out any settings you want to use the defaults for.
# 
# The values shown below are the actual defaults that plonk uses.
# You can modify any of these values to customize plonk's behavior.

settings:
  # Default package manager to use when installing packages
  # Options: homebrew, npm, cargo
  default_manager: %s
  
  # Timeout settings (in seconds)
  # Set to 0 for unlimited timeout (not recommended)
  
  # Overall operation timeout - maximum time for any single command
  operation_timeout: %d   # %d seconds = %d minutes
  
  # Individual package operation timeout - time limit for package install/uninstall
  package_timeout: %d     # %d seconds = %d minutes
  
  # Dotfile operation timeout - time limit for file copy/link operations
  dotfile_timeout: %d      # %d seconds = %d minute
  
  # Directories to expand when listing dotfiles with 'plonk dot list'
  # These directories will show individual files instead of just the directory name
  expand_directories:`,
		defaults.DefaultManager,
		defaults.OperationTimeout, defaults.OperationTimeout, defaults.OperationTimeout/60,
		defaults.PackageTimeout, defaults.PackageTimeout, defaults.PackageTimeout/60,
		defaults.DotfileTimeout, defaults.DotfileTimeout, defaults.DotfileTimeout/60)

	// Add expand directories
	for _, dir := range defaults.ExpandDirectories {
		configContent += fmt.Sprintf("\n    - %s", dir)
	}

	configContent += `

# Files and patterns to ignore when discovering dotfiles
# These patterns prevent certain files from being managed by plonk
# Supports glob patterns like *.tmp, *.backup, etc.
ignore_patterns:`

	// Add ignore patterns
	for _, pattern := range defaults.IgnorePatterns {
		configContent += fmt.Sprintf("\n  - %q", pattern)
	}

	configContent += "\n"

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
