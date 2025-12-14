package commands

import (
	"encoding/json"
	"testing"
)

func TestCLI_Dotfiles_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"dotfiles", "-o", "json"}, func(env CLITestEnv) {
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("dotfiles json failed: %v\n%s", err, out)
	}
	var payload struct {
		Summary struct {
			Missing int `json:"missing"`
		} `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	// Should show at least 1 missing (the seeded dotfile)
	if payload.Summary.Missing < 1 {
		t.Fatalf("expected missing >=1, got: %+v", payload)
	}
}
