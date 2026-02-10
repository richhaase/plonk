// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockV3_AddRemovePackage(t *testing.T) {
	l := NewLockV3()

	// Add package
	l.AddPackage("brew", "ripgrep")
	assert.True(t, l.HasPackage("brew", "ripgrep"))
	assert.Equal(t, []string{"ripgrep"}, l.GetPackages("brew"))

	// Add another package (should be sorted)
	l.AddPackage("brew", "fzf")
	assert.Equal(t, []string{"fzf", "ripgrep"}, l.GetPackages("brew"))

	// Add duplicate (should be no-op)
	l.AddPackage("brew", "ripgrep")
	assert.Equal(t, []string{"fzf", "ripgrep"}, l.GetPackages("brew"))

	// Remove package
	l.RemovePackage("brew", "fzf")
	assert.False(t, l.HasPackage("brew", "fzf"))
	assert.Equal(t, []string{"ripgrep"}, l.GetPackages("brew"))

	// Remove last package (manager key should be deleted)
	l.RemovePackage("brew", "ripgrep")
	assert.False(t, l.HasPackage("brew", "ripgrep"))
	assert.Nil(t, l.GetPackages("brew"))
}

func TestLockV3_GetAllPackages(t *testing.T) {
	l := NewLockV3()

	l.AddPackage("brew", "ripgrep")
	l.AddPackage("cargo", "bat")
	l.AddPackage("brew", "fzf")

	all := l.GetAllPackages()
	// Should be sorted: brew:fzf, brew:ripgrep, cargo:bat
	assert.Equal(t, []string{"brew:fzf", "brew:ripgrep", "cargo:bat"}, all)
}

func TestLockV3Service_ReadWriteRoundtrip(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	svc := NewLockV3Service(tmpDir)

	// Write lock
	lock := NewLockV3()
	lock.AddPackage("brew", "ripgrep")
	lock.AddPackage("brew", "fzf")
	lock.AddPackage("cargo", "bat")

	err = svc.Write(lock)
	require.NoError(t, err)

	// Read it back
	readLock, err := svc.Read()
	require.NoError(t, err)

	assert.Equal(t, 3, readLock.Version)
	assert.True(t, readLock.HasPackage("brew", "ripgrep"))
	assert.True(t, readLock.HasPackage("brew", "fzf"))
	assert.True(t, readLock.HasPackage("cargo", "bat"))
}

func TestLockV3Service_ReadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	svc := NewLockV3Service(tmpDir)

	// Read non-existent lock should return empty lock
	lock, err := svc.Read()
	require.NoError(t, err)
	assert.Equal(t, 3, lock.Version)
	assert.Empty(t, lock.Packages)
}

func TestLockV3Service_MigrateV2ToV3(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write a v2 lock file manually
	v2Content := `version: 2
resources:
  - type: package
    metadata:
      manager: brew
      name: ripgrep
  - type: package
    metadata:
      manager: brew
      name: fzf
  - type: package
    metadata:
      manager: cargo
      name: bat
  - type: package
    metadata:
      manager: go
      name: golang.org/x/tools/gopls@latest
`
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte(v2Content), 0644)
	require.NoError(t, err)

	// Read through service (should trigger migration)
	svc := NewLockV3Service(tmpDir)
	lock, err := svc.Read()
	require.NoError(t, err)

	// Verify migration
	assert.Equal(t, 3, lock.Version)
	assert.True(t, lock.HasPackage("brew", "ripgrep"))
	assert.True(t, lock.HasPackage("brew", "fzf"))
	assert.True(t, lock.HasPackage("cargo", "bat"))
	assert.True(t, lock.HasPackage("go", "golang.org/x/tools/gopls@latest"))

	// Verify the file was updated on disk
	data, err := os.ReadFile(lockPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: 3")
	assert.Contains(t, string(data), "packages:")
}

func TestLockV3Service_MigrateV2_PreservesAllPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write a more complex v2 lock file
	v2Content := `version: 2
resources:
  - type: package
    metadata:
      manager: brew
      name: ripgrep
  - type: package
    metadata:
      manager: brew
      name: fzf
  - type: package
    metadata:
      manager: brew
      name: jq
  - type: package
    metadata:
      manager: cargo
      name: bat
  - type: package
    metadata:
      manager: cargo
      name: eza
  - type: package
    metadata:
      manager: go
      name: golang.org/x/tools/gopls@latest
  - type: package
    metadata:
      manager: pnpm
      name: typescript
  - type: package
    metadata:
      manager: uv
      name: ruff
`
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte(v2Content), 0644)
	require.NoError(t, err)

	// Read through service (should trigger migration)
	svc := NewLockV3Service(tmpDir)
	lock, err := svc.Read()
	require.NoError(t, err)

	// Verify all 8 packages were migrated
	assert.Equal(t, 3, lock.Version)
	assert.Equal(t, []string{"fzf", "jq", "ripgrep"}, lock.GetPackages("brew"))
	assert.Equal(t, []string{"bat", "eza"}, lock.GetPackages("cargo"))
	assert.Equal(t, []string{"golang.org/x/tools/gopls@latest"}, lock.GetPackages("go"))
	assert.Equal(t, []string{"typescript"}, lock.GetPackages("pnpm"))
	assert.Equal(t, []string{"ruff"}, lock.GetPackages("uv"))

	// Total package count
	all := lock.GetAllPackages()
	assert.Equal(t, 8, len(all))
}

func TestLockV3Service_MigrateV2_SkipsMalformedEntries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write a v2 lock with some malformed entries
	v2Content := `version: 2
resources:
  - type: package
    metadata:
      manager: brew
      name: ripgrep
  - type: package
    metadata:
      # Missing manager
      name: orphan-package
  - type: package
    metadata:
      manager: cargo
      # Missing name
  - type: dotfile
    metadata:
      source: zshrc
      target: ~/.zshrc
  - type: package
    metadata:
      manager: go
      name: golang.org/x/tools/gopls@latest
`
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte(v2Content), 0644)
	require.NoError(t, err)

	// Read through service (should trigger migration)
	svc := NewLockV3Service(tmpDir)
	lock, err := svc.Read()
	require.NoError(t, err)

	// Verify only valid packages were migrated
	assert.Equal(t, 3, lock.Version)
	assert.True(t, lock.HasPackage("brew", "ripgrep"))
	assert.True(t, lock.HasPackage("go", "golang.org/x/tools/gopls@latest"))

	// Malformed entries should be skipped
	assert.False(t, lock.HasPackage("", "orphan-package"))
	assert.False(t, lock.HasPackage("cargo", ""))

	// Total should be 2 (the valid ones)
	all := lock.GetAllPackages()
	assert.Equal(t, 2, len(all))
}

func TestLockV3Service_MigrateV2_EmptyLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write an empty v2 lock
	v2Content := `version: 2
resources: []
`
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte(v2Content), 0644)
	require.NoError(t, err)

	svc := NewLockV3Service(tmpDir)
	lock, err := svc.Read()
	require.NoError(t, err)

	assert.Equal(t, 3, lock.Version)
	assert.Empty(t, lock.Packages)
}

func TestLockV3Service_AtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	svc := NewLockV3Service(tmpDir)
	lockPath := filepath.Join(tmpDir, LockFileName)

	// Write initial lock
	lock := NewLockV3()
	lock.AddPackage("brew", "original")
	err = svc.Write(lock)
	require.NoError(t, err)

	// Verify no temp file left behind
	tmpPath := lockPath + ".tmp"
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err), "temp file should not exist after successful write")

	// Verify original file was created
	_, err = os.Stat(lockPath)
	assert.NoError(t, err, "lock file should exist")
}

func TestLockV3Service_WriteNilLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	svc := NewLockV3Service(tmpDir)
	err = svc.Write(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot write nil lock")
}

func TestLockV3Service_ReadCorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write corrupted YAML
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte("not: valid: yaml: {{"), 0644)
	require.NoError(t, err)

	svc := NewLockV3Service(tmpDir)
	_, err = svc.Read()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse lock file")
}

func TestLockV3Service_UnsupportedVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write unsupported version
	lockPath := filepath.Join(tmpDir, LockFileName)
	err = os.WriteFile(lockPath, []byte("version: 99\npackages: {}"), 0644)
	require.NoError(t, err)

	svc := NewLockV3Service(tmpDir)
	_, err = svc.Read()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported lock version 99")
}
