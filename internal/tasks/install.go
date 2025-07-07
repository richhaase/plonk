// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package tasks

import "fmt"

func Install() error {
	fmt.Println("Installing plonk globally...")
	
	if err := Run("go", "install", "./cmd/plonk"); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	
	fmt.Println("✅ Plonk installed globally!")
	fmt.Println("Run 'plonk --help' to get started")
	return nil
}