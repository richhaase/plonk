// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
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

func TestCLI_Status_Table_GoldenSnippet(t *testing.T) {
	out, err := RunCLI(t, []string{"status"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status table failed: %v\n%s", err, out)
	}

	// Golden snippet: ensure table headers and key rows appear
	wants := []string{
		"Plonk Status\n============",
		"PACKAGE",
		"jq",
		"brew",
		"missing",
		"DOTFILE",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected output to contain %q, got:\n%s", w, out)
		}
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

func TestCLI_Install_DryRun_Table(t *testing.T) {
	// install dry run
	out, err := RunCLI(t, []string{"install", "-n", "brew:jq", "npm:typescript"}, nil)
	if err != nil {
		t.Fatalf("install -n failed: %v\n%s", err, out)
	}

	// Verify table output shows would-add entries
	if !strings.Contains(out, "would-add") {
		t.Fatalf("expected 'would-add' in output, got:\n%s", out)
	}
}

func TestCLI_Uninstall_DryRun_Table(t *testing.T) {
	// uninstall dry run
	out, err := RunCLI(t, []string{"uninstall", "-n", "brew:jq"}, nil)
	if err != nil {
		t.Fatalf("uninstall -n failed: %v\n%s", err, out)
	}

	// Verify table output shows would-remove entries
	if !strings.Contains(out, "would-remove") {
		t.Fatalf("expected 'would-remove' in output, got:\n%s", out)
	}
}
