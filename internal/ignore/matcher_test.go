// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package ignore

import "testing"

func TestMatcherShouldIgnore(t *testing.T) {
	m := NewMatcher([]string{"*.log", "!keep.log", "/root-only.txt"})

	if !m.ShouldIgnore("project/error.log", false) {
		t.Fatalf("expected project/error.log to be ignored via *.log")
	}

	if m.ShouldIgnore("keep.log", false) {
		t.Fatalf("expected keep.log to be re-included via !keep.log")
	}

	if !m.ShouldIgnore("root-only.txt", false) {
		t.Fatalf("expected root-only.txt to be ignored via /root-only.txt")
	}

	if m.ShouldIgnore("nested/root-only.txt", false) {
		t.Fatalf("expected nested/root-only.txt to be allowed by /root-only.txt")
	}
}
