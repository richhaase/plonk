package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileComparator_EqualAndDifferent(t *testing.T) {
	cmp := NewFileComparator()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	if err := os.WriteFile(a, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c, []byte("diff"), 0o644); err != nil {
		t.Fatal(err)
	}

	eq, err := cmp.CompareFiles(a, b)
	if err != nil || !eq {
		t.Fatalf("expected a==b, err=%v eq=%v", err, eq)
	}
	eq2, err := cmp.CompareFiles(a, c)
	if err != nil || eq2 {
		t.Fatalf("expected a!=c, err=%v eq=%v", err, eq2)
	}
}
