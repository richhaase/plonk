package output

import (
	"strings"
	"testing"
)

func TestInfoFormatter_Table_Managed(t *testing.T) {
	data := InfoOutput{
		Package:     "jq",
		Status:      "managed",
		Message:     "Package 'jq' is managed by plonk via brew",
		PackageInfo: map[string]any{"name": "jq", "version": "1.6", "manager": "brew", "installed": true},
	}
	out := NewInfoFormatter(data).TableOutput()
	wants := []string{"jq", "managed"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
}
