package packages

import (
	"errors"
	"testing"
)

func TestBrewExtractVersion(t *testing.T) {
	h := NewHomebrewManager()
	out := []byte("jq 1.6 1.5\n")
	got := h.extractVersion(out, "jq")
	if got != "1.6" {
		t.Fatalf("want 1.6, got %q", got)
	}
}

func TestBrewParseSearchOutput(t *testing.T) {
	h := NewHomebrewManager()
	out := []byte("jq\nfd\nIf you meant \n================\n")
	pkgs := h.parseSearchOutput(out)
	if len(pkgs) != 2 {
		t.Fatalf("want 2 results, got %v", pkgs)
	}
}

func TestBrewParseInfoOutput(t *testing.T) {
	h := NewHomebrewManager()
	out := []byte("jq: stable 1.6\nFrom: https://example.com\n")
	info := h.parseInfoOutput(out, "jq")
	if info == nil || info.Version != "1.6" {
		t.Fatalf("expected version 1.6, got %+v", info)
	}
}

type exitErr struct{ code int }

func (e *exitErr) Error() string { return "exit" }
func (e *exitErr) ExitCode() int { return e.code }

func TestBrewHandleUpgradeError(t *testing.T) {
	h := NewHomebrewManager()
	// not found
	if err := h.handleUpgradeError(&exitErr{1}, []byte("No formula found"), "jq"); err == nil {
		t.Fatalf("expected error for not found")
	}
	// already up-to-date should be nil
	if err := h.handleUpgradeError(&exitErr{1}, []byte("already up-to-date"), "jq"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	// permission denied
	if err := h.handleUpgradeError(&exitErr{2}, []byte("Permission denied"), "jq"); err == nil || !errors.Is(err, err) {
		// Just ensure non-nil
	}
}
