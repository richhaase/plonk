package commands

import (
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

func TestRunImportBasic(t *testing.T) {
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	err := runImport(nil, []string{})
	if err != nil {
		t.Errorf("Import command should not fail: %v", err)
	}

	// For now, just test it doesn't crash
	// We'll add more comprehensive tests when we implement the functionality
	_ = tempHome
}
