// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// MockCommandExecutor for testing
type MockCommandExecutor struct {
	Commands map[string]func() *exec.Cmd
	Calls    []string
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Commands: make(map[string]func() *exec.Cmd),
		Calls:    make([]string, 0),
	}
}

func (m *MockCommandExecutor) Execute(name string, args ...string) *exec.Cmd {
	key := name
	if len(args) > 0 {
		key += " " + strings.Join(args, " ")
	}
	m.Calls = append(m.Calls, key)

	if cmdFunc, exists := m.Commands[key]; exists {
		return cmdFunc()
	}

	// Return a command that will fail by default (more realistic for IsInstalled checks)
	return exec.Command("false")
}

func (m *MockCommandExecutor) SetCommand(key string, cmd *exec.Cmd) {
	m.Commands[key] = func() *exec.Cmd { return cmd }
}

func (m *MockCommandExecutor) SetCommandFunc(key string, cmdFunc func() *exec.Cmd) {
	m.Commands[key] = cmdFunc
}

func TestHomebrewManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  *exec.Cmd
		expected bool
	}{
		{
			name:     "success when command succeeds",
			command:  exec.Command("echo", "Homebrew 4.0.0"),
			expected: true,
		},
		{
			name:     "failure when command fails",
			command:  exec.Command("false"), // 'false' command always exits with code 1
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("brew --version", tt.command)

			manager := NewHomebrewManager(mockExec)
			available := manager.IsAvailable()

			if available != tt.expected {
				t.Errorf("IsAvailable() = %v, expected %v", available, tt.expected)
			}

			// Verify the right command was called
			expectedCall := "brew --version"
			if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
				t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
			}
		})
	}
}

func TestHomebrewManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the install command to succeed
	installCmd := exec.Command("echo", "installed")
	mockExec.SetCommand("brew install git", installCmd)

	manager := NewHomebrewManager(mockExec)
	err := manager.Install("git")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the right command was called
	expectedCall := "brew install git"
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

	// Mock the upgrade command to succeed
	upgradeCmd := exec.Command("echo", "upgraded")
	mockExec.SetCommand("brew upgrade git", upgradeCmd)

	manager := NewHomebrewManager(mockExec)
	err := manager.Update("git")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the right command was called
	expectedCall := "brew upgrade git"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestHomebrewManager_Update_AllPackages(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the upgrade command to succeed
	upgradeCmd := exec.Command("echo", "upgraded")
	mockExec.SetCommand("brew upgrade", upgradeCmd)

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

func TestHomebrewManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		command     *exec.Cmd
		expected    bool
	}{
		{
			name:        "returns true when package is installed",
			packageName: "git",
			command:     exec.Command("echo", "git installed"),
			expected:    true,
		},
		{
			name:        "returns false when package is not installed",
			packageName: "nonexistent",
			command:     exec.Command("false"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("brew list "+tt.packageName, tt.command)

			manager := NewHomebrewManager(mockExec)
			installed := manager.IsInstalled(tt.packageName)

			if installed != tt.expected {
				t.Errorf("IsInstalled(%s) = %v, expected %v", tt.packageName, installed, tt.expected)
			}

			// Verify the right command was called
			expectedCall := "brew list " + tt.packageName
			if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
				t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
			}
		})
	}
}

// ASDF Manager Tests
func TestAsdfManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  *exec.Cmd
		expected bool
	}{
		{
			name:     "success when command succeeds",
			command:  exec.Command("echo", "v0.11.3"),
			expected: true,
		},
		{
			name:     "failure when command fails",
			command:  exec.Command("false"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("asdf version", tt.command)

			manager := NewAsdfManager(mockExec)
			available := manager.IsAvailable()

			if available != tt.expected {
				t.Errorf("IsAvailable() = %v, expected %v", available, tt.expected)
			}

			expectedCall := "asdf version"
			if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
				t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
			}
		})
	}
}

func TestAsdfManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the install command to succeed
	installCmd := exec.Command("echo", "installed")
	mockExec.SetCommand("asdf install nodejs 20.0.0", installCmd)

	manager := NewAsdfManager(mockExec)
	err := manager.Install("nodejs 20.0.0")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedCall := "asdf install nodejs 20.0.0"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestAsdfManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	listCmd := exec.Command("echo", "nodejs\npython\ngolang")
	mockExec.SetCommand("asdf plugin list", listCmd)

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
	mockExec.SetCommand("asdf latest nodejs", latestCmd)
	// Mock the install command
	installCmd := exec.Command("echo", "installed")
	mockExec.SetCommand("asdf install nodejs 20.11.0", installCmd)

	manager := NewAsdfManager(mockExec)
	err := manager.Update("nodejs")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should call both "asdf latest nodejs" and "asdf install nodejs 20.11.0"
	if len(mockExec.Calls) != 2 {
		t.Errorf("Expected 2 calls, got %d: %v", len(mockExec.Calls), mockExec.Calls)
	}

	expectedCalls := []string{"asdf latest nodejs", "asdf install nodejs 20.11.0"}
	for i, expectedCall := range expectedCalls {
		if i >= len(mockExec.Calls) || mockExec.Calls[i] != expectedCall {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expectedCall, mockExec.Calls[i])
		}
	}
}

func TestAsdfManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		command  *exec.Cmd
		expected bool
	}{
		{
			name:     "returns true when tool has versions installed",
			toolName: "nodejs",
			command:  exec.Command("echo", "  18.0.0\n* 20.0.0"),
			expected: true,
		},
		{
			name:     "returns false when tool is not installed",
			toolName: "nonexistent",
			command:  exec.Command("false"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("asdf list "+tt.toolName, tt.command)

			manager := NewAsdfManager(mockExec)
			installed := manager.IsInstalled(tt.toolName)

			if installed != tt.expected {
				t.Errorf("IsInstalled(%s) = %v, expected %v", tt.toolName, installed, tt.expected)
			}

			expectedCall := "asdf list " + tt.toolName
			if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
				t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
			}
		})
	}
}

func TestAsdfManager_GetInstalledVersions_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// asdf list <tool> returns versions
	listCmd := exec.Command("echo", "  18.0.0\n* 20.0.0\n  21.0.0")
	mockExec.SetCommand("asdf list nodejs", listCmd)

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
func TestNpmManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  *exec.Cmd
		expected bool
	}{
		{
			name:     "success when command succeeds",
			command:  exec.Command("echo", "8.19.2"),
			expected: true,
		},
		{
			name:     "failure when command fails",
			command:  exec.Command("false"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("npm --version", tt.command)

			manager := NewNpmManager(mockExec)
			available := manager.IsAvailable()

			if available != tt.expected {
				t.Errorf("IsAvailable() = %v, expected %v", available, tt.expected)
			}

			expectedCall := "npm --version"
			if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
				t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
			}
		})
	}
}

func TestNpmManager_Install_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the install command to succeed
	installCmd := exec.Command("echo", "installed")
	mockExec.SetCommand("npm install -g typescript", installCmd)

	manager := NewNpmManager(mockExec)
	err := manager.Install("typescript")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedCall := "npm install -g typescript"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_Update_SinglePackage(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the update command to succeed
	updateCmd := exec.Command("echo", "updated")
	mockExec.SetCommand("npm update -g typescript", updateCmd)

	manager := NewNpmManager(mockExec)
	err := manager.Update("typescript")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedCall := "npm update -g typescript"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_UpdateAll_CallsCorrectCommand(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Mock the update all command to succeed
	updateAllCmd := exec.Command("echo", "updated all")
	mockExec.SetCommand("npm update -g", updateAllCmd)

	manager := NewNpmManager(mockExec)
	err := manager.UpdateAll()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedCall := "npm update -g"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected call to '%s', got %v", expectedCall, mockExec.Calls)
	}
}

func TestNpmManager_ListInstalled_ParsesOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	// npm list -g --depth=0 --parseable returns paths
	listCmd := exec.Command("echo", "/usr/local/lib/node_modules/npm\n/usr/local/lib/node_modules/typescript\n/usr/local/lib/node_modules/@vue/cli")
	mockExec.SetCommand("npm list -g --depth=0 --parseable", listCmd)

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

func TestNpmManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		command     *exec.Cmd
		expected    bool
	}{
		{
			name:        "returns true when package is installed",
			packageName: "typescript",
			command:     exec.Command("echo", "typescript@4.8.4"),
			expected:    true,
		},
		{
			name:        "returns false when package is not installed",
			packageName: "nonexistent",
			command:     exec.Command("false"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockCommandExecutor()
			mockExec.SetCommand("npm list -g --depth=0 "+tt.packageName, tt.command)

			manager := NewNpmManager(mockExec)
			installed := manager.IsInstalled(tt.packageName)

			if installed != tt.expected {
				t.Errorf("IsInstalled(%s) = %v, expected %v", tt.packageName, installed, tt.expected)
			}
		})
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

func TestAsdfManager_ListGlobalTools(t *testing.T) {
	tests := []struct {
		name          string
		toolVersions  string
		expectedTools []string
		expectError   bool
	}{
		{
			name: "successful parsing of .tool-versions",
			toolVersions: `nodejs 20.0.0
python 3.11.3
ruby 3.0.0
`,
			expectedTools: []string{
				"nodejs 20.0.0",
				"python 3.11.3",
				"ruby 3.0.0",
			},
			expectError: false,
		},
		{
			name:          "empty .tool-versions file",
			toolVersions:  "",
			expectedTools: []string{},
			expectError:   false,
		},
		{
			name: "single tool",
			toolVersions: `nodejs 18.16.0
`,
			expectedTools: []string{"nodejs 18.16.0"},
			expectError:   false,
		},
		{
			name: "tools with comments and empty lines",
			toolVersions: `# My tools
nodejs 20.0.0

python 3.11.3
# Another comment
ruby 3.0.0
`,
			expectedTools: []string{
				"nodejs 20.0.0",
				"python 3.11.3",
				"ruby 3.0.0",
			},
			expectError: false,
		},
		{
			name:          "no .tool-versions file - returns empty list",
			toolVersions:  "",
			expectedTools: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary home directory
			tempHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			// Create .tool-versions file if not testing error case
			if !tt.expectError {
				toolVersionsPath := filepath.Join(tempHome, ".tool-versions")
				err := os.WriteFile(toolVersionsPath, []byte(tt.toolVersions), 0644)
				if err != nil {
					t.Fatalf("Failed to create test .tool-versions file: %v", err)
				}
			}

			mockExec := NewMockCommandExecutor()
			manager := NewAsdfManager(mockExec)

			tools, err := manager.ListGlobalTools()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(tools) != len(tt.expectedTools) {
				t.Errorf("Expected %d tools, got %d", len(tt.expectedTools), len(tools))
				t.Errorf("Expected: %v", tt.expectedTools)
				t.Errorf("Got: %v", tools)
			}

			for i, tool := range tools {
				if i < len(tt.expectedTools) && tool != tt.expectedTools[i] {
					t.Errorf("Expected tool %s, got %s", tt.expectedTools[i], tool)
				}
			}
		})
	}
}

// Test PackageStatus string representation
func TestPackageStatusString(t *testing.T) {
	tests := []struct {
		status   PackageStatus
		expected string
	}{
		{PackageInstalled, "installed"},
		{PackageAvailable, "available"},
		{PackageUnknown, "unknown"},
		{PackageStatus(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.status.String(); got != test.expected {
			t.Errorf("PackageStatus(%d).String() = %s, want %s", test.status, got, test.expected)
		}
	}
}

// Test PackageInfo structure
func TestPackageInfo(t *testing.T) {
	info := PackageInfo{
		Name:        "git",
		Version:     "2.42.0",
		Status:      PackageInstalled,
		Manager:     "homebrew",
		Description: "Distributed version control system",
	}

	if info.Name != "git" {
		t.Errorf("Expected name to be 'git', got %s", info.Name)
	}

	if info.Status != PackageInstalled {
		t.Errorf("Expected status to be PackageInstalled, got %v", info.Status)
	}

	if info.Status.String() != "installed" {
		t.Errorf("Expected status string to be 'installed', got %s", info.Status.String())
	}
}

// Test HomebrewManager state-aware methods
func TestHomebrewManager_StateAwareMethods(t *testing.T) {
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".config", "plonk")

	// Create config directory and plonk.yaml
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}

	configContent := `settings:
  default_manager: homebrew
homebrew:
  brews:
    - name: git
    - name: curl
  casks:
    - name: docker
`
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	mockExec := NewMockCommandExecutor()

	// Mock specific IsInstalled calls - git is installed, curl and docker are not
	mockExec.SetCommandFunc("brew list git", func() *exec.Cmd {
		return exec.Command("echo", "git installed")
	})

	mockExec.SetCommandFunc("brew list curl", func() *exec.Cmd {
		return exec.Command("false")
	})

	mockExec.SetCommandFunc("brew list docker", func() *exec.Cmd {
		return exec.Command("false")
	})

	// Mock version info for git
	gitVersionCmd := exec.Command("echo", "git 2.42.0")
	mockExec.SetCommand("brew list --versions git", gitVersionCmd)

	manager := NewHomebrewManager(mockExec)
	manager.SetConfigDir(plonkDir)

	// Test ListManagedPackages
	managed, err := manager.ListManagedPackages()
	if err != nil {
		t.Errorf("ListManagedPackages failed: %v", err)
	}

	expectedManaged := 3 // git, curl, docker
	if len(managed) != expectedManaged {
		t.Errorf("Expected %d managed packages, got %d", expectedManaged, len(managed))
	}

	// Verify packages have correct manager and status
	gitFound := false
	curlFound := false
	dockerFound := false

	for _, pkg := range managed {
		if pkg.Manager != "homebrew" {
			t.Errorf("Expected manager to be 'homebrew', got %s", pkg.Manager)
		}

		switch pkg.Name {
		case "git":
			gitFound = true
			if pkg.Status != PackageInstalled {
				t.Errorf("Expected git to be installed, got status %v", pkg.Status)
			}
			if pkg.Version != "2.42.0" {
				t.Errorf("Expected git version '2.42.0', got %s", pkg.Version)
			}
		case "curl":
			curlFound = true
			if pkg.Status != PackageAvailable {
				t.Errorf("Expected curl to be available, got status %v", pkg.Status)
			}
		case "docker":
			dockerFound = true
			if pkg.Status != PackageAvailable {
				t.Errorf("Expected docker to be available, got status %v", pkg.Status)
			}
		}
	}

	if !gitFound || !curlFound || !dockerFound {
		t.Error("Expected to find git, curl, and docker in managed packages")
	}

	// Test ListMissingPackages - should be curl and docker (not installed)
	missing, err := manager.ListMissingPackages()
	if err != nil {
		t.Errorf("ListMissingPackages failed: %v", err)
	}

	expectedMissing := 2 // curl and docker are missing
	if len(missing) != expectedMissing {
		t.Errorf("Expected %d missing packages, got %d", expectedMissing, len(missing))
	}

	// Verify missing packages
	for _, pkg := range missing {
		if pkg.Name != "curl" && pkg.Name != "docker" {
			t.Errorf("Unexpected missing package: %s", pkg.Name)
		}
		if pkg.Status != PackageAvailable {
			t.Errorf("Expected missing package %s to have status PackageAvailable, got %v", pkg.Name, pkg.Status)
		}
	}
}

// Test AsdfManager state-aware methods
func TestAsdfManager_StateAwareMethods(t *testing.T) {
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".config", "plonk")

	// Create config directory and plonk.yaml
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}

	configContent := `settings:
  default_manager: asdf
asdf:
  - name: nodejs
    version: 20.0.0
  - name: python
    version: 3.11.3
`
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	mockExec := NewMockCommandExecutor()

	// Mock asdf plugin list to return nodejs and ruby
	pluginCmd := exec.Command("echo", "nodejs\nruby")
	mockExec.SetCommand("asdf plugin", pluginCmd)

	// Mock asdf list nodejs to return versions
	nodejsVersionsCmd := exec.Command("echo", "  18.0.0\n* 20.0.0")
	mockExec.SetCommand("asdf list", nodejsVersionsCmd)

	manager := NewAsdfManager(mockExec)
	manager.SetConfigDir(plonkDir)

	// Test ListManagedPackages
	managed, err := manager.ListManagedPackages()
	if err != nil {
		t.Errorf("ListManagedPackages failed: %v", err)
	}

	expectedManaged := 2 // nodejs 20.0.0, python 3.11.3
	if len(managed) != expectedManaged {
		t.Errorf("Expected %d managed packages, got %d", expectedManaged, len(managed))
	}

	// Verify nodejs 20.0.0 is found
	foundNodejs := false
	for _, pkg := range managed {
		if pkg.Name == "nodejs 20.0.0" {
			foundNodejs = true
			if pkg.Version != "20.0.0" {
				t.Errorf("Expected version to be '20.0.0', got %s", pkg.Version)
			}
			if pkg.Manager != "asdf" {
				t.Errorf("Expected manager to be 'asdf', got %s", pkg.Manager)
			}
		}
	}
	if !foundNodejs {
		t.Error("Expected to find nodejs 20.0.0 in managed packages")
	}
}

// Test NpmManager state-aware methods
func TestNpmManager_StateAwareMethods(t *testing.T) {
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".config", "plonk")

	// Create config directory and plonk.yaml
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}

	configContent := `settings:
  default_manager: npm
npm:
  - name: typescript
  - name: vue-cli
    package: "@vue/cli"
`
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	mockExec := NewMockCommandExecutor()

	// Mock specific IsInstalled calls - typescript is installed, @vue/cli is not
	mockExec.SetCommandFunc("npm list -g --depth=0 typescript", func() *exec.Cmd {
		return exec.Command("echo", "typescript@4.8.4")
	})

	mockExec.SetCommandFunc("npm list -g --depth=0 @vue/cli", func() *exec.Cmd {
		return exec.Command("false")
	})

	manager := NewNpmManager(mockExec)
	manager.SetConfigDir(plonkDir)

	// Test ListManagedPackages
	managed, err := manager.ListManagedPackages()
	if err != nil {
		t.Errorf("ListManagedPackages failed: %v", err)
	}

	expectedManaged := 2 // typescript, @vue/cli
	if len(managed) != expectedManaged {
		t.Errorf("Expected %d managed packages, got %d", expectedManaged, len(managed))
	}

	// Verify packages are found with correct manager
	foundTypescript := false
	foundVueCli := false
	for _, pkg := range managed {
		if pkg.Name == "typescript" {
			foundTypescript = true
			if pkg.Manager != "npm" {
				t.Errorf("Expected manager to be 'npm', got %s", pkg.Manager)
			}
			if pkg.Status != PackageInstalled {
				t.Errorf("Expected status to be PackageInstalled, got %v", pkg.Status)
			}
		}
		if pkg.Name == "@vue/cli" {
			foundVueCli = true
			if pkg.Manager != "npm" {
				t.Errorf("Expected manager to be 'npm', got %s", pkg.Manager)
			}
			if pkg.Status != PackageAvailable {
				t.Errorf("Expected status to be PackageAvailable, got %v", pkg.Status)
			}
		}
	}
	if !foundTypescript {
		t.Error("Expected to find typescript in managed packages")
	}
	if !foundVueCli {
		t.Error("Expected to find @vue/cli in managed packages")
	}

	// Test ListMissingPackages - should be only @vue/cli since typescript is installed
	missing, err := manager.ListMissingPackages()
	if err != nil {
		t.Errorf("ListMissingPackages failed: %v", err)
	}

	expectedMissing := 1 // @vue/cli (typescript is installed)
	if len(missing) != expectedMissing {
		t.Errorf("Expected %d missing packages, got %d", expectedMissing, len(missing))
	}

	// Verify @vue/cli is the missing package
	if len(missing) > 0 && missing[0].Name != "@vue/cli" {
		t.Errorf("Expected missing package to be '@vue/cli', got %s", missing[0].Name)
	}
}

// Test SetConfigDir method
func TestPackageManager_SetConfigDir(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	// Test HomebrewManager
	homebrewMgr := NewHomebrewManager(mockExec)
	homebrewMgr.SetConfigDir("/test/path")
	if homebrewMgr.plonkDir != "/test/path" {
		t.Errorf("Expected plonkDir to be '/test/path', got %s", homebrewMgr.plonkDir)
	}

	// Test AsdfManager
	asdfMgr := NewAsdfManager(mockExec)
	asdfMgr.SetConfigDir("/test/path")
	if asdfMgr.plonkDir != "/test/path" {
		t.Errorf("Expected plonkDir to be '/test/path', got %s", asdfMgr.plonkDir)
	}

	// Test NpmManager
	npmMgr := NewNpmManager(mockExec)
	npmMgr.SetConfigDir("/test/path")
	if npmMgr.plonkDir != "/test/path" {
		t.Errorf("Expected plonkDir to be '/test/path', got %s", npmMgr.plonkDir)
	}
}

// Test state-aware methods with no config
func TestPackageManager_StateAwareMethodsNoConfig(t *testing.T) {
	mockExec := NewMockCommandExecutor()

	manager := NewHomebrewManager(mockExec)
	// Don't set config dir - should handle gracefully

	managed, err := manager.ListManagedPackages()
	if err != nil {
		t.Errorf("ListManagedPackages should not fail without config: %v", err)
	}
	if len(managed) != 0 {
		t.Errorf("Expected 0 managed packages without config, got %d", len(managed))
	}

	missing, err := manager.ListMissingPackages()
	if err != nil {
		t.Errorf("ListMissingPackages should not fail without config: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("Expected 0 missing packages without config, got %d", len(missing))
	}
}
