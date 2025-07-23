// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package core contains the core business logic for plonk.
// This package should never import from internal/commands or internal/cli.
package core

import (
	"strings"
)

// ExtractBinaryNameFromPath extracts the binary name from a Go module path
// Examples:
//   - github.com/user/tool -> tool
//   - github.com/user/project/cmd/tool -> tool
//   - github.com/user/tool@v1.2.3 -> tool
func ExtractBinaryNameFromPath(modulePath string) string {
	// Remove version specification if present
	modulePath = strings.Split(modulePath, "@")[0]

	// Extract the last component of the path
	parts := strings.Split(modulePath, "/")
	binaryName := parts[len(parts)-1]

	// Handle special case of .../cmd/toolname pattern
	if len(parts) >= 2 && parts[len(parts)-2] == "cmd" {
		return binaryName
	}

	// For simple cases, the binary name is usually the last component
	return binaryName
}
