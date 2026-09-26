package output

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigShowFormatter_TableOutput(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew", OperationTimeout: 300}
	data := ConfigShowOutput{ConfigPath: "/tmp/plonk.yaml", Config: cfg}
	f := NewConfigShowFormatter(data)
	out := f.TableOutput()
	wants := []string{"# Configuration for plonk", "Config file:", "default_manager", "brew"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
}

func TestConfigShowFormatter_HighlightsCustomFields(t *testing.T) {
	cfg := &config.Config{
		DefaultManager:   "npm", // non-default
		OperationTimeout: 300,   // default
	}

	data := ConfigShowOutput{
		ConfigPath: "/tmp/plonk.yaml",
		Config:     cfg,
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

	out, err := formatConfigWithHighlights(&cfg)
	require.NoError(t, err)

	// Default entry should be present (uncolored).
	assert.Contains(t, out, "  - .config")

	// Custom entry should be highlighted in green.
	coloredItem := ColorAdded("  - .claude")
	assert.Contains(t, out, coloredItem)
}

func TestFormatConfigWithHighlights_RemovedAndAddedItems(t *testing.T) {
	cfg := *config.GetDefaults()
	cfg.ExpandDirectories = []string{"custom-dir"}
	cfg.IgnorePatterns = append(append([]string{}, cfg.IgnorePatterns[1:]...), "custom-pattern")
	out, err := formatConfigWithHighlights(&cfg)
	require.NoError(t, err)
	assert.Contains(t, out, ColorAdded("  - custom-dir"))
	assert.Contains(t, out, ColorAdded("  - custom-pattern"))
	assert.Contains(t, out, ColorRemoved("# removed: - .config"))
	assert.Contains(t, out, ColorRemoved("# removed: - .DS_Store"))
	assert.Equal(t, 2, strings.Count(out, "# removed:"))
}
