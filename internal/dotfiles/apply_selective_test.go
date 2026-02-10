// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplySelective_NormalizesTargetBeforeFilterLookup(t *testing.T) {
	configDir := t.TempDir()
	homeRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeRoot, "nested"), 0755))

	// Use a non-normalized home path so reconciled targets include "..".
	homeDir := filepath.Join(homeRoot, "nested", "..")

	sourcePath := filepath.Join(configDir, "zshrc")
	require.NoError(t, os.WriteFile(sourcePath, []byte("export TEST=1\n"), 0644))

	normalizedTarget := filepath.Clean(filepath.Join(homeRoot, ".zshrc"))
	filter := map[string]bool{normalizedTarget: true}

	result, err := ApplySelective(context.Background(), configDir, homeDir, &config.Config{}, ApplyFilterOptions{
		DryRun: true,
		Filter: filter,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalFiles)
	require.Len(t, result.Actions, 1)
	assert.Equal(t, "would-add", result.Actions[0].Status)
	assert.Equal(t, normalizedTarget, filepath.Clean(result.Actions[0].Destination))
}
