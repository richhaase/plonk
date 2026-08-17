// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
)

func TestLoad_DotfileRules(t *testing.T) {
	configContent := `
default_manager: brew
dotfiles:
  rules:
    - name: "pi/agent/auth.json.tmpl"
      mode: "0600"
    - name: "gitconfig"
      mode: "0400"
`
	tempDir := testutil.NewTestConfig(t, configContent)

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	want := []DotfileRule{
		{Name: "pi/agent/auth.json.tmpl", Mode: "0600"},
		{Name: "gitconfig", Mode: "0400"},
	}
	if !reflect.DeepEqual(cfg.Dotfiles.Rules, want) {
		t.Errorf("Rules = %+v, want %+v", cfg.Dotfiles.Rules, want)
	}
}

func TestLoad_DotfileRulesModeWithoutName(t *testing.T) {
	configContent := `
dotfiles:
  rules:
    - mode: "0600"
`
	tempDir := testutil.NewTestConfig(t, configContent)

	_, err := Load(tempDir)
	if err == nil {
		t.Fatal("expected validation error for rule without a name, got nil")
	}
}

func TestLoad_InvalidDeployModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "out of range", mode: "0778"},
		{name: "non-octal digit", mode: "0609"},
		{name: "too large", mode: "1000"},
		{name: "non-digit", mode: "abc"},
		{name: "negative", mode: "-1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configContent := `
dotfiles:
  rules:
    - name: "zshrc"
      mode: "` + tc.mode + `"
`
			tempDir := testutil.NewTestConfig(t, configContent)

			_, err := Load(tempDir)
			if err == nil {
				t.Errorf("expected validation error for mode %q, got nil", tc.mode)
			}
		})
	}
}

func TestDotfiles_DeployModes(t *testing.T) {
	cfg := Dotfiles{
		Rules: []DotfileRule{
			{Name: "pi/agent/auth.json.tmpl", Mode: "0600"},
			{Name: "gitconfig", Mode: "0400"},
			{Name: "noconfig", Mode: ""},
		},
	}

	modes, err := cfg.DeployModes()
	if err != nil {
		t.Fatalf("DeployModes() error = %v", err)
	}

	want := map[string]os.FileMode{
		"pi/agent/auth.json.tmpl": 0o600,
		"gitconfig":               0o400,
	}
	if !reflect.DeepEqual(modes, want) {
		t.Errorf("DeployModes() = %v, want %v", modes, want)
	}
}

func TestDotfiles_DeployModes_Empty(t *testing.T) {
	cfg := Dotfiles{}
	modes, err := cfg.DeployModes()
	if err != nil {
		t.Fatalf("DeployModes() error = %v", err)
	}
	if len(modes) != 0 {
		t.Errorf("DeployModes() = %v, want empty map", modes)
	}
}
