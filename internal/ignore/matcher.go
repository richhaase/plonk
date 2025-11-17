// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package ignore

import (
	"path"
	"path/filepath"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Matcher wraps gitignore semantics for plonk ignore_patterns.
type Matcher struct {
	matcher     gitignore.Matcher
	hasPatterns bool
}

// NewMatcher builds a gitignore matcher from the provided patterns.
func NewMatcher(patterns []string) *Matcher {
	compiled := make([]gitignore.Pattern, 0, len(patterns))

	for _, raw := range patterns {
		line := strings.TrimRight(raw, "\r\n")
		if strings.TrimSpace(line) == "" {
			continue
		}

		if len(line) > 0 && line[0] == '#' {
			continue
		}

		compiled = append(compiled, gitignore.ParsePattern(line, nil))
	}

	if len(compiled) == 0 {
		return &Matcher{}
	}

	return &Matcher{
		matcher:     gitignore.NewMatcher(compiled),
		hasPatterns: true,
	}
}

// ShouldIgnore reports whether the normalized path should be ignored.
func (m *Matcher) ShouldIgnore(relPath string, isDir bool) bool {
	if m == nil || !m.hasPatterns {
		return false
	}

	normalized := normalizePath(relPath)
	if normalized == "" {
		return false
	}

	parts := strings.Split(normalized, "/")
	return m.matcher.Match(parts, isDir)
}

func normalizePath(p string) string {
	if p == "" {
		return ""
	}

	norm := filepath.ToSlash(p)
	norm = path.Clean(norm)

	if norm == "." {
		return ""
	}

	for strings.HasPrefix(norm, "./") {
		norm = strings.TrimPrefix(norm, "./")
	}

	for len(norm) > 0 && norm[0] == '/' {
		norm = norm[1:]
	}

	return norm
}
