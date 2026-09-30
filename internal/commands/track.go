// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	GroupID: "manage",
	Use:     "track <manager:package>...",
	Short:   "Track installed packages",
	Long: `Track packages that are already installed on your system.

This command verifies that each package is installed, then adds it to your
lock file for management. Use this to record packages you want to keep
in sync across machines.

The package must already be installed - track only records existing packages.

Examples:
  plonk track brew:ripgrep           # Track a brew package
  plonk track cargo:bat go:golang.org/x/tools/gopls # Track multiple packages
  plonk track pnpm:typescript        # Track a pnpm package`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runTrack,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(trackCmd)
}

func runTrack(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	lockSvc := lock.NewLockV3Service(configDir)
	ctx := cmd.Context()

	var tracked, skipped, failed int

	// Serialize the read-modify-write cycle against concurrent plonk processes
	err := lock.WithMutationLock(ctx, configDir, func() error {
		lockFile, err := lockSvc.Read()
		if err != nil {
			return fmt.Errorf("failed to read lock file: %w", err)
		}

		for _, arg := range args {
			manager, pkg, err := packages.ParsePackageSpec(arg)
			if err != nil {
				output.PrintAction("error", arg, err.Error())
				failed++
				continue
			}

			// Check if already tracked
			if lockFile.HasPackage(manager, pkg) {
				output.PrintAction("skipped", manager+":"+pkg, "already tracked")
				skipped++
				continue
			}

			// Get manager and verify package is installed
			mgr, err := packages.GetManager(manager)
			if err != nil {
				output.PrintAction("error", arg, err.Error())
				failed++
				continue
			}

			installed, err := mgr.IsInstalled(ctx, pkg)
			if err != nil {
				output.PrintAction("error", manager+":"+pkg, err.Error())
				failed++
				continue
			}

			if !installed {
				output.PrintAction("error", manager+":"+pkg, "not installed")
				failed++
				continue
			}

			// Add to lock file
			lockFile.AddPackage(manager, pkg)
			output.PrintAction("tracked", manager+":"+pkg, "added to plonk.lock")
			tracked++
		}

		// Write updated lock file
		if tracked > 0 {
			if err := lockSvc.Write(lockFile); err != nil {
				return fmt.Errorf("failed to write lock file: %w", err)
			}
			gitops.AutoCommit(ctx, configDir, "track", args)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Summary
	if failed > 0 {
		return fmt.Errorf("tracked %d, skipped %d, failed %d", tracked, skipped, failed)
	}

	return nil
}
