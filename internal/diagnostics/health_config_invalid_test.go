package diagnostics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurationValidity_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })
	// Write invalid YAML
	if err := os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte("default_manager: [oops"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep := RunHealthChecks()
	found := false
	for _, c := range rep.Checks {
		if c.Category == "configuration" && c.Name == "Configuration Validity" && c.Status == "fail" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected configuration validity fail check")
	}
}
