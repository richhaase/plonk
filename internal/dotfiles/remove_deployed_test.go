// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDeployedPreservesSourceUntilSuccessfulDeletion(t *testing.T) {
	home, configDir := t.TempDir(), t.TempDir()
	name := "example.tmpl"
	source, target := filepath.Join(configDir, name), filepath.Join(home, ".example")
	require.NoError(t, os.WriteFile(source, []byte("template"), 0644))
	require.NoError(t, os.WriteFile(target, []byte("deployed"), 0600))
	dm := NewDotfileManager(configDir, home, nil)
	require.NoError(t, dm.RemoveDeployed(name, true))
	require.FileExists(t, source)
	require.FileExists(t, target)
	require.NoError(t, dm.RemoveDeployed(name, false))
	require.FileExists(t, source)
	require.NoFileExists(t, target)
	require.NoError(t, dm.RemoveDeployed(name, false)) // Already absent is allowed.
	require.NoError(t, dm.Remove(name))
	require.NoFileExists(t, source)
}

func TestRemoveDeployedRejectsDirectoryAndEscapingParent(t *testing.T) {
	home, configDir, outside := t.TempDir(), t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "config"), 0755))
	source := filepath.Join(configDir, "config", "example")
	require.NoError(t, os.WriteFile(source, []byte("source"), 0644))
	dm := NewDotfileManager(configDir, home, nil)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "example"), 0755))
	require.Error(t, dm.RemoveDeployed("config/example", false))
	require.FileExists(t, source)
	require.NoError(t, os.RemoveAll(filepath.Join(home, ".config")))
	outsideFile := filepath.Join(outside, "example")
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0644))
	require.NoError(t, os.Symlink(outside, filepath.Join(home, ".config")))
	require.Error(t, dm.RemoveDeployed("config/example", false))
	require.FileExists(t, source)
	require.FileExists(t, outsideFile)
}

func TestRemoveDeployedUnlinksFinalSymlinkWithoutFollowing(t *testing.T) {
	home, configDir, outside := t.TempDir(), t.TempDir(), t.TempDir()
	source := filepath.Join(configDir, "example")
	target := filepath.Join(home, ".example")
	outsideFile := filepath.Join(outside, "example")
	require.NoError(t, os.WriteFile(source, []byte("source"), 0644))
	require.NoError(t, os.WriteFile(outsideFile, []byte("outside"), 0644))
	require.NoError(t, os.Symlink(outsideFile, target))
	require.NoError(t, NewDotfileManager(configDir, home, nil).RemoveDeployed("example", false))
	_, err := os.Lstat(target)
	require.True(t, os.IsNotExist(err))
	require.FileExists(t, outsideFile)
}

func TestRemoveDeployedProtectsControlFilesAndParentAliases(t *testing.T) {
	for _, control := range []string{"plonk.lock", "plonk.yaml", ".plonk.mutlock", ".git/config"} {
		for _, alias := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/alias=%t", control, alias), func(t *testing.T) {
				home := t.TempDir()
				configDir := filepath.Join(home, ".config", "plonk")
				require.NoError(t, os.MkdirAll(configDir, 0755))
				target := filepath.Join(configDir, control)
				require.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
				require.NoError(t, os.WriteFile(target, []byte("protected"), 0600))
				name := "config/plonk/" + control + ".tmpl"
				if alias {
					require.NoError(t, os.Symlink(configDir, filepath.Join(home, ".alias")))
					name = "alias/" + control + ".tmpl"
				}
				source := filepath.Join(configDir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(source), 0755))
				require.NoError(t, os.WriteFile(source, []byte("template"), 0600))
				dm := NewDotfileManager(configDir, home, nil)
				for _, dryRun := range []bool{true, false} {
					require.ErrorContains(t, dm.RemoveDeployed(name, dryRun), "internal deployed file")
					contents, err := os.ReadFile(target)
					require.NoError(t, err)
					require.Equal(t, "protected", string(contents))
					require.FileExists(t, source)
				}
			})
		}
	}
}
