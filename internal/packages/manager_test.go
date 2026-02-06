// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "testing"

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantMgr string
		wantPkg string
		wantErr bool
	}{
		{name: "valid", spec: "brew:ripgrep", wantMgr: "brew", wantPkg: "ripgrep"},
		{name: "invalid format", spec: "ripgrep", wantErr: true},
		{name: "unsupported manager", spec: "npm:typescript", wantErr: true},
		{name: "empty package", spec: "brew:", wantErr: true},
	}

	for _, tt := range tests {
		manager, pkg, err := ParsePackageSpec(tt.spec)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
			continue
		}
		if manager != tt.wantMgr || pkg != tt.wantPkg {
			t.Errorf("%s: ParsePackageSpec(%q) = (%q, %q), want (%q, %q)", tt.name, tt.spec, manager, pkg, tt.wantMgr, tt.wantPkg)
		}
	}
}
