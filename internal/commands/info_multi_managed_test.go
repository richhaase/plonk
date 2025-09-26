package commands

import (
	"context"
	"os"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestInfo_ManagedMultipleManagersMessage(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	_ = svc.AddPackage("npm", "jq", "2.0.0", map[string]interface{}{"manager": "npm", "name": "jq", "version": "2.0.0"})

	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &fakeInfoMgr{name: "brew", installed: true} })
		r.Register("npm", func() packages.PackageManager { return &fakeInfoMgr{name: "npm", installed: true} })
	})

	res, err := Info(context.Background(), "jq")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Status != "managed" {
		t.Fatalf("expected managed, got %s", res.Status)
	}
	if res.Message == "" || res.Message[len(res.Message)-1] == ')' && res.Message[len(res.Message)-2] == '0' {
		t.Fatalf("expected message to mention other manager(s): %q", res.Message)
	}
}
