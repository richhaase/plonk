// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		config   config.ManagerConfig
		output   string
		expected []string
	}{
		{
			name: "lines parsing",
			config: config.ManagerConfig{
				Binary: "pipx",
				List: config.ListConfig{
					Command: []string{"pipx", "list", "--short"},
					Parse:   "lines",
				},
			},
			output:   "ruff\nblack\nmypy\n",
			expected: []string{"ruff", "black", "mypy"},
		},
		{
			name: "json parsing",
			config: config.ManagerConfig{
				Binary: "conda",
				List: config.ListConfig{
					Command:   []string{"conda", "list", "--json"},
					Parse:     "json",
					JSONField: "name",
				},
			},
			output:   `[{"name":"numpy"},{"name":"pandas"}]`,
			expected: []string{"numpy", "pandas"},
		},
		{
			name: "lines parsing with version",
			config: config.ManagerConfig{
				Binary: "cargo",
				List: config.ListConfig{
					Command: []string{"cargo", "install", "--list"},
					Parse:   "lines",
				},
			},
			output:   "ripgrep v14.0.0:\nfd v8.7.0:\n",
			expected: []string{"ripgrep", "fd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdKey := tt.config.List.Command[0]
			for _, arg := range tt.config.List.Command[1:] {
				cmdKey += " " + arg
			}

			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					cmdKey: {
						Output: []byte(tt.output),
					},
				},
			}

			mgr := NewGenericManager(tt.config, mock)
			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenericManager_Install_Idempotent(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		Install: config.CommandConfig{
			Command: []string{"pipx", "install", "{{.Package}}"},
			IdempotentErrors: []string{
				"already installed",
			},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx install ruff": {
				Output: []byte("Error: 'ruff' is already installed"),
				Error:  &MockExitError{Code: 1},
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Install(context.Background(), "ruff")

	assert.NoError(t, err, "should be idempotent - 'already installed' is success")
}

func TestGenericManager_Upgrade_Idempotent(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		Upgrade: config.CommandConfig{
			Command: []string{"pipx", "upgrade", "{{.Package}}"},
			IdempotentErrors: []string{
				"already up-to-date",
			},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx upgrade ruff": {
				Output: []byte("ruff is already up-to-date"),
				Error:  nil,
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Upgrade(context.Background(), []string{"ruff"})

	assert.NoError(t, err)
}

func TestGenericManager_UpgradeAll(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		UpgradeAll: config.CommandConfig{
			Command: []string{"pipx", "upgrade-all"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx upgrade-all": {
				Output: []byte("upgraded 3 packages"),
				Error:  nil,
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.UpgradeAll(context.Background())

	assert.NoError(t, err)
}

func TestGenericManager_TemplateExpansion(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		Install: config.CommandConfig{
			Command: []string{"pipx", "install", "{{.Package}}"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx install my-tool": {
				Output: []byte("installed my-tool"),
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Install(context.Background(), "my-tool")

	require.NoError(t, err)
	assert.Len(t, mock.Commands, 1)
	assert.Equal(t, "pipx", mock.Commands[0].Name)
	assert.Equal(t, []string{"install", "my-tool"}, mock.Commands[0].Args)
}

func TestGenericManager_IsAvailable_LookPathError(t *testing.T) {
	cfg := config.ManagerConfig{Binary: "nonexistent"}

	mock := &MockCommandExecutor{
		DefaultResponse: CommandResponse{Error: &MockExitError{Code: 127}},
	}

	mgr := NewGenericManager(cfg, mock)
	available, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err, "should return nil error for unavailable")
	assert.False(t, available)
}

func TestGenericManager_IsAvailable_VersionError(t *testing.T) {
	cfg := config.ManagerConfig{Binary: "broken"}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"broken --version": {Error: &MockExitError{Code: 1}},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	available, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, available)
}

func TestGenericManager_ListInstalled_EmptyCommand(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "test",
		List:   config.ListConfig{Command: []string{}},
	}

	mgr := NewGenericManager(cfg, nil)
	result, err := mgr.ListInstalled(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{}, result)
}

func TestGenericManager_ParseLines_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "leading and trailing blank lines",
			input:    "\n\nruff\nblack\n\n",
			expected: []string{"ruff", "black"},
		},
		{
			name:     "lines with versions",
			input:    "ruff 0.1.0\nblack 23.0.0\n",
			expected: []string{"ruff", "black"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   \n  \n",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &GenericManager{}
			result := mgr.parseLines([]byte(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenericManager_ParseJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		field     string
		expected  []string
		wantError bool
	}{
		{
			name:     "valid json",
			input:    `[{"name":"pkg1"},{"name":"pkg2"}]`,
			field:    "name",
			expected: []string{"pkg1", "pkg2"},
		},
		{
			name:      "invalid json",
			input:     `not json`,
			field:     "name",
			wantError: true,
		},
		{
			name:     "missing field",
			input:    `[{"foo":"bar"}]`,
			field:    "name",
			expected: nil,
		},
		{
			name:     "non-string field",
			input:    `[{"name":123}]`,
			field:    "name",
			expected: nil,
		},
		{
			name:     "empty array",
			input:    `[]`,
			field:    "name",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &GenericManager{}
			result, err := mgr.parseJSON([]byte(tt.input), tt.field)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGenericManager_ParseOutput_UnknownStrategy(t *testing.T) {
	cfg := config.ListConfig{Parse: "unknown"}
	mgr := &GenericManager{}

	_, err := mgr.parseOutput([]byte("test"), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown parse strategy")
}

func TestGenericManager_IdempotentError_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		patterns []string
		expected bool
	}{
		{
			name:     "exact match",
			output:   "already installed",
			patterns: []string{"already installed"},
			expected: true,
		},
		{
			name:     "case insensitive",
			output:   "ALREADY INSTALLED",
			patterns: []string{"already installed"},
			expected: true,
		},
		{
			name:     "substring match",
			output:   "Error: package is already installed on system",
			patterns: []string{"already installed"},
			expected: true,
		},
		{
			name:     "no match",
			output:   "network error",
			patterns: []string{"already installed"},
			expected: false,
		},
		{
			name:     "empty patterns",
			output:   "already installed",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &GenericManager{}
			result := mgr.isIdempotentError(tt.output, tt.patterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenericManager_Install_NonIdempotentError(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		Install: config.CommandConfig{
			Command:          []string{"pipx", "install", "{{.Package}}"},
			IdempotentErrors: []string{"already installed"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx install bad": {
				Output: []byte("Error: package not found"),
				Error:  &MockExitError{Code: 1},
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Install(context.Background(), "bad")

	assert.Error(t, err, "non-idempotent error should propagate")
}

func TestGenericManager_Upgrade_EmptyPackages_CallsUpgradeAll(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		UpgradeAll: config.CommandConfig{
			Command: []string{"pipx", "upgrade-all"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx upgrade-all": {Output: []byte("upgraded all")},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Upgrade(context.Background(), []string{})

	assert.NoError(t, err)
	assert.Len(t, mock.Commands, 1)
	assert.Equal(t, "pipx", mock.Commands[0].Name)
	assert.Equal(t, []string{"upgrade-all"}, mock.Commands[0].Args)
}

func TestGenericManager_UpgradeAll_IdempotentError(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "brew",
		UpgradeAll: config.CommandConfig{
			Command:          []string{"brew", "upgrade"},
			IdempotentErrors: []string{"already up-to-date"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"brew upgrade": {
				Output: []byte("All packages are already up-to-date"),
				Error:  &MockExitError{Code: 1},
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.UpgradeAll(context.Background())

	assert.NoError(t, err, "idempotent error should be suppressed")
}

func TestGenericManager_Uninstall_IdempotentError(t *testing.T) {
	cfg := config.ManagerConfig{
		Binary: "pipx",
		Uninstall: config.CommandConfig{
			Command:          []string{"pipx", "uninstall", "{{.Package}}"},
			IdempotentErrors: []string{"not installed"},
		},
	}

	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"pipx uninstall missing": {
				Output: []byte("Error: 'missing' is not installed"),
				Error:  &MockExitError{Code: 1},
			},
		},
	}

	mgr := NewGenericManager(cfg, mock)
	err := mgr.Uninstall(context.Background(), "missing")

	assert.NoError(t, err, "idempotent uninstall error should be success")
}
