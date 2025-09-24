package commands

import (
	"strings"
	"testing"
)

func TestCLI_Apply_Table_DryRun_GoldenCore(t *testing.T) {
	out, err := RunCLI(t, []string{"apply", "--dry-run"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("apply --dry-run failed: %v\n%s", err, out)
	}

	// Assert core, deterministic section of the table output
	core := []string{
		"Plonk Apply (Dry Run)",
		"Summary:",
		"Packages:",
		"Dotfiles:",
	}
	for _, s := range core {
		if !strings.Contains(out, s) {
			t.Fatalf("expected output to contain %q, got:\n%s", s, out)
		}
	}
}

func TestCLI_Apply_Table_DryRun_GoldenRows(t *testing.T) {
	out, err := RunCLI(t, []string{"apply", "--dry-run"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("apply --dry-run failed: %v\n%s", err, out)
	}

	// Expect recognizable rows
	wants := []string{
		"jq",            // package name appears
		"would install", // package dry-run marker
		".zshrc",        // dotfile destination label in display
		"would deploy",  // dotfile dry-run marker
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected output to contain %q, got:\n%s", w, out)
		}
	}
}
