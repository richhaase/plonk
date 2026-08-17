// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/template"
)

func TestCheckTemplateReadiness_Env(t *testing.T) {
	configDir := writeTempl("secret={{KEY}}")
	defer os.RemoveAll(configDir)

	classifier := templateIssueClassifier{
		ctx: context.Background(),
		env: template.NewEnvResolverFromLookup(func(key string) (string, bool) {
			if key == "KEY" {
				return "v", true
			}
			return "", false
		}),
		keychain: template.NewMockSecretResolver(template.ProviderKeychain, nil),
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "pass" {
		t.Fatalf("Status = %q, want pass (issues: %v)", check.Status, check.Issues)
	}
}

func TestCheckTemplateReadiness_MissingEnv(t *testing.T) {
	configDir := writeTempl("secret={{KEY}}")
	defer os.RemoveAll(configDir)

	classifier := templateIssueClassifier{
		ctx:      context.Background(),
		env:      template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain: template.NewMockSecretResolver(template.ProviderKeychain, nil),
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "warn" {
		t.Fatalf("Status = %q, want warn", check.Status)
	}
	if !strings.Contains(strings.Join(check.Issues, " "), "KEY") {
		t.Errorf("Issues do not name the env locator: %v", check.Issues)
	}
	if len(check.Suggestions) == 0 || !strings.Contains(strings.Join(check.Suggestions, " "), "KEY") {
		t.Errorf("Suggestions should include remediation for KEY: %v", check.Suggestions)
	}
}

func TestCheckTemplateReadiness_KeychainFound(t *testing.T) {
	configDir := writeTempl("secret={{keychain:svc/acct}}")
	defer os.RemoveAll(configDir)

	keychain := template.NewMockSecretResolver(template.ProviderKeychain, map[string]string{"svc/acct": "s3cret"})
	classifier := templateIssueClassifier{
		ctx:      context.Background(),
		env:      template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain: keychain,
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "pass" {
		t.Fatalf("Status = %q, want pass (issues: %v)", check.Status, check.Issues)
	}
}

func TestCheckTemplateReadiness_KeychainMissing(t *testing.T) {
	configDir := writeTempl("secret={{keychain:svc/acct}}")
	defer os.RemoveAll(configDir)

	classifier := templateIssueClassifier{
		ctx:      context.Background(),
		env:      template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain: template.NewMockSecretResolver(template.ProviderKeychain, nil),
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "warn" {
		t.Fatalf("Status = %q, want warn", check.Status)
	}
	if !strings.Contains(strings.Join(check.Issues, " "), "keychain:svc/acct") {
		t.Errorf("Issues should name the keychain locator only: %v", check.Issues)
	}
	found := false
	for _, s := range check.Suggestions {
		if strings.Contains(s, "security add-generic-password") && strings.Contains(s, "svc") {
			found = true
		}
	}
	if !found {
		t.Errorf("Suggestions should include provider-owned remediation: %v", check.Suggestions)
	}
}

func TestCheckTemplateReadiness_KeychainUnavailable(t *testing.T) {
	configDir := writeTempl("secret={{keychain:svc/acct}}")
	defer os.RemoveAll(configDir)

	keychain := template.NewMockSecretResolver(template.ProviderKeychain, nil)
	keychain.SetFallbackError(template.ErrProviderUnavailable)
	classifier := templateIssueClassifier{
		ctx:      context.Background(),
		env:      template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain: keychain,
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "warn" {
		t.Fatalf("Status = %q, want warn", check.Status)
	}
	if !strings.Contains(strings.Join(check.Issues, " "), "provider unavailable") {
		t.Errorf("Issues should surface provider unavailability: %v", check.Issues)
	}
}

func TestCheckTemplateReadiness_Malformed(t *testing.T) {
	configDir := writeTempl("bad={{unknown:foo}}")
	defer os.RemoveAll(configDir)

	classifier := templateIssueClassifier{
		ctx:      context.Background(),
		env:      template.NewEnvResolverFromLookup(func(string) (string, bool) { return "", false }),
		keychain: template.NewMockSecretResolver(template.ProviderKeychain, nil),
	}

	check := checkTemplateReadinessAt(configDir, classifier)
	if check.Status != "warn" {
		t.Fatalf("Status = %q, want warn", check.Status)
	}
	if len(check.Issues) == 0 {
		t.Fatal("expected a malformed-syntax issue")
	}
}

// writeTempl creates a temp config dir with secrets.tmpl and returns its path.
func writeTempl(body string) string {
	dir, err := os.MkdirTemp("", "plonk-health-tmpl-*")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secrets.tmpl"), []byte(body), 0644); err != nil {
		panic(err)
	}
	return dir
}
