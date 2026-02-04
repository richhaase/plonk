// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show differences for drifted dotfiles",
	Long: `Show differences between source and deployed dotfiles that have drifted.

With no arguments, shows diffs for all drifted dotfiles.
With a file argument, shows diff for that specific file only.

Examples:
  plonk diff                # Show all drifted files
  plonk diff ~/.vimrc       # Show diff for specific file
  plonk diff vimrc          # Use config name directly`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runDiff,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Get drifted dotfiles from reconciliation
	driftedFiles, err := getDriftedDotfileStatuses(cfg, configDir, homeDir)
	if err != nil {
		return fmt.Errorf("failed to get drifted files: %w", err)
	}

	if len(driftedFiles) == 0 {
		output.Println("No drifted dotfiles found")
		return nil
	}

	// Filter by argument if provided
	if len(args) > 0 {
		filtered := filterDriftedStatus(args[0], driftedFiles)
		if filtered == nil {
			return fmt.Errorf("dotfile not found or not drifted: %s", args[0])
		}
		driftedFiles = []dotfiles.DotfileStatus{*filtered}
	}

	// Get diff tool from config or use default
	diffTool := cfg.DiffTool
	if diffTool == "" {
		diffTool = "git diff --no-index"
	}

	// Execute diff for each drifted file
	var diffErrors []string
	for _, status := range driftedFiles {
		sourcePath := status.Source
		destPath := status.Target

		if err := executeDiffTool(diffTool, sourcePath, destPath); err != nil {
			// Report error but continue with other files
			fmt.Fprintf(os.Stderr, "Error showing diff for %s: %v\n", status.Name, err)
			diffErrors = append(diffErrors, status.Name)
		}
	}

	if len(diffErrors) > 0 {
		return fmt.Errorf("failed to show diff for %d file(s): %v", len(diffErrors), diffErrors)
	}
	return nil
}

// getDriftedDotfileStatuses reconciles dotfiles and returns only drifted ones
func getDriftedDotfileStatuses(cfg *config.Config, configDir, homeDir string) ([]dotfiles.DotfileStatus, error) {
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return nil, err
	}

	var drifted []dotfiles.DotfileStatus
	for _, s := range statuses {
		if s.State == dotfiles.SyncStateDrifted {
			drifted = append(drifted, s)
		}
	}

	return drifted, nil
}

// filterDriftedStatus finds a specific drifted file from the list
func filterDriftedStatus(arg string, driftedFiles []dotfiles.DotfileStatus) *dotfiles.DotfileStatus {
	// Normalize the argument path
	argPath, err := normalizePath(arg)
	if err != nil {
		// If we can't normalize, we won't find a match
		return nil
	}

	for i := range driftedFiles {
		status := &driftedFiles[i]
		// Compare against the target path
		if status.Target != "" {
			targetPath, err := normalizePath(status.Target)
			if err != nil {
				continue
			}
			if targetPath == argPath {
				return status
			}
		}
		// Also check against the Name for shorthand matching (e.g., "vimrc" for ~/.vimrc)
		if status.Name == arg {
			return status
		}
	}
	return nil
}

// executeDiffTool runs the configured diff tool
func executeDiffTool(tool string, source, dest string) error {
	// Split the tool command in case it has flags (e.g., "git diff --no-index")
	parts := strings.Fields(tool)
	if len(parts) == 0 {
		return fmt.Errorf("invalid diff tool: %s", tool)
	}

	// Append destination and source paths (shows $HOME on left, $PLONKDIR on right)
	args := append(parts[1:], dest, source)

	//nolint:gosec // G204: diff tool from user config (cfg.DiffTool) - intentional user control like $EDITOR
	cmd := exec.Command(parts[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command
	if err := cmd.Run(); err != nil {
		// Check if it's just a non-zero exit code (common for diff tools)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// This is expected for diff tools when files differ
			return nil
		}
		return fmt.Errorf("diff tool failed: %w", err)
	}

	return nil
}

// expandHome expands ~ to home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// normalizePath resolves a path to its absolute form, handling ~, $HOME, and relative paths
func normalizePath(path string) (string, error) {
	// First expand any environment variables (e.g., $HOME, $ZSHPATH)
	path = os.ExpandEnv(path)

	// Then expand tilde
	path = expandHome(path)

	// Finally, convert to absolute path (handles relative paths)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	// Clean the path to remove any redundant elements
	return filepath.Clean(absPath), nil
}
