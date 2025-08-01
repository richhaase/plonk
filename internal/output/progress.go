// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import "fmt"

// ProgressUpdate prints a progress update in consistent format
// Shows progress for all operations when using apply command
func ProgressUpdate(current, total int, operation, item string) {
	if total > 1 {
		fmt.Printf("[%d/%d] %s: %s\n", current, total, operation, item)
	} else if total == 1 {
		// For single items, still show what we're doing
		fmt.Printf("%s: %s\n", operation, item)
	}
}

// StageUpdate prints a stage update for multi-stage operations
func StageUpdate(stage string) {
	fmt.Printf("%s\n", stage)
}
