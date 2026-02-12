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

This stages all changes, creates a commit, and pushes to the default remote.
If there are no changes, only a push is performed.

Examples:
  plonk push    # Commit and push`,
	RunE:         runPush,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	client := gitops.New(configDir)

	if !client.IsRepo() {
		return fmt.Errorf("%s is not a git repository", configDir)
	}

	if !client.HasRemote() {
		return fmt.Errorf("no remote configured for %s", configDir)
	}

	// Commit any pending changes — no-ops if nothing to commit
	msg := gitops.CommitMessage("push", nil)
	if err := client.Commit(msg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push
	output.Println("Pushing to remote...")
	if err := client.Push(); err != nil {
		return err
	}
	output.Println("Push complete")
	return nil
}
