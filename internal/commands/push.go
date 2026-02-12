// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Commit pending changes and push to remote",
	Long: `Commit any uncommitted changes in your plonk directory and push to the remote.

If auto_commit is enabled, pending changes are committed before pushing.
If auto_commit is disabled, only already-committed work is pushed.

Examples:
  plonk push    # Commit and push`,
	RunE:         runPush,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
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

	// If dirty, commit if auto_commit is enabled; otherwise just warn and push committed work
	dirty, err := client.IsDirty(ctx)
	if err != nil {
		return err
	}
	if dirty {
		cfg := config.LoadWithDefaults(configDir)
		if cfg.AutoCommitEnabled() {
			msg := gitops.CommitMessage("push", nil)
			if err := client.Commit(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}
			output.Println("Committed pending changes")
		} else {
			output.Println("Warning: uncommitted changes exist but auto_commit is disabled; pushing committed work only")
		}
	}

	// Push
	output.Println("Pushing to remote...")
	if err := client.Push(ctx); err != nil {
		return err
	}
	output.Println("Push complete")
	return nil
}
