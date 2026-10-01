// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupRejectsInvalidConfigBeforeApply(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte("operation_timeout: invalid\n"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "private.tmpl"), []byte("ordinary fixture"), 0644))
	require.ErrorContains(t, SetupFromClonedRepo(context.Background(), dir, true), "failed to load configuration")
	require.NoFileExists(t, filepath.Join(home, ".private"))
}

func TestManagerDetectionDoesNotMigrateLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plonk.lock")
	data := []byte("version: 2\nresources: []\n")
	require.NoError(t, os.WriteFile(path, data, 0600))
	_, err := DetectRequiredManagers(path)
	require.NoError(t, err)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, after)
}
