package commands

import (
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestCLI_Upgrade_Table_Basic(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one brew package
		seedLock(env.T, env.ConfigDir)
		// Make brew available
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		// Pre-upgrade installed info
		env.Executor.Responses["brew info --installed --json=v2"] = packages.CommandResponse{Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil}
		// Upgrade command
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte(""), Error: nil}
		// Post-upgrade installed info (same version is acceptable)
		env.Executor.Responses["brew info --installed --json=v2"] = packages.CommandResponse{Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil}
	})
	if err != nil {
		t.Fatalf("upgrade table failed: %v\n%s", err, out)
	}
	wants := []string{
		"Package Upgrade Results",
		"Brew:",
		"jq",
		"Summary:",
		"Total:",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected %q in output, got:\n%s", w, out)
		}
	}
}
