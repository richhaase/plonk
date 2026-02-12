// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client wraps git CLI operations on a specific directory.
type Client struct {
	dir string
}

// New creates a git client for the given directory.
func New(dir string) *Client {
	return &Client{dir: dir}
}

// IsRepo checks if dir itself is the root of a git work tree
// (i.e., has a .git directory/file directly inside it).
func (c *Client) IsRepo() bool {
	_, err := os.Stat(filepath.Join(c.dir, ".git"))
	return err == nil
}

// HasRemote checks if the repo has at least one remote configured.
func (c *Client) HasRemote() bool {
	cmd := exec.Command("git", "-C", c.dir, "remote")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// IsDirty returns true if there are uncommitted changes (staged, unstaged, or untracked).
func (c *Client) IsDirty() (bool, error) {
	cmd := exec.Command("git", "-C", c.dir, "status", "--porcelain", "--untracked-files=normal")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// Commit stages all changes and commits with the given message.
// Returns nil if there's nothing to commit.
func (c *Client) Commit(message string) error {
	// Stage everything
	addCmd := exec.Command("git", "-C", c.dir, "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	// Commit — if nothing staged, git returns exit 1 with "nothing to commit".
	// That's not an error for us.
	commitCmd := exec.Command("git", "-C", c.dir, "commit", "-m", message)
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

// Push pushes to the default remote/branch.
// Sets GIT_TERMINAL_PROMPT=0 to avoid hanging on credential prompts.
func (c *Client) Push() error {
	cmd := exec.Command("git", "-C", c.dir, "push")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}
	return nil
}

// Pull pulls from the default remote/branch using merge (never rebase).
// Uses --no-rebase to be explicit regardless of user's global git config,
// and --no-edit to avoid opening an editor for merge commits.
// Sets GIT_TERMINAL_PROMPT=0 to avoid hanging on credential prompts.
func (c *Client) Pull() error {
	cmd := exec.Command("git", "-C", c.dir, "pull", "--no-rebase", "--no-edit")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, out)
	}
	return nil
}

// CommitMessage builds a commit message from a plonk command and its arguments.
func CommitMessage(command string, args []string) string {
	if len(args) == 0 {
		return fmt.Sprintf("plonk: %s", command)
	}
	display := args
	suffix := ""
	if len(display) > 5 {
		display = display[:5]
		suffix = fmt.Sprintf(" (+%d more)", len(args)-5)
	}
	return fmt.Sprintf("plonk: %s %s%s", command, strings.Join(display, " "), suffix)
}
