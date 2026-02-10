// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubManager struct {
	installed    map[string]bool
	isInstalledE map[string]error
	installE     map[string]error
	installedNow []string
}

func (s *stubManager) IsInstalled(_ context.Context, name string) (bool, error) {
	if err := s.isInstalledE[name]; err != nil {
		return false, err
	}
	return s.installed[name], nil
}

func (s *stubManager) Install(_ context.Context, name string) error {
	if err := s.installE[name]; err != nil {
		return err
	}
	s.installedNow = append(s.installedNow, name)
	s.installed[name] = true
	return nil
}

func setCachedManager(name string, mgr Manager) {
	managerMu.Lock()
	defer managerMu.Unlock()
	managerCache[name] = mgr
}

func writeLockFile(t *testing.T, configDir string, mutate func(*lock.LockV3)) {
	t.Helper()
	svc := lock.NewLockV3Service(configDir)
	l := lock.NewLockV3()
	mutate(l)
	require.NoError(t, svc.Write(l))
}

func TestSimpleApply_DryRun(t *testing.T) {
	ResetManagerCache()
	t.Cleanup(ResetManagerCache)

	tmpDir := t.TempDir()
	writeLockFile(t, tmpDir, func(l *lock.LockV3) {
		l.AddPackage("brew", "ripgrep")
		l.AddPackage("brew", "fd")
	})

	mgr := &stubManager{installed: map[string]bool{"ripgrep": true, "fd": false}}
	setCachedManager("brew", mgr)

	result, err := SimpleApply(context.Background(), tmpDir, true)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"brew:ripgrep"}, result.Skipped)
	assert.ElementsMatch(t, []string{"brew:fd"}, result.WouldInstall)
	assert.Empty(t, result.Installed)
	assert.Empty(t, result.Failed)
	assert.Empty(t, mgr.installedNow)
}

func TestSimpleApply_InstallSuccess(t *testing.T) {
	ResetManagerCache()
	t.Cleanup(ResetManagerCache)

	tmpDir := t.TempDir()
	writeLockFile(t, tmpDir, func(l *lock.LockV3) {
		l.AddPackage("brew", "fd")
	})

	mgr := &stubManager{installed: map[string]bool{"fd": false}}
	setCachedManager("brew", mgr)

	result, err := SimpleApply(context.Background(), tmpDir, false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"brew:fd"}, result.Installed)
	assert.ElementsMatch(t, []string{"fd"}, mgr.installedNow)
	assert.Empty(t, result.Failed)
}

func TestSimpleApply_CollectsPerPackageFailures(t *testing.T) {
	ResetManagerCache()
	t.Cleanup(ResetManagerCache)

	tmpDir := t.TempDir()
	writeLockFile(t, tmpDir, func(l *lock.LockV3) {
		l.AddPackage("brew", "ok")
		l.AddPackage("brew", "bad-check")
		l.AddPackage("brew", "bad-install")
	})

	mgr := &stubManager{
		installed:    map[string]bool{"ok": false, "bad-check": false, "bad-install": false},
		isInstalledE: map[string]error{"bad-check": errors.New("check failed")},
		installE:     map[string]error{"bad-install": errors.New("install failed")},
	}
	setCachedManager("brew", mgr)

	result, err := SimpleApply(context.Background(), tmpDir, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2 package(s) failed")
	assert.ElementsMatch(t, []string{"brew:bad-check", "brew:bad-install"}, result.Failed)
	assert.ElementsMatch(t, []string{"brew:ok"}, result.Installed)
	require.Len(t, result.Errors, 2)
	assert.Contains(t, result.Errors[0].Error()+result.Errors[1].Error(), "bad-check")
	assert.Contains(t, result.Errors[0].Error()+result.Errors[1].Error(), "bad-install")
}

func TestSimpleApply_UnsupportedManagerFailsEachPackage(t *testing.T) {
	ResetManagerCache()
	t.Cleanup(ResetManagerCache)

	tmpDir := t.TempDir()
	writeLockFile(t, tmpDir, func(l *lock.LockV3) {
		l.AddPackage("npm", "typescript")
		l.AddPackage("npm", "eslint")
	})

	result, err := SimpleApply(context.Background(), tmpDir, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2 package(s) failed")
	assert.ElementsMatch(t, []string{"npm:eslint", "npm:typescript"}, result.Failed)
	require.Len(t, result.Errors, 2)
	assert.Contains(t, result.Errors[0].Error()+result.Errors[1].Error(), "manager not available")
}
