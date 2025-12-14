package output

import (
	"strings"
	"testing"
)

// Test that drifted dotfiles are labeled correctly in status output
func TestStatusFormatter_DriftedLabel(t *testing.T) {
	// One drifted (StateDegraded) and one managed
	dotfileItems := []Item{
		{Name: ".config/nvim/lazy-lock.json", State: StateDegraded},
		{Name: ".zshrc", State: StateManaged},
	}

	data := StatusOutput{
		StateSummary: Summary{
			TotalManaged: 2,
			Results: []Result{
				{
					Domain:  "dotfile",
					Managed: dotfileItems,
				},
			},
		},
	}

	out := NewStatusFormatter(data).TableOutput()

	if !strings.Contains(out, "drifted") {
		t.Fatalf("expected output to contain 'drifted' label; got:\n%s", out)
	}

	if strings.Contains(out, "degraded") {
		t.Fatalf("did not expect output to contain 'degraded'; got:\n%s", out)
	}
}

// Test that summary excludes drifted from managed count and shows drifted count
func TestStatusFormatter_SummaryCountsExcludeDrifted(t *testing.T) {
	dotfileItems := []Item{
		{Name: ".config/nvim/lazy-lock.json", State: StateDegraded},
		{Name: ".zshrc", State: StateManaged},
	}

	data := StatusOutput{
		StateSummary: Summary{
			TotalManaged: 2, // Includes the drifted item by design
			Results: []Result{
				{Domain: "dotfile", Managed: dotfileItems},
			},
		},
	}

	out := NewStatusFormatter(data).TableOutput()

	if !strings.Contains(out, "Summary: 1 managed") {
		t.Fatalf("expected managed summary to be 1 after excluding drifted; got:\n%s", out)
	}

	if !strings.Contains(out, ", 1 drifted") {
		t.Fatalf("expected drifted summary to be present; got:\n%s", out)
	}
}
