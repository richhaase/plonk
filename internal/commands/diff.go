// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources"
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
	Args: cobra.MaximumNArgs(1),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()
	cfg := config.LoadWithDefaults(configDir)

	// Get drifted dotfiles from reconciliation
	driftedFiles, err := getDriftedDotfiles(cmd.Context(), cfg, configDir, homeDir)
	if err != nil {
		return fmt.Errorf("failed to get drifted files: %w", err)
	}

	if len(driftedFiles) == 0 {
		fmt.Println("No drifted dotfiles found")
		return nil
	}

	// Filter by argument if provided
	if len(args) > 0 {
		filtered := filterDriftedFile(args[0], driftedFiles)
		if filtered == nil {
			return fmt.Errorf("dotfile not found or not drifted: %s", args[0])
		}
		driftedFiles = []resources.Item{*filtered}
	}

	// Get diff tool from config or use default
	diffTool := cfg.DiffTool
	if diffTool == "" {
		diffTool = "git diff --no-index"
	}

	// Execute diff for each drifted file
	for _, item := range driftedFiles {
		// Get the source name (without leading dot)
		sourceName := getSourceNameFromItem(item)
		sourcePath := filepath.Join(configDir, sourceName)

		// Get destination path from metadata
		destPath := ""
		if dest, ok := item.Metadata["destination"].(string); ok {
			normalizedDest, err := normalizePath(dest)
			if err != nil {
				// Fall back to simple expansion if normalization fails
				destPath = expandHome(dest)
			} else {
				destPath = normalizedDest
			}
		} else {
			// This shouldn't happen, but provide a fallback
			destPath = expandHome("~/" + item.Name)
		}

		if err := executeDiffTool(diffTool, sourcePath, destPath); err != nil {
			// Report error but continue with other files
			fmt.Fprintf(os.Stderr, "Error showing diff for %s: %v\n", item.Name, err)
		}
	}

	return nil
}

// getDriftedDotfiles reconciles dotfiles and returns only drifted ones
func getDriftedDotfiles(ctx context.Context, cfg *config.Config, configDir, homeDir string) ([]resources.Item, error) {
	// Reconcile all domains
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return nil, err
	}

	// Convert results to summary to get all items
	summary := resources.ConvertResultsToSummary(results)

	var drifted []resources.Item
	// Find dotfile results
	for _, result := range summary.Results {
		if result.Domain == "dotfile" {
			// Check all managed items for drift
			for _, item := range result.Managed {
				if item.State == resources.StateDegraded {
					drifted = append(drifted, item)
				}
			}
		}
	}

	return drifted, nil
}

// filterDriftedFile finds a specific drifted file from the list
func filterDriftedFile(arg string, driftedFiles []resources.Item) *resources.Item {
	// Normalize the argument path
	argPath, err := normalizePath(arg)
	if err != nil {
		// If we can't normalize, we won't find a match
		return nil
	}

	for _, item := range driftedFiles {
		// Get the deployed path from metadata
		if dest, ok := item.Metadata["destination"].(string); ok {
			deployedPath, err := normalizePath(dest)
			if err != nil {
				continue
			}
			if deployedPath == argPath {
				return &item
			}
		}
	}
	return nil
}

// getSourceNameFromItem extracts the proper source name from a dotfile item
// item.Name is the dotfile name with leading dot (e.g., ".zshrc")
// The source file in PLONK_DIR has no leading dot (e.g., "zshrc")
func getSourceNameFromItem(item resources.Item) string {
	// Check metadata first
	if source, ok := item.Metadata["source"].(string); ok {
		return source
	}

	// Fallback: remove leading dot from item.Name
	if strings.HasPrefix(item.Name, ".") {
		return item.Name[1:]
	}
	return item.Name
}

// executeDiffTool runs the configured diff tool
func executeDiffTool(tool string, source, dest string) error {
	// Split the tool command in case it has flags (e.g., "git diff --no-index")
	parts := strings.Fields(tool)
	if len(parts) == 0 {
		return fmt.Errorf("invalid diff tool: %s", tool)
	}

	// Append source and destination paths
	args := append(parts[1:], source, dest)

	cmd := exec.Command(parts[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command
	if err := cmd.Run(); err != nil {
		// Check if it's just a non-zero exit code (common for diff tools)
		if _, ok := err.(*exec.ExitError); ok {
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
