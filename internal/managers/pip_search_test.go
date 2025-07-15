// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestPipManager_Search(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name:  "successful search",
			query: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "search", "requests").
					Return([]byte("requests (2.28.1) - Python HTTP for Humans.\nrequests-oauthlib (1.3.1) - OAuthlib authentication support for Requests.\nrequests-mock (1.9.3) - Mock out responses from the requests package"), nil)
			},
			want:    []string{"requests", "requests-oauthlib", "requests-mock"},
			wantErr: false,
		},
		{
			name:  "empty search results",
			query: "nonexistent-package-xyz",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "search", "nonexistent-package-xyz").
					Return([]byte(""), nil)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "XMLRPC API disabled error",
			query: "test",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "search", "test").
					Return([]byte("ERROR: XMLRPC request failed [code: -32500]\nRuntimeError: PyPI's XMLRPC API is currently disabled due to unmanageable load and will be deprecated in the near future."), &ExitError{Code: 1})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "command execution error",
			query: "test",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "search", "test").
					Return(nil, &ExitError{Code: 1})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.Search(ctx, tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalStringSlices(got, tt.want) {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &PipManager{BaseManager: &BaseManager{}}
			got := manager.parseSearchOutput(tt.output)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
