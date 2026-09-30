package output

import (
	"strings"
	"testing"
)

func TestDotfileRemovalFormatter_TableOutput_SingleAndBatch(t *testing.T) {
	// Single file removed
	d1 := DotfileRemovalOutput{
		TotalFiles: 1,
		Results: []SerializableRemovalResult{{
			Name:     "~/.vimrc",
			Status:   "removed",
			Metadata: map[string]interface{}{"source": "vimrc"},
		}},
		Summary: DotfileRemovalSummary{Removed: 1},
	}
	out := NewDotfileRemovalFormatter(d1).TableOutput()
	if !strings.Contains(out, "removed  ~/.vimrc") {
		t.Fatalf("unexpected: %s", out)
	}

	// Batch dry-run
	d2 := DotfileRemovalOutput{
		TotalFiles: 2,
		Results:    []SerializableRemovalResult{{Name: "~/.zshrc", Status: "would-remove"}, {Name: "~/.vimrc", Status: "would-remove"}},
		Summary:    DotfileRemovalSummary{},
	}
	out2 := NewDotfileRemovalFormatter(d2).TableOutput()
	if !strings.Contains(out2, "2 planned") {
		t.Fatalf("unexpected: %s", out2)
	}

	// Batch with mixed statuses
	d3 := DotfileRemovalOutput{
		TotalFiles: 3,
		Results: []SerializableRemovalResult{
			{Name: "a", Status: "removed"},
			{Name: "b", Status: "skipped", Error: "not managed"},
			{Name: "c", Status: "failed", Error: "oops"},
		},
		Summary: DotfileRemovalSummary{Removed: 1, Skipped: 1, Failed: 1},
	}
	out3 := NewDotfileRemovalFormatter(d3).TableOutput()
	if !(strings.Contains(out3, "1 removed") && strings.Contains(out3, "1 skipped") && strings.Contains(out3, "1 failed")) {
		t.Fatalf("unexpected: %s", out3)
	}
}
