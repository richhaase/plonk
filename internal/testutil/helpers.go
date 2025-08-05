// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// NewTestConfig creates a temporary directory with plonk.yaml containing the given content.
// Returns the directory path. The directory is automatically cleaned up via t.Cleanup.
func NewTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if content != "" {
		configPath := filepath.Join(dir, "plonk.yaml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}
	}
	return dir
}

// SetEnv sets an environment variable and automatically restores it after the test.
func SetEnv(t *testing.T, key, value string) {
	t.Helper()
	original := os.Getenv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if original == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, original)
		}
	})
}
