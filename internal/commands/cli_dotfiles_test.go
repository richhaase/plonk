package commands

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCLI_Dotfiles_JSON_Missing_Unmanaged(t *testing.T) {
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

	// Unmanaged (create a file in home not configured)
	out, err = RunCLI(t, []string{"dotfiles", "-o", "json", "--unmanaged"}, func(env CLITestEnv) {
		untracked := filepath.Join(env.HomeDir, ".plonk-untracked-rc")
		if err := osWriteFileAll(untracked, "# test\n"); err != nil {
			t.Fatalf("failed to seed untracked: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("dotfiles --unmanaged failed: %v\n%s", err, out)
	}
	var unmanaged struct {
		Summary struct {
			Untracked int `json:"untracked"`
		} `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &unmanaged); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if unmanaged.Summary.Untracked < 1 {
		t.Fatalf("expected untracked >=1, got: %+v", unmanaged)
	}
}
