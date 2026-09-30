package output

import (
	"strings"
	"testing"
)

func TestDoctorFormatter_TableOutput(t *testing.T) {
	data := DoctorOutput{
		Overall: HealthStatus{Status: "warning", Message: "Some checks have warnings"},
		Checks: []HealthCheck{
			{Name: "System", Category: "system", Status: "pass", Message: "OK"},
			{Name: "Homebrew", Category: "package-managers", Status: "warn", Message: "Not found", Suggestions: []string{"Install brew"}},
		},
	}
	f := NewDoctorFormatter(data)
	out := f.TableOutput()
	wants := []string{"warning", "System", "Homebrew", "Next: Install brew"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
}
