package output

import (
	"strings"
	"testing"
)

func TestWriteTitle(t *testing.T) {
	var b strings.Builder
	WriteTitle(&b, "Plonk Status")
	want := "Plonk Status\n============\n\n"
	if got := b.String(); got != want {
		t.Errorf("WriteTitle() = %q, want %q", got, want)
	}
}

func TestWriteRemoteSync(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		var b strings.Builder
		WriteRemoteSync(&b, "behind by 2 commits (run plonk pull)")
		want := "Remote: behind by 2 commits (run plonk pull)\n\n"
		if got := b.String(); got != want {
			t.Errorf("WriteRemoteSync() = %q, want %q", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var b strings.Builder
		WriteRemoteSync(&b, "")
		if got := b.String(); got != "" {
			t.Errorf("WriteRemoteSync(\"\") = %q, want empty", got)
		}
	})
}

func TestWriteErrors(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		var b strings.Builder
		errors := []Item{
			{Name: "foo", Error: "not found"},
			{Name: "bar"},
		}
		WriteErrors(&b, "package", errors)
		got := b.String()
		if !strings.Contains(got, "package errors:") {
			t.Errorf("expected domain header, got %q", got)
		}
		if !strings.Contains(got, "foo: not found") {
			t.Errorf("expected error detail, got %q", got)
		}
	})

	t.Run("no errors", func(t *testing.T) {
		var b strings.Builder
		WriteErrors(&b, "package", nil)
		if got := b.String(); got != "" {
			t.Errorf("expected empty output for nil errors, got %q", got)
		}
	})
}
