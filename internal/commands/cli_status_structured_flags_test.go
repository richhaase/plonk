package commands

import (
	"encoding/json"
	"testing"
)

func TestCLI_Status_JSON_WithFlags(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "--packages", "--dotfiles", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json with flags failed: %v\n%s", err, out)
	}

	var payload struct {
		ConfigPath   string `json:"config_path"`
		StateSummary struct {
			Results []struct {
				Domain string `json:"domain"`
			} `json:"results"`
		} `json:"state_summary"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if len(payload.StateSummary.Results) == 0 {
		t.Fatalf("expected results present, got: %+v", payload)
	}
}

func TestCLI_Status_JSON_MissingFlag(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "--missing", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json --missing failed: %v\n%s", err, out)
	}

	var payload struct {
		StateSummary struct {
			TotalMissing int `json:"total_missing"`
		} `json:"state_summary"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.StateSummary.TotalMissing < 1 {
		t.Fatalf("expected TotalMissing >=1 with --missing, got: %+v", payload)
	}
}
