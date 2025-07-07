// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package tasks

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func Build() error {
	fmt.Println("Building plonk...")

	if err := os.MkdirAll("build", 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	version := getVersion()
	gitCommit := getGitCommit()
	buildDate := getBuildDate()

	ldflags := fmt.Sprintf("-X 'plonk/internal/commands.Version=%s' -X 'plonk/internal/commands.GitCommit=%s' -X 'plonk/internal/commands.BuildDate=%s'",
		version, gitCommit, buildDate)

	if err := Run("go", "build", "-ldflags", ldflags, "-o", "build/plonk", "./cmd/plonk"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("✅ Built plonk binary to build/")
	return nil
}

func getVersion() string {
	if output, err := Output("git", "describe", "--tags", "--exact-match", "HEAD"); err == nil {
		return strings.TrimSpace(output)
	}
	if output, err := Output("git", "rev-parse", "--short", "HEAD"); err == nil {
		return "dev-" + strings.TrimSpace(output)
	}
	return "dev"
}

func getGitCommit() string {
	if output, err := Output("git", "rev-parse", "HEAD"); err == nil {
		return strings.TrimSpace(output)
	}
	return "unknown"
}

func getBuildDate() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
