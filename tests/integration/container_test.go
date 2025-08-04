//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnv provides a containerized test environment
type TestEnv struct {
	t         *testing.T
	container testcontainers.Container
	ctx       context.Context
}

// NewTestEnv creates a new containerized test environment
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	ctx := context.Background()

	// Check if we're in CI - skip for now (future: run directly)
	if os.Getenv("CI") == "true" {
		t.Skip("CI mode - container tests not yet implemented for CI")
	}

	// Set Docker host for testcontainers if not already set
	if os.Getenv("DOCKER_HOST") == "" {
		// Try common Docker socket locations
		sockets := []string{
			"unix:///var/run/docker.sock",                                  // Standard Docker
			"unix://" + os.Getenv("HOME") + "/.docker/run/docker.sock",     // Docker Desktop
			"unix://" + os.Getenv("HOME") + "/.colima/default/docker.sock", // Colima default
			"unix://" + os.Getenv("HOME") + "/.colima/docker.sock",         // Colima
		}

		for _, socket := range sockets {
			if _, err := os.Stat(strings.TrimPrefix(socket, "unix://")); err == nil {
				os.Setenv("DOCKER_HOST", socket)
				break
			}
		}
	}

	// Disable Ryuk for Colima compatibility
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	// Find plonk-linux binary - check both current dir and project root
	plonkBinary := "./plonk-linux"
	if _, err := os.Stat(plonkBinary); os.IsNotExist(err) {
		// Try project root (when running from tests/integration)
		plonkBinary = "../../plonk-linux"
		if _, err := os.Stat(plonkBinary); os.IsNotExist(err) {
			t.Fatal("Linux binary not found. Run: just build-linux")
		}
	}

	req := testcontainers.ContainerRequest{
		Image: "plonk-test:poc",
		Env: map[string]string{
			"NO_COLOR":              "1",
			"HOMEBREW_NO_ANALYTICS": "1",
		},
		Files: []testcontainers.ContainerFile{{
			HostFilePath:      plonkBinary,
			ContainerFilePath: "/usr/local/bin/plonk",
			FileMode:          0755,
		}},
		WaitingFor: wait.ForExec([]string{"test", "-f", "/tmp/ready.txt"}).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup
	t.Cleanup(func() {
		if t.Failed() {
			// Get container logs on failure
			logs, err := container.Logs(ctx)
			if err == nil {
				logBytes := make([]byte, 10000)
				n, _ := logs.Read(logBytes)
				if n > 0 {
					t.Logf("Container logs:\n%s", string(logBytes[:n]))
				}
			}
		}
		container.Terminate(ctx)
	})

	return &TestEnv{
		t:         t,
		container: container,
		ctx:       ctx,
	}
}

// Run executes plonk command in container
func (e *TestEnv) Run(args ...string) (string, error) {
	return e.Exec("plonk", args...)
}

// Exec runs any command in container
func (e *TestEnv) Exec(cmd string, args ...string) (string, error) {
	e.t.Helper()

	fullCmd := append([]string{cmd}, args...)
	exitCode, reader, err := e.container.Exec(e.ctx, fullCmd)
	if err != nil {
		return "", fmt.Errorf("exec failed: %w", err)
	}

	// Separate stdout and stderr using Docker's stdcopy
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	// Log stderr for debugging if test fails
	if e.t.Failed() && stderr.Len() > 0 {
		e.t.Logf("Command stderr: %s", stderr.String())
	}

	if exitCode != 0 {
		return stdout.String(), fmt.Errorf("command failed with exit code %d, stderr: %s", exitCode, stderr.String())
	}

	return stdout.String(), nil
}

// RunJSON runs plonk command and parses JSON output
func (e *TestEnv) RunJSON(v interface{}, args ...string) error {
	// Append -o json to args
	args = append(args, "-o", "json")

	output, err := e.Run(args...)
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, output)
	}

	if err := json.Unmarshal([]byte(output), v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}

	return nil
}

// WriteFile writes a file inside the container
func (e *TestEnv) WriteFile(path string, content []byte) error {
	e.t.Helper()

	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		_, err := e.Exec("mkdir", "-p", dir)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file using echo and shell redirection
	// Use base64 to handle special characters
	encoded := base64.StdEncoding.EncodeToString(content)
	_, err := e.Exec("sh", "-c", fmt.Sprintf("echo '%s' | base64 -d > %s", encoded, path))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func TestInstallPackage(t *testing.T) {
	env := NewTestEnv(t)

	// Test: Install a package with JSON output
	var installResult struct {
		Command    string `json:"command"`
		TotalItems int    `json:"total_items"`
		Results    []struct {
			Name    string `json:"name"`
			Manager string `json:"manager"`
			Status  string `json:"status"`
		} `json:"results"`
		Summary struct {
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
			Skipped   int `json:"skipped"`
		} `json:"summary"`
	}

	// Use 'hello' - a simple package that should work on any architecture
	err := env.RunJSON(&installResult, "install", "brew:hello")
	require.NoError(t, err, "Install should succeed")

	// Verify installation result
	assert.Equal(t, "install", installResult.Command)
	assert.Equal(t, 1, installResult.TotalItems)
	require.Len(t, installResult.Results, 1)
	assert.Equal(t, "hello", installResult.Results[0].Name)
	assert.Equal(t, "brew", installResult.Results[0].Manager)
	assert.Equal(t, "added", installResult.Results[0].Status)
	assert.Equal(t, 1, installResult.Summary.Succeeded)

	// Test: Verify status shows package
	var statusResult struct {
		Summary struct {
			TotalManaged int `json:"total_managed"`
			Domains      []struct {
				Domain       string `json:"domain"`
				ManagedCount int    `json:"managed_count"`
			} `json:"domains"`
		} `json:"summary"`
		ManagedItems []struct {
			Name    string `json:"name"`
			Domain  string `json:"domain"`
			Manager string `json:"manager"`
		} `json:"managed_items"`
	}

	err = env.RunJSON(&statusResult, "status")
	require.NoError(t, err, "Status should succeed")

	// Find hello in managed items
	found := false
	for _, item := range statusResult.ManagedItems {
		if item.Name == "hello" && item.Manager == "brew" && item.Domain == "package" {
			found = true
			break
		}
	}
	assert.True(t, found, "Hello should be in status output")
	assert.Equal(t, 1, statusResult.Summary.TotalManaged, "Should have 1 managed package")

	// Test: Verify brew actually installed it
	brewOut, err := env.Exec("brew", "list")
	require.NoError(t, err, "Brew list should work")
	assert.Contains(t, brewOut, "hello", "Hello should be in brew list")

	// Test: Verify the binary works
	helloOut, err := env.Exec("hello")
	require.NoError(t, err, "Hello command should work")
	assert.Contains(t, helloOut, "Hello, world!", "Hello should output greeting")
}
