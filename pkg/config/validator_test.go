// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

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
		name    string
		yaml    string
		wantErr string
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

func TestValidatePackageNames_ValidNames(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
	}{
		{
			name:        "simple package name",
			packageName: "git",
		},
		{
			name:        "package with hyphen",
			packageName: "font-hack-nerd-font",
		},
		{
			name:        "scoped npm package",
			packageName: "@anthropic-ai/claude-code",
		},
		{
			name:        "package with numbers",
			packageName: "node16",
		},
		{
			name:        "package with underscores",
			packageName: "my_package",
		},
		{
			name:        "package with dots",
			packageName: "some.package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.packageName)
			if err != nil {
				t.Errorf("ValidatePackageName() error = %v, expected no error for valid package name", err)
			}
		})
	}
}

func TestValidatePackageNames_InvalidNames(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		wantErr     string
	}{
		{
			name:        "empty package name",
			packageName: "",
			wantErr:     "empty",
		},
		{
			name:        "package with spaces",
			packageName: "my package",
			wantErr:     "invalid",
		},
		{
			name:        "package with special chars",
			packageName: "package!@#$",
			wantErr:     "invalid",
		},
		{
			name:        "package starting with dash",
			packageName: "-invalid",
			wantErr:     "invalid",
		},
		{
			name:        "package ending with dash",
			packageName: "invalid-",
			wantErr:     "invalid",
		},
		{
			name:        "only whitespace",
			packageName: "   ",
			wantErr:     "empty",
		},
		{
			name:        "package with newline",
			packageName: "package\nname",
			wantErr:     "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.packageName)
			if err == nil {
				t.Errorf("ValidatePackageName() error = nil, expected error containing %q", tt.wantErr)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Errorf("ValidatePackageName() error = %v, expected error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigContent_ValidPackageNames(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "valid homebrew packages",
			yaml: `homebrew:
  brews:
    - git
    - neovim
    - name: aider
      config: config/aider/
  casks:
    - font-hack-nerd-font`,
		},
		{
			name: "valid npm packages",
			yaml: `npm:
  - "@anthropic-ai/claude-code"
  - name: some-tool
    package: "@scope/different-name"`,
		},
		{
			name: "valid asdf packages",
			yaml: `asdf:
  - name: nodejs
    version: "24.2.0"
  - name: python
    version: "3.13.2"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigContent([]byte(tt.yaml))
			if err != nil {
				t.Errorf("ValidateConfigContent() error = %v, expected no error for valid config", err)
			}
		})
	}
}

func TestValidateConfigContent_InvalidPackageNames(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "empty package name in homebrew",
			yaml: `homebrew:
  brews:
    - ""
    - git`,
			wantErr: "package name",
		},
		{
			name: "invalid package name with spaces",
			yaml: `homebrew:
  brews:
    - "my package"`,
			wantErr: "package name",
		},
		{
			name: "invalid package name in npm",
			yaml: `npm:
  - "invalid!@#$"`,
			wantErr: "package name",
		},
		{
			name: "empty package name in complex structure",
			yaml: `homebrew:
  brews:
    - name: ""
      config: config/test/`,
			wantErr: "package name",
		},
		{
			name: "invalid package name in asdf",
			yaml: `asdf:
  - name: "node js"
    version: "24.2.0"`,
			wantErr: "package name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigContent([]byte(tt.yaml))
			if err == nil {
				t.Errorf("ValidateConfigContent() error = nil, expected error containing %q", tt.wantErr)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Errorf("ValidateConfigContent() error = %v, expected error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePaths_ValidPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "simple config path",
			path: "config/nvim/",
		},
		{
			name: "nested config path",
			path: "config/apps/neovim/",
		},
		{
			name: "dotfile path",
			path: "dotfiles/zshrc",
		},
		{
			name: "path with underscores",
			path: "config/my_app/",
		},
		{
			name: "path with numbers",
			path: "config/app2/",
		},
		{
			name: "relative path with dots",
			path: "../config/shared/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if err != nil {
				t.Errorf("ValidateFilePath() error = %v, expected no error for valid path", err)
			}
		})
	}
}

func TestValidateFilePaths_InvalidPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: "empty",
		},
		{
			name:    "only whitespace",
			path:    "   ",
			wantErr: "empty",
		},
		{
			name:    "absolute path",
			path:    "/absolute/path",
			wantErr: "absolute",
		},
		{
			name:    "path with spaces",
			path:    "config/my app/",
			wantErr: "invalid",
		},
		{
			name:    "path with special chars",
			path:    "config/app!@#/",
			wantErr: "invalid",
		},
		{
			name:    "path with newline",
			path:    "config/app\n/",
			wantErr: "invalid",
		},
		{
			name:    "path with colon",
			path:    "config:invalid",
			wantErr: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if err == nil {
				t.Errorf("ValidateFilePath() error = nil, expected error containing %q", tt.wantErr)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Errorf("ValidateFilePath() error = %v, expected error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigContent_ValidFilePaths(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "valid homebrew config paths",
			yaml: `homebrew:
  brews:
    - name: neovim
      config: config/nvim/
    - name: aider
      config: config/aider/`,
		},
		{
			name: "valid npm config paths",
			yaml: `npm:
  - name: some-tool
    package: "@scope/tool"
    config: config/npm-tools/`,
		},
		{
			name: "valid asdf config paths",
			yaml: `asdf:
  - name: nodejs
    version: "24.2.0"
    config: config/node/`,
		},
		{
			name: "valid dotfiles",
			yaml: `dotfiles:
  - zshrc
  - vimrc
  - config/git/gitconfig`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigContent([]byte(tt.yaml))
			if err != nil {
				t.Errorf("ValidateConfigContent() error = %v, expected no error for valid config", err)
			}
		})
	}
}

func TestValidateConfigContent_InvalidFilePaths(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "empty config path",
			yaml: `homebrew:
  brews:
    - name: neovim
      config: ""`,
			wantErr: "file path",
		},
		{
			name: "absolute config path",
			yaml: `homebrew:
  brews:
    - name: neovim
      config: "/absolute/path"`,
			wantErr: "file path",
		},
		{
			name: "config path with spaces",
			yaml: `homebrew:
  brews:
    - name: neovim
      config: "config/my app/"`,
			wantErr: "file path",
		},
		{
			name: "invalid dotfile path",
			yaml: `dotfiles:
  - "/absolute/dotfile"`,
			wantErr: "file path",
		},
		{
			name: "config path with special chars",
			yaml: `asdf:
  - name: nodejs
    version: "24.2.0"
    config: "config/node!@#/"`,
			wantErr: "file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigContent([]byte(tt.yaml))
			if err == nil {
				t.Errorf("ValidateConfigContent() error = nil, expected error containing %q", tt.wantErr)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Errorf("ValidateConfigContent() error = %v, expected error containing %q", err, tt.wantErr)
			}
		})
	}
}
