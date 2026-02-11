// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package gitops

import (
	"fmt"
	"os/exec"
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

// IsRepo checks if dir is inside a git work tree.
func (c *Client) IsRepo() bool {
	cmd := exec.Command("git", "-C", c.dir, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// HasRemote checks if the repo has at least one remote configured.
func (c *Client) HasRemote() bool {
	cmd := exec.Command("git", "-C", c.dir, "remote")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// IsDirty returns true if there are uncommitted changes (staged or unstaged).
func (c *Client) IsDirty() (bool, error) {
	cmd := exec.Command("git", "-C", c.dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// Commit stages all changes and commits with the given message.
// Returns nil if there's nothing to commit.
func (c *Client) Commit(message string) error {
	dirty, err := c.IsDirty()
	if err != nil {
		return err
	}
	if !dirty {
		return nil
	}

	addCmd := exec.Command("git", "-C", c.dir, "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	commitCmd := exec.Command("git", "-C", c.dir, "commit", "-m", message)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

// Push pushes to the default remote/branch.
func (c *Client) Push() error {
	cmd := exec.Command("git", "-C", c.dir, "push")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}
	return nil
}

// Pull pulls from the default remote/branch using merge.
func (c *Client) Pull() error {
	cmd := exec.Command("git", "-C", c.dir, "pull")
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
