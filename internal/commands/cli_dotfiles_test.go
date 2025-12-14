package commands

import (
	"encoding/json"
	"testing"
)

func TestCLI_Dotfiles_JSON_Missing(t *testing.T) {
	// Missing
	out, err := RunCLI(t, []string{"dotfiles", "-o", "json", "--missing"}, func(env CLITestEnv) {
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("dotfiles --missing failed: %v\n%s", err, out)
	}
	var missing struct {
		Summary struct {
			Missing int `json:"missing"`
		} `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &missing); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if missing.Summary.Missing < 1 {
		t.Fatalf("expected missing >=1, got: %+v", missing)
	}
}
