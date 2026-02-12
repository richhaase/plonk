// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package gitops

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
)

// autoCommitTimeout is the maximum time allowed for an auto-commit operation.
const autoCommitTimeout = 30 * time.Second

// AutoCommit is the standard post-mutation hook.
// Reads config from disk, checks if auto-commit is enabled, and commits if appropriate.
// Uses a 30-second timeout to prevent hanging on broken repos.
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

	ctx, cancel := context.WithTimeout(context.Background(), autoCommitTimeout)
	defer cancel()

	msg := CommitMessage(command, args)
	if err := client.Commit(ctx, msg); err != nil {
		output.Printf("Warning: auto-commit failed: %v\n", err)
	}
}
