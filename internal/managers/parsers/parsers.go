// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package parsers provides common output parsing utilities for package managers
package parsers

import (
	"encoding/json"
	"fmt"
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

// NormalizePackageName normalizes package names according to common rules
func NormalizePackageName(name string) string {
	normalized := strings.ToLower(name)
	// Common normalizations
	normalized = strings.ReplaceAll(normalized, "-", "_") // Python style
	return normalized
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
