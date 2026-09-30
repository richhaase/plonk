// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/stretchr/testify/require"
)

func fakeBrew(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	require.NoError(t, os.Mkdir(bin, 0755))
	state := filepath.Join(dir, "installed")
	log := filepath.Join(dir, "operations")
	script := `#!/bin/sh
case "$1" in
 list) if [ "$2" = "--formula" ]; then cat "$PACKAGE_STATE" 2>/dev/null; fi; exit 0 ;;
 install) echo "install $3" >> "$PACKAGE_LOG"; [ "$3" = broken ] && exit 1; echo "$3" >> "$PACKAGE_STATE" ;;
 uninstall) echo "uninstall $3" >> "$PACKAGE_LOG"; [ "$3" = broken ] && exit 1; sed "/^$3$/d" "$PACKAGE_STATE" > "$PACKAGE_STATE.new"; mv "$PACKAGE_STATE.new" "$PACKAGE_STATE" ;;
esac
`
	require.NoError(t, os.WriteFile(filepath.Join(bin, "brew"), []byte(script), 0755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PACKAGE_STATE", state)
	t.Setenv("PACKAGE_LOG", log)
	packages.ResetManagerCache()
	t.Cleanup(packages.ResetManagerCache)
	return state, log
}

func packageConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte("git:\n  auto_commit: false\n"), 0644))
	return dir
}

func TestPackageAddAndRemoveLifecycle(t *testing.T) {
	state, log := fakeBrew(t)
	dir := packageConfig(t)
	ctx := context.Background()
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, true, false, false))
	svc := lock.NewLockV3Service(dir)
	current, err := svc.Read()
	require.NoError(t, err)
	require.True(t, current.HasPackage("brew", "demo"))
	contents, err := os.ReadFile(log)
	require.NoError(t, err)
	require.Equal(t, "install demo\n", string(contents))
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, true, false, false))
	contents, err = os.ReadFile(log)
	require.NoError(t, err)
	require.Equal(t, "install demo\n", string(contents))
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, false, false, false))
	contents, err = os.ReadFile(state)
	require.NoError(t, err)
	require.Contains(t, string(contents), "demo")
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, true, false, false))
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, false, true, false))
	current, err = svc.Read()
	require.NoError(t, err)
	require.False(t, current.HasPackage("brew", "demo"))
	contents, err = os.ReadFile(state)
	require.NoError(t, err)
	require.Empty(t, string(contents))
	// Cache invalidation allows a subsequent add to reinstall the removed package.
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:demo"}, true, false, false))
	contents, err = os.ReadFile(state)
	require.NoError(t, err)
	require.Contains(t, string(contents), "demo")
}

func TestPackageFailuresKeepOnlySuccessfulState(t *testing.T) {
	state, _ := fakeBrew(t)
	dir := packageConfig(t)
	ctx := context.Background()
	require.Error(t, mutatePackages(ctx, dir, []string{"brew:broken", "brew:demo"}, true, false, false))
	svc := lock.NewLockV3Service(dir)
	current, err := svc.Read()
	require.NoError(t, err)
	require.True(t, current.HasPackage("brew", "demo"))
	require.False(t, current.HasPackage("brew", "broken"))
	require.NoError(t, os.WriteFile(state, []byte("demo\nbroken\n"), 0644))
	packages.ResetManagerCache()
	require.NoError(t, mutatePackages(ctx, dir, []string{"brew:broken"}, true, false, false))
	require.Error(t, mutatePackages(ctx, dir, []string{"brew:broken", "brew:demo"}, false, true, false))
	current, err = svc.Read()
	require.NoError(t, err)
	require.True(t, current.HasPackage("brew", "broken"))
	require.False(t, current.HasPackage("brew", "demo"))
}

func TestPackageDryRunPreservesLegacyLockAndInstallations(t *testing.T) {
	state, log := fakeBrew(t)
	dir := packageConfig(t)
	legacy := []byte("version: 2\nresources:\n  - id: brew:demo\n    type: package\n    metadata:\n      manager: brew\n      name: demo\n")
	lockPath := filepath.Join(dir, "plonk.lock")
	require.NoError(t, os.WriteFile(lockPath, legacy, 0644))
	require.NoError(t, mutatePackages(context.Background(), dir, []string{"brew:demo"}, true, false, true))
	require.NoFileExists(t, state)
	require.NoFileExists(t, log)
	after, err := os.ReadFile(lockPath)
	require.NoError(t, err)
	require.Equal(t, legacy, after)
	require.NoError(t, mutatePackages(context.Background(), dir, []string{"brew:demo"}, false, true, true))
	after, err = os.ReadFile(lockPath)
	require.NoError(t, err)
	require.Equal(t, legacy, after)
	require.NoFileExists(t, filepath.Join(dir, ".plonk.mutlock"))
}

func TestSplitResourceArgs(t *testing.T) {
	files, specs := splitResourceArgs([]string{"brew:demo", "unknown:demo", "go:example.com/tool", "~/.config/tool:name", "./file:name", "/tmp/file:name"})
	require.Equal(t, []string{"brew:demo", "unknown:demo", "go:example.com/tool"}, specs)
	require.Equal(t, []string{"~/.config/tool:name", "./file:name", "/tmp/file:name"}, files)
}

func TestRemovedCommandsAreNotRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		require.NotEqual(t, "track", cmd.Name())
		require.NotEqual(t, "untrack", cmd.Name())
	}
}
