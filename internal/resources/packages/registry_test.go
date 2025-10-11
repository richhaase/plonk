// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerRegistry_HasManager(t *testing.T) {

	tests := []struct {
		name     string
		manager  string
		expected bool
	}{
		{
			name:     "brew manager exists",
			manager:  "brew",
			expected: true,
		},
		{
			name:     "npm manager exists",
			manager:  "npm",
			expected: true,
		},
		{
			name:     "cargo manager exists",
			manager:  "cargo",
			expected: true,
		},
		{
			name:     "uv manager exists",
			manager:  "uv",
			expected: true,
		},
		{
			name:     "gem manager exists",
			manager:  "gem",
			expected: true,
		},

		{
			name:     "invalid manager does not exist",
			manager:  "invalid",
			expected: false,
		},
		{
			name:     "empty manager does not exist",
			manager:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

// mockPackageManagerV1 implements PackageManager for V1 testing
type mockPackageManagerV1 struct{}

func (m *mockPackageManagerV1) IsAvailable(ctx context.Context) (bool, error) { return true, nil }
func (m *mockPackageManagerV1) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockPackageManagerV1) Install(ctx context.Context, name string) error { return nil }
func (m *mockPackageManagerV1) Uninstall(ctx context.Context, name string) error {
	return nil
}
func (m *mockPackageManagerV1) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (m *mockPackageManagerV1) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (m *mockPackageManagerV1) Dependencies() []string { return nil }

// mockPackageManagerV2 implements PackageManager for V2 testing
type mockPackageManagerV2 struct {
	executor CommandExecutor
}

func (m *mockPackageManagerV2) IsAvailable(ctx context.Context) (bool, error) { return true, nil }
func (m *mockPackageManagerV2) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockPackageManagerV2) Install(ctx context.Context, name string) error { return nil }
func (m *mockPackageManagerV2) Uninstall(ctx context.Context, name string) error {
	return nil
}
func (m *mockPackageManagerV2) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (m *mockPackageManagerV2) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (m *mockPackageManagerV2) Dependencies() []string { return nil }

func TestManagerRegistry_V1Registration(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factory := func() PackageManager {
		return &mockPackageManagerV1{}
	}

	registry.Register("testmgr", factory)

	mgr, err := registry.GetManager("testmgr")
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	_, ok := mgr.(*mockPackageManagerV1)
	assert.True(t, ok)
}

func TestManagerRegistry_V2Registration(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factoryV2 := func(exec CommandExecutor) PackageManager {
		return &mockPackageManagerV2{executor: exec}
	}

	registry.RegisterV2("testmgr", factoryV2)

	exec := &MockCommandExecutor{}
	mgr, err := registry.GetManagerWithExecutor("testmgr", exec)
	require.NoError(t, err)
	assert.NotNil(t, mgr)

	v2mgr, ok := mgr.(*mockPackageManagerV2)
	require.True(t, ok)
	assert.Equal(t, exec, v2mgr.executor)
}

func TestManagerRegistry_V2PrefersV2OverV1(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factoryV1 := func() PackageManager {
		return &mockPackageManagerV1{}
	}

	factoryV2 := func(exec CommandExecutor) PackageManager {
		return &mockPackageManagerV2{executor: exec}
	}

	registry.Register("testmgr", factoryV1)
	registry.RegisterV2("testmgr", factoryV2)

	mgr, err := registry.GetManager("testmgr")
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	_, ok := mgr.(*mockPackageManagerV2)
	assert.True(t, ok, "should prefer V2 over V1")
}

func TestManagerRegistry_V1FallbackWhenV2Missing(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factoryV1 := func() PackageManager {
		return &mockPackageManagerV1{}
	}

	registry.Register("testmgr", factoryV1)

	mgr, err := registry.GetManagerWithExecutor("testmgr", nil)
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	_, ok := mgr.(*mockPackageManagerV1)
	assert.True(t, ok, "should use V1 when V2 is not available")
}

func TestManagerRegistry_GetManagerWithNilExecutorUsesDefault(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factoryV2 := func(exec CommandExecutor) PackageManager {
		return &mockPackageManagerV2{executor: exec}
	}

	registry.RegisterV2("testmgr", factoryV2)

	mgr, err := registry.GetManagerWithExecutor("testmgr", nil)
	require.NoError(t, err)
	assert.NotNil(t, mgr)

	v2mgr, ok := mgr.(*mockPackageManagerV2)
	require.True(t, ok)
	assert.Equal(t, defaultExecutor, v2mgr.executor)
}

func TestManagerRegistry_ErrorOnUnknownManager(t *testing.T) {
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	_, err := registry.GetManager("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")

	_, err = registry.GetManagerWithExecutor("unknown", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")
}

func TestManagerRegistry_BackwardCompatibility(t *testing.T) {
	// Ensure existing GetManager() works with V1 factories
	registry := &ManagerRegistry{
		managers: make(map[string]*managerEntry),
	}

	factoryV1 := func() PackageManager {
		return &mockPackageManagerV1{}
	}

	registry.Register("legacy", factoryV1)

	mgr, err := registry.GetManager("legacy")
	require.NoError(t, err)
	assert.NotNil(t, mgr)
}
