// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "github shorthand",
			input:    "user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "github shorthand with hyphen",
			input:    "my-user/my-repo",
			expected: "https://github.com/my-user/my-repo.git",
		},
		{
			name:     "github shorthand with dots",
			input:    "user.name/repo.name",
			expected: "https://github.com/user.name/repo.name.git",
		},
		{
			name:     "https url without .git",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "https url with .git",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "ssh url",
			input:    "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "git protocol",
			input:    "git://github.com/user/repo.git",
			expected: "git://github.com/user/repo.git",
		},
		{
			name:     "gitlab https url",
			input:    "https://gitlab.com/user/repo",
			expected: "https://gitlab.com/user/repo.git",
		},
		{
			name:     "bitbucket ssh url",
			input:    "git@bitbucket.org:user/repo.git",
			expected: "git@bitbucket.org:user/repo.git",
		},
		{
			name:      "empty url",
			input:     "",
			expectErr: true,
		},
		{
			name:      "whitespace only",
			input:     "   ",
			expectErr: true,
		},
		{
			name:     "url with surrounding whitespace",
			input:    "  user/repo  ",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:      "invalid format",
			input:     "not-a-url",
			expectErr: true,
		},
		{
			name:      "http url (not https)",
			input:     "http://github.com/user/repo",
			expectErr: true,
		},
		{
			name:      "ftp url",
			input:     "ftp://github.com/user/repo",
			expectErr: true,
		},
		{
			name:      "github shorthand with too many slashes",
			input:     "user/repo/extra",
			expectErr: true,
		},
		{
			name:      "github shorthand with special chars",
			input:     "user@repo/test",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitURL(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "git URL")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test error messages for clarity
func TestParseGitURLErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrMsg string
	}{
		{
			name:           "empty url error",
			input:          "",
			expectedErrMsg: "empty git URL",
		},
		{
			name:           "unsupported format error",
			input:          "not-a-url",
			expectedErrMsg: "unsupported git URL format",
		},
		{
			name:           "unsupported format with details",
			input:          "http://example.com",
			expectedErrMsg: "supported: user/repo, https://..., git@...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGitURL(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}
