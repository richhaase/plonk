// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"testing"

	managerTesting "github.com/richhaase/plonk/internal/managers/testing"
)

func TestPipManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  []byte
		want    []string
		wantErr bool
	}{
		{
			name:    "JSON format output",
			output:  []byte(`[{"name": "requests", "version": "2.28.1"}, {"name": "numpy", "version": "1.24.0"}]`),
			want:    []string{"requests", "numpy"},
			wantErr: false,
		},
		{
			name:    "empty JSON array",
			output:  []byte(`[]`),
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "malformed JSON falls back to plain text",
			output:  []byte(`requests==2.28.1\nnumpy==1.24.0`),
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "empty output",
			output:  []byte(""),
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "single package JSON",
			output:  []byte(`[{"name": "requests", "version": "2.28.1"}]`),
			want:    []string{"requests"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got, err := manager.parseListOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseListOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_parseListOutputPlainText(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard pip list output",
			output: []byte("requests==2.28.1\nnumpy==1.24.0\npandas==1.5.2"),
			want:   []string{"pandas==1.5.2"},
		},
		{
			name:   "pip list with spaces",
			output: []byte("requests 2.28.1\nnumpy 1.24.0\npandas 1.5.2"),
			want:   []string{"pandas"},
		},
		{
			name:   "mixed separators",
			output: []byte("requests==2.28.1\nnumpy 1.24.0\npandas>=1.5.2"),
			want:   []string{"pandas>=1.5.2"},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single package",
			output: []byte("requests==2.28.1"),
			want:   []string{},
		},
		{
			name:   "packages with underscores and hyphens",
			output: []byte("python-dateutil==2.8.2\nbeautifulsoup4==4.11.1\npython_magic==0.4.27"),
			want:   []string{"python_magic==0.4.27"},
		},
		{
			name:   "output with extra whitespace",
			output: []byte("  requests==2.28.1  \n  numpy==1.24.0  \n"),
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got := manager.parseListOutputPlainText(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseListOutputPlainText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard pip show output",
			output: []byte(`Name: requests
Version: 2.28.1
Summary: Python HTTP for Humans.
Home-page: https://requests.readthedocs.io
Author: Kenneth Reitz
License: Apache 2.0`),
			packageName: "requests",
			want: &PackageInfo{
				Name:        "requests",
				Version:     "2.28.1",
				Description: "Python HTTP for Humans.",
				Homepage:    "https://requests.readthedocs.io",
			},
		},
		{
			name: "minimal pip show output",
			output: []byte(`Name: numpy
Version: 1.24.0`),
			packageName: "numpy",
			want: &PackageInfo{
				Name:        "numpy",
				Version:     "1.24.0",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name: "package with URL field",
			output: []byte(`Name: django
Version: 4.1.4
Summary: A high-level Python Web framework.
URL: https://www.djangoproject.com/`),
			packageName: "django",
			want: &PackageInfo{
				Name:        "django",
				Version:     "4.1.4",
				Description: "A high-level Python Web framework.",
				Homepage:    "",
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
			manager := NewPipManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Shared integration tests using the common test suite
func TestPipManager_SharedTestSuite(t *testing.T) {
	suite := &managerTesting.ManagerTestSuite{
		Manager:     NewPipManager(),
		TestPackage: "requests",
		BinaryName:  "pip",
	}

	t.Run("IsAvailable", suite.TestIsAvailable)
	t.Run("ListInstalled", suite.TestListInstalled)
	t.Run("SupportsSearch", suite.TestSupportsSearch)
}
