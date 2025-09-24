package output

import (
	"strings"
	"testing"
)

func TestPackageOperationFormatter_Table_Install_DryRun(t *testing.T) {
	data := PackageOperationOutput{
		Command:    "install",
		TotalItems: 2,
		DryRun:     true,
		Results: []SerializableOperationResult{
			{Name: "jq", Manager: "brew", Status: "added"},
			{Name: "typescript", Manager: "npm", Status: "added"},
		},
		Summary: PackageOperationSummary{Succeeded: 0, Skipped: 0, Failed: 0},
	}
	out := NewPackageOperationFormatter(data).TableOutput()
	wants := []string{"Package Install (Dry Run)", "jq", "would-install", "typescript", "Total: 2 processed"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
}

func TestPackageOperationFormatter_Table_Uninstall(t *testing.T) {
	data := PackageOperationOutput{
		Command:    "uninstall",
		TotalItems: 1,
		Results:    []SerializableOperationResult{{Name: "jq", Manager: "brew", Status: "removed"}},
		Summary:    PackageOperationSummary{Succeeded: 1},
	}
	out := NewPackageOperationFormatter(data).TableOutput()
	if !strings.Contains(out, "Package Uninstall") || !strings.Contains(out, "jq") {
		t.Fatalf("unexpected:\n%s", out)
	}
}
