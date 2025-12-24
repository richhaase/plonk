// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import "fmt"

// ValidateBatchResults checks if all operations in a batch failed.
// It takes the number of results and a predicate that returns true if the result at index i failed.
// Returns an error if all results failed, nil otherwise.
func ValidateBatchResults(count int, operationName string, isFailed func(i int) bool) error {
	if count == 0 {
		return nil
	}

	allFailed := true
	for i := 0; i < count; i++ {
		if !isFailed(i) {
			allFailed = false
			break
		}
	}

	if allFailed {
		return fmt.Errorf("%s operation failed: all %d item(s) failed to process", operationName, count)
	}

	return nil
}
