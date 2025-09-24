// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"gopkg.in/yaml.v3"
)

// helper: seed a simple lock with one brew package
func seedLock(t *testing.T, configDir string) {
	t.Helper()
	svc := lock.NewYAMLLockService(configDir)
	meta := map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"}
	if err := svc.AddPackage("brew", "jq", "1.0.0", meta); err != nil {
		t.Fatalf("failed seeding lock: %v", err)
	}
}

// helper: write a simple dotfile source (missing in HOME by default)
func seedDotfile(t *testing.T, configDir string, name string, contents string) {
	t.Helper()
	path := filepath.Join(configDir, name)
	if err := osWriteFileAll(path, contents); err != nil {
		t.Fatalf("failed writing dotfile source: %v", err)
	}
}

// minimal wrapper to write file with dirs created
func osWriteFileAll(path, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(contents), 0o644)
}

func TestCLI_Status_JSON_and_YAML(t *testing.T) {
	out, err := RunCLI(t, []string{"status", "-o", "json", "--packages", "--dotfiles"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status json failed: %v\n%s", err, out)
	}

	// Decode JSON and assert key fields
	var payload struct {
		ConfigPath   string `json:"config_path"`
		LockPath     string `json:"lock_path"`
		ConfigExists bool   `json:"config_exists"`
		LockExists   bool   `json:"lock_exists"`
		StateSummary struct {
			TotalMissing int `json:"total_missing"`
			Results      []struct {
				Domain  string `json:"domain"`
				Missing []struct {
					Name    string `json:"name"`
					Manager string `json:"manager"`
				} `json:"missing"`
			} `json:"results"`
		} `json:"state_summary"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if payload.LockPath == "" || !payload.LockExists {
		t.Fatalf("expected lock to exist in payload: %+v", payload)
	}
	// Expect at least 2 missing (1 package + 1 dotfile)
	if payload.StateSummary.TotalMissing < 2 {
		t.Fatalf("expected >=2 missing, got %d", payload.StateSummary.TotalMissing)
	}
	// Expect both domains present
	foundPkg, foundDot := false, false
	for _, r := range payload.StateSummary.Results {
		if r.Domain == "package" {
			foundPkg = true
		}
		if r.Domain == "dotfile" {
			foundDot = true
		}
	}
	if !foundPkg || !foundDot {
		t.Fatalf("expected package and dotfile domains, got: %+v", payload.StateSummary.Results)
	}

	// YAML variant: basic shape check
	outY, err := RunCLI(t, []string{"status", "-o", "yaml", "--packages", "--dotfiles"}, nil)
	if err != nil {
		t.Fatalf("status yaml failed: %v\n%s", err, outY)
	}
	var y any
	if err := yaml.Unmarshal([]byte(outY), &y); err != nil {
		t.Fatalf("invalid yaml: %v\n%s", err, outY)
	}
}

func TestCLI_Apply_DryRun_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"apply", "--dry-run", "-o", "json"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "gitconfig", "[user]\n\tname = Test\n")
	})
	if err != nil {
		t.Fatalf("apply dry-run json failed: %v\n%s", err, out)
	}

	var payload struct {
		DryRun   bool `json:"dry_run"`
		Success  bool `json:"success"`
		Packages *struct {
			TotalWouldInstall int `json:"total_would_install"`
		} `json:"packages"`
		Dotfiles *struct {
			Summary struct {
				Added int `json:"added"`
			} `json:"summary"`
		} `json:"dotfiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if !payload.DryRun || !payload.Success {
		t.Fatalf("expected dry_run and success true: %+v", payload)
	}
	if payload.Packages == nil || payload.Packages.TotalWouldInstall < 1 {
		t.Fatalf("expected packages would-install >=1, got: %+v", payload.Packages)
	}
	if payload.Dotfiles == nil || payload.Dotfiles.Summary.Added < 1 {
		t.Fatalf("expected dotfiles added >=1, got: %+v", payload.Dotfiles)
	}
}

func TestCLI_Status_Table_GoldenSnippet(t *testing.T) {
	out, err := RunCLI(t, []string{"status"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status table failed: %v\n%s", err, out)
	}

	// Golden snippet: ensure sections and key rows appear
	wants := []string{
		"Plonk Status\n============",
		"PACKAGES\n--------",
		"jq",
		"brew",
		"missing",
		"DOTFILES\n--------",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected output to contain %q, got:\n%s", w, out)
		}
	}
}

func TestCLI_Status_Flags_Table(t *testing.T) {
	// packages only
	out, err := RunCLI(t, []string{"status", "--packages"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status --packages failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "DOTFILES\n--------") {
		t.Fatalf("expected DOTFILES section to be absent, got:\n%s", out)
	}
	if !strings.Contains(out, "PACKAGES\n--------") {
		t.Fatalf("expected PACKAGES section to be present, got:\n%s", out)
	}

	// dotfiles only
	out, err = RunCLI(t, []string{"status", "--dotfiles"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "gitconfig", "[user]\n\tname = Test\n")
	})
	if err != nil {
		t.Fatalf("status --dotfiles failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "PACKAGES\n--------") {
		t.Fatalf("expected PACKAGES section to be absent, got:\n%s", out)
	}
	if !strings.Contains(out, "DOTFILES\n--------") {
		t.Fatalf("expected DOTFILES section to be present, got:\n%s", out)
	}

	// missing only
	out, err = RunCLI(t, []string{"status", "--missing"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "bashrc", "export X=1\n")
	})
	if err != nil {
		t.Fatalf("status --missing failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "missing") {
		t.Fatalf("expected 'missing' entries in table, got:\n%s", out)
	}
	if strings.Contains(out, " managed\n") {
		t.Fatalf("unexpected 'managed' rows when --missing: \n%s", out)
	}
}

func TestCLI_Apply_Table_DryRun_GoldenSnippet(t *testing.T) {
	out, err := RunCLI(t, []string{"apply", "--dry-run"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("apply --dry-run failed: %v\n%s", err, out)
	}
	wants := []string{
		"Plonk Apply (Dry Run)",
		"Summary:",
		"Packages:",
		"Dotfiles:",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected output to contain %q, got:\n%s", w, out)
		}
	}
	// sanity: show would-install at least once for packages or dotfiles
	if !strings.Contains(out, "would install") && !strings.Contains(out, "would deploy") {
		t.Fatalf("expected 'would install' or 'would deploy' markers, got:\n%s", out)
	}
	_ = fmt.Sprintf("") // keep fmt import
}

func TestCLI_Install_Uninstall_DryRun_JSON(t *testing.T) {
	// install dry run
	out, err := RunCLI(t, []string{"install", "-n", "-o", "json", "brew:jq", "npm:typescript"}, nil)
	if err != nil {
		t.Fatalf("install -n json failed: %v\n%s", err, out)
	}

	var install struct {
		Command string `json:"command"`
		DryRun  bool   `json:"dry_run"`
		Results []struct {
			Name    string `json:"name"`
			Manager string `json:"manager"`
			Status  string `json:"status"`
		} `json:"results"`
	}
	if e := json.Unmarshal([]byte(out), &install); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if install.Command != "install" || !install.DryRun {
		t.Fatalf("unexpected install payload: %+v", install)
	}
	// Expect would-add entries present
	sawWould := false
	for _, r := range install.Results {
		if r.Status == "would-add" {
			sawWould = true
			break
		}
	}
	if !sawWould {
		t.Fatalf("expected at least one would-add, got: %+v", install.Results)
	}

	// uninstall dry run
	out, err = RunCLI(t, []string{"uninstall", "-n", "-o", "json", "brew:jq"}, nil)
	if err != nil {
		t.Fatalf("uninstall -n json failed: %v\n%s", err, out)
	}
	var uninstall struct {
		Command string `json:"command"`
		DryRun  bool   `json:"dry_run"`
		Results []struct {
			Status string `json:"status"`
		} `json:"results"`
	}
	if e := json.Unmarshal([]byte(out), &uninstall); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if uninstall.Command != "uninstall" || !uninstall.DryRun {
		t.Fatalf("unexpected uninstall payload: %+v", uninstall)
	}
	sawWouldRemove := false
	for _, r := range uninstall.Results {
		if r.Status == "would-remove" {
			sawWouldRemove = true
			break
		}
	}
	if !sawWouldRemove {
		t.Fatalf("expected would-remove, got: %+v", uninstall.Results)
	}
}
