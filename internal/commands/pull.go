package commands

import (
	"fmt"

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
	return runPull(args)
}

func runPull(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("pull command takes no arguments")
	}
	
	plonkDir := getPlonkDir()
	
	// Check if plonk directory exists and is a git repo
	if !gitClient.IsRepo(plonkDir) {
		return fmt.Errorf("no repository found in %s, use 'plonk clone <repo>' first", plonkDir)
	}
	
	err := gitClient.Pull(plonkDir)
	if err != nil {
		return err
	}
	
	fmt.Printf("Successfully pulled updates in %s\n", plonkDir)
	return nil
}