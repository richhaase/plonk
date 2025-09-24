package packages

import (
	"context"
	"testing"
)

// complianceEnv sets up a mock executor and temporary registry for manager under test.
func complianceEnv(t *testing.T, register func(*ManagerRegistry), responses map[string]CommandResponse) *MockCommandExecutor {
	t.Helper()

	// Use temporary registry to include only specified managers
	WithTemporaryRegistry(t, register)

	// Install mock executor
	mock := &MockCommandExecutor{Responses: responses}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	return mock
}

func TestCompliance_Brew_Minimal(t *testing.T) {
	// Prepare responses for Homebrew manager flows
	responses := map[string]CommandResponse{
		// Availability
		"brew --version": {Output: []byte("Homebrew 4.0.0"), Error: nil},
		// Installed packages info (JSON v2)
		"brew info --installed --json=v2": {Output: []byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), Error: nil},
		// list --versions fallback
		"brew list --versions jq": {Output: []byte("jq 1.6"), Error: nil},
		// Info output
		"brew info jq": {Output: []byte("jq: stable 1.6\nFrom: https://github.com/Homebrew/homebrew-core\n"), Error: nil},
		// Search
		"brew search jq": {Output: []byte("jq\njq-extra\n"), Error: nil},
		// Install/Uninstall
		"brew install jq":   {Output: []byte(""), Error: nil},
		"brew uninstall jq": {Output: []byte(""), Error: nil},
		// Upgrade
		"brew upgrade jq": {Output: []byte(""), Error: nil},
	}

	mock := complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return NewHomebrewManager() })
	}, responses)
	_ = mock

	mgr, err := NewManagerRegistry().GetManager("brew")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}

	ctx := context.Background()

	// IsAvailable
	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v, avail=%v", err, avail)
	}

	// ListInstalled (should include jq via info JSON)
	list, err := mgr.ListInstalled(ctx)
	if err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("ListInstalled returned empty list")
	}

	// IsInstalled
	inst, err := mgr.IsInstalled(ctx, "jq")
	if err != nil || !inst {
		t.Fatalf("IsInstalled jq: %v, inst=%v", err, inst)
	}

	// InstalledVersion
	ver, err := mgr.InstalledVersion(ctx, "jq")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion jq: %v, ver=%q", err, ver)
	}

	// Info
	info, err := mgr.Info(ctx, "jq")
	if err != nil || info == nil || info.Name != "jq" {
		t.Fatalf("Info jq: %v, info=%+v", err, info)
	}

	// Search
	res, err := mgr.Search(ctx, "jq")
	if err != nil || len(res) == 0 {
		t.Fatalf("Search jq: %v, res=%v", err, res)
	}

	// Install/Uninstall
	if err := mgr.Install(ctx, "jq"); err != nil {
		t.Fatalf("Install jq: %v", err)
	}
	if err := mgr.Uninstall(ctx, "jq"); err != nil {
		t.Fatalf("Uninstall jq: %v", err)
	}

	// Upgrade
	if err := mgr.Upgrade(ctx, []string{"jq"}); err != nil {
		t.Fatalf("Upgrade jq: %v", err)
	}
}

func TestCompliance_Npm_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"npm --version": {Output: []byte("10.0.0"), Error: nil},
		// List installed (JSON)
		"npm list -g --depth=0 --json": {Output: []byte(`{"dependencies":{"typescript":{"version":"5.4.2"}}}`), Error: nil},
		// IsInstalled checks
		"npm list -g typescript": {Output: []byte("/usr/local/lib\n└── typescript@5.4.2"), Error: nil},
		// Info
		"npm view typescript --json": {Output: []byte(`{"name":"typescript","version":"5.4.2"}`), Error: nil},
		// Search
		"npm search typescript --json": {Output: []byte(`[{"name":"typescript"},{"name":"ts-node"}]`), Error: nil},
		// Install/Uninstall
		"npm install -g typescript":   {Output: []byte(""), Error: nil},
		"npm uninstall -g typescript": {Output: []byte(""), Error: nil},
		// Upgrade specific
		"npm update -g typescript": {Output: []byte(""), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("npm", func() PackageManager { return NewNpmManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("npm")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	// IsAvailable
	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v, avail=%v", err, avail)
	}

	// ListInstalled
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, list=%v", err, list)
	}

	// IsInstalled
	inst, err := mgr.IsInstalled(ctx, "typescript")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, inst=%v", err, inst)
	}

	// InstalledVersion
	ver, err := mgr.InstalledVersion(ctx, "typescript")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion: %v, ver=%q", err, ver)
	}

	// Info
	info, err := mgr.Info(ctx, "typescript")
	if err != nil || info == nil || info.Name != "typescript" {
		t.Fatalf("Info: %v, info=%+v", err, info)
	}

	// Search
	srch, err := mgr.Search(ctx, "typescript")
	if err != nil || len(srch) == 0 {
		t.Fatalf("Search: %v, res=%v", err, srch)
	}

	// Install/Uninstall
	if err := mgr.Install(ctx, "typescript"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "typescript"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	// Upgrade
	if err := mgr.Upgrade(ctx, []string{"typescript"}); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}
