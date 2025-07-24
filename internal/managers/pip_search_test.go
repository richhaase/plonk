// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"testing"
)

func TestPipManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard search output",
			output: []byte(`requests (2.28.1) - Python HTTP for Humans.
requests-oauthlib (1.3.1) - OAuthlib authentication support for Requests.
requests-mock (1.9.3) - Mock out responses from the requests package
requests-toolbelt (0.9.1) - A utility belt for advanced users of python-requests`),
			want: []string{"requests", "requests-oauthlib", "requests-mock", "requests-toolbelt"},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "whitespace only",
			output: []byte("   \n  \n   "),
			want:   []string{},
		},
		{
			name: "output with blank lines",
			output: []byte(`requests (2.28.1) - Python HTTP for Humans.

requests-oauthlib (1.3.1) - OAuthlib authentication support for Requests.`),
			want: []string{"requests", "requests-oauthlib"},
		},
		{
			name: "malformed lines",
			output: []byte(`requests (2.28.1) - Python HTTP for Humans.
This is not a valid package line
requests-oauthlib (1.3.1) - OAuthlib authentication support for Requests.
Another invalid line without parentheses`),
			want: []string{"requests", "requests-oauthlib"},
		},
		{
			name: "single package result",
			output: []byte(`requests (2.28.1) - Python HTTP for Humans.`),
			want: []string{"requests"},
		},
		{
			name: "packages with underscores and hyphens",
			output: []byte(`python-dateutil (2.8.2) - Extensions to the standard Python datetime module.
beautifulsoup4 (4.11.1) - Screen-scraping library
python_magic (0.4.27) - File type identification using libmagic`),
			want: []string{"python-dateutil", "beautifulsoup4", "python_magic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &PipManager{BaseManager: &BaseManager{}}
			got := manager.parseSearchOutput(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

