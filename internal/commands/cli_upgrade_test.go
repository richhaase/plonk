package commands

import (
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/packages"
)

func TestCLI_Upgrade_Table_Basic(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one managed package in lock
		seedLock(env.T, env.ConfigDir)
		// Make brew available
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		// Pre-upgrade version
		env.Executor.Responses["brew info --installed --json=v2"] = packages.CommandResponse{Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil}
		// Upgrade command succeeds
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte(""), Error: nil}
	})
	if err != nil {
		t.Fatalf("upgrade failed: %v\n%s", err, out)
	}

	// Verify table output contains expected content
	if !strings.Contains(out, "Summary:") {
		t.Fatalf("expected 'Summary:' in output, got:\n%s", out)
	}
}
