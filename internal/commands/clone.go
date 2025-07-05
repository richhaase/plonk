package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// GitInterface defines git operations for dependency injection
type GitInterface interface {
	Clone(repoURL, targetDir string) error
	Pull(repoDir string) error
	IsRepo(dir string) bool
}

// RealGit implements GitInterface using go-git
type RealGit struct{}

func (g *RealGit) Clone(repoURL, targetDir string) error {
	// Create parent directory if needed
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	
	// Clone the repository
	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	return nil
}

func (g *RealGit) Pull(repoDir string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	
	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Pull updates
	err = worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull updates: %w", err)
	}
	
	return nil
}

func (g *RealGit) IsRepo(dir string) bool {
	_, err := git.PlainOpen(dir)
	return err == nil
}

// Global git instance (can be mocked for testing)
var gitClient GitInterface = &RealGit{}

// getPlonkDir returns the plonk directory location
// Defaults to ~/.config/plonk but can be overridden with PLONK_DIR environment variable
func getPlonkDir() string {
	if dir := os.Getenv("PLONK_DIR"); dir != "" {
		return dir
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return ".plonk"
	}
	
	return filepath.Join(homeDir, ".config", "plonk")
}

var cloneCmd = &cobra.Command{
	Use:   "clone <repository>",
	Short: "Clone dotfiles repository",
	Long: `Clone a dotfiles repository to the plonk directory.

The clone location defaults to ~/.config/plonk/ but can be customized by setting 
the PLONK_DIR environment variable.

Examples:
  plonk clone git@github.com/user/dotfiles.git    # Clone to ~/.config/plonk/
  PLONK_DIR=~/my-dotfiles plonk clone <repo>       # Clone to custom location`,
	RunE: cloneCmdRun,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func cloneCmdRun(cmd *cobra.Command, args []string) error {
	return runClone(args)
}

func runClone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("repository URL is required")
	}
	
	plonkDir := getPlonkDir()
	
	// Check if target directory already exists
	if _, err := os.Stat(plonkDir); err == nil {
		return fmt.Errorf("target directory %s already exists, use 'plonk pull' to update", plonkDir)
	}
	
	repoURL := args[0]
	err := gitClient.Clone(repoURL, plonkDir)
	if err != nil {
		return err
	}
	
	fmt.Printf("Successfully cloned %s to %s\n", repoURL, plonkDir)
	return nil
}