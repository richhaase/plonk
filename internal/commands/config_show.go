// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config" // getConfigPath returns the path to the main configuration file
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func getConfigPath(configDir string) string {
	return filepath.Join(configDir, "plonk.yaml")
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display effective configuration",
	Long: `Display the effective plonk configuration (defaults merged with user settings).

Shows the complete configuration that plonk is actually using, including all default
values merged with any user-specified overrides from the config file.

Examples:
  plonk config show           # Show effective configuration
  plonk config show -o json   # Show as JSON
  plonk config show -o yaml   # Show as YAML (default)`,
	RunE: runConfigShow,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get config directory
	configDir := config.GetDefaultConfigDirectory()
	configPath := getConfigPath(configDir)

	// Load configuration (this handles missing files gracefully due to zero-config)
	cfg := config.LoadWithDefaults(configDir)

	// Build output data
	outputData := ConfigShowOutput{
		ConfigPath: configPath,
		Config:     cfg,
	}

	return RenderOutput(outputData, format)
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string         `json:"config_path" yaml:"config_path"`
	Config     *config.Config `json:"config" yaml:"config"`
}

// TableOutput generates human-friendly table output for config show
func (c ConfigShowOutput) TableOutput() string {
	output := fmt.Sprintf("Config file: %s\n\n", c.ConfigPath)

	if c.Config != nil {
		// Convert config to YAML for display (shows effective config with defaults)
		yamlBytes, err := yaml.Marshal(c.Config)
		if err != nil {
			return fmt.Sprintf("Error formatting configuration: %v\n", err)
		}
		output += string(yamlBytes)
	}

	return output
}

// StructuredData returns the structured data for serialization
func (c ConfigShowOutput) StructuredData() any {
	return c
}
