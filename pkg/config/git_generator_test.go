package config

import (
	"strings"
	"testing"
)

func TestGenerateGitconfig_BasicUserConfig(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":     "Rich Haase",
			"email":    "haaserdh@gmail.com",
			"username": "richhaase",
		},
		Core: map[string]string{
			"pager":       "delta",
			"excludesfile": "~/.gitignore_global",
		},
	}

	result := GenerateGitconfig(config)

	// Should contain user section
	if !strings.Contains(result, "[user]") {
		t.Error("Expected [user] section in gitconfig")
	}
	if !strings.Contains(result, `name = "Rich Haase"`) {
		t.Error("Expected user name in gitconfig")
	}
	if !strings.Contains(result, `email = "haaserdh@gmail.com"`) {
		t.Error("Expected user email in gitconfig")
	}
	if !strings.Contains(result, `username = "richhaase"`) {
		t.Error("Expected username in gitconfig")
	}

	// Should contain core section
	if !strings.Contains(result, "[core]") {
		t.Error("Expected [core] section in gitconfig")
	}
	if !strings.Contains(result, `pager = "delta"`) {
		t.Error("Expected core pager in gitconfig")
	}
	if !strings.Contains(result, `excludesfile = "~/.gitignore_global"`) {
		t.Error("Expected core excludesfile in gitconfig")
	}
}

func TestGenerateGitconfig_WithAliases(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":  "Test User",
			"email": "test@example.com",
		},
		Aliases: map[string]string{
			"conflicts": "diff --name-only --diff-filter=U",
			"ls":        `log --pretty=format:"%C(yellow)%h%Cred%d\\ %Creset%s%Cblue\\ [%cn]" --decorate`,
			"sb":        "status -sb",
			"undo":      "reset HEAD~1 --mixed",
		},
	}

	result := GenerateGitconfig(config)

	// Should contain alias section
	if !strings.Contains(result, "[alias]") {
		t.Error("Expected [alias] section in gitconfig")
	}
	if !strings.Contains(result, `conflicts = "diff --name-only --diff-filter=U"`) {
		t.Error("Expected conflicts alias in gitconfig")
	}
	if !strings.Contains(result, `ls = "log --pretty=format:"%C(yellow)%h%Cred%d\\ %Creset%s%Cblue\\ [%cn]" --decorate"`) {
		t.Error("Expected ls alias in gitconfig")
	}
	if !strings.Contains(result, `sb = "status -sb"`) {
		t.Error("Expected sb alias in gitconfig")
	}
	if !strings.Contains(result, `undo = "reset HEAD~1 --mixed"`) {
		t.Error("Expected undo alias in gitconfig")
	}
}

func TestGenerateGitconfig_WithColorAndBehaviorSettings(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":  "Test User",
			"email": "test@example.com",
		},
		Color: map[string]string{
			"ui":     "auto",
			"branch": "auto",
			"diff":   "auto",
			"status": "auto",
		},
		Push: map[string]string{
			"default":         "current",
			"autoSetupRemote": "true",
			"followTags":      "true",
		},
		Pull: map[string]string{
			"rebase": "merges",
		},
		Init: map[string]string{
			"defaultBranch": "trunk",
		},
	}

	result := GenerateGitconfig(config)

	// Should contain color section
	if !strings.Contains(result, "[color]") {
		t.Error("Expected [color] section in gitconfig")
	}
	if !strings.Contains(result, `ui = "auto"`) {
		t.Error("Expected color ui setting in gitconfig")
	}
	if !strings.Contains(result, `branch = "auto"`) {
		t.Error("Expected color branch setting in gitconfig")
	}

	// Should contain push section
	if !strings.Contains(result, "[push]") {
		t.Error("Expected [push] section in gitconfig")
	}
	if !strings.Contains(result, `default = "current"`) {
		t.Error("Expected push default setting in gitconfig")
	}
	if !strings.Contains(result, `autoSetupRemote = "true"`) {
		t.Error("Expected push autoSetupRemote setting in gitconfig")
	}

	// Should contain pull section
	if !strings.Contains(result, "[pull]") {
		t.Error("Expected [pull] section in gitconfig")
	}
	if !strings.Contains(result, `rebase = "merges"`) {
		t.Error("Expected pull rebase setting in gitconfig")
	}

	// Should contain init section
	if !strings.Contains(result, "[init]") {
		t.Error("Expected [init] section in gitconfig")
	}
	if !strings.Contains(result, `defaultBranch = "trunk"`) {
		t.Error("Expected init defaultBranch setting in gitconfig")
	}
}

func TestGenerateGitconfig_WithAdvancedSettings(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":  "Test User",
			"email": "test@example.com",
		},
		Delta: map[string]string{
			"navigate":     "true",
			"side-by-side": "true",
		},
		Fetch: map[string]string{
			"prune":             "true",
			"recurseSubmodules": "on-demand",
		},
		Status: map[string]string{
			"submoduleSummary":  "true",
			"showUntrackedFiles": "all",
			"short":             "false",
		},
		Diff: map[string]string{
			"submodule":   "log",
			"colorMoved":  "zebra",
		},
		Log: map[string]string{
			"abbrevCommit": "true",
		},
		Rerere: map[string]string{
			"enabled": "true",
		},
		Branch: map[string]string{
			"sort": "-committerdate",
		},
		Rebase: map[string]string{
			"autosquash": "true",
		},
		Merge: map[string]string{
			"conflictstyle": "diff3",
		},
	}

	result := GenerateGitconfig(config)

	// Should contain delta section
	if !strings.Contains(result, "[delta]") {
		t.Error("Expected [delta] section in gitconfig")
	}
	if !strings.Contains(result, `navigate = "true"`) {
		t.Error("Expected delta navigate setting in gitconfig")
	}
	if !strings.Contains(result, `side-by-side = "true"`) {
		t.Error("Expected delta side-by-side setting in gitconfig")
	}

	// Should contain fetch section
	if !strings.Contains(result, "[fetch]") {
		t.Error("Expected [fetch] section in gitconfig")
	}
	if !strings.Contains(result, `prune = "true"`) {
		t.Error("Expected fetch prune setting in gitconfig")
	}
	if !strings.Contains(result, `recurseSubmodules = "on-demand"`) {
		t.Error("Expected fetch recurseSubmodules setting in gitconfig")
	}

	// Should contain various other sections
	if !strings.Contains(result, "[status]") {
		t.Error("Expected [status] section in gitconfig")
	}
	if !strings.Contains(result, "[diff]") {
		t.Error("Expected [diff] section in gitconfig")
	}
	if !strings.Contains(result, "[log]") {
		t.Error("Expected [log] section in gitconfig")
	}
	if !strings.Contains(result, "[rerere]") {
		t.Error("Expected [rerere] section in gitconfig")
	}
	if !strings.Contains(result, "[branch]") {
		t.Error("Expected [branch] section in gitconfig")
	}
	if !strings.Contains(result, "[rebase]") {
		t.Error("Expected [rebase] section in gitconfig")
	}
	if !strings.Contains(result, "[merge]") {
		t.Error("Expected [merge] section in gitconfig")
	}
}

func TestGenerateGitconfig_WithSpecialSections(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":  "Test User",
			"email": "test@example.com",
		},
		Filter: map[string]map[string]string{
			"lfs": {
				"clean":    "git-lfs clean -- %f",
				"smudge":   "git-lfs smudge -- %f",
				"process":  "git-lfs filter-process",
				"required": "true",
			},
		},
	}

	result := GenerateGitconfig(config)

	// Should contain filter "lfs" section
	if !strings.Contains(result, `[filter "lfs"]`) {
		t.Error("Expected [filter \"lfs\"] section in gitconfig")
	}
	if !strings.Contains(result, `clean = "git-lfs clean -- %f"`) {
		t.Error("Expected filter lfs clean setting in gitconfig")
	}
	if !strings.Contains(result, `smudge = "git-lfs smudge -- %f"`) {
		t.Error("Expected filter lfs smudge setting in gitconfig")
	}
	if !strings.Contains(result, `process = "git-lfs filter-process"`) {
		t.Error("Expected filter lfs process setting in gitconfig")
	}
	if !strings.Contains(result, `required = "true"`) {
		t.Error("Expected filter lfs required setting in gitconfig")
	}
}

func TestGenerateGitconfig_EmptyConfig(t *testing.T) {
	config := &GitConfig{}

	result := GenerateGitconfig(config)

	// Should contain header comment
	if !strings.Contains(result, "# Generated by plonk - Do not edit manually") {
		t.Error("Expected header comment in gitconfig")
	}

	// Should not contain any sections for empty config
	if strings.Contains(result, "[user]") {
		t.Error("Empty config should not contain [user] section")
	}
	if strings.Contains(result, "[core]") {
		t.Error("Empty config should not contain [core] section")
	}
}

func TestGenerateGitconfig_SectionOrder(t *testing.T) {
	config := &GitConfig{
		User: map[string]string{
			"name":  "Test User",
			"email": "test@example.com",
		},
		Core: map[string]string{
			"pager": "delta",
		},
		Aliases: map[string]string{
			"st": "status",
		},
		Color: map[string]string{
			"ui": "auto",
		},
	}

	result := GenerateGitconfig(config)
	lines := strings.Split(result, "\n")

	var userIndex, coreIndex, aliasIndex, colorIndex int

	for i, line := range lines {
		if strings.Contains(line, "[user]") {
			userIndex = i
		}
		if strings.Contains(line, "[core]") {
			coreIndex = i
		}
		if strings.Contains(line, "[alias]") {
			aliasIndex = i
		}
		if strings.Contains(line, "[color]") {
			colorIndex = i
		}
	}

	// Verify sections appear in expected order
	if userIndex == 0 || coreIndex == 0 || aliasIndex == 0 || colorIndex == 0 {
		t.Error("Could not find expected sections")
	}

	// User should come first
	if userIndex >= coreIndex {
		t.Error("User section should come before core section")
	}
	
	// Core should come before aliases
	if coreIndex >= aliasIndex {
		t.Error("Core section should come before alias section")
	}
}