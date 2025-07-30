// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import "fmt"

// ProgressUpdate prints a progress update in consistent format
// Only shows progress for multi-item operations (total > 1)
func ProgressUpdate(current, total int, operation, item string) {
	if total <= 1 {
		return // No progress for single items
	}
	fmt.Printf("[%d/%d] %s: %s\n", current, total, operation, item)
}

// StageUpdate prints a stage update for multi-stage operations
func StageUpdate(stage string) {
	fmt.Printf("%s\n", stage)
}
