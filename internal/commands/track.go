// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track <manager:package>...",
	Short: "Track installed packages",
	Long: `Track packages that are already installed on your system.

This command verifies that each package is installed, then adds it to your
lock file for management. Use this to record packages you want to keep
in sync across machines.

The package must already be installed - track only records existing packages.

Examples:
  plonk track brew:ripgrep           # Track a brew package
  plonk track cargo:bat go:gopls     # Track multiple packages
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

	lockFile, err := lockSvc.Read()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	ctx := context.Background()
	var tracked, skipped, failed int

	for _, arg := range args {
		manager, pkg, err := parsePackageSpec(arg)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		// Check if already tracked
		if lockFile.HasPackage(manager, pkg) {
			fmt.Printf("Skipping %s:%s (already tracked)\n", manager, pkg)
			skipped++
			continue
		}

		// Get manager and verify package is installed
		mgr, err := packages.GetManager(manager)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		installed, err := mgr.IsInstalled(ctx, pkg)
		if err != nil {
			fmt.Printf("Error checking %s:%s: %v\n", manager, pkg, err)
			failed++
			continue
		}

		if !installed {
			fmt.Printf("Error: %s:%s is not installed\n", manager, pkg)
			failed++
			continue
		}

		// Add to lock file
		lockFile.AddPackage(manager, pkg)
		fmt.Printf("Tracking %s:%s\n", manager, pkg)
		tracked++
	}

	// Write updated lock file
	if tracked > 0 {
		if err := lockSvc.Write(lockFile); err != nil {
			return fmt.Errorf("failed to write lock file: %w", err)
		}
	}

	// Summary
	if failed > 0 {
		return fmt.Errorf("tracked %d, skipped %d, failed %d", tracked, skipped, failed)
	}

	return nil
}

// parsePackageSpec parses "manager:package" format
func parsePackageSpec(spec string) (manager, pkg string, err error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected manager:package")
	}

	manager = parts[0]
	pkg = parts[1]

	if !packages.IsSupportedManager(manager) {
		return "", "", fmt.Errorf("unsupported manager: %s (supported: %v)", manager, packages.SupportedManagers)
	}

	if pkg == "" {
		return "", "", fmt.Errorf("package name cannot be empty")
	}

	return manager, pkg, nil
}
