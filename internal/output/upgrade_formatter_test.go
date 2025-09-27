package output

import (
	"strings"
	"testing"
)

func TestUpgradeFormatter_TableAndStructured(t *testing.T) {
	data := UpgradeOutput{
		Command:    "upgrade",
		TotalItems: 3,
		Results: []UpgradeResult{
			{Manager: "brew", Package: "jq", FromVersion: "1.6", ToVersion: "1.6.1", Status: "upgraded"},
			{Manager: "brew", Package: "fd", FromVersion: "8.7.0", ToVersion: "8.7.0", Status: "skipped"},
			{Manager: "npm", Package: "typescript", Status: "failed", Error: "timeout"},
		},
		Summary: UpgradeSummary{Total: 3, Upgraded: 1, Failed: 1, Skipped: 1},
	}
	f := NewUpgradeFormatter(data)
	out := f.TableOutput()
	// Core snippets
	wants := []string{"Package Upgrade Results", "Brew:", "jq (upgraded", "fd (already up-to-date)", "Npm:", "typescript (failed", "Summary:", "Total: 3"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
	if f.StructuredData().(UpgradeOutput).Command != "upgrade" {
		t.Fatalf("structured mismatch")
	}
}

// nothing
