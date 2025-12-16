package commands

import (
	"strings"
	"testing"
)

func TestCLI_Dotfiles_Table(t *testing.T) {
	out, err := RunCLI(t, []string{"dotfiles"}, func(env CLITestEnv) {
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("dotfiles failed: %v\n%s", err, out)
	}
	// Verify table output contains expected sections
	if !strings.Contains(out, "Dotfiles Status") {
		t.Fatalf("expected 'Dotfiles Status' in output, got:\n%s", out)
	}
}
