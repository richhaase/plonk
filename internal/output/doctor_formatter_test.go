package output

import (
	"strings"
	"testing"
)

func TestDoctorFormatter_TableAndStructured(t *testing.T) {
	data := DoctorOutput{
		Overall: HealthStatus{Status: "warning", Message: "Some checks have warnings"},
		Checks: []HealthCheck{
			{Name: "System", Category: "system", Status: "pass", Message: "OK"},
			{Name: "Homebrew", Category: "package-managers", Status: "warn", Message: "Not found", Suggestions: []string{"Install brew"}},
		},
	}
	f := NewDoctorFormatter(data)
	out := f.TableOutput()
	wants := []string{"Overall Status: WARNING", "System", "Homebrew", "Suggestions"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in:\n%s", w, out)
		}
	}
	if f.StructuredData().(DoctorOutput).Overall.Status != "warning" {
		t.Fatalf("structured mismatch")
	}
}
