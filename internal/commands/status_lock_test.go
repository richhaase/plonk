// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPackageStatus_NoLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := getPackageStatus(context.Background(), tmpDir)
	require.NoError(t, err)
	assert.Empty(t, result.Managed)
	assert.Empty(t, result.Missing)
	assert.Empty(t, result.Errors)
}

func TestGetPackageStatus_MalformedLockFileReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "plonk.lock")
	require.NoError(t, os.WriteFile(lockPath, []byte("version: ["), 0644))

	_, err := getPackageStatus(context.Background(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read lock file")
}

func TestGetPackageStatus_UnsupportedManagerIsReportedPerPackage(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "plonk.lock")
	content := `version: 3
packages:
  npm:
    - typescript
`
	require.NoError(t, os.WriteFile(lockPath, []byte(content), 0644))

	result, err := getPackageStatus(context.Background(), tmpDir)
	require.NoError(t, err)
	require.Len(t, result.Errors, 1)
	assert.Equal(t, "typescript", result.Errors[0].Name)
	assert.Equal(t, "npm", result.Errors[0].Manager)
	assert.Contains(t, result.Errors[0].Error, "unsupported manager")
}
