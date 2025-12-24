package diagnostics

import (
	"testing"

	packages "github.com/richhaase/plonk/internal/packages"
)

func TestPackageManagerEcosystem_Check(t *testing.T) {
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version": {Output: []byte("Homebrew 4.0"), Error: nil},
		"npm --version":  {Output: []byte("10.0.0"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })
	rep := RunHealthChecks()
	found := false
	for _, c := range rep.Checks {
		if c.Name == "Package Managers" && c.Category == "package-managers" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected package managers check present")
	}
}
