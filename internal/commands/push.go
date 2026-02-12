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
	Short: "Push committed changes to remote",
	Long: `Push committed changes in your plonk directory to the remote.

Warns if there are uncommitted changes in the working tree.
Use mutation commands (add, rm, track, untrack) with auto_commit enabled
to commit changes automatically, or commit manually before pushing.

Examples:
  plonk push    # Push to remote`,
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

	// Warn if there are uncommitted changes
	dirty, err := client.IsDirty(ctx)
	if err != nil {
		return err
	}
	if dirty {
		output.Println("Warning: uncommitted changes in working tree")
	}

	// Push
	output.Println("Pushing to remote...")
	if err := client.Push(ctx); err != nil {
		return err
	}
	output.Println("Push complete")
	return nil
}
