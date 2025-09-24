package output

import (
	"strings"
	"testing"
)

func TestSearchFormatter_FoundMultiple_Table(t *testing.T) {
	data := SearchOutput{
		Package: "jq",
		Status:  "found-multiple",
		Message: "Found 2 result(s) for 'jq' across 2 manager(s): brew, npm",
		Results: []SearchResultEntry{
			{Manager: "brew", Packages: []string{"jq"}},
			{Manager: "npm", Packages: []string{"jq-cli"}},
		},
	}
	out := NewSearchFormatter(data).TableOutput()
	wants := []string{"Results by manager:", "brew:", "npm:", "Install examples:"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
}
