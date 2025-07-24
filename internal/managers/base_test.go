// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"testing"
)

// TestBaseManager_Configuration tests the basic configuration setup
func TestBaseManager_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config ManagerConfig
		want   string
	}{
		{
			name: "basic configuration",
			config: ManagerConfig{
				BinaryName: "npm",
			},
			want: "npm",
		},
		{
			name: "configuration with fallbacks",
			config: ManagerConfig{
				BinaryName:       "pip",
				FallbackBinaries: []string{"pip3"},
			},
			want: "pip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBaseManager(tt.config)
			if base.Config.BinaryName != tt.want {
				t.Errorf("NewBaseManager() BinaryName = %v, want %v", base.Config.BinaryName, tt.want)
			}
			if base.GetBinary() != tt.want {
				t.Errorf("GetBinary() = %v, want %v", base.GetBinary(), tt.want)
			}
		})
	}
}

// TestManagerConfig_Functions tests the configuration function builders
func TestManagerConfig_Functions(t *testing.T) {
	config := ManagerConfig{
		BinaryName: "npm",
		ListArgs: func() []string {
			return []string{"list", "-g", "--depth=0"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", "-g", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", "-g", pkg}
		},
	}

	// Test list args
	listArgs := config.ListArgs()
	expected := []string{"list", "-g", "--depth=0"}
	if len(listArgs) != len(expected) {
		t.Errorf("ListArgs() length = %v, want %v", len(listArgs), len(expected))
	}
	for i, arg := range listArgs {
		if arg != expected[i] {
			t.Errorf("ListArgs()[%d] = %v, want %v", i, arg, expected[i])
		}
	}

	// Test install args
	installArgs := config.InstallArgs("typescript")
	expectedInstall := []string{"install", "-g", "typescript"}
	if len(installArgs) != len(expectedInstall) {
		t.Errorf("InstallArgs() length = %v, want %v", len(installArgs), len(expectedInstall))
	}
	for i, arg := range installArgs {
		if arg != expectedInstall[i] {
			t.Errorf("InstallArgs()[%d] = %v, want %v", i, arg, expectedInstall[i])
		}
	}

	// Test uninstall args
	uninstallArgs := config.UninstallArgs("typescript")
	expectedUninstall := []string{"uninstall", "-g", "typescript"}
	if len(uninstallArgs) != len(expectedUninstall) {
		t.Errorf("UninstallArgs() length = %v, want %v", len(uninstallArgs), len(expectedUninstall))
	}
	for i, arg := range uninstallArgs {
		if arg != expectedUninstall[i] {
			t.Errorf("UninstallArgs()[%d] = %v, want %v", i, arg, expectedUninstall[i])
		}
	}
}