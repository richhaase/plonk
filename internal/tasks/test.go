// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package tasks

import "fmt"

func Test() error {
	fmt.Println("Running tests...")

	if err := Run("go", "test", "./..."); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("✅ Tests passed!")
	return nil
}
