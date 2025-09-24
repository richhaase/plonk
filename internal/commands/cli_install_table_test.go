package commands

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// seedLockPackage writes a lock entry for a given manager:name
func seedLockPackage(t *testing.T, configDir, manager, name, version string) {
	t.Helper()
	svc := lock.NewYAMLLockService(configDir)
	meta := map[string]interface{}{"manager": manager, "name": name, "version": version}
	if err := svc.AddPackage(manager, name, version, meta); err != nil {
		t.Fatalf("failed seeding lock pkg: %v", err)
	}
}

func TestCLI_Install_Table_MixedResults(t *testing.T) {
	out, err := RunCLI(t, []string{"install", "npm:okpkg", "npm:failpkg", "npm:skippkg"}, func(env CLITestEnv) {
		// Seed skippkg so it is marked as already managed (skipped)
		seedLockPackage(env.T, env.ConfigDir, "npm", "skippkg", "1.0.0")

		// Make npm available
		env.Executor.Responses["npm --version"] = packages.CommandResponse{Output: []byte("10.0.0"), Error: nil}

		// okpkg install succeeds
		env.Executor.Responses["npm install -g okpkg"] = packages.CommandResponse{Output: []byte(""), Error: nil}
		// InstalledVersion path for npm
		env.Executor.Responses["npm list -g okpkg --depth=0 --json"] = packages.CommandResponse{Output: []byte(`{"dependencies":{"okpkg":{"version":"1.2.3"}}}`), Error: nil}

		// failpkg install fails
		env.Executor.Responses["npm install -g failpkg"] = packages.CommandResponse{Output: []byte(""), Error: &packages.MockExitError{Code: 1}}
	})
	if err != nil {
		t.Fatalf("install table mixed failed: %v\n%s", err, out)
	}

	// Basic table shape + rows for each status
	wants := []string{
		"Package Install",          // title
		"PACKAGE  MANAGER  STATUS", // headers
		"okpkg",                    // package names
		"failpkg",
		"skippkg",
		"npm",     // manager column
		"added",   // success status
		"failed",  // failed status
		"skipped", // skipped status
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected output to contain %q, got:\n%s", w, out)
		}
	}
}
