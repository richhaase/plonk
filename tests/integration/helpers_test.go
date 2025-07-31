//go:build integration
// +build integration

package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// buildPlonk builds the plonk binary for testing
func buildPlonk(t *testing.T) string {
	t.Helper()

	// Build the binary to a temporary location
	tempDir := t.TempDir()
	plonkBinary := filepath.Join(tempDir, "plonk")

	cmd := exec.Command("go", "build", "-o", plonkBinary, "../../cmd/plonk")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build plonk: %v\nOutput: %s", err, output)
	}

	require.FileExists(t, plonkBinary)
	return plonkBinary
}
