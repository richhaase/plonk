// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"
	"plonk/internal/errors"

	"github.com/spf13/cobra"
)

var dotAddCmd = &cobra.Command{
	Use:   "add <dotfile>",
	Short: "Add a dotfile to plonk configuration and import it",
	Long: `Import an existing dotfile into your plonk configuration.

This command will:
- Copy the dotfile from its current location to your plonk dotfiles directory
- Add it to your plonk.yaml configuration
- Preserve the original file in case you need to revert

Examples:
  plonk dot add ~/.zshrc           # Add .zshrc to plonk management
  plonk dot add ~/.config/nvim/    # Add nvim config directory
  plonk dot add ~/.gitconfig       # Add git config file`,
	RunE: runDotAdd,
	Args: cobra.ExactArgs(1),
}

func init() {
	dotCmd.AddCommand(dotAddCmd)
}

func runDotAdd(cmd *cobra.Command, args []string) error {
	dotfilePath := args[0]

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "dot-add", "output-format", "invalid output format")
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "dot-add", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Resolve and validate dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path")
	}

	// Check if dotfile exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile does not exist: %s", resolvedPath))
	}

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			cfg = &config.Config{
				Settings: config.Settings{
					DefaultManager: "homebrew",
				},
			}
		} else {
			return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", "failed to load config")
		}
	}

	// Generate source and destination paths
	source, destination := generatePaths(resolvedPath, homeDir)

	// Check if already managed by checking if source file exists in config dir
	adapter := config.NewConfigAdapter(cfg)
	dotfileTargets := adapter.GetDotfileTargets()
	if _, exists := dotfileTargets[source]; exists {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile is already managed: %s", source))
	}

	// Copy dotfile to plonk config directory
	sourcePath := filepath.Join(configDir, source)
	if err := copyDotfile(resolvedPath, sourcePath); err != nil {
		return errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
	}

	// Configuration doesn't need to be updated since we use auto-discovery
	// The dotfile will be automatically detected once it's in the config directory

	// Prepare output
	outputData := DotfileAddOutput{
		Source:      source,
		Destination: destination,
		Action:      "added",
		Path:        resolvedPath,
	}

	return RenderOutput(outputData, format)
}

// resolveDotfilePath resolves relative paths and validates the dotfile path
func resolveDotfilePath(path, homeDir string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "resolve", "failed to resolve path")
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(absPath, homeDir) {
		return "", errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate", fmt.Sprintf("dotfile must be within home directory: %s", absPath))
	}

	return absPath, nil
}

// generatePaths generates source and destination paths for the dotfile
func generatePaths(resolvedPath, homeDir string) (string, string) {
	// Get relative path from home directory
	relPath, err := filepath.Rel(homeDir, resolvedPath)
	if err != nil {
		// Fallback to just the filename
		relPath = filepath.Base(resolvedPath)
	}

	// Generate destination (always relative to home with ~ prefix)
	destination := "~/" + relPath

	// Generate source path using our naming convention
	source := targetToSource(destination)

	return source, destination
}

// targetToSource converts a target path to source path using our convention
func targetToSource(target string) string {
	// Use the config package implementation
	return config.TargetToSource(target)
}

// copyDotfile copies a dotfile (file or directory) to the destination
func copyDotfile(src, dst string) error {
	// For dot add, we need to copy from the actual file system location
	// to the plonk config directory, so we use a simple file copy approach

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDirectoryContents(src, dst)
	}
	return copyFileContents(src, dst)
}

// copyFileContents copies a file from src to dst
func copyFileContents(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0600)
}

// copyDirectoryContents copies a directory from src to dst
func copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFileContents(path, destPath)
	})
}

// DotfileAddOutput represents the output structure for dotfile add command
type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
}

// TableOutput generates human-friendly table output for dotfile add
func (d DotfileAddOutput) TableOutput() string {
	output := "Dotfile Add\n===========\n\n"
	output += fmt.Sprintf("✅ Added dotfile to plonk configuration\n")
	output += fmt.Sprintf("   Source: %s\n", d.Source)
	output += fmt.Sprintf("   Destination: %s\n", d.Destination)
	output += fmt.Sprintf("   Original: %s\n", d.Path)
	output += "\nThe dotfile has been copied to your plonk config directory and added to plonk.yaml\n"
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileAddOutput) StructuredData() any {
	return d
}
