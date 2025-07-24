// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package parsers provides common output parsing utilities for package managers
package parsers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Package represents a parsed package from manager output
type Package struct {
	Name    string
	Version string
}

// ParseJSON parses JSON output from package managers
func ParseJSON(output []byte, extractor func([]byte) ([]Package, error)) ([]string, error) {
	packages, err := extractor(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	names := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if pkg.Name != "" {
			names = append(names, strings.ToLower(pkg.Name))
		}
	}

	return names, nil
}

// ParseLines parses line-based output from package managers
func ParseLines(output []byte, parser LineParser) []string {
	lines := strings.Split(string(output), "\n")
	var packages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if parser.ShouldSkip(line) {
			continue
		}

		pkg := parser.ParseLine(line)
		if pkg != "" {
			packages = append(packages, strings.ToLower(pkg))
		}
	}

	return packages
}

// LineParser defines how to parse individual lines
type LineParser interface {
	ShouldSkip(line string) bool
	ParseLine(line string) string
}

// SimpleLineParser extracts the first word from each line
type SimpleLineParser struct {
	SkipHeaders  int
	SkipPatterns []string
}

func (p *SimpleLineParser) ShouldSkip(line string) bool {
	if line == "" {
		return true
	}

	for _, pattern := range p.SkipPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

func (p *SimpleLineParser) ParseLine(line string) string {
	parts := strings.Fields(line)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ColumnLineParser extracts a specific column from each line
type ColumnLineParser struct {
	Column       int
	Delimiter    string
	SkipHeaders  int
	SkipPatterns []string
}

func (p *ColumnLineParser) ShouldSkip(line string) bool {
	if line == "" {
		return true
	}

	for _, pattern := range p.SkipPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

func (p *ColumnLineParser) ParseLine(line string) string {
	var parts []string
	if p.Delimiter != "" {
		parts = strings.Split(line, p.Delimiter)
	} else {
		parts = strings.Fields(line)
	}

	if p.Column < len(parts) {
		return strings.TrimSpace(parts[p.Column])
	}
	return ""
}

// ParseSimpleList parses output where each line is a package name
func ParseSimpleList(output []byte) []string {
	parser := &SimpleLineParser{}
	return ParseLines(output, parser)
}

// ParseTableOutput parses table-formatted output (like pip list)
func ParseTableOutput(output []byte, skipHeaders int) []string {
	if len(output) == 0 {
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	packages := []string{}

	for i, line := range lines {
		// Skip header lines
		if i < skipHeaders {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "---") {
			continue
		}

		// Extract first column (package name)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, strings.ToLower(parts[0]))
		}
	}

	return packages
}

// ExtractVersion extracts version from various formats
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

// ExtractKeyValue extracts a value for a key in "Key: Value" format.
// This is more robust than ExtractVersion for general key-value parsing.
func ExtractKeyValue(output []byte, key string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+":") {
			value := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
			// Clean up quotes if present
			value = strings.Trim(value, `"'`)
			return value
		}
	}
	return ""
}

// ParseKeyValuePairs parses multiple key-value pairs from output.
// Useful for commands like apt show, pip show, gem specification, etc.
func ParseKeyValuePairs(output []byte, keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		if value := ExtractKeyValue(output, key); value != "" {
			result[key] = value
		}
	}
	return result
}

// ParseDependencies extracts dependency lists from various formats.
// Handles comma-separated lists and removes version constraints.
func ParseDependencies(deps string) []string {
	if deps == "" {
		return []string{}
	}

	var result []string
	// Split by comma for most package managers
	parts := strings.Split(deps, ",")
	for _, dep := range parts {
		dep = strings.TrimSpace(dep)
		// Remove version constraints like (>= 1.0), [1.0], ~> 1.0
		if idx := strings.IndexAny(dep, "([{<>=~"); idx > 0 {
			dep = strings.TrimSpace(dep[:idx])
		}
		if dep != "" {
			result = append(result, dep)
		}
	}
	return result
}

// ParseIndentedList parses lists where items are indented (like gem dependencies).
func ParseIndentedList(output []byte, indent string) []string {
	var result []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// Check indentation before trimming
		if strings.HasPrefix(line, indent) && len(line) > len(indent) {
			// Make sure it's exactly the indent we want (not deeper)
			remainder := line[len(indent):]
			if !strings.HasPrefix(remainder, " ") && !strings.HasPrefix(remainder, "\t") {
				item := strings.TrimSpace(remainder)
				// Extract just the package name if there's additional info
				if idx := strings.IndexAny(item, " \t("); idx > 0 {
					item = item[:idx]
				}
				if item != "" {
					result = append(result, item)
				}
			}
		}
	}

	return result
}

// ExtractVersionFromPackageHeader extracts version from headers like "package v1.2.3:" or "package@1.2.3".
func ExtractVersionFromPackageHeader(header string, packageName string) string {
	// Check if header starts with package name
	if !strings.HasPrefix(header, packageName) {
		return ""
	}

	// Look for version patterns
	remainder := strings.TrimPrefix(header, packageName)
	remainder = strings.TrimSpace(remainder)

	// Handle "v1.2.3" or "@1.2.3"
	if strings.HasPrefix(remainder, "v") || strings.HasPrefix(remainder, "@") {
		version := strings.TrimPrefix(remainder, "v")
		version = strings.TrimPrefix(version, "@")
		version = strings.TrimSpace(version)

		// Remove trailing characters like ":" or additional info
		if idx := strings.IndexAny(version, ":, "); idx > 0 {
			version = version[:idx]
		}

		return version
	}

	// Handle space separated "package 2.1.0"
	if remainder != "" && !strings.HasPrefix(remainder, "(") {
		// Take the first word as version
		parts := strings.Fields(remainder)
		if len(parts) > 0 {
			// Check if it looks like a version (starts with digit)
			if len(parts[0]) > 0 && parts[0][0] >= '0' && parts[0][0] <= '9' {
				return parts[0]
			}
		}
	}

	// Handle parentheses format like "package (1.2.3)"
	if strings.HasPrefix(remainder, "(") {
		if endIdx := strings.Index(remainder, ")"); endIdx > 1 {
			return strings.TrimSpace(remainder[1:endIdx])
		}
	}

	return ""
}

// CleanVersionString removes common version prefixes and suffixes.
func CleanVersionString(version string) string {
	// Remove trailing comma first
	version = strings.TrimSuffix(version, ",")
	// Remove build info after space or parentheses
	if idx := strings.IndexAny(version, " \t("); idx > 0 {
		version = version[:idx]
	}
	// Remove outer quotes
	version = strings.Trim(version, `"'`)
	// Remove v prefix if present
	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
		// Remove quotes again if they were after the v
		version = strings.Trim(version, `"'`)
	}
	// Remove trailing punctuation
	version = strings.TrimSuffix(version, ":")
	return version
}

// NormalizePackageName normalizes package names according to common rules
func NormalizePackageName(name string) string {
	normalized := strings.ToLower(name)
	// Common normalizations
	normalized = strings.ReplaceAll(normalized, "-", "_") // Python style
	return normalized
}

// CleanJSONValue removes quotes and trailing commas from a JSON value extracted from text.
func CleanJSONValue(value string) string {
	// First trim the trailing comma if present
	value = strings.TrimSuffix(value, ",")
	// Then remove quotes
	value = strings.Trim(value, `"'`)
	return value
}

// Common JSON extractors for different package manager formats

// NPMPackage represents an npm package in JSON output
type NPMPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ExtractNPMPackages extracts packages from npm JSON output
func ExtractNPMPackages(output []byte) ([]Package, error) {
	var npmPkgs []NPMPackage
	if err := json.Unmarshal(output, &npmPkgs); err != nil {
		return nil, err
	}

	packages := make([]Package, 0, len(npmPkgs))
	for _, pkg := range npmPkgs {
		packages = append(packages, Package{
			Name:    pkg.Name,
			Version: pkg.Version,
		})
	}
	return packages, nil
}

// PipPackage represents a pip package in JSON output
type PipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ExtractPipPackages extracts packages from pip JSON output
func ExtractPipPackages(output []byte) ([]Package, error) {
	var pipPkgs []PipPackage
	if err := json.Unmarshal(output, &pipPkgs); err != nil {
		return nil, err
	}

	packages := make([]Package, 0, len(pipPkgs))
	for _, pkg := range pipPkgs {
		packages = append(packages, Package{
			Name:    pkg.Name,
			Version: pkg.Version,
		})
	}
	return packages, nil
}

// Common parsing utilities

// ParseVersionOutput extracts version from output using a specific prefix
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
	return "", fmt.Errorf("version not found with prefix %s", prefix)
}

// ParsePackageList parses a simple list of packages separated by newlines
func ParsePackageList(output []byte, separator string) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	if separator == "" {
		separator = "\n"
	}

	lines := strings.Split(result, separator)
	var packages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, strings.ToLower(line))
		}
	}

	return packages
}

// CleanPackageOutput removes common noise from package manager output
func CleanPackageOutput(output []byte) []byte {
	cleaned := string(output)

	// Remove common warning patterns
	patterns := []string{
		`WARNING:.*\n`,
		`DEPRECATION:.*\n`,
		`ERROR:.*\n`,
		`Note:.*\n`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	return []byte(strings.TrimSpace(cleaned))
}

// SplitAndFilterLines splits output by lines and filters using a predicate
func SplitAndFilterLines(output []byte, filter func(string) bool) []string {
	lines := strings.Split(string(output), "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && filter(line) {
			result = append(result, line)
		}
	}

	return result
}

// ExtractFirstWord extracts the first word from each line
func ExtractFirstWord(lines []string) []string {
	var result []string

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			result = append(result, strings.ToLower(fields[0]))
		}
	}

	return result
}

// ParseInfoKeyValue extracts key-value pairs from output for package info
func ParseInfoKeyValue(output []byte, keys []string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, key := range keys {
			prefix := key + ":"
			if strings.HasPrefix(line, prefix) {
				value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if value != "" {
					result[key] = value
				}
			}
		}
	}

	return result
}
