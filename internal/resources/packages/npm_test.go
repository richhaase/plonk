// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestNpmManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard npm list output",
			output: []byte(`{
  "dependencies": {
    "typescript": {
      "version": "4.9.4"
    },
    "eslint": {
      "version": "8.30.0"
    },
    "prettier": {
      "version": "2.8.1"
    }
  }
}`),
			want: []string{"eslint", "prettier", "typescript"},
		},
		{
			name:   "empty dependencies",
			output: []byte(`{"dependencies": {}}`),
			want:   []string{},
		},
		{
			name:   "no dependencies field",
			output: []byte(`{}`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single package",
			output: []byte(`{
  "dependencies": {
    "lodash": {
      "version": "4.17.21"
    }
  }
}`),
			want: []string{"lodash"},
		},
		{
			name: "packages with scoped names",
			output: []byte(`{
  "dependencies": {
    "@types/node": {
      "version": "18.11.18"
    },
    "@typescript-eslint/parser": {
      "version": "5.48.0"
    },
    "react": {
      "version": "18.2.0"
    }
  }
}`),
			want: []string{"@types/node", "@typescript-eslint/parser", "react"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNpmManager()
			got := manager.parseListOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestNpmManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard npm search output",
			output: []byte(`[
  {
    "name": "react",
    "version": "18.2.0",
    "description": "React is a JavaScript library for building user interfaces."
  },
  {
    "name": "react-dom",
    "version": "18.2.0",
    "description": "React package for working with the DOM."
  }
]`),
			want: []string{"react", "react-dom"},
		},
		{
			name:   "empty search results",
			output: []byte(`[]`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single result",
			output: []byte(`[
  {
    "name": "lodash",
    "version": "4.17.21",
    "description": "Lodash modular utilities."
  }
]`),
			want: []string{"lodash"},
		},
		{
			name: "scoped packages",
			output: []byte(`[
  {
    "name": "@types/node",
    "version": "18.11.18",
    "description": "TypeScript definitions for Node.js"
  },
  {
    "name": "@typescript-eslint/parser",
    "version": "5.48.0",
    "description": "An ESLint custom parser for TypeScript"
  }
]`),
			want: []string{"@types/node", "@typescript-eslint/parser"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNpmManager()
			got := manager.parseSearchOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestNpmManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard npm view output",
			output: []byte(`{
  "name": "react",
  "version": "18.2.0",
  "description": "React is a JavaScript library for building user interfaces.",
  "homepage": "https://reactjs.org/"
}`),
			packageName: "react",
			want: &PackageInfo{
				Name:        "react",
				Version:     "18.2.0",
				Description: "React is a JavaScript library for building user interfaces.",
				Homepage:    "https://reactjs.org/",
			},
		},
		{
			name: "minimal npm view output",
			output: []byte(`{
  "name": "lodash",
  "version": "4.17.21"
}`),
			packageName: "lodash",
			want: &PackageInfo{
				Name:        "lodash",
				Version:     "4.17.21",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name: "scoped package",
			output: []byte(`{
  "name": "@types/node",
  "version": "18.11.18",
  "description": "TypeScript definitions for Node.js",
  "homepage": "https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node"
}`),
			packageName: "@types/node",
			want: &PackageInfo{
				Name:        "@types/node",
				Version:     "18.11.18",
				Description: "TypeScript definitions for Node.js",
				Homepage:    "https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node",
			},
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "unknown",
			want: &PackageInfo{
				Name:        "unknown",
				Version:     "",
				Description: "",
				Homepage:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNpmManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
