// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import "fmt"

// ValidateBatchResults enforces the documented partial-failure exit policy:
// any failed item in a batch produces a non-zero exit status, matching the
// behavior of track/untrack/apply. This keeps exit codes consistent across
// all batch commands: success (exit 0) only when every requested item
// succeeded (skips do not count as failures).
//
// It takes the number of results and a predicate that returns true if the
// result at index i failed.
func ValidateBatchResults(count int, operationName string, isFailed func(i int) bool) error {
	for i := 0; i < count; i++ {
		if isFailed(i) {
			return fmt.Errorf("%s operation failed: %d of %d item(s) failed to process", operationName, failedCount(count, isFailed), count)
		}
	}

	return nil
}

// failedCount counts failed items for the error message
func failedCount(count int, isFailed func(i int) bool) int {
	n := 0
	for i := 0; i < count; i++ {
		if isFailed(i) {
			n++
		}
	}
	return n
}
