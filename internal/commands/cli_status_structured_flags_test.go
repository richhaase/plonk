package commands

import (
	"strings"
	"testing"
)

// TestCLI_Status_DefaultShowsBothDomains verifies that running
// `plonk status` without domain flags returns both package and dotfile results.
func TestCLI_Status_DefaultShowsBothDomains(t *testing.T) {
	out, err := RunCLI(t, []string{"status"}, func(env CLITestEnv) {
		seedLock(env.T, env.ConfigDir)
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
	})
	if err != nil {
		t.Fatalf("status failed: %v\n%s", err, out)
	}

	// Should have both package and dotfile table headers in output
	if !strings.Contains(out, "PACKAGE") {
		t.Errorf("expected PACKAGE table header in default output, got:\n%s", out)
	}
	if !strings.Contains(out, "DOTFILE") {
		t.Errorf("expected DOTFILE table header in default output, got:\n%s", out)
	}
}
