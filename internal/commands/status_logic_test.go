// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStatusFlags(t *testing.T) {
	tests := []struct {
		name          string
		showUnmanaged bool
		showMissing   bool
		wantErr       bool
		errContains   string
	}{
		{
			name:          "both false is valid",
			showUnmanaged: false,
			showMissing:   false,
			wantErr:       false,
		},
		{
			name:          "unmanaged only is valid",
			showUnmanaged: true,
			showMissing:   false,
			wantErr:       false,
		},
		{
			name:          "missing only is valid",
			showUnmanaged: false,
			showMissing:   true,
			wantErr:       false,
		},
		{
			name:          "both true is invalid",
			showUnmanaged: true,
			showMissing:   true,
			wantErr:       true,
			errContains:   "mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStatusFlags(tt.showUnmanaged, tt.showMissing)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeDisplayFlags(t *testing.T) {
	tests := []struct {
		name         string
		showPackages bool
		showDotfiles bool
		wantPackages bool
		wantDotfiles bool
	}{
		{
			name:         "both false returns both true",
			showPackages: false,
			showDotfiles: false,
			wantPackages: true,
			wantDotfiles: true,
		},
		{
			name:         "packages only",
			showPackages: true,
			showDotfiles: false,
			wantPackages: true,
			wantDotfiles: false,
		},
		{
			name:         "dotfiles only",
			showPackages: false,
			showDotfiles: true,
			wantPackages: false,
			wantDotfiles: true,
		},
		{
			name:         "both true stays both true",
			showPackages: true,
			showDotfiles: true,
			wantPackages: true,
			wantDotfiles: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages, dotfiles := normalizeDisplayFlags(tt.showPackages, tt.showDotfiles)
			assert.Equal(t, tt.wantPackages, packages)
			assert.Equal(t, tt.wantDotfiles, dotfiles)
		})
	}
}
