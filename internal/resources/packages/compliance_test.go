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
		// InstalledVersion JSON query
		"npm list -g typescript --depth=0 --json": {Output: []byte(`{"dependencies":{"typescript":{"version":"5.4.2"}}}`), Error: nil},
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

func TestCompliance_Pnpm_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"pnpm --version": {Output: []byte("9.0.0"), Error: nil},
		// List installed (array JSON with dependencies)
		"pnpm list -g --json": {Output: []byte(`[{"path":"/usr/local","private":false,"dependencies":{"typescript":{"version":"5.4.2"}}}]`), Error: nil},
		// Info view
		"pnpm view typescript --json": {Output: []byte(`{"name":"typescript","version":"5.4.2","description":"TS"}`), Error: nil},
		// Install/Uninstall/Upgrade
		"pnpm add -g typescript":    {Output: []byte(""), Error: nil},
		"pnpm remove -g typescript": {Output: []byte(""), Error: nil},
		"pnpm update -g typescript": {Output: []byte(""), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("pnpm", func() PackageManager { return NewPnpmManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("pnpm")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v", err)
	}
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, %v", err, list)
	}
	inst, err := mgr.IsInstalled(ctx, "typescript")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, %v", err, inst)
	}
	ver, err := mgr.InstalledVersion(ctx, "typescript")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion: %v, %q", err, ver)
	}
	info, err := mgr.Info(ctx, "typescript")
	if err != nil || info == nil {
		t.Fatalf("Info: %v, %v", err, info)
	}
	if err := mgr.Install(ctx, "typescript"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "typescript"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if err := mgr.Upgrade(ctx, []string{"typescript"}); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}

func TestCompliance_Pipx_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"pipx --version": {Output: []byte("1.4.3"), Error: nil},
		// List installed
		"pipx list --short": {Output: []byte("httpx 0.27.0\n"), Error: nil},
		// Install/Uninstall/Upgrade
		"pipx install httpx":   {Output: []byte(""), Error: nil},
		"pipx uninstall httpx": {Output: []byte(""), Error: nil},
		"pipx upgrade httpx":   {Output: []byte(""), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("pipx", func() PackageManager { return NewPipxManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("pipx")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v", err)
	}
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, %v", err, list)
	}
	inst, err := mgr.IsInstalled(ctx, "httpx")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, %v", err, inst)
	}
	if err := mgr.Install(ctx, "httpx"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "httpx"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if err := mgr.Upgrade(ctx, []string{"httpx"}); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}

func TestCompliance_Cargo_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"cargo --version": {Output: []byte("cargo 1.78"), Error: nil},
		// List installed
		"cargo install --list": {Output: []byte("ripgrep v14.0.3:\n    rg"), Error: nil},
		// Search/info
		"cargo search ripgrep --limit 1": {Output: []byte("ripgrep = \"14.0.3\"  # fast search"), Error: nil},
		// Install/Uninstall
		"cargo install ripgrep":   {Output: []byte(""), Error: nil},
		"cargo uninstall ripgrep": {Output: []byte(""), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("cargo", func() PackageManager { return NewCargoManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("cargo")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v", err)
	}
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, %v", err, list)
	}
	inst, err := mgr.IsInstalled(ctx, "ripgrep")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, %v", err, inst)
	}
	ver, err := mgr.InstalledVersion(ctx, "ripgrep")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion: %v, %q", err, ver)
	}
	info, err := mgr.Info(ctx, "ripgrep")
	if err != nil || info == nil {
		t.Fatalf("Info: %v, %v", err, info)
	}
	if err := mgr.Install(ctx, "ripgrep"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "ripgrep"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
}

func TestCompliance_Uv_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"uv --version": {Output: []byte("0.2.0"), Error: nil},
		// List installed
		"uv tool list": {Output: []byte("httpx v0.27.0"), Error: nil},
		// Install/Uninstall/Upgrade
		"uv tool install httpx":   {Output: []byte(""), Error: nil},
		"uv tool uninstall httpx": {Output: []byte(""), Error: nil},
		"uv tool upgrade httpx":   {Output: []byte(""), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("uv", func() PackageManager { return NewUvManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("uv")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v", err)
	}
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, %v", err, list)
	}
	inst, err := mgr.IsInstalled(ctx, "httpx")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, %v", err, inst)
	}
	ver, err := mgr.InstalledVersion(ctx, "httpx")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion: %v, %q", err, ver)
	}
	if err := mgr.Install(ctx, "httpx"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "httpx"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if err := mgr.Upgrade(ctx, []string{"httpx"}); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}

func TestCompliance_Conda_Minimal(t *testing.T) {
	responses := map[string]CommandResponse{
		// Availability
		"conda --version": {Output: []byte("conda 24.7.0"), Error: nil},
		// List installed (base env)
		"conda list -n base --json": {Output: []byte(`[{"name":"numpy","version":"1.26.0","build":"","channel":"conda-forge"}]`), Error: nil},
		// Search --json
		"conda search numpy --json": {Output: []byte(`{"numpy":[{"name":"numpy","version":"1.26.0","summary":"NumPy","home":"https://numpy.org","depends":["python >=3.9"]}]}`), Error: nil},
		// Info --json
		"conda info --json": {Output: []byte(`{"platform":"osx-64","conda_version":"24.7.0","base_environment":"/opt/conda"}`), Error: nil},
		// Install/Remove/Update
		"conda install -n base -y numpy": {Output: []byte(""), Error: nil},
		"conda remove -n base -y numpy":  {Output: []byte(""), Error: nil},
		"conda update -n base -y numpy":  {Output: []byte(""), Error: nil},
		// Info package --info path
		"conda search numpy --info --json": {Output: []byte(`{"numpy":[{"name":"numpy","version":"1.26.0","summary":"NumPy","home":"https://numpy.org","depends":["python >=3.9"]}]}`), Error: nil},
	}

	complianceEnv(t, func(r *ManagerRegistry) {
		r.Register("conda", func() PackageManager { return NewCondaManager() })
	}, responses)

	mgr, err := NewManagerRegistry().GetManager("conda")
	if err != nil {
		t.Fatalf("get manager: %v", err)
	}
	ctx := context.Background()

	avail, err := mgr.IsAvailable(ctx)
	if err != nil || !avail {
		t.Fatalf("IsAvailable: %v", err)
	}
	list, err := mgr.ListInstalled(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListInstalled: %v, %v", err, list)
	}
	inst, err := mgr.IsInstalled(ctx, "numpy")
	if err != nil || !inst {
		t.Fatalf("IsInstalled: %v, %v", err, inst)
	}
	ver, err := mgr.InstalledVersion(ctx, "numpy")
	if err != nil || ver == "" {
		t.Fatalf("InstalledVersion: %v, %q", err, ver)
	}
	info, err := mgr.Info(ctx, "numpy")
	if err != nil || info == nil {
		t.Fatalf("Info: %v, %v", err, info)
	}
	if err := mgr.Install(ctx, "numpy"); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := mgr.Uninstall(ctx, "numpy"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if err := mgr.Upgrade(ctx, []string{"numpy"}); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}
