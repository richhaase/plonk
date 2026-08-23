// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
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

	// Create DotfileManager for rendering templates
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)

	// Execute diff for each drifted file
	var diffErrors []string
	for _, status := range driftedFiles {
		sourcePath := status.Source
		destPath := status.Target
		var cleanupPaths []string

		// For template files, render to a temp file so the external diff tool
		// sees rendered content instead of raw {{VAR}} placeholders. Secret-bearing
		// templates are never written unmasked: both sides are masked with
		// [REDACTED_SECRET] before being handed to the diff tool.
		if strings.HasSuffix(status.Name, ".tmpl") {
			hasSecrets, err := dm.HasSecrets(status.Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error inspecting template %s: %v\n", status.Name, err)
				diffErrors = append(diffErrors, status.Name)
				continue
			}
			if hasSecrets {
				targetContent, err := os.ReadFile(destPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading target %s: %v\n", destPath, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				maskedSource, maskedTarget, err := dm.RenderForDiff(status.Name, targetContent)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error rendering template %s: %v\n", status.Name, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				tmp, err := writeTempDiffFile(status.Name, maskedSource)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating temp file for %s: %v\n", status.Name, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				cleanupPaths = append(cleanupPaths, tmp)
				sourcePath = tmp
				tmp, err = writeTempDiffFile(status.Name, maskedTarget)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating temp file for %s: %v\n", status.Name, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				cleanupPaths = append(cleanupPaths, tmp)
				destPath = tmp
			} else {
				rendered, err := dm.RenderSource(status.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error rendering template %s: %v\n", status.Name, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				tmp, err := writeTempDiffFile(status.Name, rendered)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating temp file for %s: %v\n", status.Name, err)
					diffErrors = append(diffErrors, status.Name)
					continue
				}
				cleanupPaths = append(cleanupPaths, tmp)
				sourcePath = tmp
			}
		}

		if err := executeDiffTool(cmd.Context(), diffTool, sourcePath, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error showing diff for %s: %v\n", status.Name, err)
			diffErrors = append(diffErrors, status.Name)
		}

		for _, p := range cleanupPaths {
			removeTempFile(p)
		}
	}

	if len(diffErrors) > 0 {
		return fmt.Errorf("failed to show diff for %d file(s): %v", len(diffErrors), diffErrors)
	}
	return nil
}

// writeTempDiffFile writes content to a temp file for use by an external diff tool.
// The caller is responsible for removing the returned path.
func writeTempDiffFile(name string, content []byte) (string, error) {
	base := strings.TrimSuffix(name, ".tmpl")
	base = strings.ReplaceAll(base, string(os.PathSeparator), "-")
	if base == "" {
		base = "file"
	}
	tmpFile, err := os.CreateTemp("", "plonk-diff-"+base+"-*.rendered")
	if err != nil {
		return "", err
	}
	tmpPath := filepath.Clean(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func removeTempFile(path string) {
	//nolint:gosec // G703: path comes from os.CreateTemp in writeTempDiffFile
	os.Remove(path)
}

// getDriftedDotfileStatuses reconciles dotfiles and returns only drifted ones.
// Files that failed reconciliation are reported to stderr so users know
// why certain files are absent from the diff output.
func getDriftedDotfileStatuses(cfg *config.Config, configDir, homeDir string) ([]dotfiles.DotfileStatus, error) {
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return nil, err
	}

	var drifted []dotfiles.DotfileStatus
	var errorCount int
	for _, s := range statuses {
		switch s.State {
		case dotfiles.SyncStateDrifted:
			drifted = append(drifted, s)
		case dotfiles.SyncStateError:
			fmt.Fprintf(os.Stderr, "Warning: could not check %s: %v\n", s.Name, s.Error)
			errorCount++
		}
	}

	if errorCount > 0 && len(drifted) == 0 {
		return nil, fmt.Errorf("%d file(s) could not be checked for drift", errorCount)
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
		// For template files, also match without the .tmpl suffix (e.g., "gitconfig" matches "gitconfig.tmpl")
		if status.Name == arg || strings.TrimSuffix(status.Name, ".tmpl") == arg {
			return status
		}
	}
	return nil
}

// executeDiffTool runs the configured diff tool. Per the documented diff
// convention (diff(1), git diff --no-index), exit status 1 means the files
// differ; any other non-zero exit (e.g. diff(1) status 2 for trouble, a tool
// crash, or signal death) is treated as a failure and propagated.
func executeDiffTool(ctx context.Context, tool string, source, dest string) error {
	// Split the tool command in case it has flags (e.g., "git diff --no-index")
	parts := strings.Fields(tool)
	if len(parts) == 0 {
		return fmt.Errorf("invalid diff tool: %s", tool)
	}

	// Append destination and source paths (shows $HOME on left, $PLONKDIR on right)
	args := append(parts[1:], dest, source)

	//nolint:gosec // G204: diff tool from user config (cfg.DiffTool) - intentional user control like $EDITOR
	cmd := exec.CommandContext(ctx, parts[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// Exit status 1 is the documented "files differ" result
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
