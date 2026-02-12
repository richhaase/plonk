// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package gitops

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
)

// AutoCommit is the standard post-mutation hook.
// Reads config from disk, checks if auto-commit is enabled, and commits if appropriate.
// Errors are warnings, never fatal.
func AutoCommit(configDir string, command string, args []string) {
	cfg := config.LoadWithDefaults(configDir)
	if !cfg.AutoCommitEnabled() {
		return
	}

	client := New(configDir)

	if !client.IsRepo() {
		output.Printf("Warning: %s is not a git repository; changes not committed. Set git.auto_commit: false in plonk.yaml to silence this warning.\n", configDir)
		return
	}

	msg := CommitMessage(command, args)
	if err := client.Commit(context.Background(), msg); err != nil {
		output.Printf("Warning: auto-commit failed: %v\n", err)
	}
}
