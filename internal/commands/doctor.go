// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/diagnostics"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

// No flags needed for doctor command anymore

var doctorCmd = &cobra.Command{
	GroupID: "inspect",
	Use:     "doctor",
	Short:   "Check system readiness for using plonk",
	Long: `Perform health checks to ensure your system is properly configured
for plonk. This includes checking for required package managers,
configuration files, and system compatibility.

Shows:
- System information (OS, arch, etc.)
- Package manager availability
- Configuration file status and location
- Environment variables (PLONK_DIR, etc.)
- Any issues that would prevent plonk from working

Doctor reports issues with suggestions on how to fix them.
Install any missing package managers and ensure their executables are on PATH.

Examples:
  plonk doctor    # Run health checks`,
	RunE:         runDoctor,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Build a context with configured operation timeout
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)
	t := config.GetTimeouts(cfg)
	ctx, cancel := context.WithTimeout(cmd.Context(), t.Operation)
	defer cancel()

	// Run comprehensive health checks using diagnostics with context
	healthReport := diagnostics.RunHealthChecksWithContext(ctx)

	output.RenderOutput(output.NewDoctorFormatter(healthReport))
	return nil
}
