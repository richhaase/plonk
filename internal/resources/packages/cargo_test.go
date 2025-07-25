// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"testing"
)

func TestCargoManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard cargo install --list output",
			output: []byte(`ripgrep v13.0.0:
    rg
serde_json v1.0.91:
    json_pretty
tokio v1.24.1:
    tokio-util`),
			want: []string{"ripgrep", "serde_json", "tokio"},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single package",
			output: []byte(`ripgrep v13.0.0:
    rg`),
			want: []string{"ripgrep"},
		},
		{
			name: "packages with underscores and hyphens",
			output: []byte(`serde_json v1.0.91:
    json_pretty
tokio-util v0.7.4:
    util-tool
my-custom-tool v1.0.0:
    custom`),
			want: []string{"serde_json", "tokio-util", "my-custom-tool"},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  ripgrep v13.0.0:
    rg
  tokio v1.24.1:
    tokio-util`),
			want: []string{"ripgrep", "tokio"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseListOutput(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCargoManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard cargo search output",
			output: []byte(`serde = "1.0.152"    # A generic serialization/deserialization framework
serde_json = "1.0.91"    # A JSON serialization file format
serde_derive = "1.0.152"    # Macros 1.1 implementation of #[derive(Serialize, Deserialize)]`),
			want: []string{"serde", "serde_json", "serde_derive"},
		},
		{
			name:   "no results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single result",
			output: []byte(`ripgrep = "13.0.0"    # ripgrep recursively searches directories for a regex pattern while respecting your gitignore`),
			want:   []string{"ripgrep"},
		},
		{
			name: "packages with hyphens",
			output: []byte(`tokio-util = "0.7.4"    # Additional utilities for working with Tokio
my-tool = "1.0.0"    # A custom tool`),
			want: []string{"tokio-util", "my-tool"},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  serde = "1.0.152"    # A generic framework
  tokio = "1.24.1"    # An event-driven framework  `),
			want: []string{"serde", "tokio"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseSearchOutput(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCargoManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard cargo search output for single package",
			output: []byte(`serde = "1.0.152"    # A generic serialization/deserialization framework
serde_json = "1.0.91"    # A JSON serialization file format`),
			packageName: "serde",
			want: &PackageInfo{
				Name:        "serde",
				Version:     "1.0.152",
				Description: "A generic serialization/deserialization framework",
				Homepage:    "",
			},
		},
		{
			name:        "exact match for requested package",
			output:      []byte(`ripgrep = "13.0.0"    # ripgrep recursively searches directories for a regex pattern while respecting your gitignore`),
			packageName: "ripgrep",
			want: &PackageInfo{
				Name:        "ripgrep",
				Version:     "13.0.0",
				Description: "ripgrep recursively searches directories for a regex pattern while respecting your gitignore",
				Homepage:    "",
			},
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "unknown",
			want:        nil,
		},
		{
			name:        "package with underscores",
			output:      []byte(`tokio_util = "0.7.4"    # Additional utilities for working with Tokio`),
			packageName: "tokio_util",
			want: &PackageInfo{
				Name:        "tokio_util",
				Version:     "0.7.4",
				Description: "Additional utilities for working with Tokio",
				Homepage:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
