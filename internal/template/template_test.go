// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseDirectives(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int
		providers []string
		locators  []string
		wantErr   bool
	}{
		{
			name:      "legacy env var",
			input:     "email = {{EMAIL}}",
			want:      1,
			providers: []string{""},
			locators:  []string{"EMAIL"},
		},
		{
			name: "explicit env", input: "{{env:HOME}}",
			want: 1, providers: []string{ProviderEnv}, locators: []string{"HOME"},
		},
		{
			name: "keychain service account", input: `key={{keychain:mysvc/myacct}}`,
			want: 1, providers: []string{ProviderKeychain}, locators: []string{"mysvc/myacct"},
		},
		{
			name: "keychain service only", input: `key={{keychain:mysvc}}`,
			want: 1, providers: []string{ProviderKeychain}, locators: []string{"mysvc"},
		},
		{
			name:      "multiple mixed",
			input:     `{{A}} {{env:B}} {{keychain:s/x-y.z}}`,
			want:      3,
			providers: []string{"", ProviderEnv, ProviderKeychain},
			locators:  []string{"A", "B", "s/x-y.z"},
		},
		{
			name: "pass through non-directive brace", input: `{{#if}}`,
			want: 0,
		},
		{
			name:  "empty keychain locator",
			input: `{{keychain:}}`, want: 1, wantErr: true,
		},
		{
			name:  "unknown provider",
			input: `{{unknown:foo}}`, want: 1, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := Directives([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("Directives() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Directives() error = %v", err)
			}
			if len(ds) != tt.want {
				t.Fatalf("Directives() got %d, want %d", len(ds), tt.want)
			}
			for i, d := range ds {
				if d.Provider != tt.providers[i] {
					t.Errorf("directive[%d].Provider = %q, want %q", i, d.Provider, tt.providers[i])
				}
				if d.Locator != tt.locators[i] {
					t.Errorf("directive[%d].Locator = %q, want %q", i, d.Locator, tt.locators[i])
				}
			}
		})
	}
}

func TestRenderEnv(t *testing.T) {
	r := NewRenderer(NewEnvResolverFromLookup(func(key string) (string, bool) {
		vars := map[string]string{"EMAIL": "u@e.com", "NAME": "Test"}
		v, ok := vars[key]
		return v, ok
	}))

	out, err := r.Render(context.Background(), []byte("email={{EMAIL}} name={{env:NAME}}"), RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if want := "email=u@e.com name=Test"; string(out) != want {
		t.Errorf("Render() = %q, want %q", string(out), want)
	}
}

func TestRenderMissing(t *testing.T) {
	keychain := NewMockSecretResolver(ProviderKeychain, map[string]string{"s/a": "v"})
	r := NewRenderer(
		NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain,
	)
	_, err := r.Render(context.Background(), []byte("a={{MISSING_ONE}} b={{env:B}} c={{keychain:s/a}}"), RenderOptions{})
	if err == nil {
		t.Fatal("expected missing error")
	}
	if !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("error does not wrap ErrSecretNotFound: %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "MISSING_ONE") || !strings.Contains(msg, "B") {
		t.Errorf("missing error messages should list locators, got %q", msg)
	}
	if strings.Contains(msg, "ok") {
		t.Errorf("missing error should not include resolved value, got %q", msg)
	}
}

func TestRenderMaskSecrets(t *testing.T) {
	mock := NewMockSecretResolver(ProviderKeychain, map[string]string{"svc/acct": "supersecret"})
	r := NewRenderer(mock)
	out, err := r.Render(context.Background(), []byte("key={{keychain:svc/acct}}"), RenderOptions{MaskSecrets: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if want := "key=" + RedactedMarker; string(out) != want {
		t.Errorf("Render(masked) = %q, want %q", string(out), want)
	}
}

func TestRenderWithSecretsCollects(t *testing.T) {
	mock := NewMockSecretResolver(ProviderKeychain, map[string]string{"svc/acct": "token-123"})
	r := NewRenderer(mock)
	out, secrets, err := r.RenderWithSecrets(context.Background(), []byte("key={{keychain:svc/acct}}"), RenderOptions{})
	if err != nil {
		t.Fatalf("RenderWithSecrets() error = %v", err)
	}
	if string(out) != "key=token-123" {
		t.Errorf("RenderWithSecrets() = %q", string(out))
	}
	if len(secrets) != 1 || secrets[0] != "token-123" {
		t.Errorf("secrets = %v, want [token-123]", secrets)
	}
}
func TestKeychainResolverNonDarwinUnavailable(t *testing.T) {
	res := NewMacOSKeychainResolver()
	res.goos = "linux"
	_, err := res.Resolve(context.Background(), "svc/acct")
	if err == nil {
		t.Fatal("expected error on non-darwin")
	}
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got %v", err)
	}
}

func TestClassifyKeychainError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   error
	}{
		{"not found", "security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.", ErrSecretNotFound},
		{"locked interaction", "security: User interaction is not allowed.", ErrKeychainLocked},
		{"locked auth", "The user name or passphrase you entered is not correct.", ErrKeychainLocked},
		{"access denied", "some random message", ErrAccessDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyKeychainError(errors.New("exit status 1"), tt.stderr, "svc/acct")
			if !errors.Is(err, tt.want) {
				t.Errorf("classifyKeychainError() = %v, want %v", err, tt.want)
			}
			if strings.Contains(err.Error(), "token-123") {
				t.Fatal("error leaked a value")
			}
		})
	}
}

func TestClassifyKeychainTimeout(t *testing.T) {
	err := classifyKeychainError(context.DeadlineExceeded, "", "svc/acct")
	if !errors.Is(err, ErrKeychainLocked) {
		t.Errorf("expected ErrKeychainLocked, got %v", err)
	}
}

func TestParseKeychainLocator(t *testing.T) {
	svc, acct := ParseKeychainLocator("svc/acct")
	if svc != "svc" || acct != "acct" {
		t.Errorf("ParseKeychainLocator(svc/acct) = %q, %q", svc, acct)
	}
	svc, acct = ParseKeychainLocator("svc")
	if svc != "svc" || acct != "" {
		t.Errorf("ParseKeychainLocator(svc) = %q, %q", svc, acct)
	}
}

func TestRendererMaskedDoesNotCallSecretBackend(t *testing.T) {
	mock := NewMockSecretResolver(ProviderKeychain, map[string]string{"svc/acct": "x-that-leaks"})
	mock.SetFallbackError(errors.New("backend should not be called"))
	r := NewRenderer(mock)
	_, err := r.Render(context.Background(), []byte("a={{env:X}} b={{keychain:svc/acct}}"), RenderOptions{MaskSecrets: true})
	if err == nil {
		t.Fatal("expected env missing error")
	}
	_, err = r.Render(context.Background(), []byte("b={{keychain:svc/acct}}"), RenderOptions{MaskSecrets: true})
	if err != nil {
		t.Fatalf("masked render should not call backend, got %v", err)
	}
}

func TestParseKeychainLocatorRegexp(t *testing.T) {
	ds, err := Directives([]byte(`{{keychain:service.with-dots/under_score-1}}`))
	if err != nil {
		t.Fatalf("Directives() error = %v", err)
	}
	if len(ds) != 1 || ds[0].Locator != "service.with-dots/under_score-1" {
		t.Errorf("unexpected directive parsing: %+v", ds)
	}
}
