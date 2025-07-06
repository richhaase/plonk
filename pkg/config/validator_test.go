package config

import (
	"strings"
	"testing"
)

func TestValidateYAML_ValidSyntax(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "minimal valid config",
			yaml: `settings:
  default_manager: homebrew`,
		},
		{
			name: "complete valid config",
			yaml: `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
  - vimrc

homebrew:
  brews:
    - name: neovim
      config: config/nvim/
    - git
  casks:
    - font-hack-nerd-font

asdf:
  - name: nodejs
    version: "24.2.0"`,
		},
		{
			name: "config with all sections",
			yaml: `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5

dotfiles:
  - zshrc

homebrew:
  brews:
    - git

asdf:
  - name: python
    version: "3.13.2"

npm:
  - "@vue/cli"

zsh:
  aliases:
    ll: "ls -la"

git:
  user:
    name: "Test User"
    email: "test@example.com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAML([]byte(tt.yaml))
			if err != nil {
				t.Errorf("ValidateYAML() error = %v, expected no error for valid YAML", err)
			}
		})
	}
}

func TestValidateYAML_InvalidSyntax(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     string
	}{
		{
			name: "invalid indentation",
			yaml: `settings:
  default_manager: homebrew
    extra_indent: value`,
			wantErr: "yaml",
		},
		{
			name: "unclosed quote",
			yaml: `settings:
  default_manager: "homebrew`,
			wantErr: "yaml",
		},
		{
			name: "invalid list format",
			yaml: `dotfiles:
  - zshrc
  vimrc`,
			wantErr: "yaml",
		},
		{
			name: "duplicate keys",
			yaml: `settings:
  default_manager: homebrew
  default_manager: asdf`,
			wantErr: "duplicate",
		},
		{
			name: "tabs instead of spaces",
			yaml: `settings:
	default_manager: homebrew`,
			wantErr: "yaml",
		},
		{
			name: "invalid nesting",
			yaml: `homebrew:
  brews:
    - name: git
      - config: test`,
			wantErr: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAML([]byte(tt.yaml))
			if err == nil {
				t.Errorf("ValidateYAML() error = nil, expected error containing %q", tt.wantErr)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Errorf("ValidateYAML() error = %v, expected error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateYAML_EmptyConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "empty string",
			yaml:    "",
			wantErr: false, // Empty is technically valid YAML
		},
		{
			name:    "only comments",
			yaml:    "# This is a comment\n# Another comment",
			wantErr: false,
		},
		{
			name:    "only whitespace",
			yaml:    "   \n\n   ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAML([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}