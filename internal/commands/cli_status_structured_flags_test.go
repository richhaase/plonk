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

// TestCLI_Status_JSON_PackagesOnlyFilter verifies that --packages flag
// filters JSON output to only include package results.
func TestCLI_Status_JSON_PackagesOnlyFilter(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "--packages", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json --packages failed: %v\n%s", err, out)
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

	// Should only have package domain
	for _, r := range payload.StateSummary.Results {
		if r.Domain == "dotfile" {
			t.Errorf("expected no dotfile domain with --packages flag, got: %v", payload.StateSummary.Results)
		}
	}
	// Should have at least the package domain
	hasPackage := false
	for _, r := range payload.StateSummary.Results {
		if r.Domain == "package" {
			hasPackage = true
			break
		}
	}
	if !hasPackage {
		t.Errorf("expected package domain with --packages flag, got: %v", payload.StateSummary.Results)
	}
}

// TestCLI_Status_JSON_DotfilesOnlyFilter verifies that --dotfiles flag
// filters JSON output to only include dotfile results.
func TestCLI_Status_JSON_DotfilesOnlyFilter(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "--dotfiles", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json --dotfiles failed: %v\n%s", err, out)
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

	// Should only have dotfile domain
	for _, r := range payload.StateSummary.Results {
		if r.Domain == "package" {
			t.Errorf("expected no package domain with --dotfiles flag, got: %v", payload.StateSummary.Results)
		}
	}
	// Should have at least the dotfile domain
	hasDotfile := false
	for _, r := range payload.StateSummary.Results {
		if r.Domain == "dotfile" {
			hasDotfile = true
			break
		}
	}
	if !hasDotfile {
		t.Errorf("expected dotfile domain with --dotfiles flag, got: %v", payload.StateSummary.Results)
	}
}

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
