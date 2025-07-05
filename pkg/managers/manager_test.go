package managers

import (
	"os/exec"
	"testing"
)

// MockCommandExecutor for testing
type MockCommandExecutor struct {
	Commands map[string]*exec.Cmd
	Calls    []string
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Commands: make(map[string]*exec.Cmd),
		Calls:    make([]string, 0),
	}
}

func (m *MockCommandExecutor) Execute(name string, args ...string) *exec.Cmd {
	key := name
	if len(args) > 0 {
		key += " " + args[0]
	}
	m.Calls = append(m.Calls, key)
	
	if cmd, exists := m.Commands[key]; exists {
		return cmd
	}
	
	// Return a command that will succeed by default
	return exec.Command("echo", "mock success")
}

func (m *MockCommandExecutor) SetCommand(key string, cmd *exec.Cmd) {
	m.Commands[key] = cmd
}

// Test that we can check if Homebrew is available
func TestHomebrewManager_IsAvailable_Success(t *testing.T) {
	// Create a mock executor that returns success for "brew --version"
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "Homebrew 4.0.0")
	mockExec.SetCommand("brew --version", successCmd)
	
	manager := NewHomebrewManager(mockExec)
	
	available := manager.IsAvailable()
	
	if !available {
		t.Error("Expected Homebrew to be available when command succeeds")
	}
	
	// Verify the right command was called
	expectedCall := "brew --version"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Set up a command that will fail
	failCmd := exec.Command("false") // 'false' command always exits with code 1
	mockExec.SetCommand("brew --version", failCmd)
	
	manager := NewHomebrewManager(mockExec)
	
	available := manager.IsAvailable()
	
	if available {
		t.Error("Expected Homebrew to be unavailable when command fails")
	}
}

func TestHomebrewManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewHomebrewManager(mockExec)
	err := manager.Install("git")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the right command was called
	expectedCall := "brew install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Set up a command that returns a list of packages
	listCmd := exec.Command("echo", "git\ncurl\njq")
	mockExec.SetCommand("brew list", listCmd)
	
	manager := NewHomebrewManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []string{"git", "curl", "jq"}
	if len(packages) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packages))
	}
	
	for i, pkg := range expected {
		if i >= len(packages) || packages[i] != pkg {
			t.Errorf("Expected package '%s' at index %d, got '%s'", pkg, i, packages[i])
		}
	}
}

func TestHomebrewManager_Install_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Set up a command that will fail
	failCmd := exec.Command("false")
	mockExec.SetCommand("brew install", failCmd)
	
	manager := NewHomebrewManager(mockExec)
	err := manager.Install("nonexistent-package")
	
	if err == nil {
		t.Error("Expected error when install fails, got nil")
	}
}

func TestHomebrewManager_ListInstalled_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Set up a command that will fail
	failCmd := exec.Command("false")
	mockExec.SetCommand("brew list", failCmd)
	
	manager := NewHomebrewManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err == nil {
		t.Error("Expected error when list fails, got nil")
	}
	
	if packages != nil {
		t.Error("Expected nil packages when error occurs")
	}
}

func TestHomebrewManager_Update_SinglePackage(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewHomebrewManager(mockExec)
	err := manager.Update("git")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the right command was called
	expectedCall := "brew upgrade"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_Update_AllPackages(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewHomebrewManager(mockExec)
	err := manager.UpdateAll()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the right command was called
	expectedCall := "brew upgrade"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_IsInstalled_True(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// brew list <package> succeeds when package is installed
	successCmd := exec.Command("echo", "git installed")
	mockExec.SetCommand("brew list", successCmd)
	
	manager := NewHomebrewManager(mockExec)
	installed := manager.IsInstalled("git")
	
	if !installed {
		t.Error("Expected package to be installed when command succeeds")
	}
	
	// Verify the right command was called
	expectedCall := "brew list"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_IsInstalled_False(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// brew list <package> fails when package is not installed
	failCmd := exec.Command("false")
	mockExec.SetCommand("brew list", failCmd)
	
	manager := NewHomebrewManager(mockExec)
	installed := manager.IsInstalled("nonexistent")
	
	if installed {
		t.Error("Expected package to not be installed when command fails")
	}
}

// ASDF Manager Tests
func TestAsdfManager_IsAvailable_Success(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "v0.11.3")
	mockExec.SetCommand("asdf version", successCmd)
	
	manager := NewAsdfManager(mockExec)
	available := manager.IsAvailable()
	
	if !available {
		t.Error("Expected ASDF to be available when command succeeds")
	}
	
	expectedCall := "asdf version"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestAsdfManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("asdf version", failCmd)
	
	manager := NewAsdfManager(mockExec)
	available := manager.IsAvailable()
	
	if available {
		t.Error("Expected ASDF to be unavailable when command fails")
	}
}

func TestAsdfManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewAsdfManager(mockExec)
	err := manager.Install("nodejs 20.0.0")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "asdf install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestAsdfManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	listCmd := exec.Command("echo", "nodejs\npython\ngolang")
	mockExec.SetCommand("asdf plugin", listCmd)
	
	manager := NewAsdfManager(mockExec)
	plugins, err := manager.ListInstalled()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []string{"nodejs", "python", "golang"}
	if len(plugins) != len(expected) {
		t.Errorf("Expected %d plugins, got %d", len(expected), len(plugins))
	}
	
	for i, plugin := range expected {
		if i >= len(plugins) || plugins[i] != plugin {
			t.Errorf("Expected plugin '%s' at index %d, got '%s'", plugin, i, plugins[i])
		}
	}
}

func TestAsdfManager_Update_WithToolAndVersion(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Mock the "asdf latest" command to return a version
	latestCmd := exec.Command("echo", "20.11.0")
	mockExec.SetCommand("asdf latest", latestCmd)
	// Mock the install command
	installCmd := exec.Command("echo", "installed")
	mockExec.SetCommand("asdf install", installCmd)
	
	manager := NewAsdfManager(mockExec)
	err := manager.Update("nodejs")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should call both "asdf latest nodejs" and "asdf install nodejs 20.11.0"
	if len(mockExec.Calls) != 2 {
		t.Errorf("Expected 2 calls, got %d: %v", len(mockExec.Calls), mockExec.Calls)
	}
	
	expectedCalls := []string{"asdf latest", "asdf install"}
	for i, expectedCall := range expectedCalls {
		if i >= len(mockExec.Calls) || mockExec.Calls[i] != expectedCall {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expectedCall, mockExec.Calls[i])
		}
	}
}

func TestAsdfManager_IsInstalled_True(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// asdf list <tool> succeeds when tool has versions installed
	successCmd := exec.Command("echo", "  18.0.0\n* 20.0.0")
	mockExec.SetCommand("asdf list", successCmd)
	
	manager := NewAsdfManager(mockExec)
	installed := manager.IsInstalled("nodejs")
	
	if !installed {
		t.Error("Expected tool to be installed when command succeeds")
	}
	
	expectedCall := "asdf list"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestAsdfManager_IsInstalled_False(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// asdf list <tool> fails when tool is not installed
	failCmd := exec.Command("false")
	mockExec.SetCommand("asdf list", failCmd)
	
	manager := NewAsdfManager(mockExec)
	installed := manager.IsInstalled("nonexistent")
	
	if installed {
		t.Error("Expected tool to not be installed when command fails")
	}
}

func TestAsdfManager_GetInstalledVersions_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// asdf list <tool> returns versions
	listCmd := exec.Command("echo", "  18.0.0\n* 20.0.0\n  21.0.0")
	mockExec.SetCommand("asdf list", listCmd)
	
	manager := NewAsdfManager(mockExec)
	versions, err := manager.GetInstalledVersions("nodejs")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []string{"18.0.0", "20.0.0", "21.0.0"}
	if len(versions) != len(expected) {
		t.Errorf("Expected %d versions, got %d", len(expected), len(versions))
	}
	
	for i, version := range expected {
		if i >= len(versions) || versions[i] != version {
			t.Errorf("Expected version '%s' at index %d, got '%s'", version, i, versions[i])
		}
	}
}

func TestAsdfManager_Update_ReturnsErrorOnLatestFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Mock the "asdf latest" command to fail
	failCmd := exec.Command("false")
	mockExec.SetCommand("asdf latest", failCmd)
	
	manager := NewAsdfManager(mockExec)
	err := manager.Update("nonexistent-tool")
	
	if err == nil {
		t.Error("Expected error when latest command fails")
	}
}

func TestAsdfManager_GetInstalledVersions_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// asdf list <tool> fails
	failCmd := exec.Command("false")
	mockExec.SetCommand("asdf list", failCmd)
	
	manager := NewAsdfManager(mockExec)
	versions, err := manager.GetInstalledVersions("nonexistent")
	
	if err == nil {
		t.Error("Expected error when list command fails")
	}
	
	if versions != nil {
		t.Error("Expected nil versions when error occurs")
	}
}

// NPM Manager Tests
func TestNpmManager_IsAvailable_Success(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "8.19.2")
	mockExec.SetCommand("npm --version", successCmd)
	
	manager := NewNpmManager(mockExec)
	available := manager.IsAvailable()
	
	if !available {
		t.Error("Expected NPM to be available when command succeeds")
	}
	
	expectedCall := "npm --version"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("npm --version", failCmd)
	
	manager := NewNpmManager(mockExec)
	available := manager.IsAvailable()
	
	if available {
		t.Error("Expected NPM to be unavailable when command fails")
	}
}

func TestNpmManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewNpmManager(mockExec)
	err := manager.Install("typescript")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "npm install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_Update_SinglePackage(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewNpmManager(mockExec)
	err := manager.Update("typescript")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "npm update"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_UpdateAll_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewNpmManager(mockExec)
	err := manager.UpdateAll()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "npm update"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// npm list -g --depth=0 --parseable returns paths
	listCmd := exec.Command("echo", "/usr/local/lib/node_modules/npm\n/usr/local/lib/node_modules/typescript\n/usr/local/lib/node_modules/@vue/cli")
	mockExec.SetCommand("npm list", listCmd)
	
	manager := NewNpmManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should exclude npm itself and parse package names from paths
	expected := []string{"typescript", "@vue/cli"}
	if len(packages) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packages))
	}
	
	for i, pkg := range expected {
		if i >= len(packages) || packages[i] != pkg {
			t.Errorf("Expected package '%s' at index %d, got '%s'", pkg, i, packages[i])
		}
	}
}

func TestNpmManager_IsInstalled_True(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// npm list -g typescript succeeds when package is installed
	successCmd := exec.Command("echo", "typescript@4.8.4")
	mockExec.SetCommand("npm list", successCmd)
	
	manager := NewNpmManager(mockExec)
	installed := manager.IsInstalled("typescript")
	
	if !installed {
		t.Error("Expected package to be installed when command succeeds")
	}
}

func TestNpmManager_IsInstalled_False(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// npm list -g nonexistent fails when package is not installed
	failCmd := exec.Command("false")
	mockExec.SetCommand("npm list", failCmd)
	
	manager := NewNpmManager(mockExec)
	installed := manager.IsInstalled("nonexistent")
	
	if installed {
		t.Error("Expected package to not be installed when command fails")
	}
}

func TestNpmManager_ListInstalled_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("npm list", failCmd)
	
	manager := NewNpmManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err == nil {
		t.Error("Expected error when list command fails")
	}
	
	if packages != nil {
		t.Error("Expected nil packages when error occurs")
	}
}

func TestNpmManager_Update_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("npm update", failCmd)
	
	manager := NewNpmManager(mockExec)
	err := manager.Update("nonexistent-package")
	
	if err == nil {
		t.Error("Expected error when update fails")
	}
}

// Pip Manager Tests
func TestPipManager_IsAvailable_Success(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "pip 22.3.1")
	mockExec.SetCommand("pip --version", successCmd)
	
	manager := NewPipManager(mockExec)
	available := manager.IsAvailable()
	
	if !available {
		t.Error("Expected Pip to be available when command succeeds")
	}
	
	expectedCall := "pip --version"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestPipManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewPipManager(mockExec)
	err := manager.Install("requests")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "pip install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestPipManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("pip --version", failCmd)
	
	manager := NewPipManager(mockExec)
	available := manager.IsAvailable()
	
	if available {
		t.Error("Expected Pip to be unavailable when command fails")
	}
}

func TestPipManager_Update_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewPipManager(mockExec)
	err := manager.Update("requests")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "pip install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestPipManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// pip list --format=freeze returns package==version format
	listCmd := exec.Command("echo", "requests==2.28.1\nnumpy==1.24.0\npandas==1.5.2")
	mockExec.SetCommand("pip list", listCmd)
	
	manager := NewPipManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []string{"requests", "numpy", "pandas"}
	if len(packages) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packages))
	}
	
	for i, pkg := range expected {
		if i >= len(packages) || packages[i] != pkg {
			t.Errorf("Expected package '%s' at index %d, got '%s'", pkg, i, packages[i])
		}
	}
}

func TestPipManager_IsInstalled_True(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// pip show <package> succeeds when package is installed
	successCmd := exec.Command("echo", "Name: requests")
	mockExec.SetCommand("pip show", successCmd)
	
	manager := NewPipManager(mockExec)
	installed := manager.IsInstalled("requests")
	
	if !installed {
		t.Error("Expected package to be installed when command succeeds")
	}
}

func TestPipManager_IsInstalled_False(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// pip show <package> fails when package is not installed
	failCmd := exec.Command("false")
	mockExec.SetCommand("pip show", failCmd)
	
	manager := NewPipManager(mockExec)
	installed := manager.IsInstalled("nonexistent")
	
	if installed {
		t.Error("Expected package to not be installed when command fails")
	}
}

func TestPipManager_UpdateAll_CallsCorrectCommands(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Mock outdated packages list
	outdatedCmd := exec.Command("echo", "requests==2.28.0\nnumpy==1.23.0")
	mockExec.SetCommand("pip list", outdatedCmd)
	
	manager := NewPipManager(mockExec)
	err := manager.UpdateAll()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should call pip list --outdated and then pip install --upgrade for each package
	if len(mockExec.Calls) < 3 {
		t.Errorf("Expected at least 3 calls (list + 2 updates), got %d: %v", len(mockExec.Calls), mockExec.Calls)
	}
	
	// First call should be the list command
	if mockExec.Calls[0] != "pip list" {
		t.Errorf("Expected first call to be 'pip list', got '%s'", mockExec.Calls[0])
	}
}

func TestPipManager_ListInstalled_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("pip list", failCmd)
	
	manager := NewPipManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err == nil {
		t.Error("Expected error when list command fails")
	}
	
	if packages != nil {
		t.Error("Expected nil packages when error occurs")
	}
}

// Cargo Manager Tests
func TestCargoManager_IsAvailable_Success(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "cargo 1.70.0")
	mockExec.SetCommand("cargo --version", successCmd)
	
	manager := NewCargoManager(mockExec)
	available := manager.IsAvailable()
	
	if !available {
		t.Error("Expected Cargo to be available when command succeeds")
	}
	
	expectedCall := "cargo --version"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestCargoManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewCargoManager(mockExec)
	err := manager.Install("ripgrep")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "cargo install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestCargoManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("cargo --version", failCmd)
	
	manager := NewCargoManager(mockExec)
	available := manager.IsAvailable()
	
	if available {
		t.Error("Expected Cargo to be unavailable when command fails")
	}
}

func TestCargoManager_Update_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	
	manager := NewCargoManager(mockExec)
	err := manager.Update("ripgrep")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedCall := "cargo install"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestCargoManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// cargo install --list returns package names with versions and binaries
	listCmd := exec.Command("echo", "ripgrep v13.0.0:\n    rg\nbat v0.22.1:\n    bat\nexa v0.10.1:\n    exa")
	mockExec.SetCommand("cargo install", listCmd)
	
	manager := NewCargoManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []string{"ripgrep", "bat", "exa"}
	if len(packages) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packages))
	}
	
	for i, pkg := range expected {
		if i >= len(packages) || packages[i] != pkg {
			t.Errorf("Expected package '%s' at index %d, got '%s'", pkg, i, packages[i])
		}
	}
}

func TestCargoManager_IsInstalled_True(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// cargo install --list shows the package
	listCmd := exec.Command("echo", "ripgrep v13.0.0:\n    rg")
	mockExec.SetCommand("cargo install", listCmd)
	
	manager := NewCargoManager(mockExec)
	installed := manager.IsInstalled("ripgrep")
	
	if !installed {
		t.Error("Expected package to be installed when found in list")
	}
}

func TestCargoManager_IsInstalled_False(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// cargo install --list doesn't show the package
	listCmd := exec.Command("echo", "ripgrep v13.0.0:\n    rg")
	mockExec.SetCommand("cargo install", listCmd)
	
	manager := NewCargoManager(mockExec)
	installed := manager.IsInstalled("nonexistent")
	
	if installed {
		t.Error("Expected package to not be installed when not found in list")
	}
}

func TestCargoManager_UpdateAll_WithEmptyList(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// Mock empty list of installed packages
	listCmd := exec.Command("echo", "")
	mockExec.SetCommand("cargo install", listCmd)
	
	manager := NewCargoManager(mockExec)
	err := manager.UpdateAll()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should call cargo install --list once
	if len(mockExec.Calls) != 1 {
		t.Errorf("Expected exactly 1 call, got %d: %v", len(mockExec.Calls), mockExec.Calls)
	}
}

func TestCargoManager_ListInstalled_ReturnsErrorOnFailure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("cargo install", failCmd)
	
	manager := NewCargoManager(mockExec)
	packages, err := manager.ListInstalled()
	
	if err == nil {
		t.Error("Expected error when list command fails")
	}
	
	if packages != nil {
		t.Error("Expected nil packages when error occurs")
	}
}