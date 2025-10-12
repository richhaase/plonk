package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestInstallPackagesWith_NpmScopedMetadata(t *testing.T) {
	cfg := &config.Config{DefaultManager: "npm"}
	lockSvc := NewMockLockService()
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"npm --version":             {Output: []byte("10.0.0"), Error: nil},
		"npm install -g @scope/pkg": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
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
	cfg := &config.Config{DefaultManager: "go", Managers: map[string]config.ManagerConfig{
		"go": {
			Binary:  "go",
			Install: config.CommandConfig{Command: []string{"go", "install", "{{.Package}}@latest"}},
		},
	}}
	lockSvc := NewMockLockService()
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"go --version":                           {Output: []byte("go1.22"), Error: nil},
		"go install github.com/acme/tool@latest": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
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
