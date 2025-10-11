package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

// fakeSimpleMgr returns fixed version and succeeds
type fakeSimpleMgr struct{}

func (f *fakeSimpleMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeSimpleMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeSimpleMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeSimpleMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeSimpleMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (f *fakeSimpleMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (f *fakeSimpleMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeSimpleMgr) Dependencies() []string                           { return nil }

func TestInstallPackagesWith_NpmScopedMetadata(t *testing.T) {
	cfg := &config.Config{DefaultManager: "npm"}
	lockSvc := NewMockLockService()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("npm", func() PackageManager { return &fakeSimpleMgr{} }) })
	reg := NewManagerRegistry()
	_, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"@scope/tool"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(lockSvc.addCalls) != 1 {
		t.Fatalf("expected 1 add call")
	}
	meta := lockSvc.addCalls[0].Metadata
	if meta["full_name"].(string) != "@scope/tool" || meta["scope"].(string) != "@scope" {
		t.Fatalf("expected scoped metadata, got %#v", meta)
	}
}

func TestInstallPackagesWith_GoSourcePathMetadata(t *testing.T) {
	cfg := &config.Config{DefaultManager: "go"}
	lockSvc := NewMockLockService()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("go", func() PackageManager { return &fakeSimpleMgr{} }) })
	reg := NewManagerRegistry()
	_, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"github.com/user/project/cmd/tool"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(lockSvc.addCalls) != 1 {
		t.Fatalf("expected 1 add call")
	}
	meta := lockSvc.addCalls[0].Metadata
	if meta["source_path"].(string) != "github.com/user/project/cmd/tool" {
		t.Fatalf("expected source_path recorded, got %#v", meta)
	}
}
