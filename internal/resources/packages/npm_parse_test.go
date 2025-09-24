package packages

import (
	"context"
	"testing"
)

func TestNpmParseListOutput(t *testing.T) {
	n := &NpmManager{}
	out := []byte(`{"dependencies":{"typescript":{},"eslint":{}}}`)
	got := n.parseListOutput(out)
	if len(got) != 2 {
		t.Fatalf("want 2, got %v", got)
	}
}

func TestNpmParseInfoOutput(t *testing.T) {
	n := &NpmManager{}
	out := []byte(`{"name":"typescript","version":"5.4.2","description":"TS"}`)
	info := n.parseInfoOutput(out, "typescript")
	if info == nil || info.Version != "5.4.2" {
		t.Fatalf("bad info: %+v", info)
	}
}

func TestNpmParseSearchOutput(t *testing.T) {
	n := &NpmManager{}
	out := []byte(`[{"name":"typescript"},{"name":"ts-node"}]`)
	res := n.parseSearchOutput(out)
	if len(res) != 2 {
		t.Fatalf("want 2, got %v", res)
	}
}

func TestNpmInstalledVersion_JSONError(t *testing.T) {
	// Mock executor
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	// Make npm available
	mock.Responses["npm --version"] = CommandResponse{Output: []byte("10.0.0"), Error: nil}
	// IsInstalled check succeeds
	mock.Responses["npm list -g typescript"] = CommandResponse{Output: []byte("/usr/local\n└── typescript@5.4.2"), Error: nil}
	// InstalledVersion JSON call returns invalid JSON
	mock.Responses["npm list -g typescript --depth=0 --json"] = CommandResponse{Output: []byte("{"), Error: nil}

	n := &NpmManager{}
	if _, err := n.InstalledVersion(context.Background(), "typescript"); err == nil {
		t.Fatalf("expected JSON parse error")
	}
}
