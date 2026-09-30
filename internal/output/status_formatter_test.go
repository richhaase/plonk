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
		ShowAll: true,
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

func TestStatusFormatter_ActionableDefault(t *testing.T) {
	data := StatusOutput{
		RemoteSync: "up to date",
		HomeDir:    "/home/test",
		StateSummary: Summary{TotalManaged: 3, TotalMissing: 2, TotalErrors: 2, Results: []Result{
			{Domain: "package", Managed: []Item{{Name: "healthy-package", Manager: "brew", State: StateManaged}}, Missing: []Item{{Name: "missing-package", Manager: "brew", State: StateMissing}}, Errors: []Item{{Name: "broken-package", Error: "manager unavailable"}}},
			{Domain: "dotfile", Managed: []Item{{Name: ".healthy", State: StateManaged}, {Name: ".drifted", State: StateDegraded, Metadata: map[string]interface{}{"destination": "/home/test/.drifted", "source_type": "template"}}}, Missing: []Item{{Name: ".missing", State: StateMissing}}, Errors: []Item{{Name: ".broken", Error: "cannot render"}}},
		}},
	}
	for _, all := range []bool{false, true} {
		data.ShowAll = all
		out := NewStatusFormatter(data).TableOutput()
		for _, want := range []string{"Remote: up to date", "missing-package", "~/.drifted", "template", ".missing", "broken-package", "manager unavailable", ".broken", "cannot render"} {
			if !strings.Contains(out, want) {
				t.Errorf("all=%v: missing %q in output:\n%s", all, want, out)
			}
		}
		for _, healthy := range []string{"healthy-package", ".healthy", "Summary:"} {
			if strings.Contains(out, healthy) != all {
				t.Errorf("all=%v: unexpected visibility of %q:\n%s", all, healthy, out)
			}
		}
	}
	if data.StateSummary.Results[0].Managed[0].Name != "healthy-package" || len(data.StateSummary.Results[1].Managed) != 2 {
		t.Fatal("formatter changed input")
	}
}

func TestStatusFormatter_HealthyDefault(t *testing.T) {
	out := NewStatusFormatter(StatusOutput{RemoteSync: "up to date", StateSummary: Summary{TotalManaged: 1, Results: []Result{{Domain: "package", Managed: []Item{{Name: "healthy", State: StateManaged}}}}}}).TableOutput()
	if strings.Contains(out, "healthy") || strings.Contains(out, "Summary:") || strings.Contains(out, "PACKAGE") {
		t.Fatalf("unexpected healthy details:\n%s", out)
	}
	if !strings.Contains(out, "Remote: up to date") {
		t.Fatalf("missing remote status:\n%s", out)
	}
}

func TestStatusFormatter_ErrorsOnly(t *testing.T) {
	out := NewStatusFormatter(StatusOutput{StateSummary: Summary{TotalErrors: 1, Results: []Result{{Domain: "package", Errors: []Item{{Name: "broken", Error: "unsupported manager"}}}}}}).TableOutput()
	if !strings.Contains(out, "broken") || !strings.Contains(out, "unsupported manager") || strings.Contains(out, "No managed items") {
		t.Fatalf("expected errors:\n%s", out)
	}
}
