package output

import "testing"

func TestDotfileListFormatter_Table_And_Structured(t *testing.T) {
	managed := []DotfileItem{{Name: ".zshrc", Path: "/home/u/.zshrc", State: "managed", Target: "~/.zshrc", Source: "zshrc"}}
	missing := []DotfileItem{{Name: ".vimrc", Path: "/home/u/.vimrc", State: "missing"}}
	untracked := []DotfileItem{{Name: ".foo", Path: "/home/u/.foo", State: "untracked", Metadata: map[string]any{"path": "/home/u/.foo"}}}

	// Non-verbose
	f := NewDotfileListFormatter(DotfileStatusOutput{Managed: managed, Missing: missing, Untracked: untracked, Verbose: false})
	out := f.TableOutput()
	if !contains(out, "Dotfiles Summary") {
		t.Fatalf("expected title")
	}
	sd := f.StructuredData()
	if sd == nil {
		t.Fatalf("expected structured data")
	}

	// Verbose (includes untracked)
	f2 := NewDotfileListFormatter(DotfileStatusOutput{Managed: managed, Missing: missing, Untracked: untracked, Verbose: true})
	out2 := f2.TableOutput()
	if !(contains(out2, "untracked") || contains(out2, "?")) {
		t.Fatalf("expected untracked entries in verbose mode")
	}
}
