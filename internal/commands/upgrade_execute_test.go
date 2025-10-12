package commands

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// fake managers for upgrade execution tests

func seedUpgradeLock(t *testing.T, dir string) {
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	_ = svc.AddPackage("npm", "typescript", "1.0.0", map[string]interface{}{"manager": "npm", "name": "typescript", "version": "1.0.0"})
	_ = svc.AddPackage("foo", "bar", "1.0.0", map[string]interface{}{"manager": "foo", "name": "bar", "version": "1.0.0"})
}

func TestUpgrade_ManagerUnavailableAndMixedResults(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		seedUpgradeLock(env.T, env.ConfigDir)
		// Rely on default manager configs; make brew and npm available; omit "foo" to trigger error path
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["npm --version"] = packages.CommandResponse{Output: []byte("10.0.0"), Error: nil}
		// Treat upgrades as success for brew and npm
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte("upgraded"), Error: nil}
		env.Executor.Responses["npm update typescript -g"] = packages.CommandResponse{Output: []byte("upgraded"), Error: nil}
	})
	if err == nil {
		t.Fatalf("expected non-nil error due to failures, got nil. Output:\n%s", out)
	}
	// Ensure summary line present and mentions failed/skipped/upgraded in some combination
	if !containsAll(out, []string{"Summary:", "Failed:", "Skipped:"}) {
		t.Fatalf("expected summary with failed and skipped, got:\n%s", out)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
