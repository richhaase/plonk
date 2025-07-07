// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package tasks

import (
	"fmt"
	"os"
)

func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	
	if err := os.RemoveAll("build"); err != nil {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}
	
	if err := Run("go", "clean"); err != nil {
		return fmt.Errorf("go clean failed: %w", err)
	}
	
	fmt.Println("✅ Build artifacts cleaned")
	return nil
}