package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"plonk/internal/directories"
)

// repoCmd represents the repo command (convenience command).
var repoCmd = &cobra.Command{
	Use:   "repo <repository>",
	Short: "Complete setup: clone/pull repository, install packages, and apply configurations",
	Long: `Convenience command that performs the complete setup process:

1. Clone repository (or pull if already exists)
2. Install all packages from configuration
3. Apply all configuration files

This is equivalent to running:
  plonk clone <repository>  (or plonk pull if repo exists)
  plonk install
  plonk apply

Examples:
  plonk repo git@github.com/user/dotfiles.git     # Complete setup from repository
  plonk git@github.com/user/dotfiles.git          # Same as above (convenience)`,
	RunE: repoCmdRun,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(repoCmd)
}

func repoCmdRun(cmd *cobra.Command, args []string) error {
	return runRepo(args)
}

func runRepo(args []string) error {
	if err := ValidateExactArgs("repo", 1, args); err != nil {
		return err
	}

	repoURL := args[0]

	// Ensure directory structure exists and handle migration if needed.
	if err := directories.Default.EnsureStructure(); err != nil {
		return fmt.Errorf("failed to setup directory structure: %w", err)
	}

	repoDir := directories.Default.RepoDir()

	// Step 1: Clone or pull repository.
	fmt.Println("Step 1: Setting up repository...")
	if gitClient.IsRepo(repoDir) {
		// Repository exists, pull updates.
		fmt.Printf("Repository exists, pulling updates...\n")
		if err := gitClient.Pull(repoDir); err != nil {
			return fmt.Errorf("failed to pull repository: %w", err)
		}
		fmt.Printf("Successfully pulled updates in %s\n", repoDir)
	} else {
		// No repository, clone it.
		fmt.Printf("Cloning repository %s...\n", repoURL)
		if err := gitClient.Clone(repoURL, repoDir); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		fmt.Printf("Successfully cloned %s to %s\n", repoURL, repoDir)
	}

	// Step 2: Install packages.
	fmt.Println("\nStep 2: Installing packages...")
	if err := runInstall([]string{}); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	// Step 3: Apply configurations (only global dotfiles, as install already applied package configs).
	fmt.Println("\nStep 3: Applying remaining configurations...")
	if err := runApply([]string{}); err != nil {
		return fmt.Errorf("failed to apply configurations: %w", err)
	}

	fmt.Printf("\n✅ Repository setup complete! Your environment is ready.\n")
	return nil
}
