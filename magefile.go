//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

var Default = Build

// Build the plonk binary
func Build() error {
	fmt.Println("Building plonk...")
	if err := os.MkdirAll("build", 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	if err := sh.Run("go", "build", "-o", "build/plonk", "./cmd/plonk"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println("✅ Built plonk binary to build/")
	return nil
}

// Run all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "./...")
}

// Run linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=10m")
}

// Format code (gofmt)
func Format() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Clean build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	if err := sh.Rm("build"); err != nil {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}
	if err := sh.Run("go", "clean"); err != nil {
		return fmt.Errorf("go clean failed: %w", err)
	}
	fmt.Println("✅ Build artifacts cleaned")
	return nil
}
