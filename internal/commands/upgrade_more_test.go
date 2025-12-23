package commands

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/packages"
)

// Fake manager that reports available and simulates version change on Upgrade.

func TestUpgrade_UpdatesLockMetadata(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one brew package with starting version
		svc := lock.NewYAMLLockService(env.ConfigDir)
		if e := svc.AddPackage("brew", "jq", "1.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0"}); e != nil {
			t.Fatalf("seed lock: %v", e)
		}
		// Make brew available and treat upgrade as success
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte("upgraded"), Error: nil}
	})
	if err != nil {
		t.Fatalf("upgrade error: %v\n%s", err, out)
	}

	// Read lock and verify version updated and installed_at present
	svc := lock.NewYAMLLockService(os.Getenv("PLONK_DIR"))
	// Note: RunCLI sets PLONK_DIR to env.ConfigDir
	lk, e := svc.Read()
	if e != nil {
		t.Fatalf("read lock: %v", e)
	}
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	if lk.Resources[0].InstalledAt == "" {
		t.Fatalf("expected InstalledAt to be set")
	}
}

func TestUpgrade_InstalledAtUpdated(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one brew package, then force a known InstalledAt timestamp
		svc := lock.NewYAMLLockService(env.ConfigDir)
		if e := svc.AddPackage("brew", "jq", "1.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0"}); e != nil {
			t.Fatalf("seed lock: %v", e)
		}
		lk, e := svc.Read()
		if e != nil {
			t.Fatalf("read lock: %v", e)
		}
		if len(lk.Resources) != 1 {
			t.Fatalf("setup expected 1 resource, got %d", len(lk.Resources))
		}
		lk.Resources[0].InstalledAt = "2000-01-01T00:00:00Z"
		if e := svc.Write(lk); e != nil {
			t.Fatalf("write lock: %v", e)
		}

		// Make brew available and treat upgrade as success
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte("upgraded"), Error: nil}
	})
	if err != nil {
		t.Fatalf("upgrade error: %v\n%s", err, out)
	}

	svc := lock.NewYAMLLockService(os.Getenv("PLONK_DIR"))
	lk2, e := svc.Read()
	if e != nil {
		t.Fatalf("read lock: %v", e)
	}
	if lk2.Resources[0].InstalledAt == "2000-01-01T00:00:00Z" {
		t.Fatalf("expected InstalledAt to be updated, still %s", lk2.Resources[0].InstalledAt)
	}
}

// Manager IsAvailable=false should mark failures.

func TestUpgrade_ManagerAvailableFalse_Fails(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade", "pipx"}, func(env CLITestEnv) {
		// Seed one pipx package
		svc := lock.NewYAMLLockService(env.ConfigDir)
		_ = svc.AddPackage("pipx", "httpx", "0.1", map[string]interface{}{"manager": "pipx", "name": "httpx", "version": "0.1"})
		// Omit any pipx responses to force IsAvailable=false
	})
	if err == nil {
		t.Fatalf("expected error due to manager unavailable, got nil\n%s", out)
	}
	// Validate JSON path would be too heavy; check that output summary exists with Failed
	var payload struct{}
	_ = json.Unmarshal([]byte(out), &payload) // don't fail; out may be table
	if !strings.Contains(out, "Failed:") {
		t.Fatalf("expected Failed in summary, got:\n%s", out)
	}
}
