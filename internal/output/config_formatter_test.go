package output

import (
	"strings"
	"testing"
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
