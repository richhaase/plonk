// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"testing"

	"github.com/richhaase/plonk/internal/template"
)

func TestDotfileManager_Deploy_ConfiguredMode(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/pi/agent/auth.json.tmpl"] = []byte(`{"token": "{{TOKEN}}"}`)
	fs.Dirs["/config/pi"] = true
	fs.Dirs["/config/pi/agent"] = true

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetResolvers(template.NewEnvResolverFromLookup(func(key string) (string, bool) {
		if key == "TOKEN" {
			return "secret", true
		}
		return "", false
	}))

	// Default source mode is 0644 (from memFileInfo.Mode).
	m.SetDeployModes(map[string]os.FileMode{
		"pi/agent/auth.json.tmpl": 0o600,
	})

	err := m.Deploy("pi/agent/auth.json.tmpl")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.pi/agent/auth.json"]
	if mode != 0o600 {
		t.Errorf("deployed mode = %v (0o%o), want 0o600", mode, mode)
	}
}

func TestDotfileManager_Deploy_DefaultModeWhenUnconfigured(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("content")
	// Only configure a mode for a different dotfile.
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetDeployModes(map[string]os.FileMode{"other": 0o600})

	err := m.Deploy("zshrc")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Source mode in MemoryFS is 0644 and no mode is configured for zshrc.
	mode := fs.ChmodCalls["/home/user/.zshrc"]
	if mode != 0o644 {
		t.Errorf("deployed mode = %v (0o%o), want default 0o644", mode, mode)
	}
}

func TestDotfileManager_Deploy_NoModesConfigured(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Deploy("zshrc")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.zshrc"]
	if mode != 0o644 {
		t.Errorf("deployed mode = %v (0o%o), want default 0o644", mode, mode)
	}
}

func TestDotfileManager_ApplyAll_ConfiguredMode(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/missing"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.SetDeployModes(map[string]os.FileMode{"missing": 0o600})

	_, err := m.ApplyAll(false)
	if err != nil {
		t.Fatalf("ApplyAll() error = %v", err)
	}

	mode := fs.ChmodCalls["/home/user/.missing"]
	if mode != 0o600 {
		t.Errorf("AppliedAll deployed mode = %v (0o%o), want 0o600", mode, mode)
	}
}
