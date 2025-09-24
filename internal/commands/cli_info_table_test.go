package commands

import (
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestCLI_Info_Table_Managed_Brew(t *testing.T) {
	out, err := RunCLI(t, []string{"info", "brew:jq"}, func(env CLITestEnv) {
		// Mark as managed and provide minimal brew responses
		seedLock(env.T, env.ConfigDir)
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["brew list jq"] = packages.CommandResponse{Output: []byte("jq"), Error: nil}
		env.Executor.Responses["brew info jq"] = packages.CommandResponse{Output: []byte("jq: stable 1.6\nhttps://stedolan.github.io/jq/\n"), Error: nil}
	})
	if err != nil {
		t.Fatalf("info table failed: %v\n%s", err, out)
	}
	wants := []string{
		"Package:",
		"jq",
		"Manager:",
		"brew",
		"Status:",
		"Managed by plonk",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected %q in output, got:\n%s", w, out)
		}
	}
}
