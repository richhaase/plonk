package commands

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCLI_Dotfiles_JSON_Missing_Managed_Untracked(t *testing.T) {
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

	// Managed (create same file in home target)
	out, err = RunCLI(t, []string{"dotfiles", "-o", "json", "--managed"}, func(env CLITestEnv) {
		seedDotfile(env.T, env.ConfigDir, "gitconfig", "[user]\n\tname = Test\n")
		// Create deployed file
		homeTarget := filepath.Join(env.HomeDir, ".gitconfig")
		if err := osWriteFileAll(homeTarget, "[user]\n\tname = Test\n"); err != nil {
			t.Fatalf("failed to seed home target: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("dotfiles --managed failed: %v\n%s", err, out)
	}
	var managed struct {
		Summary struct {
			Managed int `json:"managed"`
		} `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &managed); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if managed.Summary.Managed < 1 {
		t.Fatalf("expected managed >=1, got: %+v", managed)
	}

	// Untracked (create a file in home not configured)
	out, err = RunCLI(t, []string{"dotfiles", "-o", "json", "--untracked", "-v"}, func(env CLITestEnv) {
		untracked := filepath.Join(env.HomeDir, ".plonk-untracked-rc")
		if err := osWriteFileAll(untracked, "# test\n"); err != nil {
			t.Fatalf("failed to seed untracked: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("dotfiles --untracked failed: %v\n%s", err, out)
	}
	var untracked struct {
		Summary struct {
			Untracked int `json:"untracked"`
		} `json:"summary"`
	}
	if e := json.Unmarshal([]byte(out), &untracked); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if untracked.Summary.Untracked < 1 {
		t.Fatalf("expected untracked >=1, got: %+v", untracked)
	}
}
