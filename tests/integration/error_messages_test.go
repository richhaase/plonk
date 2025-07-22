//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerErrorMessages(t *testing.T) {
	// Setup test environment
	testDir := t.TempDir()
	os.Setenv("PLONK_DIR", testDir)
	defer os.Unsetenv("PLONK_DIR")

	// Build plonk
	mustRun(t, "go", "build", "-o", "plonk", "../../cmd/plonk")

	t.Run("ValidManagerInstallSuggestions", func(t *testing.T) {
		// Test that each manager provides helpful install suggestions
		managers := []string{"homebrew", "npm", "cargo", "gem", "go", "pip"}

		for _, manager := range managers {
			// Skip if manager is already available
			output := run(t, "./plonk", "doctor", "-o", "json")
			if strings.Contains(output, manager+": ✅") {
				continue
			}

			// Create config with this manager as default
			configContent := `version: 1
default_manager: ` + manager
			os.WriteFile(filepath.Join(testDir, "plonk.yaml"), []byte(configContent), 0644)

			// Try to use the manager
			output, _ = runWithError("./plonk", "install", "test-package", "--force")

			// Should provide installation instructions
			lowerOutput := strings.ToLower(output)
			if !strings.Contains(lowerOutput, "install") {
				t.Errorf("Error message for %s should contain installation instructions", manager)
			}
		}
	})
}
