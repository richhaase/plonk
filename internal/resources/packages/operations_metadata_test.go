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
	// Go is no longer a built-in manager and does not have special-case
	// metadata handling in core. Custom managers should be covered by
	// configuration-driven behavior instead.
}
