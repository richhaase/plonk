package commands

import (
	"encoding/json"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestCLI_Upgrade_JSON_Basic(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade", "-o", "json"}, func(env CLITestEnv) {
		// Seed one managed package in lock
		seedLock(env.T, env.ConfigDir)
		// Make brew available
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		// Pre-upgrade version (via info JSON path or list --versions fallback)
		// Use JSON path: provide installed info so InstalledVersion can parse from v2 info
		env.Executor.Responses["brew info --installed --json=v2"] = packages.CommandResponse{Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil}
		// Upgrade command succeeds
		env.Executor.Responses["brew upgrade jq"] = packages.CommandResponse{Output: []byte(""), Error: nil}
		// Post-upgrade version (unchanged for simplicity -> skipped or upgraded without version change acceptable)
		env.Executor.Responses["brew info --installed --json=v2"] = packages.CommandResponse{Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil}
	})
	if err != nil {
		t.Fatalf("upgrade json failed: %v\n%s", err, out)
	}

	var payload struct {
		Command    string                                         `json:"command"`
		TotalItems int                                            `json:"total_items"`
		Results    []struct{ Manager, Package, Status string }    `json:"results"`
		Summary    struct{ Total, Upgraded, Failed, Skipped int } `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Command != "upgrade" || payload.TotalItems < 1 || len(payload.Results) == 0 {
		t.Fatalf("unexpected upgrade payload: %+v", payload)
	}
}
