// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"testing"

	"github.com/richhaase/plonk/internal/ignore"
)

func TestPathValidator_ShouldSkipPathGitignoreSemantics(t *testing.T) {
	pv := NewPathValidator("/home/user", "/home/user/.config/plonk")

	rootMatcher := ignore.NewMatcher([]string{"/AGENTS.md"})
	if !pv.ShouldSkipPath("AGENTS.md", mockFileInfo{name: "AGENTS.md", isDir: false}, rootMatcher) {
		t.Fatalf("expected root-level AGENTS.md to be skipped by /AGENTS.md")
	}
	if pv.ShouldSkipPath("codex/AGENTS.md", mockFileInfo{name: "AGENTS.md", isDir: false}, rootMatcher) {
		t.Fatalf("did not expect nested codex/AGENTS.md to be skipped by /AGENTS.md")
	}

	globalMatcher := ignore.NewMatcher([]string{"AGENTS.md"})
	if !pv.ShouldSkipPath("codex/AGENTS.md", mockFileInfo{name: "AGENTS.md", isDir: false}, globalMatcher) {
		t.Fatalf("expected AGENTS.md pattern to match nested codex/AGENTS.md")
	}

	dirMatcher := ignore.NewMatcher([]string{"codex/"})
	if !pv.ShouldSkipPath("codex", mockFileInfo{name: "codex", isDir: true}, dirMatcher) {
		t.Fatalf("expected directory codex to be skipped by codex/")
	}
	if !pv.ShouldSkipPath("codex/notes.txt", mockFileInfo{name: "notes.txt", isDir: false}, dirMatcher) {
		t.Fatalf("expected files inside ignored directory to be skipped")
	}
}
