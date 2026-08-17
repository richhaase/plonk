// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/template"
)

func TestDotfileManager_Deploy_Keychain(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("password = {{keychain:svc/acct}}")

	keychain := template.NewMockSecretResolver(template.ProviderKeychain, map[string]string{"svc/acct": "supersecret"})
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }), keychain)

	if err := m.Deploy("gitconfig.tmpl"); err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	content, ok := fs.Files["/home/user/.gitconfig"]
	if !ok {
		t.Fatal("Deploy() did not create /home/user/.gitconfig")
	}
	if want := "password = supersecret"; string(content) != want {
		t.Errorf("Deploy() content = %q, want %q", string(content), want)
	}

	if _, ok := fs.Files["/home/user/.gitconfig.plonk.tmp"]; ok {
		t.Error("Deploy() left a temp file behind")
	}
}

func TestDotfileManager_Deploy_KeychainMissing(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("password = {{keychain:svc/missing}}")

	keychain := template.NewMockSecretResolver(template.ProviderKeychain, nil)
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }), keychain)

	err := m.Deploy("gitconfig.tmpl")
	if err == nil {
		t.Fatal("Deploy() expected error for missing keychain secret")
	}
	if !strings.Contains(err.Error(), "svc/missing") {
		t.Errorf("Deploy() error should name the locator only: %v", err)
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatal("Deploy() error leaked a secret value")
	}
}

func TestDotfileManager_HasSecrets(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     bool
	}{
		{"keychain", "key={{keychain:svc/acct}}", true},
		{"env only", "key={{env:SOMETHING}}", false},
		{"legacy env", "key={{SOMETHING}}", false},
		{"mixed", "a={{A}} b={{keychain:s/acct}}", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMemoryFS()
			fs.Dirs["/config"] = true
			fs.Files["/config/x.tmpl"] = []byte(tt.template)
			m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
			got, err := m.HasSecrets("x.tmpl")
			if err != nil {
				t.Fatalf("HasSecrets() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("HasSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDotfileManager_RenderForDiff_MasksSecrets(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/creds.tmpl"] = []byte("password = {{keychain:svc/acct}}\nhost = example.com")

	// The keychain now returns a NEW value; the deployed target still holds an older
	// secret plonk can no longer resolve. Both must be masked in the diff output.
	keychain := template.NewMockSecretResolver(template.ProviderKeychain, map[string]string{"svc/acct": "new-secret-88"})
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }), keychain)

	target := []byte("password = stale-old-secret-7\nhost = example.com")
	src, dst, err := m.RenderForDiff("creds.tmpl", target)
	if err != nil {
		t.Fatalf("RenderForDiff() error = %v", err)
	}

	if !strings.Contains(string(src), template.RedactedMarker) {
		t.Errorf("masked source missing %q: %q", template.RedactedMarker, string(src))
	}
	if !strings.Contains(string(dst), template.RedactedMarker) {
		t.Errorf("masked target missing %q: %q", template.RedactedMarker, string(dst))
	}
	for _, leaked := range []string{"new-secret-88", "stale-old-secret-7"} {
		if strings.Contains(string(src), leaked) || strings.Contains(string(dst), leaked) {
			t.Errorf("RenderForDiff() leaked %q", leaked)
		}
	}
	if !strings.Contains(string(src), "example.com") {
		t.Errorf("non-secret content should be preserved: %q", string(src))
	}
}

func TestDotfileManager_IsDrifted_KeychainInMemory(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/creds.tmpl"] = []byte("password = {{keychain:svc/acct}}")
	fs.Files["/home/user/.creds"] = []byte("password = tok-1")

	keychain := template.NewMockSecretResolver(template.ProviderKeychain, map[string]string{"svc/acct": "tok-1"})
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }), keychain)

	d := Dotfile{Name: "creds.tmpl", Source: "/config/creds.tmpl", Target: "/home/user/.creds"}
	drifted, err := m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if drifted {
		t.Error("IsDrifted() = true, want false when rendered value matches target")
	}

	fs.Files["/home/user/.creds"] = []byte("password = tok-2")
	changed, err := m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if !changed {
		t.Error("IsDrifted() = false, want true when target differs")
	}
}
