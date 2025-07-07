// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os"
	"os/exec"
	"path/filepath"
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
			mockExec.SetCommand("brew list", tt.command)

			manager := NewHomebrewManager(mockExec)
			installed := manager.IsInstalled(tt.packageName)

			if installed != tt.expected {
				t.Errorf("IsInstalled(%s) = %v, expected %v", tt.packageName, installed, tt.expected)
			}

			// Verify the right command was called
			expectedCall := "brew list"
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
			mockExec.SetCommand("asdf list", tt.command)

			manager := NewAsdfManager(mockExec)
			installed := manager.IsInstalled(tt.toolName)

			if installed != tt.expected {
				t.Errorf("IsInstalled(%s) = %v, expected %v", tt.toolName, installed, tt.expected)
			}

			expectedCall := "asdf list"
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
			mockExec.SetCommand("npm list", tt.command)

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
