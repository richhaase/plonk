// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingDeployFS struct {
	*MemoryFS
}

func (f failingDeployFS) WriteFile(path string, data []byte, mode os.FileMode) error {
	if strings.HasPrefix(path, "/home/user/.broken.") {
		return errors.New("write denied")
	}
	return f.MemoryFS.WriteFile(path, data, mode)
}

func TestApplyStatusesPartialFailure(t *testing.T) {
	for _, state := range []SyncState{SyncStateMissing, SyncStateDrifted} {
		for _, dryRun := range []bool{false, true} {
			name := string(state)
			if dryRun {
				name += "/dry-run"
			}
			t.Run(name, func(t *testing.T) {
				fs := failingDeployFS{NewMemoryFS()}
				fs.Dirs["/config"], fs.Dirs["/home/user"] = true, true
				fs.Files["/config/broken"] = []byte("new")
				fs.Files["/config/good"] = []byte("new")
				if state == SyncStateDrifted {
					fs.Files["/home/user/.broken"] = []byte("old")
					fs.Files["/home/user/.good"] = []byte("old")
				}
				manager := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
				statuses, err := manager.Reconcile()
				require.NoError(t, err)
				result, err := applyStatuses(context.Background(), manager, statuses, dryRun)
				require.Len(t, result.Actions, 2)
				if dryRun {
					require.NoError(t, err)
					assert.Equal(t, 0, result.Summary.Failed)
					assert.Equal(t, 2, result.Summary.Added+result.Summary.Updated)
					assert.NotEqual(t, "new", string(fs.Files["/home/user/.good"]))
				} else {
					require.EqualError(t, err, "failed to deploy 1 file(s)")
					assert.Equal(t, 1, result.Summary.Failed)
					assert.Equal(t, 1, result.Summary.Added+result.Summary.Updated)
					assert.Equal(t, "new", string(fs.Files["/home/user/.good"]))
					for _, action := range result.Actions {
						if action.Destination == "/home/user/.broken" {
							assert.Equal(t, "failed", action.Status)
							assert.Equal(t, "failed to write temp file: write denied", action.Error)
						}
					}
				}
				assert.NotEqual(t, "new", string(fs.Files["/home/user/.broken"]))
			})
		}
	}
}

func TestApplyStatusesCancellation(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"], fs.Dirs["/home/user"] = true, true
	fs.Files["/config/zshrc"] = []byte("content")
	manager := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	statuses, err := manager.Reconcile()
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := applyStatuses(ctx, manager, statuses, false)
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, result.Actions)
	assert.NotContains(t, fs.Files, "/home/user/.zshrc")
}
