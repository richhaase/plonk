// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull remote changes into plonk directory",
	Long: `Pull remote changes into your plonk directory.

If there are uncommitted local changes and auto_commit is enabled, they are
committed first to avoid conflicts. If auto_commit is disabled and there are
uncommitted changes, the pull is refused.

Use --apply to automatically run 'plonk apply' after pulling.

Examples:
  plonk pull            # Pull remote changes
  plonk pull --apply    # Pull and apply`,
	RunE:         runPull,
	SilenceUsage: true,
}

func init() {
	pullCmd.Flags().BoolP("apply", "a", false, "Run plonk apply after pulling")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	applyAfter, _ := cmd.Flags().GetBool("apply")
	ctx := cmd.Context()
	configDir := config.GetDefaultConfigDirectory()
	client := gitops.New(configDir)

	if !client.IsRepo() {
		return fmt.Errorf("%s is not a git repository", configDir)
	}

	hasRemote, err := client.HasRemote(ctx)
	if err != nil {
		return err
	}
	if !hasRemote {
		return fmt.Errorf("no remote configured for %s", configDir)
	}

	// Check for dirty state — commit or refuse depending on config
	dirty, err := client.IsDirty(ctx)
	if err != nil {
		return err
	}
	if dirty {
		cfg := config.LoadWithDefaults(configDir)
		if !cfg.AutoCommitEnabled() {
			return fmt.Errorf("uncommitted changes in %s; commit manually or enable git.auto_commit", configDir)
		}
		msg := gitops.CommitMessage("pre-pull snapshot", nil)
		if err := client.Commit(ctx, msg); err != nil {
			return fmt.Errorf("failed to commit local changes: %w", err)
		}
		output.Println("Committed local changes before pull")
	}

	// Pull
	output.Println("Pulling from remote...")
	if err := client.Pull(ctx); err != nil {
		return err
	}
	output.Println("Pull complete")

	// Optionally apply
	if applyAfter {
		output.Println("Applying configuration...")
		homeDir, err := config.GetHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		cfg := config.LoadWithDefaults(configDir)

		orch := orchestrator.New(
			orchestrator.WithConfig(cfg),
			orchestrator.WithConfigDir(configDir),
			orchestrator.WithHomeDir(homeDir),
			orchestrator.WithDryRun(false),
		)

		result, err := orch.Apply(ctx)
		output.RenderOutput(result)
		if err != nil {
			return err
		}
	}

	return nil
}
