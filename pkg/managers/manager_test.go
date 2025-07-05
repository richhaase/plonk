package managers

import (
	"os/exec"
	"testing"
)

// CommandExecutor interface for dependency injection
type CommandExecutor interface {
	Execute(name string, args ...string) *exec.Cmd
}

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