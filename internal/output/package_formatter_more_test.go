package output

import "testing"

func TestPackageOperationFormatter_TableOutput_DryRunMapping(t *testing.T) {
	data := PackageOperationOutput{
		Command:    "install",
		TotalItems: 2,
		DryRun:     true,
		Results: []SerializableOperationResult{
			{Name: "a", Manager: "brew", Status: "added"},
			{Name: "b", Manager: "npm", Status: "removed"},
		},
		Summary: PackageOperationSummary{Succeeded: 2},
	}
	out := NewPackageOperationFormatter(data).TableOutput()
	if !(contains(out, "would-install") && contains(out, "would-remove")) {
		t.Fatalf("expected dry-run mappings in output, got:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || (len(s) > 0 && (stringIndex(s, sub) >= 0)))
}

// simple substring search to avoid importing strings
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
