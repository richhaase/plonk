// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"

	"github.com/spf13/cobra"
)

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display the current plonk configuration with validation status.

Shows:
- Configuration file location
- Validation status  
- Package counts by manager
- Dotfiles count
- Settings summary

Examples:
  plonk config show           # Show configuration summary
  plonk config show -o json   # Show as JSON
  plonk config show -o yaml   # Show as YAML`,
	RunE: runConfigShow,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// Handle missing config gracefully
		if strings.Contains(err.Error(), "config file not found") {
			outputData := ConfigShowOutput{
				ConfigPath: filepath.Join(configDir, "plonk.yaml"),
				Status:     "missing",
				Message:    "Configuration file not found. Run 'plonk config init' to create one.",
			}
			return RenderOutput(outputData, format)
		}
		
		// Handle validation errors
		if strings.Contains(err.Error(), "validation failed") {
			outputData := ConfigShowOutput{
				ConfigPath: getConfigPath(configDir),
				Status:     "invalid",
				Message:    err.Error(),
			}
			return RenderOutput(outputData, format)
		}
		
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Build output data
	outputData := ConfigShowOutput{
		ConfigPath: getConfigPath(configDir),
		Status:     "valid",
		Message:    "Configuration is valid",
		Config:     cfg,
		Summary: ConfigSummary{
			HomebrewPackages: len(cfg.Homebrew.Brews) + len(cfg.Homebrew.Casks),
			ASDFTools:        len(cfg.ASDF),
			NPMPackages:      len(cfg.NPM),
			Dotfiles:         len(cfg.Dotfiles),
			DefaultManager:   cfg.Settings.DefaultManager,
		},
	}

	return RenderOutput(outputData, format)
}

// getConfigPath finds the actual config file path
func getConfigPath(configDir string) string {
	mainPath := filepath.Join(configDir, "plonk.yaml")
	if _, err := os.Stat(mainPath); err == nil {
		return mainPath
	}
	
	repoPath := filepath.Join(configDir, "repo", "plonk.yaml")
	if _, err := os.Stat(repoPath); err == nil {
		return repoPath
	}
	
	return mainPath // Default to main path
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string         `json:"config_path" yaml:"config_path"`
	Status     string         `json:"status" yaml:"status"`
	Message    string         `json:"message" yaml:"message"`
	Config     *config.Config `json:"config,omitempty" yaml:"config,omitempty"`
	Summary    ConfigSummary  `json:"summary,omitempty" yaml:"summary,omitempty"`
}

// ConfigSummary represents configuration summary counts
type ConfigSummary struct {
	HomebrewPackages int    `json:"homebrew_packages" yaml:"homebrew_packages"`
	ASDFTools        int    `json:"asdf_tools" yaml:"asdf_tools"`
	NPMPackages      int    `json:"npm_packages" yaml:"npm_packages"`
	Dotfiles         int    `json:"dotfiles" yaml:"dotfiles"`
	DefaultManager   string `json:"default_manager" yaml:"default_manager"`
}

// TableOutput generates human-friendly table output for config show
func (c ConfigShowOutput) TableOutput() string {
	output := "Configuration Status\n===================\n\n"
	
	output += fmt.Sprintf("📁 Config File: %s\n", c.ConfigPath)
	
	switch c.Status {
	case "valid":
		output += "✅ Status: Valid\n\n"
		output += "Package Summary:\n"
		output += fmt.Sprintf("  • Homebrew: %d packages\n", c.Summary.HomebrewPackages)
		output += fmt.Sprintf("  • ASDF: %d tools\n", c.Summary.ASDFTools)
		output += fmt.Sprintf("  • NPM: %d packages\n", c.Summary.NPMPackages)
		output += fmt.Sprintf("  • Dotfiles: %d files\n", c.Summary.Dotfiles)
		output += fmt.Sprintf("  • Default Manager: %s\n", c.Summary.DefaultManager)
	case "invalid":
		output += "❌ Status: Invalid\n"
		output += fmt.Sprintf("💬 %s\n", c.Message)
	case "missing":
		output += "📋 Status: Missing\n"
		output += fmt.Sprintf("💬 %s\n", c.Message)
	}
	
	return output
}

// StructuredData returns the structured data for serialization
func (c ConfigShowOutput) StructuredData() any {
	return c
}