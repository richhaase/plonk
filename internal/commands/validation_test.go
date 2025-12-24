// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"
)

func TestValidateBatchResults(t *testing.T) {
	tests := []struct {
		name          string
		count         int
		operationName string
		isFailed      func(i int) bool
		wantErr       bool
		errContains   string
	}{
		{
			name:          "empty results returns nil",
			count:         0,
			operationName: "test",
			isFailed:      func(i int) bool { return true },
			wantErr:       false,
		},
		{
			name:          "all failed returns error",
			count:         3,
			operationName: "install packages",
			isFailed:      func(i int) bool { return true },
			wantErr:       true,
			errContains:   "install packages operation failed: all 3 item(s) failed to process",
		},
		{
			name:          "some succeeded returns nil",
			count:         3,
			operationName: "test",
			isFailed:      func(i int) bool { return i == 0 }, // Only first failed
			wantErr:       false,
		},
		{
			name:          "none failed returns nil",
			count:         5,
			operationName: "test",
			isFailed:      func(i int) bool { return false },
			wantErr:       false,
		},
		{
			name:          "single item failed returns error",
			count:         1,
			operationName: "add dotfiles",
			isFailed:      func(i int) bool { return true },
			wantErr:       true,
			errContains:   "add dotfiles operation failed: all 1 item(s) failed to process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBatchResults(tt.count, tt.operationName, tt.isFailed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBatchResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errContains {
				t.Errorf("ValidateBatchResults() error = %q, want %q", err.Error(), tt.errContains)
			}
		})
	}
}
