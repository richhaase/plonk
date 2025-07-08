// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"
	"plonk/internal/dotfiles"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		return err
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Resolve and validate dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		return err
	}

	// Check if dotfile exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return fmt.Errorf("dotfile does not exist: %s", resolvedPath)
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
				Dotfiles: []config.DotfileEntry{},
			}
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Generate source and destination paths
	source, destination := generatePaths(resolvedPath, homeDir)

	// Check if already managed
	for _, entry := range cfg.Dotfiles {
		if entry.Destination == destination || (entry.Source != "" && entry.Source == source) {
			return fmt.Errorf("dotfile is already managed: %s", destination)
		}
	}

	// Copy dotfile to plonk config directory
	sourcePath := filepath.Join(configDir, source)
	if err := copyDotfile(resolvedPath, sourcePath); err != nil {
		return fmt.Errorf("failed to copy dotfile: %w", err)
	}

	// Add to configuration
	newEntry := config.DotfileEntry{
		Source:      source,
		Destination: destination,
	}
	cfg.Dotfiles = append(cfg.Dotfiles, newEntry)

	// Save configuration
	if err := saveDotfileConfig(cfg, configDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

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
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(absPath, homeDir) {
		return "", fmt.Errorf("dotfile must be within home directory: %s", absPath)
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
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
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
	return os.WriteFile(dst, content, 0644)
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

// saveDotfileConfig saves the configuration to plonk.yaml atomically
func saveDotfileConfig(cfg *config.Config, configDir string) error {
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write to file atomically
	atomicWriter := dotfiles.NewAtomicFileWriter()
	return atomicWriter.WriteFile(configPath, data, 0644)
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