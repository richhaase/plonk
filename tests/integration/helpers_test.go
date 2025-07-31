//go:build integration
// +build integration

package integration_test

import (
	"os"
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

	// Check if we should build with coverage
	buildArgs := []string{"build", "-modcacherw"}
	if os.Getenv("INTEGRATION_COVERAGE") == "true" {
		// Build with coverage instrumentation for Go 1.20+
		buildArgs = append(buildArgs, "-cover")
	}
	buildArgs = append(buildArgs, "-o", plonkBinary, "../../cmd/plonk")

	// Build with explicit module cache to avoid polluting test temp dir
	// This prevents "permission denied" errors during cleanup
	cmd := exec.Command("go", buildArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build plonk: %v\nOutput: %s", err, output)
	}

	require.FileExists(t, plonkBinary)
	return plonkBinary
}
