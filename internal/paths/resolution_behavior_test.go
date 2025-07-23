// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BehaviorTestCase represents a test case for path resolution behavior
type BehaviorTestCase struct {
	name    string
	input   string
	homeDir string
	workDir string // current working directory for relative path tests
}

// BehaviorResult captures the result of a path resolution implementation
type BehaviorResult struct {
	output string
	err    error
	panic  bool // some implementations might panic
}

// TestPathResolutionBehaviors documents the current behavior of all path resolution implementations
func TestPathResolutionBehaviors(t *testing.T) {
	// Test setup
	homeDir := t.TempDir()
	workDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Create some test directories and files
	os.MkdirAll(filepath.Join(homeDir, ".config", "nvim"), 0755)
	os.MkdirAll(filepath.Join(homeDir, ".ssh"), 0755)
	os.MkdirAll(filepath.Join(workDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(workDir, "local.conf"), []byte("test"), 0644)

	testCases := []BehaviorTestCase{
		// Tilde expansion tests
		{name: "tilde_only", input: "~", homeDir: homeDir, workDir: workDir},
		{name: "tilde_slash", input: "~/", homeDir: homeDir, workDir: workDir},
		{name: "tilde_file", input: "~/.zshrc", homeDir: homeDir, workDir: workDir},
		{name: "tilde_subdir", input: "~/.config/nvim", homeDir: homeDir, workDir: workDir},
		{name: "tilde_nonexistent", input: "~/does-not-exist", homeDir: homeDir, workDir: workDir},

		// Absolute path tests
		{name: "abs_home_file", input: filepath.Join(homeDir, ".zshrc"), homeDir: homeDir, workDir: workDir},
		{name: "abs_outside_home", input: "/etc/passwd", homeDir: homeDir, workDir: workDir},
		{name: "abs_root", input: "/", homeDir: homeDir, workDir: workDir},
		{name: "abs_nonexistent", input: "/does/not/exist", homeDir: homeDir, workDir: workDir},

		// Relative path tests
		{name: "rel_current_dir", input: ".", homeDir: homeDir, workDir: workDir},
		{name: "rel_parent_dir", input: "..", homeDir: homeDir, workDir: workDir},
		{name: "rel_file", input: "local.conf", homeDir: homeDir, workDir: workDir},
		{name: "rel_subdir", input: "./subdir", homeDir: homeDir, workDir: workDir},
		{name: "rel_subdir_no_dot", input: "subdir", homeDir: homeDir, workDir: workDir},

		// Edge cases
		{name: "empty_string", input: "", homeDir: homeDir, workDir: workDir},
		{name: "single_dot", input: ".", homeDir: homeDir, workDir: workDir},
		{name: "double_dot", input: "..", homeDir: homeDir, workDir: workDir},
		{name: "triple_dot", input: "...", homeDir: homeDir, workDir: workDir},
		{name: "path_traversal", input: "../../../etc/passwd", homeDir: homeDir, workDir: workDir},
		{name: "double_slash", input: "//tmp//test", homeDir: homeDir, workDir: workDir},
		{name: "trailing_slash", input: "~/.config/", homeDir: homeDir, workDir: workDir},

		// Special characters (where applicable)
		{name: "space_in_path", input: "~/My Documents", homeDir: homeDir, workDir: workDir},
		{name: "special_chars", input: "~/test@file#.txt", homeDir: homeDir, workDir: workDir},
	}

	// Implementation 1: PathResolver.ResolveDotfilePath
	t.Run("PathResolver.ResolveDotfilePath", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Change to work directory for relative path tests
				oldWd, _ := os.Getwd()
				os.Chdir(tc.workDir)
				defer os.Chdir(oldWd)

				resolver := NewPathResolver(tc.homeDir, configDir)
				result := capturePathResolverResult(resolver, tc.input)
				logBehavior(t, "PathResolver.ResolveDotfilePath", tc, result)
			})
		}
	})

	// Implementation 2: dotfiles.Manager.ExpandPath (simulated)
	t.Run("dotfiles.Manager.ExpandPath", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate the dotfiles.Manager.ExpandPath behavior
				result := simulateDotfilesManagerExpandPath(tc.input, tc.homeDir)
				logBehavior(t, "dotfiles.Manager.ExpandPath", tc, result)
			})
		}
	})

	// Implementation 3: services.ResolveDotfilePath (simulated)
	t.Run("services.ResolveDotfilePath", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Change to work directory for relative path tests
				oldWd, _ := os.Getwd()
				os.Chdir(tc.workDir)
				defer os.Chdir(oldWd)

				// Simulate the services.ResolveDotfilePath behavior
				result := simulateServicesResolveDotfilePath(tc.input, tc.homeDir)
				logBehavior(t, "services.ResolveDotfilePath", tc, result)
			})
		}
	})

	// Implementation 4: Simple tilde expansion from yaml_config
	t.Run("yaml_config_tilde_expansion", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := captureYamlConfigResult(tc.input, tc.homeDir)
				logBehavior(t, "yaml_config_tilde_expansion", tc, result)
			})
		}
	})

	// Generate comparison report
	t.Run("behavior_comparison", func(t *testing.T) {
		t.Log("\n=== BEHAVIOR DIFFERENCES SUMMARY ===")
		t.Log("This test documents current behavior, not correctness")
		t.Log("Use this data to ensure consolidation maintains compatibility")
	})
}

// Helper functions to capture results from each implementation

func capturePathResolverResult(resolver *PathResolver, input string) BehaviorResult {
	defer func() {
		if r := recover(); r != nil {
			// PathResolver shouldn't panic, but capture if it does
		}
	}()

	output, err := resolver.ResolveDotfilePath(input)
	return BehaviorResult{output: output, err: err}
}

// simulateDotfilesManagerExpandPath simulates dotfiles.Manager.ExpandPath behavior
// Based on: internal/dotfiles/operations.go lines 113-118
func simulateDotfilesManagerExpandPath(path, homeDir string) BehaviorResult {
	if strings.HasPrefix(path, "~/") {
		return BehaviorResult{
			output: filepath.Join(homeDir, path[2:]),
			err:    nil,
		}
	}
	return BehaviorResult{output: path, err: nil}
}

// simulateServicesResolveDotfilePath simulates services.ResolveDotfilePath behavior
// Based on: internal/services/dotfile_operations.go lines 354-362
func simulateServicesResolveDotfilePath(path, homeDir string) BehaviorResult {
	if strings.HasPrefix(path, "~/") {
		return BehaviorResult{
			output: filepath.Join(homeDir, path[2:]),
			err:    nil,
		}
	}
	if !filepath.IsAbs(path) {
		return BehaviorResult{
			output: filepath.Join(homeDir, path),
			err:    nil,
		}
	}
	return BehaviorResult{output: path, err: nil}
}

func captureYamlConfigResult(input, homeDir string) BehaviorResult {
	// Simulate the inline expansion from yaml_config.go
	if strings.HasPrefix(input, "~/") {
		return BehaviorResult{
			output: filepath.Join(homeDir, input[2:]),
			err:    nil,
		}
	}
	return BehaviorResult{output: input, err: nil}
}

func logBehavior(t *testing.T, impl string, tc BehaviorTestCase, result BehaviorResult) {
	t.Helper()
	if result.panic {
		t.Logf("[%s] Input: %q -> PANIC", impl, tc.input)
	} else if result.err != nil {
		t.Logf("[%s] Input: %q -> ERROR: %v", impl, tc.input, result.err)
	} else {
		t.Logf("[%s] Input: %q -> Output: %q", impl, tc.input, result.output)
	}
}

// TestPathResolutionDifferences highlights specific behavioral differences
func TestPathResolutionDifferences(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Create work directory structure
	os.Chdir(workDir)
	os.WriteFile("testfile", []byte("test"), 0644)

	t.Run("relative_path_handling", func(t *testing.T) {
		// PathResolver tries current directory first, then home
		resolver := NewPathResolver(homeDir, configDir)
		result1, err1 := resolver.ResolveDotfilePath("testfile")

		// services.ResolveDotfilePath joins with home directory
		result2 := simulateServicesResolveDotfilePath("testfile", homeDir)

		t.Logf("PathResolver: %q (err: %v)", result1, err1)
		t.Logf("services: %q (err: %v)", result2.output, result2.err)
		t.Logf("Different behavior: %v", result1 != result2.output)
	})

	t.Run("validation_differences", func(t *testing.T) {
		// PathResolver validates paths are within home directory
		resolver := NewPathResolver(homeDir, configDir)
		_, err1 := resolver.ResolveDotfilePath("/etc/passwd")

		// services.ResolveDotfilePath has no validation
		result2 := simulateServicesResolveDotfilePath("/etc/passwd", homeDir)

		t.Logf("PathResolver blocks /etc/passwd: %v", err1 != nil)
		t.Logf("services allows /etc/passwd: %v (result: %q)", result2.err == nil, result2.output)
	})

	t.Run("empty_string_handling", func(t *testing.T) {
		resolver := NewPathResolver(homeDir, configDir)
		result1, err1 := resolver.ResolveDotfilePath("")

		result2 := simulateServicesResolveDotfilePath("", homeDir)

		t.Logf("PathResolver empty string: %q (err: %v)", result1, err1)
		t.Logf("services empty string: %q (err: %v)", result2.output, result2.err)
	})
}

// TestPathResolutionPerformance compares performance of implementations
func TestPathResolutionPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "plonk")

	testPath := "~/.config/nvim/init.lua"

	// Run benchmarks to compare performance
	if !testing.Short() {
		t.Run("PathResolver_performance", func(t *testing.T) {
			resolver := NewPathResolver(homeDir, configDir)
			start := testing.Benchmark(func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					resolver.ResolveDotfilePath(testPath)
				}
			})
			t.Logf("PathResolver: %d ns/op", start.NsPerOp())
		})

		t.Run("services_performance", func(t *testing.T) {
			start := testing.Benchmark(func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					simulateServicesResolveDotfilePath(testPath, homeDir)
				}
			})
			t.Logf("services: %d ns/op", start.NsPerOp())
		})
	}
}
