// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUninstallCommandsAndCacheInvalidation(t *testing.T) {
	for _, name := range []string{"brew", "cargo", "pnpm", "uv"} {
		t.Run(name, func(t *testing.T) {
			ResetManagerCache()
			t.Cleanup(ResetManagerCache)
			bin := t.TempDir()
			log := filepath.Join(bin, "args")
			t.Setenv("UNINSTALL_LOG", log)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			require.NoError(t, os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$UNINSTALL_LOG\"\nexit \"${UNINSTALL_EXIT:-0}\"\n"), 0755))
			mgr, err := GetManager(name)
			require.NoError(t, err)
			// Primed caches must survive a failed uninstall and reset on success.
			switch m := mgr.(type) {
			case *BrewSimple:
				m.installed = map[string]bool{"demo": true}
			case *CargoSimple:
				m.installed = map[string]bool{"demo": true}
			case *PNPMSimple:
				m.installed = map[string]bool{"demo": true}
			case *UVSimple:
				m.installed = map[string]bool{"demo": true}
			}
			t.Setenv("UNINSTALL_EXIT", "1")
			require.Error(t, mgr.Uninstall(context.Background(), "demo"))
			installed, err := mgr.IsInstalled(context.Background(), "demo")
			require.NoError(t, err)
			require.True(t, installed)
			t.Setenv("UNINSTALL_EXIT", "0")
			require.NoError(t, mgr.Uninstall(context.Background(), "demo"))
			args, err := os.ReadFile(log)
			require.NoError(t, err)
			want := map[string]string{"brew": "uninstall\n--\ndemo\n", "cargo": "uninstall\n--\ndemo\n", "pnpm": "remove\n-g\n--\ndemo\n", "uv": "tool\nuninstall\n--\ndemo\n"}
			require.Equal(t, want[name], string(args))
			switch m := mgr.(type) {
			case *BrewSimple:
				require.Nil(t, m.installed)
			case *CargoSimple:
				require.Nil(t, m.installed)
			case *PNPMSimple:
				require.Nil(t, m.installed)
			case *UVSimple:
				require.Nil(t, m.installed)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			require.Error(t, mgr.Uninstall(ctx, "demo"))
		})
	}
}

func TestGoUninstallVerifiesImportPathAndRespectsGOBIN(t *testing.T) {
	dir, bin := t.TempDir(), t.TempDir()
	t.Setenv("GOBIN", bin)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/plonk/tool\n\ngo 1.26.5\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644))
	target := filepath.Join(bin, "tool")
	cmd := exec.Command("go", "build", "-o", target, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	result, err := cmd.CombinedOutput()
	require.NoError(t, err, string(result))
	g := NewGoSimple()
	require.Error(t, g.Uninstall(context.Background(), "example.com/other/tool"))
	require.FileExists(t, target)
	require.NoError(t, g.Uninstall(context.Background(), "example.com/plonk/tool@latest"))
	require.NoFileExists(t, target)
	require.NoError(t, os.WriteFile(target, []byte("unrelated file"), 0755))
	require.Error(t, g.Uninstall(context.Background(), "example.com/plonk/tool"))
	require.FileExists(t, target)
	require.Error(t, g.Uninstall(context.Background(), "example.com/.."))
	require.DirExists(t, bin)
}
