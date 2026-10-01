// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func auditCommand() *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(context.Background())
	for _, f := range []string{"dry-run", "force", "sync-drifted", "packages", "dotfiles", "all", "apply"} {
		c.Flags().Bool(f, false, "")
	}
	return c
}

func captureAuditOutput(t *testing.T, run func() error) (string, error) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "output")
	require.NoError(t, err)
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr; f.Close() }()
	runErr := run()
	require.NoError(t, f.Sync())
	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	return string(data), runErr
}

func TestInvalidConfigBlocksMutations(t *testing.T) {
	for _, operation := range []string{"add file", "add package", "add mixed", "rm file", "rm package", "rm mixed", "apply", "sync drifted"} {
		t.Run(operation, func(t *testing.T) {
			home, dir, bin := t.TempDir(), t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("PLONK_DIR", dir)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			log := filepath.Join(bin, "calls")
			t.Setenv("AUDIT_CALLS", log)
			require.NoError(t, os.WriteFile(filepath.Join(bin, "brew"), []byte("#!/bin/sh\ntouch \"$AUDIT_CALLS\"\nexit 1\n"), 0755))
			packages.ResetManagerCache()
			t.Cleanup(packages.ResetManagerCache)
			invalid := []byte("git:\n  auto_commit: false\ndotfiles:\n  rules:\n    - name: test\n      mode: \"0600\"\noperation_timeout: nope\n")
			require.NoError(t, os.WriteFile(filepath.Join(dir, "plonk.yaml"), invalid, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "test"), []byte("source"), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(home, ".test"), []byte("local"), 0600))
			lockData := []byte("version: 3\npackages:\n  brew: [demo]\n")
			lockPath := filepath.Join(dir, "plonk.lock")
			require.NoError(t, os.WriteFile(lockPath, lockData, 0600))
			c := auditCommand()
			require.NoError(t, c.Flags().Set("force", "true"))
			var err error
			switch operation {
			case "add file":
				err = runAdd(c, []string{filepath.Join(home, ".test")})
			case "add package":
				err = runAdd(c, []string{"brew:demo"})
			case "add mixed":
				err = runAdd(c, []string{"brew:demo", filepath.Join(home, ".test")})
			case "rm file":
				err = runRm(c, []string{filepath.Join(home, ".test")})
			case "rm package":
				err = runRm(c, []string{"brew:demo"})
			case "rm mixed":
				err = runRm(c, []string{"brew:demo", filepath.Join(home, ".test")})
			case "apply":
				err = runApply(c, nil)
			case "sync drifted":
				require.NoError(t, c.Flags().Set("sync-drifted", "true"))
				err = runAdd(c, nil)
			}
			require.Error(t, err)
			require.NoFileExists(t, log)
			for p, want := range map[string]string{lockPath: string(lockData), filepath.Join(dir, "test"): "source", filepath.Join(home, ".test"): "local", filepath.Join(dir, "plonk.yaml"): string(invalid)} {
				got, e := os.ReadFile(p)
				require.NoError(t, e)
				require.Equal(t, want, string(got))
			}
			info, e := os.Stat(filepath.Join(home, ".test"))
			require.NoError(t, e)
			require.Equal(t, os.FileMode(0600), info.Mode().Perm())
		})
	}
}

func TestExplicitRelativeRemovalUsesCurrentDirectory(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	cwd := filepath.Join(home, ".config", "nvim")
	require.NoError(t, os.MkdirAll(cwd, 0755))
	t.Chdir(cwd)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "config", "nvim"), 0755))
	intended := filepath.Join(cwd, "init.lua")
	other := filepath.Join(home, ".init.lua")
	for _, p := range []string{intended, other, filepath.Join(dir, "config", "nvim", "init.lua"), filepath.Join(dir, "init.lua")} {
		require.NoError(t, os.WriteFile(p, []byte("test"), 0600))
	}
	dm := dotfiles.NewDotfileManager(dir, home, nil)
	result := removeDotfiles(dm, dir, home, []string{"./init.lua"}, RemoveOptions{Force: true, DryRun: true})
	require.Equal(t, RemoveStatusWouldRemove, result[0].Status)
	require.Equal(t, intended, result[0].Destination)
	require.FileExists(t, other)
	require.FileExists(t, intended)
	result = removeDotfiles(dm, dir, home, []string{"./init.lua"}, RemoveOptions{Force: true})
	require.Equal(t, RemoveStatusRemoved, result[0].Status)
	require.NoFileExists(t, intended)
	require.FileExists(t, other)
	require.Equal(t, "init.lua", resolveDotfileNameForRemoval("init.lua", home))
	t.Chdir(home)
	result = removeDotfiles(dm, dir, home, []string{"./init.lua"}, RemoveOptions{Force: true})
	require.Equal(t, RemoveStatusFailed, result[0].Status)
	require.FileExists(t, other)
	// Explicit paths outside HOME must not resolve to the otherwise-managed other file.
	t.Chdir(t.TempDir())
	result = removeDotfiles(dm, dir, home, []string{"./init.lua"}, RemoveOptions{Force: true})
	require.Equal(t, RemoveStatusFailed, result[0].Status)
	require.FileExists(t, other)
}

func TestPermissionDriftVisibleWithoutReverseSync(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PLONK_DIR", dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte("git:\n  auto_commit: false\ndotfiles:\n  rules:\n    - name: test\n      mode: \"0600\"\n"), 0600))
	source, target := filepath.Join(dir, "test"), filepath.Join(home, ".test")
	for _, p := range []string{source, target} {
		require.NoError(t, os.WriteFile(p, []byte("same"), 0644))
		require.NoError(t, os.Chmod(p, 0644))
	}
	for _, run := range []func(*cobra.Command, []string) error{runStatus, runDotfiles} {
		out, err := captureAuditOutput(t, func() error { return run(auditCommand(), nil) })
		require.NoError(t, err)
		require.Contains(t, out, "drifted")
		require.NotContains(t, out, "All managed items are in sync")
	}
	out, err := captureAuditOutput(t, func() error { return runDiff(auditCommand(), nil) })
	require.NoError(t, err)
	require.Contains(t, out, "permissions 0644 -> 0600")
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	drifted, err := getDriftedDotfileStatuses(cfg, dir, home)
	require.NoError(t, err)
	require.Empty(t, drifted)
	before, err := os.Stat(source)
	require.NoError(t, err)
	c := auditCommand()
	require.NoError(t, c.Flags().Set("sync-drifted", "true"))
	require.NoError(t, runAdd(c, nil))
	after, err := os.Stat(source)
	require.NoError(t, err)
	require.Equal(t, before.ModTime(), after.ModTime())
}

func TestCaskInventoryFailurePreservesTracking(t *testing.T) {
	for _, cancelDuring := range []bool{false, true} {
		t.Run(map[bool]string{false: "failure", true: "cancellation"}[cancelDuring], func(t *testing.T) {
			dir, bin := packageConfig(t), t.TempDir()
			lockPath := filepath.Join(dir, "plonk.lock")
			before := []byte("version: 3\npackages:\n  brew: [demo-cask]\n")
			require.NoError(t, os.WriteFile(lockPath, before, 0600))
			marker := filepath.Join(bin, "checking")
			calls := filepath.Join(bin, "calls")
			t.Setenv("AUDIT_MARKER", marker)
			t.Setenv("AUDIT_CALLS", calls)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			script := "#!/bin/sh\n[ \"$1 $2\" = 'list --formula' ] && exit 0\nif [ \"$1 $2\" = 'list --cask' ]; then touch \"$AUDIT_MARKER\"; "
			if cancelDuring {
				script += "exec sleep 30; "
			} else {
				script += "exit 1; "
			}
			script += "fi\ntouch \"$AUDIT_CALLS\"\nexit 0\n"
			require.NoError(t, os.WriteFile(filepath.Join(bin, "brew"), []byte(script), 0755))
			packages.ResetManagerCache()
			t.Cleanup(packages.ResetManagerCache)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- mutatePackages(ctx, dir, []string{"brew:demo-cask"}, false, true, false) }()
			if cancelDuring {
				require.Eventually(t, func() bool { _, e := os.Stat(marker); return e == nil }, 3*time.Second, 10*time.Millisecond)
				cancel()
			}
			select {
			case err := <-done:
				require.Error(t, err)
			case <-time.After(3 * time.Second):
				t.Fatal("removal did not finish")
			}
			after, err := os.ReadFile(lockPath)
			require.NoError(t, err)
			require.Equal(t, before, after)
			require.NoFileExists(t, calls)
		})
	}
}

func TestInspectionAndApplyPreserveLegacyLock(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PLONK_DIR", dir)
	path := filepath.Join(dir, "plonk.lock")
	before := []byte("version: 2\nresources: []\n")
	require.NoError(t, os.WriteFile(path, before, 0600))
	c := auditCommand()
	require.NoError(t, c.Flags().Set("dry-run", "true"))
	require.NoError(t, runApply(c, nil))
	require.NoError(t, runStatus(auditCommand(), nil))
	require.NoError(t, runPackages(auditCommand(), nil))
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.NoFileExists(t, filepath.Join(dir, ".plonk.mutlock"))
}
