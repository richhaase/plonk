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

func TestDotfileManagerAddAllowsRelativeSymlinkWithinHome(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".source"), []byte("inside"), 0600))
	require.NoError(t, os.Symlink(".source", filepath.Join(homeDir, ".linked")))

	err := NewDotfileManager(configDir, homeDir, nil).Add(filepath.Join(homeDir, ".linked"))
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(configDir, "linked"))
	require.NoError(t, err)
	require.Equal(t, []byte("inside"), content)
}

func TestDotfileManagerAddRejectsAbsoluteSymlinkWithinHome(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	sourcePath := filepath.Join(homeDir, ".source")
	require.NoError(t, os.WriteFile(sourcePath, []byte("inside"), 0600))
	require.NoError(t, os.Symlink(sourcePath, filepath.Join(homeDir, ".linked")))

	err := NewDotfileManager(configDir, homeDir, nil).Add(filepath.Join(homeDir, ".linked"))
	require.ErrorContains(t, err, "cannot access")
	require.NotContains(t, err.Error(), "does not exist")
	_, err = os.Stat(filepath.Join(configDir, "linked"))
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

func TestDotfileManagerDeployUsesHomeBoundaryWhenConfigContainsHome(t *testing.T) {
	configDir := t.TempDir()
	homeDir := filepath.Join(configDir, "home")
	escapeDir := filepath.Join(configDir, "escape")
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "config", "nvim"), 0755))
	require.NoError(t, os.MkdirAll(homeDir, 0755))
	require.NoError(t, os.MkdirAll(escapeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config", "nvim", "init.lua"), []byte("managed"), 0600))
	require.NoError(t, os.Symlink(filepath.Join("..", "escape"), filepath.Join(homeDir, ".config")))

	err := NewDotfileManager(configDir, homeDir, nil).Deploy(filepath.Join("config", "nvim", "init.lua"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(escapeDir, "nvim", "init.lua"))
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestRootedOSFileSystemSelectsMostSpecificRoot(t *testing.T) {
	baseDir := t.TempDir()
	tests := []struct {
		name        string
		configDir   string
		homeDir     string
		path        string
		wantRoot    string
		wantRelPath string
	}{
		{
			name:        "config under home",
			configDir:   filepath.Join(baseDir, "home", "config"),
			homeDir:     filepath.Join(baseDir, "home"),
			path:        filepath.Join(baseDir, "home", "config", "zshrc"),
			wantRoot:    filepath.Join(baseDir, "home", "config"),
			wantRelPath: "zshrc",
		},
		{
			name:        "home under config",
			configDir:   filepath.Join(baseDir, "config"),
			homeDir:     filepath.Join(baseDir, "config", "home"),
			path:        filepath.Join(baseDir, "config", "home", ".zshrc"),
			wantRoot:    filepath.Join(baseDir, "config", "home"),
			wantRelPath: ".zshrc",
		},
		{
			name:        "sibling config root",
			configDir:   filepath.Join(baseDir, "config"),
			homeDir:     filepath.Join(baseDir, "home"),
			path:        filepath.Join(baseDir, "config", "zshrc"),
			wantRoot:    filepath.Join(baseDir, "config"),
			wantRelPath: "zshrc",
		},
		{
			name:        "sibling home root",
			configDir:   filepath.Join(baseDir, "config"),
			homeDir:     filepath.Join(baseDir, "home"),
			path:        filepath.Join(baseDir, "home", ".zshrc"),
			wantRoot:    filepath.Join(baseDir, "home"),
			wantRelPath: ".zshrc",
		},
		{
			name:        "identical roots",
			configDir:   filepath.Join(baseDir, "shared"),
			homeDir:     filepath.Join(baseDir, "shared"),
			path:        filepath.Join(baseDir, "shared", "zshrc"),
			wantRoot:    filepath.Join(baseDir, "shared"),
			wantRelPath: "zshrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath, relPath, err := NewRootedOSFileSystem(tt.configDir, tt.homeDir).rootPath(tt.path)
			require.NoError(t, err)
			require.Equal(t, tt.wantRoot, rootPath)
			require.Equal(t, tt.wantRelPath, relPath)
		})
	}
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
