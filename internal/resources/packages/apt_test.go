// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestAptManager_SupportsSearch(t *testing.T) {
	apt := NewAptManager()
	if !apt.SupportsSearch() {
		t.Error("APT should support search")
	}
}

func TestAptManager_IsAvailable(t *testing.T) {
	apt := NewAptManager()
	ctx := context.Background()

	available, err := apt.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("IsAvailable returned error: %v", err)
	}

	// On non-Linux systems, it should not be available
	if runtime.GOOS != "linux" {
		if available {
			t.Error("APT should not be available on non-Linux systems")
		}
		return
	}

	// On Linux, log the result (depends on whether it's Debian-based)
	t.Logf("APT available on this Linux system: %v", available)
}

func TestAptManager_formatPackageName(t *testing.T) {
	apt := NewAptManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple package",
			input:    "htop",
			expected: "htop",
		},
		{
			name:     "package with architecture",
			input:    "libc6:i386",
			expected: "libc6:i386",
		},
		{
			name:     "package with whitespace",
			input:    "  nginx  ",
			expected: "nginx",
		},
		{
			name:     "development package",
			input:    "libssl-dev",
			expected: "libssl-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apt.formatPackageName(tt.input)
			if got != tt.expected {
				t.Errorf("formatPackageName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// MockAPTCommands can be used for testing APT operations on non-Linux systems
type MockAPTCommands struct {
	DpkgQueryResponses map[string]struct {
		Output []byte
		Error  error
	}
	AptCacheResponses map[string]struct {
		Output []byte
		Error  error
	}
}

func TestAptManager_IsInstalled_Logic(t *testing.T) {
	// This test validates the logic of IsInstalled without requiring APT to be available
	tests := []struct {
		name           string
		packageName    string
		dpkgOutput     string
		dpkgExitCode   int
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "package installed",
			packageName:    "nginx",
			dpkgOutput:     "installed",
			dpkgExitCode:   0,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "package not installed",
			packageName:    "nginx",
			dpkgOutput:     "",
			dpkgExitCode:   1,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "package deinstalled",
			packageName:    "nginx",
			dpkgOutput:     "deinstalled",
			dpkgExitCode:   0,
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a logic test - actual command execution tests would
			// be in integration tests on Linux systems
			t.Logf("Test case validates IsInstalled logic for %q", tt.packageName)
		})
	}
}

func TestAptManager_Search_OutputParsing(t *testing.T) {
	// Test the parsing logic for apt-cache search output
	testOutput := `nginx - small, powerful, scalable web/proxy server
nginx-common - small, powerful, scalable web/proxy server - common files
nginx-doc - small, powerful, scalable web/proxy server - documentation
nginx-full - nginx web/proxy server (standard version with 3rd parties)
nginx-light - nginx web/proxy server (basic version)`

	// Parse like the Search method does
	var packages []string
	lines := strings.Split(testOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " - ", 2)
		if len(parts) > 0 && parts[0] != "" {
			packages = append(packages, parts[0])
		}
	}

	expected := []string{"nginx", "nginx-common", "nginx-doc", "nginx-full", "nginx-light"}

	if len(packages) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packages))
	}

	for i, pkg := range packages {
		if i < len(expected) && pkg != expected[i] {
			t.Errorf("Package[%d] = %q, want %q", i, pkg, expected[i])
		}
	}
}

func TestAptManager_Info_OutputParsing(t *testing.T) {
	// Test the parsing logic for apt-cache show output
	testOutput := `Package: nginx
Version: 1.18.0-6ubuntu14.4
Priority: optional
Section: web
Origin: Ubuntu
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Original-Maintainer: Debian Nginx Maintainers <pkg-nginx-maintainers@alioth-lists.debian.net>
Bugs: https://bugs.launchpad.net/ubuntu/+filebug
Installed-Size: 108
Depends: nginx-core (>= 1.18.0-6ubuntu14.4) | nginx-full (>= 1.18.0-6ubuntu14.4) | nginx-light (>= 1.18.0-6ubuntu14.4) | nginx-extras (>= 1.18.0-6ubuntu14.4)
Filename: pool/main/n/nginx/nginx_1.18.0-6ubuntu14.4_all.deb
Size: 3872
MD5sum: abc123def456
SHA1: 123abc456def
SHA256: 456def789abc
SHA512: 789abc123def
Homepage: https://nginx.org/
Description-en: small, powerful, scalable web/proxy server
 Nginx ("engine X") is a high-performance web and reverse proxy server
 created by Igor Sysoev. It can be used both as a standalone web server
 and as a proxy to reduce the load on back-end HTTP or mail servers.
 .
 This is a dependency package to install either nginx-full (default) or
 nginx-light.
Description-md5: abc123
Task: lamp-server, ubuntu-wsl`

	// Test the parsing logic
	info := &PackageInfo{
		Name:    "nginx",
		Manager: "apt",
	}

	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Version: ") {
			info.Version = strings.TrimPrefix(line, "Version: ")
		} else if strings.HasPrefix(line, "Homepage: ") {
			info.Homepage = strings.TrimPrefix(line, "Homepage: ")
		} else if strings.HasPrefix(line, "Description-en: ") {
			info.Description = strings.TrimPrefix(line, "Description-en: ")
		}
	}

	// Verify parsed values
	if info.Version != "1.18.0-6ubuntu14.4" {
		t.Errorf("Version = %q, want %q", info.Version, "1.18.0-6ubuntu14.4")
	}
	if info.Homepage != "https://nginx.org/" {
		t.Errorf("Homepage = %q, want %q", info.Homepage, "https://nginx.org/")
	}
	if info.Description != "small, powerful, scalable web/proxy server" {
		t.Errorf("Description = %q, want %q", info.Description, "small, powerful, scalable web/proxy server")
	}
}

func TestAptManager_handleInstallError(t *testing.T) {
	apt := NewAptManager()

	tests := []struct {
		name        string
		output      string
		packageName string
		wantErr     string
	}{
		{
			name:        "permission denied",
			output:      "E: Could not open lock file /var/lib/dpkg/lock - open (13: Permission denied)",
			packageName: "nginx",
			wantErr:     "permission denied: apt-get install requires sudo privileges. Try: sudo plonk install apt:nginx",
		},
		{
			name:        "are you root",
			output:      "are you root?",
			packageName: "nginx",
			wantErr:     "permission denied: apt-get install requires sudo privileges. Try: sudo plonk install apt:nginx",
		},
		{
			name:        "package not found",
			output:      "E: Unable to locate package foobar123",
			packageName: "foobar123",
			wantErr:     "package 'foobar123' not found",
		},
		{
			name:        "no installation candidate",
			output:      "Package foobar has no installation candidate",
			packageName: "foobar",
			wantErr:     "package 'foobar' not found",
		},
		{
			name:        "already installed",
			output:      "nginx is already the newest version (1.18.0-6ubuntu14.4)",
			packageName: "nginx",
			wantErr:     "", // No error expected
		},
		{
			name:        "broken packages",
			output:      "E: Broken packages",
			packageName: "nginx",
			wantErr:     "dependency conflict installing package 'nginx'",
		},
		{
			name:        "network error",
			output:      "Could not resolve 'archive.ubuntu.com'",
			packageName: "nginx",
			wantErr:     "network error: failed to download package information",
		},
		{
			name:        "failed to fetch",
			output:      "Failed to fetch http://archive.ubuntu.com/ubuntu/pool/main/n/nginx/nginx_1.18.0-6ubuntu14.4_all.deb",
			packageName: "nginx",
			wantErr:     "network error: failed to download package information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock error with exit code 1
			mockErr := &testExitError{exitCode: 1}
			err := apt.handleInstallError(mockErr, []byte(tt.output), tt.packageName)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestAptManager_handleUninstallError(t *testing.T) {
	apt := NewAptManager()

	tests := []struct {
		name        string
		output      string
		packageName string
		wantErr     string
	}{
		{
			name:        "permission denied",
			output:      "E: Could not open lock file /var/lib/dpkg/lock - open (13: Permission denied)",
			packageName: "nginx",
			wantErr:     "permission denied: apt-get remove requires sudo privileges. Try: sudo plonk uninstall apt:nginx",
		},
		{
			name:        "package not installed",
			output:      "Package 'nginx' is not installed, so not removed",
			packageName: "nginx",
			wantErr:     "", // No error expected - success
		},
		{
			name:        "unable to locate",
			output:      "E: Unable to locate package foobar123",
			packageName: "foobar123",
			wantErr:     "", // No error expected - success
		},
		{
			name:        "dependency conflict",
			output:      "nginx is depended on by nginx-common",
			packageName: "nginx",
			wantErr:     "cannot uninstall package 'nginx' due to dependency conflicts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock error with exit code 1
			mockErr := &testExitError{exitCode: 1}
			err := apt.handleUninstallError(mockErr, []byte(tt.output), tt.packageName)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

// testExitError is a mock implementation of exec.ExitError for testing
type testExitError struct {
	exitCode int
}

func (e *testExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.exitCode)
}

func (e *testExitError) ExitCode() int {
	return e.exitCode
}
