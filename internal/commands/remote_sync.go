// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/gitops"
)

const fetchTimeout = 5 * time.Second

// getRemoteSyncStatus fetches from the remote and returns a human-readable
// sync status string. Returns "" if not a git repo, no remote is configured,
// no upstream tracking branch exists, or the fetch times out (graceful degradation).
func getRemoteSyncStatus(ctx context.Context, configDir string) string {
	client := gitops.New(configDir)
	if !client.IsRepo() {
		return ""
	}

	hasRemote, err := client.HasRemote(ctx)
	if err != nil || !hasRemote {
		return ""
	}

	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	status, err := client.RemoteStatus(fetchCtx)
	if err != nil || status == nil {
		return ""
	}

	return status.String()
}
