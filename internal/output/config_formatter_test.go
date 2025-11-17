package output

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Expect the colored version of the default_manager line.
	coloredLine := ColorInfo("default_manager: npm")
	assert.Contains(t, out, coloredLine)
}

func TestFormatConfigWithHighlights_ListItems(t *testing.T) {
	defaults := config.GetDefaults()
	cfg := *defaults
	cfg.ExpandDirectories = []string{".config", ".claude"}

	configDir := testutil.NewTestConfig(t, "")
	checker := config.NewUserDefinedChecker(configDir)

	out, err := formatConfigWithHighlights(&cfg, checker)
	require.NoError(t, err)

	// Default entry should be present (uncolored).
	assert.Contains(t, out, "  - .config")

	// Custom entry should be highlighted in green.
	coloredItem := ColorAdded("  - .claude")
	assert.Contains(t, out, coloredItem)
}
