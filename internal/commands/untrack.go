// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack <manager:package>...",
	Short: "Stop tracking packages",
	Long: `Stop tracking packages without uninstalling them.

This command removes packages from your lock file but does NOT uninstall
them from your system. The packages remain installed, they're just no
longer managed by plonk.

Examples:
  plonk untrack brew:ripgrep           # Stop tracking a brew package
  plonk untrack cargo:bat go:gopls     # Stop tracking multiple packages`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runUntrack,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(untrackCmd)
}

func runUntrack(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	lockSvc := lock.NewLockV3Service(configDir)

	lockFile, err := lockSvc.Read()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	var untracked, skipped, failed int

	for _, arg := range args {
		manager, pkg, err := packages.ParsePackageSpec(arg)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		// Check if tracked
		if !lockFile.HasPackage(manager, pkg) {
			fmt.Printf("Skipping %s:%s (not tracked)\n", manager, pkg)
			skipped++
			continue
		}

		// Remove from lock file
		lockFile.RemovePackage(manager, pkg)
		fmt.Printf("Untracking %s:%s\n", manager, pkg)
		untracked++
	}

	// Write updated lock file
	if untracked > 0 {
		if err := lockSvc.Write(lockFile); err != nil {
			return fmt.Errorf("failed to write lock file: %w", err)
		}
	}

	// Summary
	if failed > 0 {
		return fmt.Errorf("untracked %d, skipped %d, failed %d", untracked, skipped, failed)
	}

	return nil
}
