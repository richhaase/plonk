package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/directories"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

// GitInterface defines git operations for dependency injection
type GitInterface interface {
	Clone(repoURL, targetDir string) error
	CloneBranch(repoURL, targetDir, branch string) error
	Pull(repoDir string) error
	IsRepo(dir string) bool
}

// RealGit implements GitInterface using go-git
type RealGit struct{}

func (g *RealGit) Clone(repoURL, targetDir string) error {
	return g.CloneBranch(repoURL, targetDir, "")
}

func (g *RealGit) CloneBranch(repoURL, targetDir, branch string) error {
	// Create parent directory if needed
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Setup clone options
	cloneOptions := &git.CloneOptions{
		URL: repoURL,
	}

	// Add branch reference if specified
	if branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + branch)
		cloneOptions.SingleBranch = true
	}

	// Clone the repository
	_, err := git.PlainClone(targetDir, false, cloneOptions)
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

var cloneCmd = &cobra.Command{
	Use:   "clone <repository>",
	Short: "Clone dotfiles repository",
	Long: `Clone a dotfiles repository to the plonk directory.

The clone location defaults to ~/.config/plonk/ but can be customized by setting 
the PLONK_DIR environment variable.

Branch Support:
  plonk clone repo.git --branch develop           # Clone specific branch
  plonk clone repo.git#feature-branch             # Branch in URL
  plonk clone repo.git#branch --branch main       # Flag overrides URL

Examples:
  plonk clone git@github.com/user/dotfiles.git    # Clone to ~/.config/plonk/
  PLONK_DIR=~/my-dotfiles plonk clone <repo>       # Clone to custom location`,
	RunE: cloneCmdRun,
	Args: cobra.ExactArgs(1),
}

var branchFlag string

func init() {
	cloneCmd.Flags().StringVar(&branchFlag, "branch", "", "Branch to clone")
	rootCmd.AddCommand(cloneCmd)
}

func cloneCmdRun(cmd *cobra.Command, args []string) error {
	return runCloneWithBranch(args, branchFlag)
}

// parseRepoURL extracts the repository URL and branch from a URL that may contain #branch
func parseRepoURL(input string) (url, branch string) {
	parts := strings.SplitN(input, "#", 2)
	url = parts[0]
	if len(parts) > 1 {
		branch = parts[1]
	}
	return url, branch
}

func runClone(args []string) error {
	return runCloneWithBranch(args, "")
}

func runCloneWithBranch(args []string, flagBranch string) error {
	if err := ValidateExactArgs("clone", 1, args); err != nil {
		return err
	}

	// Ensure directory structure exists
	if err := directories.Default.EnsureStructure(); err != nil {
		return fmt.Errorf("failed to setup directory structure: %w", err)
	}

	repoDir := directories.Default.RepoDir()

	// Check if target directory already exists and has content
	if entries, err := os.ReadDir(repoDir); err == nil && len(entries) > 0 {
		return fmt.Errorf("target directory %s already contains files, use 'plonk pull' to update", repoDir)
	}

	// Parse URL for branch information
	repoURL, urlBranch := parseRepoURL(args[0])

	// Flag takes precedence over URL branch
	branch := flagBranch
	if branch == "" {
		branch = urlBranch
	}

	// Use appropriate clone method
	var err error
	if branch != "" {
		err = gitClient.CloneBranch(repoURL, repoDir, branch)
	} else {
		err = gitClient.Clone(repoURL, repoDir)
	}

	if err != nil {
		return err
	}

	if branch != "" {
		fmt.Printf("Successfully cloned %s (branch: %s) to %s\n", repoURL, branch, repoDir)
	} else {
		fmt.Printf("Successfully cloned %s to %s\n", repoURL, repoDir)
	}
	return nil
}
