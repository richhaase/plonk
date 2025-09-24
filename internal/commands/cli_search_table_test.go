package commands

import (
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestCLI_Search_Table_Found_Brew(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "brew:jq"}, func(env CLITestEnv) {
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["brew search jq"] = packages.CommandResponse{Output: []byte("jq\njq-extra\n"), Error: nil}
	})
	if err != nil {
		t.Fatalf("search table failed: %v\n%s", err, out)
	}
	// Expect message and package list lines
	wants := []string{
		"Matching packages:",
		"jq",
		"Install with: plonk install brew:jq",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected %q in output, got:\n%s", w, out)
		}
	}
}

func TestCLI_Search_Table_FoundMultiple_Brew_Npm(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "jq"}, func(env CLITestEnv) {
		// Make both managers available and return results
		env.Executor.Responses["brew --version"] = packages.CommandResponse{Output: []byte("Homebrew 4.0"), Error: nil}
		env.Executor.Responses["npm --version"] = packages.CommandResponse{Output: []byte("10.0.0"), Error: nil}
		env.Executor.Responses["brew search jq"] = packages.CommandResponse{Output: []byte("jq\n"), Error: nil}
		env.Executor.Responses["npm search jq --json"] = packages.CommandResponse{Output: []byte(`[{"name":"jq"}]`), Error: nil}
	})
	if err != nil {
		t.Fatalf("search table multiple failed: %v\n%s", err, out)
	}
	// Expect grouped results section
	wants := []string{
		"Results by manager:",
		"Brew:",
		"Npm:",
		"plonk install brew:jq",
		"plonk install npm:jq",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected %q in output, got:\n%s", w, out)
		}
	}
}
