package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportCommandExists(t *testing.T) {
	if importCmd == nil {
		t.Error("Import command should not be nil")
	}

	if importCmd.Use != "import" {
		t.Errorf("Expected command use to be 'import', got '%s'", importCmd.Use)
	}

	if importCmd.Short == "" {
		t.Error("Import command should have a short description")
	}

	if importCmd.RunE == nil {
		t.Error("Import command should have a RunE function")
	}
}

func TestImportCommandRegistered(t *testing.T) {
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Use == "import" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Import command should be registered with root command")
	}
}

func TestRunImportIntegration(t *testing.T) {
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test environment with packages and dotfiles
	// Create .zshrc file
	zshrcPath := filepath.Join(tempHome, ".zshrc")
	err := os.WriteFile(zshrcPath, []byte("# Test zshrc\nexport PATH=/usr/local/bin:$PATH"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .zshrc: %v", err)
	}

	// Create .gitconfig file
	gitconfigPath := filepath.Join(tempHome, ".gitconfig")
	err = os.WriteFile(gitconfigPath, []byte("[user]\n  name = Test User\n  email = test@example.com"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .gitconfig: %v", err)
	}

	// Create .tool-versions for ASDF
	toolVersionsPath := filepath.Join(tempHome, ".tool-versions")
	err = os.WriteFile(toolVersionsPath, []byte("nodejs 20.0.0\npython 3.11.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .tool-versions: %v", err)
	}

	// Run import command
	err = runImport(nil, []string{})
	if err != nil {
		t.Errorf("Import command should not fail: %v", err)
	}

	// Check that plonk.yaml was created
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	plonkYamlPath := filepath.Join(plonkDir, "plonk.yaml")
	if _, err := os.Stat(plonkYamlPath); os.IsNotExist(err) {
		t.Error("Expected plonk.yaml to be created")
	}

	// TODO: Add more detailed checks once we have mock command execution
}
