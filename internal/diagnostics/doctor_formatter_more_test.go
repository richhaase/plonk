package diagnostics

import (
	"testing"

	out "github.com/richhaase/plonk/internal/output"
)

func TestDoctorFormatter_TableOutput_Categories(t *testing.T) {
	d := out.DoctorOutput{
		Overall: out.HealthStatus{Status: "healthy", Message: "OK"},
		Checks: []out.HealthCheck{
			{Name: "Go version", Category: "system", Status: "pass", Message: "ok"},
			{Name: "PLONK_DIR", Category: "environment", Status: "pass", Message: "ok"},
			{Name: "Config file", Category: "configuration", Status: "warn", Message: "missing"},
			{Name: "Homebrew", Category: "package-managers", Status: "fail", Message: "missing"},
		},
	}
	outStr := out.NewDoctorFormatter(d).TableOutput()
	if !contains(outStr, "Plonk Doctor Report") {
		t.Fatalf("missing header")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || (len(s) > 0 && (stringIndex(s, sub) >= 0)))
}
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
