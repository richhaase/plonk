package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

// lockSvc that reports package as managed and can fail RemovePackage
type failingRemoveLock struct{ removeErr error }

func (f *failingRemoveLock) Read() (*lock.Lock, error) {
	return &lock.Lock{Version: lock.LockFileVersion}, nil
}
func (f *failingRemoveLock) Write(l *lock.Lock) error { return nil }
func (f *failingRemoveLock) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	return nil
}
func (f *failingRemoveLock) RemovePackage(manager, name string) error { return f.removeErr }
func (f *failingRemoveLock) GetPackages(manager string) ([]lock.ResourceEntry, error) {
	return nil, nil
}
func (f *failingRemoveLock) HasPackage(manager, name string) bool         { return true }
func (f *failingRemoveLock) FindPackage(name string) []lock.ResourceEntry { return nil }

func TestUninstall_Managed_UninstallErrorButRemovedFromLock(t *testing.T) {
	cfg := &config.Config{}
	l := &failingRemoveLock{removeErr: nil}
	// Mock uninstall error
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("boom"), Error: &MockExitError{Code: 1}},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" {
		t.Fatalf("expected removed with error, got %+v", res)
	}
}

func TestUninstall_Managed_RemoveFromLockFails(t *testing.T) {
	cfg := &config.Config{}
	l := &failingRemoveLock{removeErr: errors.New("lock fail")}
	// Success uninstall
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with error due to lock removal, got %+v", res)
	}
}

func TestUninstall_Unmanaged_PassThrough(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	// Use real YAML lock in temp dir (empty means unmanaged)
	l := lock.NewYAMLLockService(t.TempDir())
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" {
		t.Fatalf("expected removed pass-through, got %+v", res)
	}
}
