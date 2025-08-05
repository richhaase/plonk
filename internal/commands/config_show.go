// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
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
	RunE:         runConfigShow,
	SilenceUsage: true,
	Args:         cobra.NoArgs,
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
		Checker:    config.NewUserDefinedChecker(configDir),
		ConfigDir:  configDir,
	}

	// Convert to output package type and create formatter
	formatterData := output.ConfigShowOutput{
		ConfigPath: outputData.ConfigPath,
		Config:     outputData.Config,
		Checker:    outputData.Checker,
		ConfigDir:  outputData.ConfigDir,
	}
	formatter := output.NewConfigShowFormatter(formatterData)
	return RenderOutput(formatter, format)
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string                     `json:"config_path" yaml:"config_path"`
	Config     *config.Config             `json:"config" yaml:"config"`
	Checker    *config.UserDefinedChecker `json:"-" yaml:"-"` // Not included in JSON/YAML
	ConfigDir  string                     `json:"-" yaml:"-"` // Not included in JSON/YAML
}
