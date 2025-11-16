package clone

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestSetupFromClonedRepo_NoManagers_NoApply(t *testing.T) {
	dir := t.TempDir()
	// Create minimal plonk.yaml so hasConfig=true
	if err := createDefaultConfig(dir); err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// No lock file => no detected managers
	if err := SetupFromClonedRepo(context.Background(), dir, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupFromClonedRepo_InstallsDetectedManagers(t *testing.T) {
	dir := t.TempDir()
	if err := createDefaultConfig(dir); err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// Seed lock with two managers
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	_ = svc.AddPackage("npm", "typescript", "1.0.0", map[string]interface{}{"manager": "npm", "name": "typescript", "version": "1.0.0"})

	// v2: use default managers and mock executor
	// Use defaults; mock with empty responses so managers appear unavailable
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	// Run setup without apply - should now fail since self-install is not supported
	err := SetupFromClonedRepo(context.Background(), dir, true, true)
	if err == nil {
		t.Fatal("expected error for missing package managers, got nil")
	}

	// Verify error message mentions automatic installation is not supported
	expectedMsg := "automatic installation of package managers is not supported"
	if err.Error() == "" || len(err.Error()) < len(expectedMsg) {
		t.Fatalf("expected error message to mention automatic installation, got: %v", err)
	}

	// Sanity: lock still present
	if _, err := svc.Read(); err != nil {
		t.Fatalf("failed reading lock: %v", err)
	}
	_ = filepath.Join // silence import linters
}

func TestInstallDetectedManagers_LoadsRepoConfig(t *testing.T) {
	dir := t.TempDir()
	configContent := `
managers:
  custom:
    binary: custom-pm
    install:
      command: ["custom-pm", "install", "{{.Package}}"]
    upgrade:
      command: ["custom-pm", "upgrade", "{{.Package}}"]
    upgrade_all:
      command: ["custom-pm", "upgrade-all"]
    uninstall:
      command: ["custom-pm", "remove", "{{.Package}}"]
`
	if err := os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte(strings.TrimSpace(configContent)), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	err := installDetectedManagers(context.Background(), dir, []string{"custom"}, Config{})
	if err == nil {
		t.Fatal("expected error due to missing custom manager binary")
	}
	if !strings.Contains(err.Error(), "automatic installation of package managers is not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
