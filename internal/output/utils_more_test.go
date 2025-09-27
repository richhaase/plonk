package output

import "testing"

func TestUtils_StatusIcons_And_Truncate(t *testing.T) {
	// Icons
	cases := []string{"managed", "added", "installed", "removed", "success", "completed", "deployed", "missing", "failed", "error", "skipped", "untracked", "info", "unknown"}
	for _, c := range cases {
		_ = GetStatusIcon(c)
	}
	// Truncate
	if TruncateString("abcd", 2) != "ab" {
		t.Fatalf("truncate basic")
	}
	if TruncateString("abcdef", 5) != "ab..." {
		t.Fatalf("truncate ellipsis")
	}
	if TruncateString("abc", 10) != "abc" {
		t.Fatalf("no truncate")
	}
}

func TestUtils_FormatErrors(t *testing.T) {
	if s := FormatValidationError("field", "x", "expect"); !contains(s, "invalid") {
		t.Fatalf("validation format: %s", s)
	}
	if s := FormatNotFoundError("item", "name", []string{"a", "b"}); !contains(s, "Valid options") {
		t.Fatalf("notfound format: %s", s)
	}
}

func TestPrintAndProgressHelpers(t *testing.T) {
	Printf("%s", "hello")
	Println("world")
	StageUpdate("Cloning...")
	ProgressUpdate(1, 1, "Installing", "jq")
	ProgressUpdate(1, 2, "Installing", "jq")
}
