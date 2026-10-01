// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectoryAddPreflightsTemplateTargets(t *testing.T) {
	for _, dryRun := range []bool{true, false} {
		t.Run(map[bool]string{true: "dry run", false: "copy"}[dryRun], func(t *testing.T) {
			home, configDir := t.TempDir(), t.TempDir()
			dir := filepath.Join(home, ".demo")
			require.NoError(t, os.MkdirAll(dir, 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(configDir, "demo"), 0755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "a-ordinary"), []byte("ordinary"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "z-auth"), []byte("DUMMY_VALUE"), 0600))
			templatePath := filepath.Join(configDir, "demo", "z-auth.tmpl")
			templateContent := []byte("{{keychain:svc/acct}}")
			require.NoError(t, os.WriteFile(templatePath, templateContent, 0600))
			dm := NewDotfileManager(configDir, home, nil)
			var err error
			if dryRun {
				err = dm.ValidateAdd(dir)
			} else {
				err = dm.Add(dir)
			}
			require.ErrorContains(t, err, "managed as a template")
			require.NoFileExists(t, filepath.Join(configDir, "demo", "a-ordinary"))
			require.NoFileExists(t, filepath.Join(configDir, "demo", "z-auth"))
			after, err := os.ReadFile(templatePath)
			require.NoError(t, err)
			require.Equal(t, templateContent, after)
			// An existing accidental plain counterpart does not bypass protection.
			plain := filepath.Join(configDir, "demo", "z-auth")
			require.NoError(t, os.WriteFile(plain, []byte("prior"), 0600))
			require.ErrorContains(t, dm.Add(filepath.Join(dir, "z-auth")), "managed as a template")
			after, err = os.ReadFile(plain)
			require.NoError(t, err)
			require.Equal(t, "prior", string(after))
		})
	}
}

func TestDirectoryAddExcludesConfigurationAndAliases(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "plonk")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "nvim"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "plonk.yaml"), []byte("git: {}"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".config", "nvim", "init.lua"), []byte("ordinary"), 0600))
	require.NoError(t, os.Symlink("plonk", filepath.Join(home, ".config", "alias")))
	dm := NewDotfileManager(configDir, home, nil)
	require.NoError(t, dm.ValidateAdd(filepath.Join(home, ".config")))
	require.NoError(t, dm.Add(filepath.Join(home, ".config")))
	require.NoFileExists(t, filepath.Join(configDir, "config", "plonk", "plonk.yaml"))
	require.NoDirExists(t, filepath.Join(configDir, "config", "alias"))
	require.FileExists(t, filepath.Join(configDir, "config", "nvim", "init.lua"))
	require.ErrorContains(t, dm.Add(filepath.Join(home, ".config", "alias")), "config directory")
	// A parent alias still skips the repository encountered beneath it.
	require.NoError(t, os.Symlink(".config", filepath.Join(home, ".parent-alias")))
	require.NoError(t, dm.Add(filepath.Join(home, ".parent-alias")))
	require.NoDirExists(t, filepath.Join(configDir, "parent-alias", "plonk"))
}

func TestDeployRejectsOldSelfCopiesAndControlAliases(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "plonk")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	target := filepath.Join(configDir, "plonk.yaml")
	require.NoError(t, os.WriteFile(target, []byte("current"), 0600))
	require.NoError(t, os.Symlink(filepath.Join(".config", "plonk"), filepath.Join(home, ".alias")))
	for _, name := range []string{"config/plonk/plonk.yaml", "alias/plonk.yaml.tmpl"} {
		source := filepath.Join(configDir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(source), 0755))
		require.NoError(t, os.WriteFile(source, []byte("stale"), 0600))
		dm := NewDotfileManager(configDir, home, nil)
		require.ErrorContains(t, dm.Deploy(name), "cannot deploy into config directory")
		state, err := dm.getState(Dotfile{Name: name, Source: source, Target: dm.toTarget(name)})
		require.ErrorContains(t, err, "cannot deploy into config directory")
		require.Equal(t, SyncStateError, state)
		after, err := os.ReadFile(target)
		require.NoError(t, err)
		require.Equal(t, "current", string(after))
	}
}

func TestDirectoryAddRejectsTemplateFileAliasesAndSkipsConfigFileAliases(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "plonk")
	dir := filepath.Join(home, ".config", "app")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "plonk.yaml"), []byte("protected"), 0600))
	require.NoError(t, os.Symlink("../plonk/plonk.yaml", filepath.Join(dir, "settings")))
	dm := NewDotfileManager(configDir, home, nil)
	require.NoError(t, dm.Add(dir))
	require.NoFileExists(t, filepath.Join(configDir, "config", "app", "settings"))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "auth.tmpl"), []byte("{{keychain:svc/acct}}"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".auth"), []byte("DUMMY_VALUE"), 0600))
	require.NoError(t, os.Symlink("../../.auth", filepath.Join(dir, "credential-alias")))
	require.ErrorContains(t, dm.ValidateAdd(dir), "managed as a template")
	require.ErrorContains(t, dm.Add(dir), "managed as a template")
	require.NoFileExists(t, filepath.Join(configDir, "config", "app", "credential-alias"))
}

func TestAddRejectsHardLinkedSources(t *testing.T) {
	for _, protected := range []string{"template target", "configuration file"} {
		t.Run(protected, func(t *testing.T) {
			home := t.TempDir()
			configDir := filepath.Join(home, ".config", "plonk")
			dir := filepath.Join(home, ".config", "app")
			require.NoError(t, os.MkdirAll(configDir, 0755))
			require.NoError(t, os.MkdirAll(dir, 0755))
			target := filepath.Join(home, ".auth")
			if protected == "configuration file" {
				target = filepath.Join(configDir, "plonk.yaml")
			} else {
				require.NoError(t, os.WriteFile(filepath.Join(configDir, "auth.tmpl"), []byte("{{keychain:svc/acct}}"), 0600))
			}
			require.NoError(t, os.WriteFile(target, []byte("DUMMY_VALUE"), 0600))
			alias := filepath.Join(dir, "alias")
			require.NoError(t, os.Link(target, alias))
			dm := NewDotfileManager(configDir, home, nil)
			require.ErrorContains(t, dm.ValidateAdd(alias), "hard-linked")
			require.ErrorContains(t, dm.Add(alias), "hard-linked")
			require.ErrorContains(t, dm.ValidateAdd(dir), "hard-linked")
			require.ErrorContains(t, dm.Add(dir), "hard-linked")
			require.NoFileExists(t, filepath.Join(configDir, "config", "app", "alias"))
		})
	}
}
