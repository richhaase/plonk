package commands

import (
	"encoding/json"
	"testing"
)

// TestCLI_Status_JSON_DefaultShowsBothDomains verifies that running
// `plonk status -o json` without domain flags returns both package and dotfile results.
// This is the default behavior matching table output.
func TestCLI_Status_JSON_DefaultShowsBothDomains(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json (default) failed: %v\n%s", err, out)
	}

	var payload struct {
		StateSummary struct {
			Results []struct {
				Domain string `json:"domain"`
			} `json:"results"`
		} `json:"state_summary"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}

	// Should have both package and dotfile domains
	domains := make(map[string]bool)
	for _, r := range payload.StateSummary.Results {
		domains[r.Domain] = true
	}
	if !domains["package"] {
		t.Errorf("expected package domain in default output, got domains: %v", domains)
	}
	if !domains["dotfile"] {
		t.Errorf("expected dotfile domain in default output, got domains: %v", domains)
	}
}
