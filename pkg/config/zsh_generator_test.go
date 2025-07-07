// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"strings"
	"testing"
)

func TestGenerateZshrc_BasicStructure(t *testing.T) {
	config := &ZSHConfig{
		EnvVars: map[string]string{
			"EDITOR": "nvim",
			"PAGER":  "bat",
		},
		Aliases: map[string]string{
			"ll":  "eza -la",
			"cat": "bat",
		},
		Inits: []string{
			`eval "$(starship init zsh)"`,
			`eval "$(zoxide init zsh)"`,
		},
	}

	result := GenerateZshrc(config)

	// Should contain env vars
	if !strings.Contains(result, "export EDITOR='nvim'") {
		t.Error("Expected EDITOR export in zshrc")
	}
	if !strings.Contains(result, "export PAGER='bat'") {
		t.Error("Expected PAGER export in zshrc")
	}

	// Should contain aliases
	if !strings.Contains(result, "alias ll='eza -la'") {
		t.Error("Expected ll alias in zshrc")
	}
	if !strings.Contains(result, "alias cat='bat'") {
		t.Error("Expected cat alias in zshrc")
	}

	// Should contain inits
	if !strings.Contains(result, `eval "$(starship init zsh)"`) {
		t.Error("Expected starship init in zshrc")
	}
	if !strings.Contains(result, `eval "$(zoxide init zsh)"`) {
		t.Error("Expected zoxide init in zshrc")
	}
}

func TestGenerateZshrc_WithCompletionsAndFunctions(t *testing.T) {
	config := &ZSHConfig{
		Completions: []string{
			`source <(kubectl completion zsh)`,
			`source <(gh completion -s zsh)`,
		},
		Functions: map[string]string{
			"mkcd": "mkdir -p \"$1\" && cd \"$1\"",
			"extract": `case "$1" in
  *.tar.gz) tar -xzf "$1" ;;
  *.zip) unzip "$1" ;;
esac`,
		},
		ShellOptions: []string{
			"AUTO_MENU",
			"COMPLETE_IN_WORD",
		},
	}

	result := GenerateZshrc(config)

	// Should contain completions
	if !strings.Contains(result, `source <(kubectl completion zsh)`) {
		t.Error("Expected kubectl completion in zshrc")
	}
	if !strings.Contains(result, `source <(gh completion -s zsh)`) {
		t.Error("Expected gh completion in zshrc")
	}

	// Should contain functions
	if !strings.Contains(result, "function mkcd() {") {
		t.Error("Expected mkcd function definition in zshrc")
	}
	if !strings.Contains(result, "mkdir -p \"$1\" && cd \"$1\"") {
		t.Error("Expected mkcd function body in zshrc")
	}
	if !strings.Contains(result, "function extract() {") {
		t.Error("Expected extract function definition in zshrc")
	}

	// Should contain shell options
	if !strings.Contains(result, "setopt AUTO_MENU") {
		t.Error("Expected AUTO_MENU setopt in zshrc")
	}
	if !strings.Contains(result, "setopt COMPLETE_IN_WORD") {
		t.Error("Expected COMPLETE_IN_WORD setopt in zshrc")
	}
}

func TestGenerateZshenv_OnlyEnvVars(t *testing.T) {
	config := &ZSHConfig{
		EnvVars: map[string]string{
			"EDITOR": "nvim",
			"PAGER":  "bat",
			"PATH":   "$PATH:/usr/local/bin",
		},
		Aliases: map[string]string{
			"ll": "eza -la", // Should not appear in .zshenv
		},
		Inits: []string{
			`eval "$(starship init zsh)"`, // Should not appear in .zshenv
		},
	}

	result := GenerateZshenv(config)

	// Should contain env vars
	if !strings.Contains(result, "export EDITOR='nvim'") {
		t.Error("Expected EDITOR export in zshenv")
	}
	if !strings.Contains(result, "export PAGER='bat'") {
		t.Error("Expected PAGER export in zshenv")
	}
	if !strings.Contains(result, "export PATH='$PATH:/usr/local/bin'") {
		t.Error("Expected PATH export in zshenv")
	}

	// Should NOT contain aliases or inits (those belong in .zshrc)
	if strings.Contains(result, "alias") {
		t.Error("zshenv should not contain aliases")
	}
	if strings.Contains(result, "eval") {
		t.Error("zshenv should not contain initialization commands")
	}
}

func TestGenerateZshrc_WithSourceBeforeAfter(t *testing.T) {
	config := &ZSHConfig{
		SourceBefore: []string{
			"source ~/.config/zsh/custom-before.zsh",
			"source ~/.local/share/before.zsh",
		},
		Aliases: map[string]string{
			"ll": "eza -la",
		},
		SourceAfter: []string{
			"source ~/.config/zsh/custom-after.zsh",
		},
	}

	result := GenerateZshrc(config)

	// Should source before files at the beginning (after header)
	lines := strings.Split(result, "\n")
	var beforeIndex, aliasIndex, afterIndex int

	for i, line := range lines {
		if strings.Contains(line, "source ~/.config/zsh/custom-before.zsh") {
			beforeIndex = i
		}
		if strings.Contains(line, "alias ll='eza -la'") {
			aliasIndex = i
		}
		if strings.Contains(line, "source ~/.config/zsh/custom-after.zsh") {
			afterIndex = i
		}
	}

	// Verify ordering: before < alias < after
	if beforeIndex == 0 || aliasIndex == 0 || afterIndex == 0 {
		t.Error("Could not find expected source statements or alias")
	}

	if beforeIndex >= aliasIndex {
		t.Error("SourceBefore should come before main config")
	}

	if afterIndex <= aliasIndex {
		t.Error("SourceAfter should come after main config")
	}

	// Should contain both source statements
	if !strings.Contains(result, "source ~/.config/zsh/custom-before.zsh") {
		t.Error("Expected SourceBefore statement in zshrc")
	}
	if !strings.Contains(result, "source ~/.local/share/before.zsh") {
		t.Error("Expected second SourceBefore statement in zshrc")
	}
	if !strings.Contains(result, "source ~/.config/zsh/custom-after.zsh") {
		t.Error("Expected SourceAfter statement in zshrc")
	}
}
