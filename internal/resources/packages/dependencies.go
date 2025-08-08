// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sort"
)

// DependencyResolver handles package manager dependency resolution
type DependencyResolver struct {
	registry *ManagerRegistry
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(registry *ManagerRegistry) *DependencyResolver {
	return &DependencyResolver{registry: registry}
}

// ResolveDependencyOrder performs topological sort to determine installation order
// Returns managers ordered so dependencies are installed before dependents
func (r *DependencyResolver) ResolveDependencyOrder(managers []string) ([]string, error) {
	// Build dependency graph
	graph, err := r.buildDependencyGraph(managers)
	if err != nil {
		return nil, err
	}

	// Perform topological sort
	ordered, err := r.topologicalSort(graph, managers)
	if err != nil {
		return nil, err
	}

	return ordered, nil
}

// buildDependencyGraph creates a dependency graph from the list of managers
func (r *DependencyResolver) buildDependencyGraph(managers []string) (map[string][]string, error) {
	graph := make(map[string][]string)
	allManagers := make(map[string]bool)

	// Add all requested managers to the graph
	for _, mgr := range managers {
		graph[mgr] = []string{}
		allManagers[mgr] = true
	}

	// Build dependency relationships
	for _, mgr := range managers {
		packageManager, err := r.registry.GetManager(mgr)
		if err != nil {
			return nil, fmt.Errorf("unknown package manager '%s': %w", mgr, err)
		}

		dependencies := packageManager.Dependencies()
		for _, dep := range dependencies {
			// Add dependency to graph if not already present
			if !allManagers[dep] {
				graph[dep] = []string{}
				allManagers[dep] = true
			}

			// Add edge: dep -> mgr (dependency relationship)
			graph[dep] = append(graph[dep], mgr)
		}
	}

	return graph, nil
}

// topologicalSort performs Kahn's algorithm for topological sorting
func (r *DependencyResolver) topologicalSort(graph map[string][]string, requestedManagers []string) ([]string, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}

	for _, dependencies := range graph {
		for _, dep := range dependencies {
			inDegree[dep]++
		}
	}

	// Initialize queue with nodes having zero in-degree
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// Sort queue for deterministic results
	sort.Strings(queue)

	var result []string

	// Process queue
	for len(queue) > 0 {
		// Remove node with zero in-degree
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Update in-degrees of dependent nodes
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sort.Strings(queue) // Keep queue sorted
			}
		}
	}

	// Check for cycles
	if len(result) != len(graph) {
		return nil, fmt.Errorf("circular dependency detected in package managers")
	}

	return result, nil
}

// GetAllDependencies returns all managers needed (including transitive dependencies)
func (r *DependencyResolver) GetAllDependencies(managers []string) ([]string, error) {
	allManagers := make(map[string]bool)

	var collectDependencies func(string) error
	collectDependencies = func(mgr string) error {
		if allManagers[mgr] {
			return nil // Already processed
		}

		allManagers[mgr] = true

		packageManager, err := r.registry.GetManager(mgr)
		if err != nil {
			return fmt.Errorf("unknown package manager '%s': %w", mgr, err)
		}

		// Recursively collect dependencies
		for _, dep := range packageManager.Dependencies() {
			if err := collectDependencies(dep); err != nil {
				return err
			}
		}

		return nil
	}

	// Collect all dependencies for requested managers
	for _, mgr := range managers {
		if err := collectDependencies(mgr); err != nil {
			return nil, err
		}
	}

	// Convert to sorted slice
	var result []string
	for mgr := range allManagers {
		result = append(result, mgr)
	}
	sort.Strings(result)

	return result, nil
}
