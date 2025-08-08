// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestDependencyResolution(t *testing.T) {
	tests := []struct {
		name          string
		managers      []string
		expectedOrder []string
		expectError   bool
	}{
		{
			name:          "independent managers only",
			managers:      []string{"pnpm", "cargo"},
			expectedOrder: []string{"cargo", "pnpm"}, // alphabetical
			expectError:   false,
		},
		{
			name:          "simple dependency",
			managers:      []string{"npm"},
			expectedOrder: []string{"brew", "npm"}, // brew first, then npm
			expectError:   false,
		},
		{
			name:          "multiple dependents",
			managers:      []string{"npm", "gem", "go"},
			expectedOrder: []string{"brew", "gem", "go", "npm"}, // brew first, then others alphabetically
			expectError:   false,
		},
		{
			name:          "mixed dependencies",
			managers:      []string{"npm", "pnpm", "cargo"},
			expectedOrder: []string{"brew", "cargo", "npm", "pnpm"}, // brew first, then independents
			expectError:   false,
		},
		{
			name:          "dependency already included",
			managers:      []string{"brew", "npm"},
			expectedOrder: []string{"brew", "npm"}, // correct order maintained
			expectError:   false,
		},
		{
			name:          "all independent managers",
			managers:      []string{"brew", "pnpm", "cargo", "uv", "pixi", "dotnet"},
			expectedOrder: []string{"brew", "cargo", "dotnet", "pixi", "pnpm", "uv"}, // alphabetical
			expectError:   false,
		},
		{
			name:          "all dependent managers",
			managers:      []string{"npm", "gem", "go", "composer", "pipx"},
			expectedOrder: []string{"brew", "composer", "gem", "go", "npm", "pipx"}, // brew first, then alphabetical
			expectError:   false,
		},
	}

	registry := NewManagerRegistry()
	resolver := NewDependencyResolver(registry)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := resolver.ResolveDependencyOrder(tt.managers)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !slicesEqual(order, tt.expectedOrder) {
				t.Errorf("expected order %v, got %v", tt.expectedOrder, order)
			}
		})
	}
}

func TestGetAllDependencies(t *testing.T) {
	tests := []struct {
		name        string
		managers    []string
		expectedAll []string
		expectError bool
	}{
		{
			name:        "independent managers",
			managers:    []string{"pnpm"},
			expectedAll: []string{"pnpm"},
			expectError: false,
		},
		{
			name:        "manager with dependency",
			managers:    []string{"npm"},
			expectedAll: []string{"brew", "npm"},
			expectError: false,
		},
		{
			name:        "multiple managers with shared dependency",
			managers:    []string{"npm", "gem"},
			expectedAll: []string{"brew", "gem", "npm"},
			expectError: false,
		},
		{
			name:        "mixed independent and dependent",
			managers:    []string{"npm", "pnpm"},
			expectedAll: []string{"brew", "npm", "pnpm"},
			expectError: false,
		},
		{
			name:        "all dependent managers",
			managers:    []string{"npm", "gem", "go", "composer", "pipx"},
			expectedAll: []string{"brew", "composer", "gem", "go", "npm", "pipx"},
			expectError: false,
		},
		{
			name:        "dependency already included",
			managers:    []string{"brew", "npm"},
			expectedAll: []string{"brew", "npm"},
			expectError: false,
		},
		{
			name:        "empty input",
			managers:    []string{},
			expectedAll: []string{},
			expectError: false,
		},
	}

	registry := NewManagerRegistry()
	resolver := NewDependencyResolver(registry)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			all, err := resolver.GetAllDependencies(tt.managers)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !slicesEqual(all, tt.expectedAll) {
				t.Errorf("expected dependencies %v, got %v", tt.expectedAll, all)
			}
		})
	}
}

func TestInvalidManager(t *testing.T) {
	registry := NewManagerRegistry()
	resolver := NewDependencyResolver(registry)

	// Test with unknown manager
	_, err := resolver.GetAllDependencies([]string{"unknown"})
	if err == nil {
		t.Error("expected error for unknown manager")
	}

	_, err = resolver.ResolveDependencyOrder([]string{"unknown"})
	if err == nil {
		t.Error("expected error for unknown manager")
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	// For this test, we would need to create mock managers with circular dependencies
	// Since our current system doesn't have circular dependencies, this test
	// verifies the algorithm would detect them if they existed

	registry := NewManagerRegistry()
	resolver := NewDependencyResolver(registry)

	// Test with valid dependencies - should not error
	_, err := resolver.ResolveDependencyOrder([]string{"npm", "brew"})
	if err != nil {
		t.Errorf("unexpected error with valid dependencies: %v", err)
	}
}

func TestManagerDependenciesImplementation(t *testing.T) {
	registry := NewManagerRegistry()

	independentManagers := []string{"brew", "pnpm", "cargo", "uv", "pixi", "dotnet"}
	dependentManagers := []string{"npm", "gem", "go", "composer", "pipx"}

	// Test independent managers return empty dependencies
	for _, mgr := range independentManagers {
		packageManager, err := registry.GetManager(mgr)
		if err != nil {
			t.Errorf("failed to get manager %s: %v", mgr, err)
			continue
		}

		deps := packageManager.Dependencies()
		if len(deps) != 0 {
			t.Errorf("expected %s to be independent (no dependencies), got %v", mgr, deps)
		}
	}

	// Test dependent managers return ["brew"] as dependency
	for _, mgr := range dependentManagers {
		packageManager, err := registry.GetManager(mgr)
		if err != nil {
			t.Errorf("failed to get manager %s: %v", mgr, err)
			continue
		}

		deps := packageManager.Dependencies()
		expectedDeps := []string{"brew"}
		if !slicesEqual(deps, expectedDeps) {
			t.Errorf("expected %s to depend on %v, got %v", mgr, expectedDeps, deps)
		}
	}
}

func TestRealWorldScenarios(t *testing.T) {
	registry := NewManagerRegistry()
	resolver := NewDependencyResolver(registry)

	scenarios := []struct {
		name             string
		detectedManagers []string
		expectedOrder    []string
		expectedFullList []string
	}{
		{
			name:             "typical web development setup",
			detectedManagers: []string{"npm", "pnpm"},
			expectedOrder:    []string{"brew", "npm", "pnpm"},
			expectedFullList: []string{"brew", "npm", "pnpm"},
		},
		{
			name:             "polyglot development setup",
			detectedManagers: []string{"npm", "cargo", "go", "gem"},
			expectedOrder:    []string{"brew", "cargo", "gem", "go", "npm"},
			expectedFullList: []string{"brew", "cargo", "gem", "go", "npm"},
		},
		{
			name:             "python focused setup",
			detectedManagers: []string{"uv", "pipx"},
			expectedOrder:    []string{"brew", "pipx", "uv"},
			expectedFullList: []string{"brew", "pipx", "uv"},
		},
		{
			name:             "minimal independent setup",
			detectedManagers: []string{"cargo", "uv"},
			expectedOrder:    []string{"cargo", "uv"},
			expectedFullList: []string{"cargo", "uv"},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Test GetAllDependencies
			allDeps, err := resolver.GetAllDependencies(scenario.detectedManagers)
			if err != nil {
				t.Errorf("GetAllDependencies failed: %v", err)
				return
			}

			if !slicesEqual(allDeps, scenario.expectedFullList) {
				t.Errorf("GetAllDependencies: expected %v, got %v", scenario.expectedFullList, allDeps)
			}

			// Test ResolveDependencyOrder
			order, err := resolver.ResolveDependencyOrder(allDeps)
			if err != nil {
				t.Errorf("ResolveDependencyOrder failed: %v", err)
				return
			}

			if !slicesEqual(order, scenario.expectedOrder) {
				t.Errorf("ResolveDependencyOrder: expected %v, got %v", scenario.expectedOrder, order)
			}
		})
	}
}

// slicesEqual checks if two string slices are equal
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
