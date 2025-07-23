// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package cli contains helpers specific to CLI interaction,
// such as argument parsing, confirmation prompts, and command helpers.
package cli

import "github.com/richhaase/plonk/internal/operations"

// GetMetadataString safely extracts string metadata from operation results
func GetMetadataString(result operations.OperationResult, key string) string {
	if result.Metadata == nil {
		return ""
	}
	if value, ok := result.Metadata[key].(string); ok {
		return value
	}
	return ""
}
