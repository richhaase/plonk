// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// ResolvedConfig represents the final computed configuration values
// after merging defaults with user overrides. This is what commands should use.
type ResolvedConfig struct {
	DefaultManager    string   `yaml:"default_manager" json:"default_manager"`
	OperationTimeout  int      `yaml:"operation_timeout" json:"operation_timeout"`
	PackageTimeout    int      `yaml:"package_timeout" json:"package_timeout"`
	DotfileTimeout    int      `yaml:"dotfile_timeout" json:"dotfile_timeout"`
	ExpandDirectories []string `yaml:"expand_directories" json:"expand_directories"`
	IgnorePatterns    []string `yaml:"ignore_patterns" json:"ignore_patterns"`
}

// GetOperationTimeout returns the operation timeout in seconds
func (r *ResolvedConfig) GetOperationTimeout() int {
	return r.OperationTimeout
}

// GetPackageTimeout returns the package timeout in seconds
func (r *ResolvedConfig) GetPackageTimeout() int {
	return r.PackageTimeout
}

// GetDotfileTimeout returns the dotfile timeout in seconds
func (r *ResolvedConfig) GetDotfileTimeout() int {
	return r.DotfileTimeout
}

// GetDefaultManager returns the default package manager
func (r *ResolvedConfig) GetDefaultManager() string {
	return r.DefaultManager
}

// GetExpandDirectories returns the directories to expand in dot list output
func (r *ResolvedConfig) GetExpandDirectories() []string {
	return r.ExpandDirectories
}

// GetIgnorePatterns returns the ignore patterns for dotfile discovery
func (r *ResolvedConfig) GetIgnorePatterns() []string {
	return r.IgnorePatterns
}

// Helper functions for creating pointers in tests and configuration

// StringPtr returns a pointer to the given string value
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int value
func IntPtr(i int) *int {
	return &i
}

// StringSlicePtr returns a pointer to the given string slice
func StringSlicePtr(s []string) *[]string {
	return &s
}
