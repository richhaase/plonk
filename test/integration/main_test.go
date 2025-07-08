// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"testing"
)

const (
	dockerImageName = "plonk-test"
	dockerfilePath  = "test/integration/docker/Dockerfile"
)

// TestMain sets up the integration test environment
func TestMain(m *testing.M) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		os.Exit(0) // Skip integration tests if Docker not available
	}

	// Build Docker image for integration tests
	if err := buildDockerImage(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup (optional - Docker images can be left for reuse)
	// cleanupDockerImage()

	os.Exit(code)
}

// buildDockerImage builds the Docker image for integration tests
func buildDockerImage() error {
	// Change to the root directory to build from the project root
	cmd := exec.Command("docker", "build", "-t", dockerImageName, "-f", dockerfilePath, ".")
	cmd.Dir = "../.." // Go to project root from test/integration/
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cleanupDockerImage removes the Docker image (optional)
func cleanupDockerImage() error {
	cmd := exec.Command("docker", "rmi", dockerImageName)
	return cmd.Run()
}