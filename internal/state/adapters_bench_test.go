// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/interfaces"
)

// mockDirectManager implements PackageManager directly
type mockDirectManager struct{}

func (m *mockDirectManager) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}
func (m *mockDirectManager) ListInstalled(ctx context.Context) ([]string, error) {
	return []string{"pkg1", "pkg2", "pkg3"}, nil
}
func (m *mockDirectManager) Install(ctx context.Context, name string) error {
	return nil
}
func (m *mockDirectManager) Uninstall(ctx context.Context, name string) error {
	return nil
}
func (m *mockDirectManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (m *mockDirectManager) Search(ctx context.Context, query string) ([]string, error) {
	return []string{"result1", "result2"}, nil
}
func (m *mockDirectManager) Info(ctx context.Context, name string) (*interfaces.PackageInfo, error) {
	return &interfaces.PackageInfo{
		Name:        name,
		Version:     "1.0.0",
		Description: "Test package",
	}, nil
}
func (m *mockDirectManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}

// BenchmarkDirectCall tests calling the manager directly
func BenchmarkDirectCall(b *testing.B) {
	ctx := context.Background()
	manager := &mockDirectManager{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.IsAvailable(ctx)
		_, _ = manager.ListInstalled(ctx)
		_, _ = manager.IsInstalled(ctx, "test")
		_, _ = manager.GetInstalledVersion(ctx, "test")
	}
}

// BenchmarkAdapterCall tests calling through the adapter
func BenchmarkAdapterCall(b *testing.B) {
	ctx := context.Background()
	manager := &mockDirectManager{}
	adapter := NewManagerAdapter(manager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.IsAvailable(ctx)
		_, _ = adapter.ListInstalled(ctx)
		_, _ = adapter.IsInstalled(ctx, "test")
		_, _ = adapter.GetInstalledVersion(ctx, "test")
	}
}

// mockConfigImpl implements ConfigInterface for benchmarking
type mockConfigImpl struct{}

func (m *mockConfigImpl) GetDotfileTargets() map[string]string {
	return map[string]string{
		"vimrc":  "~/.vimrc",
		"bashrc": "~/.bashrc",
		"zshrc":  "~/.zshrc",
	}
}
func (m *mockConfigImpl) GetHomebrewBrews() []string {
	return []string{"git", "vim", "tmux"}
}
func (m *mockConfigImpl) GetHomebrewCasks() []string {
	return []string{"firefox", "slack"}
}
func (m *mockConfigImpl) GetNPMPackages() []string {
	return []string{"typescript", "prettier", "eslint"}
}
func (m *mockConfigImpl) GetIgnorePatterns() []string {
	return []string{".git", ".DS_Store", "*.tmp"}
}
func (m *mockConfigImpl) GetExpandDirectories() []string {
	return []string{"config"}
}

// BenchmarkConfigAdapter tests the ConfigAdapter performance
func BenchmarkConfigAdapter(b *testing.B) {
	config := &mockConfigImpl{}
	adapter := NewConfigAdapter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.GetDotfileTargets()
		_ = adapter.GetIgnorePatterns()
		_ = adapter.GetExpandDirectories()
		_, _ = adapter.GetPackagesForManager("homebrew")
		_, _ = adapter.GetPackagesForManager("npm")
	}
}
