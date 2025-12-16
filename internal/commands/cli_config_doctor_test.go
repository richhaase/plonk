package commands

import (
	"strings"
	"testing"
)

func TestCLI_ConfigShow_Table(t *testing.T) {
	out, err := RunCLI(t, []string{"config", "show"}, nil)
	if err != nil {
		t.Fatalf("config show failed: %v\n%s", err, out)
	}
	// Verify table output contains expected sections
	if !strings.Contains(out, "Configuration for plonk") {
		t.Fatalf("expected 'Configuration for plonk' in output, got:\n%s", out)
	}
}

func TestCLI_Doctor_Table(t *testing.T) {
	out, err := RunCLI(t, []string{"doctor"}, nil)
	if err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, out)
	}
	// Verify table output contains expected sections
	if !strings.Contains(out, "Plonk Doctor Report") {
		t.Fatalf("expected 'Plonk Doctor Report' in output, got:\n%s", out)
	}
}
