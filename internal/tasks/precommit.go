// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package tasks

import "fmt"

func Precommit() error {
	fmt.Println("Running pre-commit checks...")

	if err := formatCode(); err != nil {
		return err
	}

	if err := runLinter(); err != nil {
		return err
	}

	if err := Test(); err != nil {
		return err
	}

	if err := runSecurity(); err != nil {
		return err
	}

	fmt.Println("✅ Pre-commit checks passed!")
	return nil
}

func formatCode() error {
	fmt.Println("🔧 Formatting Go code and organizing imports...")

	if err := Run("go", "run", "golang.org/x/tools/cmd/goimports", "-w", "."); err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}
	return nil
}

func runLinter() error {
	fmt.Println("🔍 Running linter...")

	if err := Run("go", "run", "github.com/golangci/golangci-lint/cmd/golangci-lint", "run", "--timeout=10m"); err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}
	return nil
}

func runSecurity() error {
	fmt.Println("🔍 Running govulncheck...")
	if err := Run("go", "run", "golang.org/x/vuln/cmd/govulncheck", "./..."); err != nil {
		return fmt.Errorf("govulncheck failed: %w", err)
	}

	fmt.Println("🔐 Running gosec...")
	if err := Run("go", "run", "github.com/securego/gosec/v2/cmd/gosec", "./..."); err != nil {
		return fmt.Errorf("gosec failed: %w", err)
	}

	return nil
}
