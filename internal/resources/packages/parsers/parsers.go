// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package parsers provides common output parsing utilities for package managers
package parsers

import (
	"fmt"
	"strings"
)

// ExtractVersion extracts version information from command output
func ExtractVersion(output []byte, prefix string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			version := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			return version
		}
	}
	return ""
}

// CleanJSONValue removes quotes and trailing commas from JSON values
func CleanJSONValue(value string) string {
	// First trim the trailing comma if present
	value = strings.TrimSuffix(value, ",")
	// Then remove quotes
	value = strings.Trim(value, `"'`)
	return value
}

// ParseVersionOutput parses version output with a specific prefix
func ParseVersionOutput(output []byte, prefix string) (string, error) {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			version := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			if version != "" {
				return CleanVersionString(version), nil
			}
		}
	}
	return "", fmt.Errorf("version not found with prefix '%s'", prefix)
}

// CleanVersionString removes common version prefixes and suffixes
func CleanVersionString(version string) string {
	// Remove common prefixes
	prefixes := []string{"v", "version", "Version"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(version, prefix) {
			version = strings.TrimSpace(strings.TrimPrefix(version, prefix))
			break
		}
	}

	// Remove common suffixes and extra information
	if idx := strings.Index(version, " "); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "\t"); idx > 0 {
		version = version[:idx]
	}

	return strings.TrimSpace(version)
}
