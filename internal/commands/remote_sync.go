// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/gitops"
)

// getRemoteSyncStatus fetches from the remote and returns a human-readable
// sync status string. Returns "" if not a git repo, no remote is configured,
// or no upstream tracking branch exists (graceful degradation).
func getRemoteSyncStatus(ctx context.Context, configDir string) string {
	client := gitops.New(configDir)
	if !client.IsRepo() {
		return ""
	}

	hasRemote, err := client.HasRemote(ctx)
	if err != nil || !hasRemote {
		return ""
	}

	status, err := client.RemoteStatus(ctx)
	if err != nil {
		return ""
	}

	return status.String()
}
