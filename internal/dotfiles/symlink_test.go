// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDotfileManagerAddRejectsSymlinkOutsideHome(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret")
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0600))
	require.NoError(t, os.Symlink(outsideFile, filepath.Join(homeDir, ".linked-secret")))

	err := NewDotfileManager(configDir, homeDir, nil).Add(filepath.Join(homeDir, ".linked-secret"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(configDir, "linked-secret"))
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestDotfileManagerAddRejectsDirectorySymlinkOutsideHome(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outsideDir, "secret"), []byte("outside"), 0600))
	require.NoError(t, os.Symlink(outsideDir, filepath.Join(homeDir, ".linked-dir")))

	err := NewDotfileManager(configDir, homeDir, nil).Add(filepath.Join(homeDir, ".linked-dir"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(configDir, "linked-dir", "secret"))
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestDotfileManagerDeployRejectsSymlinkOutsideConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret")
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0600))
	require.NoError(t, os.Symlink(outsideFile, filepath.Join(configDir, "linked-secret")))

	err := NewDotfileManager(configDir, homeDir, nil).Deploy("linked-secret")
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(homeDir, ".linked-secret"))
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestDotfileManagerDeployRejectsSymlinkedDestinationParent(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "config", "nvim"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config", "nvim", "init.lua"), []byte("managed"), 0600))
	require.NoError(t, os.Symlink(outsideDir, filepath.Join(homeDir, ".config")))

	err := NewDotfileManager(configDir, homeDir, nil).Deploy(filepath.Join("config", "nvim", "init.lua"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(outsideDir, "nvim", "init.lua"))
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestDotfileManagerRejectsChainedAndBrokenSymlinks(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret")
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0600))
	require.NoError(t, os.Symlink(".second", filepath.Join(homeDir, ".first")))
	require.NoError(t, os.Symlink(outsideFile, filepath.Join(homeDir, ".second")))
	require.NoError(t, os.Symlink("missing", filepath.Join(homeDir, ".broken")))

	manager := NewDotfileManager(configDir, homeDir, nil)
	require.Error(t, manager.Add(filepath.Join(homeDir, ".first")))
	require.Error(t, manager.Add(filepath.Join(homeDir, ".broken")))
}

func TestDotfileManagerRemoveRejectsSymlinkedConfigParent(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "managed")
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0600))
	require.NoError(t, os.Symlink(outsideDir, filepath.Join(configDir, "escape")))

	err := NewDotfileManager(configDir, homeDir, nil).Remove(filepath.Join("escape", "managed"))
	require.Error(t, err)
	content, err := os.ReadFile(outsideFile)
	require.NoError(t, err)
	require.Equal(t, []byte("outside"), content)
}

func TestDotfileManagerAddNeverReadsSymlinkSwappedSource(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	outsideDir := t.TempDir()
	sourcePath := filepath.Join(homeDir, ".victim")
	stagedPath := sourcePath + ".staged"
	outsideFile := filepath.Join(outsideDir, "secret")
	require.NoError(t, os.WriteFile(sourcePath, []byte("inside"), 0600))
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0600))

	done := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			_ = os.Rename(sourcePath, stagedPath)
			_ = os.Symlink(outsideFile, sourcePath)
			_ = os.Remove(sourcePath)
			_ = os.Rename(stagedPath, sourcePath)
		}
	}()

	manager := NewDotfileManager(configDir, homeDir, nil)
	for range 100 {
		_ = manager.Add(sourcePath)
		content, err := os.ReadFile(filepath.Join(configDir, "victim"))
		if err == nil {
			require.Equal(t, []byte("inside"), content)
		}
	}
	close(done)
	workers.Wait()
}
