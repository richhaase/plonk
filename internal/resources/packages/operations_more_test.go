package packages

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// fake manager variants for targeted branches

func TestInstall_ManagerUnavailable_Suggestion(t *testing.T) {
	configDir := t.TempDir()
	// v2-only: no responses for npm → unavailable
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := InstallPackages(context.Background(), configDir, []string{"prettier"}, InstallOptions{Manager: "npm"})
	if err != nil {
		t.Fatalf("InstallPackages error: %v", err)
	}
	if len(res) != 1 || res[0].Status != "failed" || res[0].Error == nil {
		t.Fatalf("unexpected result: %+v", res)
	}
	if !strings.Contains(res[0].Error.Error(), "install Node.js from") {
		t.Fatalf("expected suggestion in error, got: %v", res[0].Error)
	}
}

func TestInstall_NpmScoped_MetadataSaved(t *testing.T) {
	configDir := t.TempDir()
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"npm --version":                    {Output: []byte("10.0.0"), Error: nil},
		"npm install -g @scope/typescript": {Output: []byte("added"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	pkg := "@scope/typescript"
	res, err := InstallPackages(context.Background(), configDir, []string{pkg}, InstallOptions{Manager: "npm"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "added" {
		t.Fatalf("expected added, got: %+v", res[0])
	}

	// Verify lock metadata includes scope and full_name
	svc := lock.NewYAMLLockService(configDir)
	lk, _ := svc.Read()
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	md := lk.Resources[0].Metadata
	if md["scope"] != "@scope" || md["full_name"] != pkg {
		t.Fatalf("expected npm metadata saved, got: %#v", md)
	}
}

func TestInstall_GoSourcePath_SavedAndBinaryNamedInLock(t *testing.T) {
	configDir := t.TempDir()
	// Define custom v2 manager for go in config file and mock executor
	// Note: Install command follows typical pattern for Go tools
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"go --version":                         {Output: []byte("go1.22"), Error: nil},
		"go install github.com/foo/bar@latest": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	// Write a minimal plonk.yaml to inject the go manager for this test
	cfgPath := filepath.Join(configDir, "plonk.yaml")
	_ = os.WriteFile(cfgPath, []byte("managers:\n  go:\n    binary: go\n    install:\n      command: [\"go\", \"install\", \"{{.Package}}@latest\"]\n"), 0o644)

	src := "github.com/foo/bar"
	res, err := InstallPackages(context.Background(), configDir, []string{src}, InstallOptions{Manager: "go"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "added" {
		t.Fatalf("expected added, got: %+v", res[0])
	}

	svc := lock.NewYAMLLockService(configDir)
	lk, _ := svc.Read()
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	if lk.Resources[0].ID != "go:bar" {
		t.Fatalf("expected ID go:bar, got %s", lk.Resources[0].ID)
	}
	if lk.Resources[0].Metadata["source_path"] != src {
		t.Fatalf("missing source_path")
	}
}

func TestInstall_LockWriteFailure(t *testing.T) {
	configDir := t.TempDir()
	// Make directory read-only to trigger writer failure
	_ = os.Chmod(configDir, 0500)
	t.Cleanup(func() { _ = os.Chmod(configDir, 0700) })
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "failed" || res[0].Error == nil {
		t.Fatalf("expected failed due to lock write failure, got: %+v", res[0])
	}
}

func TestUninstall_PartialSuccess_WhenSystemUninstallFails(t *testing.T) {
	configDir := t.TempDir()
	// Seed lock as managed item
	svc := lock.NewYAMLLockService(configDir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq"})
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("fail"), Error: &MockExitError{Code: 1}},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with error detail, got: %+v", res)
	}
	if !strings.Contains(res[0].Error.Error(), "system uninstall failed") {
		t.Fatalf("expected system uninstall failed detail, got: %v", res[0].Error)
	}
}

func TestUninstall_NotManaged_PassThrough(t *testing.T) {
	configDir := t.TempDir()
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if res[0].Status != "removed" || res[0].Error != nil {
		t.Fatalf("expected removed without error, got: %+v", res[0])
	}
}

func TestUninstall_LockWriteFailure_Error(t *testing.T) {
	configDir := t.TempDir()
	// Seed lock as managed item
	svc := lock.NewYAMLLockService(configDir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq"})

	// Make directory read-only so RemovePackage's write fails
	_ = os.Chmod(configDir, 0500)
	t.Cleanup(func() { _ = os.Chmod(configDir, 0700) })
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with lock update error, got: %+v", res[0])
	}
	if !strings.Contains(res[0].Error.Error(), "failed to update lock") {
		t.Fatalf("expected failed to update lock detail, got: %v", res[0].Error)
	}
}
