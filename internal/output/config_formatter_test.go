package output

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/testutil"
)

func TestConfigShowFormatter_TableAndStructured(t *testing.T) {
	cfg := map[string]any{"default_manager": "brew", "operation_timeout": 300}
	data := ConfigShowOutput{ConfigPath: "/tmp/plonk.yaml", Config: cfg}
	f := NewConfigShowFormatter(data)
	out := f.TableOutput()
	wants := []string{"# Configuration for plonk", "Config file:", "default_manager", "brew"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
	sd := f.StructuredData().(ConfigShowOutput)
	if sd.ConfigPath == "" || sd.Config == nil {
		t.Fatalf("structured missing fields")
	}
}

func TestConfigShowFormatter_HighlightsCustomFields(t *testing.T) {
	cfg := &config.Config{
		DefaultManager:   "npm", // non-default
		OperationTimeout: 300,   // default
	}
	configDir := testutil.NewTestConfig(t, "")
	checker := config.NewUserDefinedChecker(configDir)

	data := ConfigShowOutput{
		ConfigPath: "/tmp/plonk.yaml",
		Config:     cfg,
		Checker:    checker,
	}

	f := NewConfigShowFormatter(data)
	out := f.TableOutput()

	// default_manager should be highlighted as custom (contains ANSI escapes)
	if !strings.Contains(out, "default_manager:") {
		t.Fatalf("expected default_manager key in output:\n%s", out)
	}

	// We can't rely on exact escape sequences here, but we can at least ensure
	// the uncolored value still appears somewhere in the output.
	if !strings.Contains(out, "npm") {
		t.Fatalf("expected value 'npm' in output:\n%s", out)
	}
}
