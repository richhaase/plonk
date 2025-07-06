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
