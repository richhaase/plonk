package diagnostics

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPermissions_FailOnUnwritableConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits differ on windows")
	}
	dir := t.TempDir()
	// Make directory read-only
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })

	rep := RunHealthChecks()
	found := false
	for _, c := range rep.Checks {
		if c.Category == "permissions" && c.Status == "fail" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected permissions fail when config dir not writable: %s", filepath.Base(dir))
	}
}
