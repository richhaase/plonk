// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"testing"
)

func TestGemManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard gem list output",
			output: []byte("rails\nrubocop\nbundler"),
			want:   []string{"rails", "rubocop", "bundler"},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single gem",
			output: []byte("rails"),
			want:   []string{"rails"},
		},
		{
			name:   "output with extra whitespace",
			output: []byte("  rails  \n  rubocop  \n  bundler  "),
			want:   []string{"rails", "rubocop", "bundler"},
		},
		{
			name:   "output with blank lines",
			output: []byte("rails\n\nrubocop\n\nbundler"),
			want:   []string{"rails", "rubocop", "bundler"},
		},
		{
			name:   "gems with hyphens and underscores",
			output: []byte("active-record\ntest_unit\nrspec-core"),
			want:   []string{"active-record", "test_unit", "rspec-core"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
			got := manager.parseListOutput(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGemManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard gem search output",
			output: []byte(`rails (7.0.4, 7.0.3, 6.1.7)
    A full-stack web framework optimized for programmer happiness
railties (7.0.4, 7.0.3, 6.1.7)
    Rails internals: application bootup, plugins, generators
rspec-rails (6.0.1, 5.1.2)
    Testing framework for Rails`),
			want: []string{"rails", "railties", "rspec-rails"},
		},
		{
			name:   "no results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single result",
			output: []byte(`rails (7.0.4)
    A full-stack web framework optimized for programmer happiness`),
			want: []string{"rails"},
		},
		{
			name: "gems with underscores and hyphens",
			output: []byte(`active_record (7.0.4)
    Object-relational mapping layer for Rails
rspec-core (3.12.0)
    RSpec core runner`),
			want: []string{"active_record", "rspec-core"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
			got := manager.parseSearchOutput(tt.output)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGemManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard gem specification output",
			output: []byte(`--- !ruby/object:Gem::Specification
name: rails
version: !ruby/object:Gem::Version
  version: 7.0.4
summary: A full-stack web framework optimized for programmer happiness and beautiful code
homepage: https://rubyonrails.org
description: Ruby on Rails is a full-stack web framework optimized for programmer happiness and beautiful code.`),
			packageName: "rails",
			want: &PackageInfo{
				Name:        "rails",
				Version:     "7.0.4",
				Description: "A full-stack web framework optimized for programmer happiness and beautiful code",
				Homepage:    "https://rubyonrails.org",
			},
		},
		{
			name: "minimal gem specification output",
			output: []byte(`--- !ruby/object:Gem::Specification
name: bundler
version: !ruby/object:Gem::Version
  version: 2.4.1`),
			packageName: "bundler",
			want: &PackageInfo{
				Name:        "bundler",
				Version:     "2.4.1",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "unknown",
			want: &PackageInfo{
				Name:        "unknown",
				Version:     "",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name: "gem with description field",
			output: []byte(`--- !ruby/object:Gem::Specification
name: rspec
version: !ruby/object:Gem::Version
  version: 3.12.0
description: BDD for Ruby
homepage: https://github.com/rspec/rspec`),
			packageName: "rspec",
			want: &PackageInfo{
				Name:        "rspec",
				Version:     "3.12.0",
				Description: "BDD for Ruby",
				Homepage:    "https://github.com/rspec/rspec",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}