// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// DockerRunner helps run commands in Docker containers
type DockerRunner struct {
	ImageName string
	WorkDir   string
}

// NewDockerRunner creates a new DockerRunner
func NewDockerRunner() *DockerRunner {
	return &DockerRunner{
		ImageName: dockerImageName,
		WorkDir:   "/workspace",
	}
}

// RunCommand executes a command in a fresh Docker container
func (dr *DockerRunner) RunCommand(t *testing.T, command string) ([]byte, error) {
	t.Helper()
	
	// Get the current working directory and go up to project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// From test/integration, go up two levels to project root
	projectRoot := filepath.Join(wd, "../..")
	
	cmd := exec.Command("docker", "run", "--rm",
		"-v", projectRoot+":/workspace",
		"-w", dr.WorkDir,
		dr.ImageName,
		"/bin/bash", "-c", command)
	
	return cmd.CombinedOutput()
}

// BuildPlonkBinary builds the plonk binary inside the container
func (dr *DockerRunner) BuildPlonkBinary(t *testing.T) error {
	t.Helper()
	
	output, err := dr.RunCommand(t, "go build -buildvcs=false -o plonk ./cmd/plonk")
	if err != nil {
		t.Logf("Build output: %s", output)
		return err
	}
	
	t.Logf("Build successful: %s", output)
	return nil
}

// RequireDockerImage ensures the Docker image is available
func RequireDockerImage(t *testing.T) {
	t.Helper()
	
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}
	
	// Check if image exists
	cmd := exec.Command("docker", "image", "inspect", dockerImageName)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker image %s not found. Run integration test setup first.", dockerImageName)
	}
}

// CreateTempConfigFile creates a temporary config file with the given content
func CreateTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	
	tempFile := filepath.Join(t.TempDir(), "plonk.yaml")
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	
	return tempFile
}

// CleanupBuildArtifacts removes build artifacts after tests
func CleanupBuildArtifacts(t *testing.T) {
	t.Helper()
	
	t.Cleanup(func() {
		if _, err := os.Stat("plonk"); err == nil {
			os.Remove("plonk")
		}
	})
}