package commands

import (
	"path/filepath"
	"testing"
)

func TestCLI_Diff_WithEchoTool(t *testing.T) {
	out, err := RunCLI(t, []string{"diff"}, func(env CLITestEnv) {
		// Write config specifying a safe diff tool
		cfgPath := filepath.Join(env.ConfigDir, "plonk.yaml")
		if err := osWriteFileAll(cfgPath, "diff_tool: \"echo\"\n"); err != nil {
			t.Fatalf("failed to seed config: %v", err)
		}
		// Seed a source file and a different deployed file to cause drift
		seedDotfile(env.T, env.ConfigDir, "zshrc", "export TEST=1\n")
		if err := osWriteFileAll(filepath.Join(env.HomeDir, ".zshrc"), "export TEST=2\n"); err != nil {
			t.Fatalf("failed to seed home drift: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("diff failed: %v\n%s", err, out)
	}
	// Basic smoke: echo should have printed paths; ensure no error and some output
	if len(out) == 0 {
		t.Fatalf("expected some diff output via echo tool, got empty")
	}
}
