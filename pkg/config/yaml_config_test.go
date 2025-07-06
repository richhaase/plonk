package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_BasicStructure(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
  - zshenv
  - plugins.zsh

homebrew:
  brews:
    - aichat
    - aider
    - name: neovim
      config: config/nvim/
  casks:
    - font-hack-nerd-font

asdf:
  - name: nodejs
    version: "24.2.0"
    config: config/npm/
  - name: python
    version: "3.13.2"

npm:
  - "@anthropic-ai/claude-code"
  - name: some-tool
    package: "@scope/different-name"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load configuration.
	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify settings.
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager 'homebrew', got '%s'", config.Settings.DefaultManager)
	}

	// Verify dotfiles.
	expectedDotfiles := []string{"zshrc", "zshenv", "plugins.zsh"}
	if len(config.Dotfiles) != len(expectedDotfiles) {
		t.Errorf("Expected %d dotfiles, got %d", len(expectedDotfiles), len(config.Dotfiles))
	}
	for i, expected := range expectedDotfiles {
		if i >= len(config.Dotfiles) || config.Dotfiles[i] != expected {
			t.Errorf("Expected dotfile '%s', got '%s'", expected, config.Dotfiles[i])
		}
	}

	// Verify homebrew packages.
	if len(config.Homebrew.Brews) != 3 {
		t.Errorf("Expected 3 homebrew brews, got %d", len(config.Homebrew.Brews))
	}

	// Check simple brew.
	if config.Homebrew.Brews[0].Name != "aichat" {
		t.Errorf("Expected first brew 'aichat', got '%s'", config.Homebrew.Brews[0].Name)
	}

	// Check brew with config.
	neovim := config.Homebrew.Brews[2]
	if neovim.Name != "neovim" {
		t.Errorf("Expected neovim name 'neovim', got '%s'", neovim.Name)
	}
	if neovim.Config != "config/nvim/" {
		t.Errorf("Expected neovim config 'config/nvim/', got '%s'", neovim.Config)
	}

	// Verify asdf tools.
	if len(config.ASDF) != 2 {
		t.Errorf("Expected 2 asdf tools, got %d", len(config.ASDF))
	}

	nodejs := config.ASDF[0]
	if nodejs.Name != "nodejs" || nodejs.Version != "24.2.0" || nodejs.Config != "config/npm/" {
		t.Errorf("nodejs tool not parsed correctly: %+v", nodejs)
	}

	python := config.ASDF[1]
	if python.Name != "python" || python.Version != "3.13.2" || python.Config != "" {
		t.Errorf("python tool not parsed correctly: %+v", python)
	}

	// Verify npm packages.
	if len(config.NPM) != 2 {
		t.Errorf("Expected 2 npm packages, got %d", len(config.NPM))
	}

	claudeCode := config.NPM[0]
	if claudeCode.Name != "@anthropic-ai/claude-code" {
		t.Errorf("Expected claude-code name '@anthropic-ai/claude-code', got '%s'", claudeCode.Name)
	}

	someTool := config.NPM[1]
	if someTool.Name != "some-tool" || someTool.Package != "@scope/different-name" {
		t.Errorf("some-tool not parsed correctly: %+v", someTool)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	_, err := LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Test ASDF tool without version should fail.
	configContent := `settings:
  default_manager: homebrew

asdf:
  - name: nodejs
    # Missing version for ASDF tool
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for ASDF tool without version")
	}
}

func TestSourceToTarget(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"zshrc", "~/.zshrc"},
		{"zshenv", "~/.zshenv"},
		{"config/nvim/", "~/.config/nvim/"},
		{"config/mcfly/config.yaml", "~/.config/mcfly/config.yaml"},
		{"dot_gitconfig", "~/.gitconfig"},
		{"editorconfig", "~/.editorconfig"},
	}

	for _, test := range tests {
		result := sourceToTarget(test.source)
		if result != test.expected {
			t.Errorf("sourceToTarget(%s) = %s, expected %s", test.source, result, test.expected)
		}
	}
}

func TestGetDotfileTargets(t *testing.T) {
	config := &Config{
		Dotfiles: []string{"zshrc", "config/nvim/", "dot_gitconfig"},
	}

	targets := config.GetDotfileTargets()

	expected := map[string]string{
		"zshrc":         "~/.zshrc",
		"config/nvim/":  "~/.config/nvim/",
		"dot_gitconfig": "~/.gitconfig",
	}

	for source, expectedTarget := range expected {
		if target, exists := targets[source]; !exists {
			t.Errorf("Missing target for source %s", source)
		} else if target != expectedTarget {
			t.Errorf("Target for %s = %s, expected %s", source, target, expectedTarget)
		}
	}
}

func TestZSHConfig_BasicParsing(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `settings:
  default_manager: homebrew

zsh:
  env_vars:
    EDITOR: nvim
    LANG: en_US.UTF-8
    FZF_DEFAULT_COMMAND: 'fd --type f --hidden --follow --exclude .git'
  
  shell_options:
    - AUTO_MENU
    - COMPLETE_IN_WORD
    
  inits:
    - 'eval "$(starship init zsh)"'
    - 'eval "$(some-custom-tool init zsh)"'
  
  completions:
    - 'source <(kubectl completion zsh)'
    
  plugins:
    - zsh-users/zsh-syntax-highlighting
    - zsh-users/zsh-autosuggestions
    
  aliases:
    '..': 'cd ..'
    ll: "eza -la --icons --group-directories-first"
    vim: nvim
    
  functions:
    y: |
      local tmp="$(mktemp -t "yazi-cwd.XXXXXX")" cwd
      yazi "$@" --cwd-file="$tmp"
      if cwd="$(command cat -- "$tmp")" && [ -n "$cwd" ] && [ "$cwd" != "$PWD" ]; then
        builtin cd -- "$cwd"
      fi
      rm -f -- "$tmp"
      
  source_after:
    - "$HOME/.zshrc.local"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test env vars.
	if config.ZSH.EnvVars["EDITOR"] != "nvim" {
		t.Errorf("Expected EDITOR=nvim, got %s", config.ZSH.EnvVars["EDITOR"])
	}

	if config.ZSH.EnvVars["LANG"] != "en_US.UTF-8" {
		t.Errorf("Expected LANG=en_US.UTF-8, got %s", config.ZSH.EnvVars["LANG"])
	}

	// Test shell options.
	expectedOptions := []string{"AUTO_MENU", "COMPLETE_IN_WORD"}
	if len(config.ZSH.ShellOptions) != len(expectedOptions) {
		t.Fatalf("Expected %d shell options, got %d", len(expectedOptions), len(config.ZSH.ShellOptions))
	}

	for i, option := range expectedOptions {
		if config.ZSH.ShellOptions[i] != option {
			t.Errorf("Expected shell option %s, got %s", option, config.ZSH.ShellOptions[i])
		}
	}

	// Test inits.
	expectedInits := []string{
		`eval "$(starship init zsh)"`,
		`eval "$(some-custom-tool init zsh)"`,
	}
	if len(config.ZSH.Inits) != len(expectedInits) {
		t.Fatalf("Expected %d inits, got %d", len(expectedInits), len(config.ZSH.Inits))
	}

	for i, expectedInit := range expectedInits {
		if config.ZSH.Inits[i] != expectedInit {
			t.Errorf("Expected init %d to be '%s', got '%s'", i, expectedInit, config.ZSH.Inits[i])
		}
	}

	// Test completions.
	expectedCompletions := []string{`source <(kubectl completion zsh)`}
	if len(config.ZSH.Completions) != len(expectedCompletions) {
		t.Fatalf("Expected %d completions, got %d", len(expectedCompletions), len(config.ZSH.Completions))
	}

	if config.ZSH.Completions[0] != expectedCompletions[0] {
		t.Errorf("Expected completion to be '%s', got '%s'", expectedCompletions[0], config.ZSH.Completions[0])
	}

	// Test plugins.
	expectedPlugins := []string{"zsh-users/zsh-syntax-highlighting", "zsh-users/zsh-autosuggestions"}
	if len(config.ZSH.Plugins) != len(expectedPlugins) {
		t.Fatalf("Expected %d plugins, got %d", len(expectedPlugins), len(config.ZSH.Plugins))
	}

	for i, plugin := range expectedPlugins {
		if config.ZSH.Plugins[i] != plugin {
			t.Errorf("Expected plugin %s, got %s", plugin, config.ZSH.Plugins[i])
		}
	}

	// Test aliases.
	if config.ZSH.Aliases[".."] != "cd .." {
		t.Errorf("Expected alias .. = 'cd ..', got %s", config.ZSH.Aliases[".."])
	}

	if config.ZSH.Aliases["vim"] != "nvim" {
		t.Errorf("Expected alias vim = nvim, got %s", config.ZSH.Aliases["vim"])
	}

	// Test functions.
	if _, exists := config.ZSH.Functions["y"]; !exists {
		t.Error("Expected function 'y' to exist")
	}

	// Test source after.
	if len(config.ZSH.SourceAfter) != 1 {
		t.Fatalf("Expected 1 source_after entry, got %d", len(config.ZSH.SourceAfter))
	}

	if config.ZSH.SourceAfter[0] != "$HOME/.zshrc.local" {
		t.Errorf("Expected source_after '$HOME/.zshrc.local', got %s", config.ZSH.SourceAfter[0])
	}
}

func TestZSHConfig_EmptyConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `settings:
  default_manager: homebrew

# No ZSH configuration section
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that ZSH config is empty but not nil.
	if config.ZSH.EnvVars == nil {
		config.ZSH.EnvVars = make(map[string]string)
	}

	if len(config.ZSH.EnvVars) != 0 {
		t.Errorf("Expected empty env vars, got %d entries", len(config.ZSH.EnvVars))
	}

	if len(config.ZSH.Plugins) != 0 {
		t.Errorf("Expected empty plugins, got %d entries", len(config.ZSH.Plugins))
	}
}

func TestZSHConfig_MinimalConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `settings:
  default_manager: homebrew

zsh:
  plugins:
    - zsh-users/zsh-syntax-highlighting
  aliases:
    vim: nvim
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test minimal config.
	if len(config.ZSH.Plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(config.ZSH.Plugins))
	}

	if config.ZSH.Plugins[0] != "zsh-users/zsh-syntax-highlighting" {
		t.Errorf("Expected plugin 'zsh-users/zsh-syntax-highlighting', got %s", config.ZSH.Plugins[0])
	}

	if config.ZSH.Aliases["vim"] != "nvim" {
		t.Errorf("Expected alias vim = nvim, got %s", config.ZSH.Aliases["vim"])
	}

	// Test defaults (empty init and completions lists).
	if len(config.ZSH.Inits) != 0 {
		t.Errorf("Expected no inits by default, got %d", len(config.ZSH.Inits))
	}

	if len(config.ZSH.Completions) != 0 {
		t.Errorf("Expected no completions by default, got %d", len(config.ZSH.Completions))
	}
}
