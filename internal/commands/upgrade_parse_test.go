package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// buildLock constructs a lock with representative entries for parsing tests.
func buildLock(entries []lock.ResourceEntry) *lock.Lock {
	return &lock.Lock{Version: lock.LockFileVersion, Resources: entries}
}

func TestParseUpgradeArgs_InvalidManagerColon(t *testing.T) {
	l := buildLock(nil)
	_, err := parseUpgradeArgs([]string{"brew:"}, l)
	if err == nil || err.Error() == "" {
		t.Fatalf("expected error for invalid 'manager:' syntax, got: %v", err)
	}
}

func TestParseUpgradeArgs_UnknownPackage(t *testing.T) {
	l := buildLock(nil)
	_, err := parseUpgradeArgs([]string{"doesnotexist"}, l)
	if err == nil || err.Error() == "" {
		t.Fatalf("expected error for unknown package, got: %v", err)
	}
}

func TestParseUpgradeArgs_ManagerOnlyCollectsFromLock(t *testing.T) {
	entries := []lock.ResourceEntry{
		{Type: "package", ID: "brew:jq", Metadata: map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"}},
		{Type: "package", ID: "npm:typescript", Metadata: map[string]interface{}{"manager": "npm", "name": "typescript", "version": "5.4.2"}},
	}
	l := buildLock(entries)
	spec, err := parseUpgradeArgs([]string{"brew"}, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := spec.ManagerTargets["brew"]
	if len(got) != 1 || got[0] != "jq" {
		t.Fatalf("expected brew target [jq], got: %#v", got)
	}
}

func TestParseUpgradeArgs_ManagerPackageSpecific(t *testing.T) {
	entries := []lock.ResourceEntry{{Type: "package", ID: "brew:jq", Metadata: map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"}}}
	l := buildLock(entries)
	spec, err := parseUpgradeArgs([]string{"brew:jq"}, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	targets := spec.ManagerTargets["brew"]
	if len(targets) != 1 || targets[0] != "jq" {
		t.Fatalf("expected target jq, got: %#v", targets)
	}
}

func TestParseUpgradeArgs_BarePackageMatchesAllManagers(t *testing.T) {
	entries := []lock.ResourceEntry{
		{Type: "package", ID: "brew:ripgrep", Metadata: map[string]interface{}{"manager": "brew", "name": "ripgrep", "version": "14.1.1"}},
	}
	l := buildLock(entries)
	spec, err := parseUpgradeArgs([]string{"ripgrep"}, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := spec.ManagerTargets["brew"]; !ok {
		t.Fatalf("expected ripgrep to be targeted under brew")
	}
}

func TestParseUpgradeArgs_GoSourcePathTarget(t *testing.T) {
	// go packages store name (binary) and source_path; target should be source_path
	entries := []lock.ResourceEntry{
		{Type: "package", ID: "go:bar", Metadata: map[string]interface{}{"manager": "go", "name": "bar", "source_path": "github.com/foo/bar", "version": "0.1.0"}},
	}
	l := buildLock(entries)
	spec, err := parseUpgradeArgs([]string{"bar"}, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	targets := spec.ManagerTargets["go"]
	if len(targets) != 1 || targets[0] != "github.com/foo/bar" {
		t.Fatalf("expected go target to be source path, got: %#v", targets)
	}
}

func TestParseUpgradeArgs_NpmScopedUsesFullName(t *testing.T) {
	entries := []lock.ResourceEntry{
		{Type: "package", ID: "npm:typescript", Metadata: map[string]interface{}{"manager": "npm", "name": "typescript", "full_name": "@scope/typescript", "version": "5.4.2"}},
	}
	l := buildLock(entries)
	spec, err := parseUpgradeArgs([]string{"@scope/typescript"}, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	targets := spec.ManagerTargets["npm"]
	if len(targets) != 1 || targets[0] != "@scope/typescript" {
		t.Fatalf("expected npm target to be full name, got: %#v", targets)
	}
}
