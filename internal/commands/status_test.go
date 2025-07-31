// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestSortItems(t *testing.T) {
	tests := []struct {
		name     string
		input    []resources.Item
		expected []resources.Item
	}{
		{
			name: "sorts by name case-insensitive",
			input: []resources.Item{
				{Name: "Zebra"},
				{Name: "apple"},
				{Name: "Banana"},
				{Name: "cherry"},
			},
			expected: []resources.Item{
				{Name: "apple"},
				{Name: "Banana"},
				{Name: "cherry"},
				{Name: "Zebra"},
			},
		},
		{
			name:     "handles empty slice",
			input:    []resources.Item{},
			expected: []resources.Item{},
		},
		{
			name: "handles single item",
			input: []resources.Item{
				{Name: "single"},
			},
			expected: []resources.Item{
				{Name: "single"},
			},
		},
		{
			name: "preserves other fields during sort",
			input: []resources.Item{
				{Name: "z-package", Manager: "brew", State: resources.StateManaged},
				{Name: "a-package", Manager: "npm", State: resources.StateMissing},
			},
			expected: []resources.Item{
				{Name: "a-package", Manager: "npm", State: resources.StateMissing},
				{Name: "z-package", Manager: "brew", State: resources.StateManaged},
			},
		},
		{
			name: "handles special characters",
			input: []resources.Item{
				{Name: "@scope/package"},
				{Name: "regular-package"},
				{Name: "_underscore"},
				{Name: "123-numbers"},
			},
			expected: []resources.Item{
				{Name: "123-numbers"},
				{Name: "@scope/package"},
				{Name: "_underscore"},
				{Name: "regular-package"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			items := make([]resources.Item, len(tt.input))
			copy(items, tt.input)

			sortItems(items)

			assert.Equal(t, tt.expected, items)
		})
	}
}

func TestSortItemsByManager(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]resources.Item
		expected []string
	}{
		{
			name: "sorts managers alphabetically",
			input: map[string][]resources.Item{
				"npm":  {{Name: "typescript"}, {Name: "react"}},
				"brew": {{Name: "git"}, {Name: "vim"}},
				"gem":  {{Name: "rails"}},
			},
			expected: []string{"brew", "gem", "npm"},
		},
		{
			name:     "handles empty map",
			input:    map[string][]resources.Item{},
			expected: []string{},
		},
		{
			name: "handles single manager",
			input: map[string][]resources.Item{
				"cargo": {{Name: "ripgrep"}},
			},
			expected: []string{"cargo"},
		},
		{
			name: "handles go before npm alphabetically",
			input: map[string][]resources.Item{
				"pip":  {{Name: "black"}},
				"go":   {{Name: "gopls"}},
				"npm":  {{Name: "prettier"}},
				"brew": {{Name: "jq"}},
			},
			expected: []string{"brew", "go", "npm", "pip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortItemsByManager(tt.input)
			assert.Equal(t, tt.expected, result)

			// Also verify that items within each manager are sorted
			for _, manager := range result {
				items := tt.input[manager]
				// Check if items are sorted by name
				for i := 1; i < len(items); i++ {
					assert.LessOrEqual(t, items[i-1].Name, items[i].Name,
						"Items within manager %s should be sorted", manager)
				}
			}
		})
	}
}

func TestSortItemsByManager_SortsItemsWithinManager(t *testing.T) {
	input := map[string][]resources.Item{
		"brew": {
			{Name: "zsh"},
			{Name: "git"},
			{Name: "vim"},
			{Name: "bat"},
		},
	}

	managers := sortItemsByManager(input)

	assert.Equal(t, []string{"brew"}, managers)

	// Verify items are sorted within the manager
	brewItems := input["brew"]
	expectedOrder := []string{"bat", "git", "vim", "zsh"}

	for i, item := range brewItems {
		assert.Equal(t, expectedOrder[i], item.Name)
	}
}
