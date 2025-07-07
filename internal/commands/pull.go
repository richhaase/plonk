// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"plonk/internal/directories"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull updates to dotfiles repository",
	Long: `Pull updates to the existing dotfiles repository.

Requires an existing repository in the plonk directory. Use 'plonk clone' first
if you haven't cloned a repository yet.

Examples:
  plonk pull                                      # Pull updates`,
	RunE: pullCmdRun,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func pullCmdRun(cmd *cobra.Command, args []string) error {
	dryRun := IsDryRun(cmd)
	return runPullWithOptions(args, dryRun)
}

func runPull(args []string) error {
	return runPullWithOptions(args, false)
}

func runPullWithOptions(args []string, dryRun bool) error {
	if err := ValidateNoArgs("pull", args); err != nil {
		return err
	}

	// Ensure directory structure exists and handle migration if needed
	if err := directories.Default.EnsureStructure(); err != nil {
		return fmt.Errorf("failed to setup directory structure: %w", err)
	}

	repoDir := directories.Default.RepoDir()

	// Check if repo directory exists and is a git repo
	if !gitClient.IsRepo(repoDir) {
		if dryRun {
			fmt.Printf("Dry-run mode: Showing what would happen when pulling repository updates\n\n")
			fmt.Printf("❌ No repository found in %s\n", repoDir)
			fmt.Printf("💡 Would need to use 'plonk clone <repo>' first\n")
			return nil
		}
		return fmt.Errorf("no repository found in %s, use 'plonk clone <repo>' first", repoDir)
	}

	if dryRun {
		fmt.Printf("Dry-run mode: Showing what would happen when pulling repository updates\n\n")
		fmt.Printf("📁 Repository directory: %s\n", repoDir)
		fmt.Printf("📥 Would pull latest updates from remote repository\n")
		fmt.Printf("🔄 Would update local files with any changes from remote\n")
		fmt.Printf("\nDry-run complete. No updates were pulled.\n")
		return nil
	}

	err := gitClient.Pull(repoDir)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully pulled updates in %s\n", repoDir)
	return nil
}
