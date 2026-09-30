// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"

	"github.com/richhaase/plonk/internal/template"
	"github.com/stretchr/testify/require"
)

func TestDotfileManager_Deploy_ConfiguredMode(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/pi/agent/auth.json.tmpl"] = []byte(`{"token": "{{TOKEN}}"}`)
	fs.Dirs["/config/pi"] = true
	fs.Dirs["/config/pi/agent"] = true

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(key string) (string, bool) {
		if key == "TOKEN" {
			return "secret", true
		}
		return "", false
	}))

	// Default source mode is 0644 (from memFileInfo.Mode).
	m.SetDeployModes(map[string]os.FileMode{
		"pi/agent/auth.json.tmpl": 0o600,
	})

	err := m.Deploy("pi/agent/auth.json.tmpl")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.pi/agent/auth.json"]
	if mode != 0o600 {
		t.Errorf("deployed mode = %v (0o%o), want 0o600", mode, mode)
	}
}

func TestDotfileManager_Deploy_DefaultModeWhenUnconfigured(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("content")
	// Only configure a mode for a different dotfile.
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetDeployModes(map[string]os.FileMode{"other": 0o600})

	err := m.Deploy("zshrc")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Source mode in MemoryFS is 0644 and no mode is configured for zshrc.
	mode := fs.ChmodCalls["/home/user/.zshrc"]
	if mode != 0o644 {
		t.Errorf("deployed mode = %v (0o%o), want default 0o644", mode, mode)
	}
}

func TestDotfileManager_Deploy_NoModesConfigured(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Deploy("zshrc")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.zshrc"]
	if mode != 0o644 {
		t.Errorf("deployed mode = %v (0o%o), want default 0o644", mode, mode)
	}
}

func TestApplyStatuses_ConfiguredMode(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/missing"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetDeployModes(map[string]os.FileMode{"missing": 0o600})

	statuses, err := m.Reconcile()
	if err != nil {
		t.Fatal(err)
	}
	_, err = applyStatuses(context.Background(), m, statuses, false)
	if err != nil {
		t.Fatalf("applyStatuses() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.missing"]
	if mode != 0o600 {
		t.Errorf("applyStatuses deployed mode = %v (0o%o), want 0o600", mode, mode)
	}
}

func TestApply_ReconcilesConfiguredMode(t *testing.T) {
	for _, templateFile := range []bool{false, true} {
		name := "zshrc"
		content := []byte("export EDITOR=vim\n")
		source := content
		if templateFile {
			name += ".tmpl"
			source = []byte("export EDITOR={{PLONK_DEPLOY_MODE_TEST}}\n")
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("PLONK_DEPLOY_MODE_TEST", "vim")
			for _, tc := range []struct {
				name    string
				rules   []config.DotfileRule
				updated int
				mode    os.FileMode
			}{
				{"configured", []config.DotfileRule{{Name: name, Mode: "0600"}}, 1, 0600},
				{"unconfigured", nil, 0, 0640},
				{"unmatched", []config.DotfileRule{{Name: "other", Mode: "0600"}}, 0, 0640},
				{"no mode", []config.DotfileRule{{Name: name}}, 0, 0640},
			} {
				t.Run(tc.name, func(t *testing.T) {
					configDir, homeDir := t.TempDir(), t.TempDir()
					target := filepath.Join(homeDir, ".zshrc")
					require.NoError(t, os.WriteFile(filepath.Join(configDir, name), source, 0644))
					require.NoError(t, os.WriteFile(target, content, 0600))
					require.NoError(t, os.Chmod(target, 0640))
					before, err := os.Stat(target)
					require.NoError(t, err)
					cfg := config.LoadWithDefaults(configDir)
					cfg.Dotfiles.Rules = tc.rules

					dryResult, err := Apply(context.Background(), configDir, homeDir, cfg, true)
					require.NoError(t, err)
					require.Equal(t, tc.updated, dryResult.Summary.Updated)
					require.Equal(t, 1-tc.updated, dryResult.Summary.Unchanged)
					if tc.updated == 1 {
						require.Len(t, dryResult.Actions, 1)
						require.Equal(t, "would-update", dryResult.Actions[0].Status)
					}
					afterDryRun, err := os.Stat(target)
					require.NoError(t, err)
					require.Equal(t, os.FileMode(0640), afterDryRun.Mode().Perm())
					require.True(t, os.SameFile(before, afterDryRun), "dry-run must not replace the target")
					afterContent, err := os.ReadFile(target)
					require.NoError(t, err)
					require.Equal(t, content, afterContent)

					result, err := Apply(context.Background(), configDir, homeDir, cfg, false)
					require.NoError(t, err)
					require.Equal(t, tc.updated, result.Summary.Updated)
					require.Equal(t, 1-tc.updated, result.Summary.Unchanged)
					afterApply, err := os.Stat(target)
					require.NoError(t, err)
					require.Equal(t, tc.mode, afterApply.Mode().Perm())
					afterContent, err = os.ReadFile(target)
					require.NoError(t, err)
					require.Equal(t, content, afterContent)

					repeated, err := Apply(context.Background(), configDir, homeDir, cfg, false)
					require.NoError(t, err)
					require.Equal(t, 1, repeated.Summary.Unchanged)
					require.Empty(t, repeated.Actions)
				})
			}
		})
	}
}

func TestApplySelective_ReconcilesOnlySelectedMode(t *testing.T) {
	configDir, homeDir := t.TempDir(), t.TempDir()
	cfg := config.LoadWithDefaults(configDir)
	for _, name := range []string{"zshrc", "vimrc"} {
		require.NoError(t, os.WriteFile(filepath.Join(configDir, name), []byte("content"), 0644))
		target := filepath.Join(homeDir, "."+name)
		require.NoError(t, os.WriteFile(target, []byte("content"), 0600))
		require.NoError(t, os.Chmod(target, 0644))
		cfg.Dotfiles.Rules = append(cfg.Dotfiles.Rules, config.DotfileRule{Name: name, Mode: "0600"})
	}

	result, err := ApplySelective(context.Background(), configDir, homeDir, cfg, ApplyFilterOptions{
		Filter: map[string]bool{filepath.Join(homeDir, ".zshrc"): true},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Summary.Updated)
	require.Equal(t, 1, result.TotalFiles)
	selected, err := os.Stat(filepath.Join(homeDir, ".zshrc"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), selected.Mode().Perm())
	unselected, err := os.Stat(filepath.Join(homeDir, ".vimrc"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0644), unselected.Mode().Perm())
}
