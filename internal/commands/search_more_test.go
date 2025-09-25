package commands

import (
	"context"
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// Manager unavailable for all managers → no-managers status
func TestCLI_Search_NoManagers_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "-o", "json", "foo"}, func(env CLITestEnv) {
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {}) // empty registry
	})
	if err != nil {
		t.Fatalf("search no-managers error: %v\n%s", err, out)
	}
	if !contains(out, `"status": "no-managers"`) {
		t.Fatalf("expected no-managers, got: %s", out)
	}
}

// Specific manager returns not-found
func TestCLI_Search_ManagerNotFound_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "-o", "json", "brew:foo"}, func(env CLITestEnv) {
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("brew", func() packages.PackageManager { return &fakeSearchNoResults{} })
		})
	})
	if err != nil {
		t.Fatalf("search manager not-found error: %v\n%s", err, out)
	}
	if !contains(out, `"status": "not-found"`) {
		t.Fatalf("expected not-found, got: %s", out)
	}
}

type fakeSearchNoResults struct{}

func (f *fakeSearchNoResults) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeSearchNoResults) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeSearchNoResults) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeSearchNoResults) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeSearchNoResults) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (f *fakeSearchNoResults) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeSearchNoResults) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "brew"}, nil
}
func (f *fakeSearchNoResults) Search(ctx context.Context, q string) ([]string, error) {
	return []string{}, nil
}
func (f *fakeSearchNoResults) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "brew"}, nil
}
func (f *fakeSearchNoResults) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeSearchNoResults) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeSearchNoResults) Dependencies() []string                           { return nil }

// local helpers
func contains(s, sub string) bool { return strings.Contains(s, sub) }
