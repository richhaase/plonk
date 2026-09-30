package output

import (
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/richhaase/plonk/internal/testutil"
)

func TestPresentationColorStreams(t *testing.T) {
	originalWriter, originalProgress, originalNoColor := writer, progressWriter, color.NoColor
	t.Cleanup(func() { writer, progressWriter, color.NoColor = originalWriter, originalProgress, originalNoColor })
	t.Setenv("TERM", "xterm-256color")
	for _, tc := range []struct {
		name           string
		stdout, stderr bool
		noColor, term  string
	}{
		{"both terminals", true, true, "", "xterm-256color"},
		{"stdout redirected", false, true, "", "xterm-256color"},
		{"stderr redirected", true, false, "", "xterm-256color"},
		{"both redirected", false, false, "", "xterm-256color"},
		{"NO_COLOR", true, true, "1", "xterm-256color"},
		{"dumb terminal", true, true, "", "dumb"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tc.noColor)
			t.Setenv("TERM", tc.term)
			writer = testutil.NewBufferWriter(tc.stdout)
			progress := testutil.NewBufferWriter(tc.stderr)
			progressWriter = progress
			InitColors()
			var w strings.Builder
			WriteAction(&w, "missing", "/home/test/file with spaces", "not deployed", true)
			PrintAction("would-install", "brew:ripgrep", "")
			disabled := tc.noColor != "" || tc.term == "dumb"
			if strings.Contains(w.String(), "\x1b[") != (tc.stdout && !disabled) {
				t.Fatalf("unexpected stdout color: %q", w.String())
			}
			if strings.Contains(progress.String(), "\x1b[") != (tc.stderr && !disabled) {
				t.Fatalf("unexpected stderr color: %q", progress.String())
			}
			if !strings.Contains(w.String(), "  /home/test/file with spaces\n") {
				t.Fatalf("path was colored or split: %q", w.String())
			}
		})
	}
}

func TestPresentationStateColorsAndAlignment(t *testing.T) {
	original := color.NoColor
	t.Cleanup(func() { color.NoColor = original })
	color.NoColor = false
	for state, code := range map[string]string{"installed": "32", "missing": "33", "drifted": "33", "failed": "31", "would install": "34"} {
		if !strings.Contains(ColorState(state), "\x1b["+code+"m") {
			t.Errorf("wrong color for %s: %q", state, ColorState(state))
		}
	}
	var w strings.Builder
	WriteAction(&w, "missing", "brew:ripgrep", "", true)
	if !strings.Contains(w.String(), "\x1b[0m        brew:ripgrep") {
		t.Fatalf("ANSI codes affected alignment: %q", w.String())
	}
}

func TestApplyPresentationRetainsPartialAndDomainFailures(t *testing.T) {
	r := ApplyResult{Packages: &PackageResults{TotalInstalled: 1, TotalFailed: 1, Managers: []ManagerResults{{Name: "brew", Packages: []PackageOperation{{Name: "ripgrep", Status: "installed"}, {Name: "bad", Status: "failed", Error: "download failed"}}}}}}
	out := r.TableOutput()
	for _, want := range []string{"brew:ripgrep", "brew:bad", "download failed", "1 installed, 0 deployed, 1 failed", "Completed with errors."} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
	r = ApplyResult{}
	r.AddDotfileError(errors.New("cannot list source directory"))
	out = r.TableOutput()
	if !strings.Contains(out, "cannot list source directory") || strings.Contains(out, "Already up to date") {
		t.Fatalf("domain failure hidden: %s", out)
	}
}

func TestDoctorPresentationRetainsUnknownCategoriesAndRemediation(t *testing.T) {
	out := NewDoctorFormatter(DoctorOutput{Checks: []HealthCheck{{Category: "custom", Name: "Check", Status: "fail", Message: "broken", Details: []string{"detail"}, Issues: []string{"issue"}, Suggestions: []string{"repair"}}}}).TableOutput()
	for _, want := range []string{"Custom", "Check", "broken", "detail", "issue", "Next: repair"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "**") || strings.Contains(out, "##") {
		t.Fatalf("literal Markdown: %s", out)
	}
}

func TestInventoryErrorWithoutDetailKeepsItsErrorState(t *testing.T) {
	result := Result{Errors: []Item{{Name: "broken", Manager: "brew"}}}
	var w strings.Builder
	writePackagesTable(&w, result)
	if !strings.Contains(w.String(), "brew:broken") || strings.Contains(w.String(), "managed") {
		t.Fatalf("error mislabeled: %s", w.String())
	}
	if result.Errors[0].State != "" {
		t.Fatal("formatter mutated input")
	}
}
