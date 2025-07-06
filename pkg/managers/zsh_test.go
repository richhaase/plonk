package managers

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestZSHManager_IsAvailable_Success(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	successCmd := exec.Command("echo", "zsh 5.8 (x86_64-apple-darwin20.0)")
	mockExec.SetCommand("zsh --version", successCmd)

	zshMgr := NewZSHManager(mockExec)
	if !zshMgr.IsAvailable() {
		t.Error("Expected ZSH to be available")
	}
}

func TestZSHManager_IsAvailable_Failure(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	failCmd := exec.Command("false")
	mockExec.SetCommand("zsh --version", failCmd)

	zshMgr := NewZSHManager(mockExec)
	if zshMgr.IsAvailable() {
		t.Error("Expected ZSH to be unavailable")
	}
}

func TestZSHManager_Install_ClonesPlugin(t *testing.T) {
	tempDir := t.TempDir()

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	err := zshMgr.Install("zsh-users/zsh-syntax-highlighting")
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify git clone command was called
	expectedCall := "git clone"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected git clone command not executed. Got: %v", mockExec.Calls)
	}
}

func TestZSHManager_Install_SkipsExistingPlugin(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing plugin directory
	pluginDir := filepath.Join(tempDir, "zsh-syntax-highlighting")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create existing plugin dir: %v", err)
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	err = zshMgr.Install("zsh-users/zsh-syntax-highlighting")
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify no git clone command was executed
	if len(mockExec.Calls) > 0 {
		t.Errorf("Expected no commands to be executed, but got: %v", mockExec.Calls)
	}
}

func TestZSHManager_ListInstalled_ReturnsPlugins(t *testing.T) {
	tempDir := t.TempDir()

	// Create some plugin directories
	plugins := []string{"zsh-syntax-highlighting", "zsh-autosuggestions", "powerlevel10k"}
	for _, plugin := range plugins {
		err := os.MkdirAll(filepath.Join(tempDir, plugin), 0755)
		if err != nil {
			t.Fatalf("Failed to create plugin dir %s: %v", plugin, err)
		}
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	installed, err := zshMgr.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}

	if len(installed) != len(plugins) {
		t.Fatalf("Expected %d plugins, got %d", len(plugins), len(installed))
	}

	for _, expectedPlugin := range plugins {
		if !contains(installed, expectedPlugin) {
			t.Errorf("Expected plugin %s not found in: %v", expectedPlugin, installed)
		}
	}
}

func TestZSHManager_ListInstalled_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	installed, err := zshMgr.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}

	if len(installed) != 0 {
		t.Errorf("Expected no plugins, got %d: %v", len(installed), installed)
	}
}

func TestZSHManager_ListInstalled_NonExistentDirectory(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "nonexistent")

	// Set custom plugin directory to non-existent path
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	installed, err := zshMgr.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}

	if len(installed) != 0 {
		t.Errorf("Expected no plugins for non-existent directory, got %d: %v", len(installed), installed)
	}
}

func TestZSHManager_Update_SinglePlugin(t *testing.T) {
	tempDir := t.TempDir()

	// Create plugin directory
	pluginDir := filepath.Join(tempDir, "zsh-syntax-highlighting")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	err = zshMgr.Update("zsh-syntax-highlighting")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify git pull command was called
	expectedCall := "git -C"
	if len(mockExec.Calls) != 1 || mockExec.Calls[0] != expectedCall {
		t.Errorf("Expected git pull command not executed. Got: %v", mockExec.Calls)
	}
}

func TestZSHManager_Update_AllPlugins(t *testing.T) {
	tempDir := t.TempDir()

	// Create plugin directories
	plugins := []string{"zsh-syntax-highlighting", "zsh-autosuggestions"}
	for _, plugin := range plugins {
		err := os.MkdirAll(filepath.Join(tempDir, plugin), 0755)
		if err != nil {
			t.Fatalf("Failed to create plugin dir %s: %v", plugin, err)
		}
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	// Update all plugins (empty string means all)
	err := zshMgr.Update("")
	if err != nil {
		t.Fatalf("Update all failed: %v", err)
	}

	// Verify git pull commands were called for both plugins
	if len(mockExec.Calls) != len(plugins) {
		t.Errorf("Expected %d git pull commands, got %d: %v", len(plugins), len(mockExec.Calls), mockExec.Calls)
	}
}

func TestZSHManager_IsInstalled_True(t *testing.T) {
	tempDir := t.TempDir()

	// Create plugin directory
	pluginDir := filepath.Join(tempDir, "zsh-syntax-highlighting")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	if !zshMgr.IsInstalled("zsh-users/zsh-syntax-highlighting") {
		t.Error("Expected plugin to be installed")
	}
}

func TestZSHManager_IsInstalled_False(t *testing.T) {
	tempDir := t.TempDir()

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	if zshMgr.IsInstalled("zsh-users/zsh-syntax-highlighting") {
		t.Error("Expected plugin to not be installed")
	}
}

func TestZSHManager_Remove_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create plugin directory
	pluginDir := filepath.Join(tempDir, "zsh-syntax-highlighting")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	err = zshMgr.Remove("zsh-syntax-highlighting")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify plugin directory was removed
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("Expected plugin directory to be removed")
	}
}

func TestZSHManager_Remove_NonExistent(t *testing.T) {
	tempDir := t.TempDir()

	// Set custom plugin directory
	originalZPluginDir := os.Getenv("ZPLUGINDIR")
	defer os.Setenv("ZPLUGINDIR", originalZPluginDir)
	os.Setenv("ZPLUGINDIR", tempDir)

	mockExec := NewMockCommandExecutor()
	zshMgr := NewZSHManager(mockExec)

	err := zshMgr.Remove("non-existent-plugin")
	if err == nil {
		t.Error("Expected error when removing non-existent plugin")
	}
}

func TestGetPluginName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"zsh-users/zsh-syntax-highlighting", "zsh-syntax-highlighting"},
		{"romkatv/powerlevel10k", "powerlevel10k"},
		{"simple-plugin", "simple-plugin"},
		{"org/user/plugin", "plugin"},
	}

	for _, test := range tests {
		result := getPluginName(test.input)
		if result != test.expected {
			t.Errorf("getPluginName(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
