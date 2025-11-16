// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"reflect"
	"testing"
)

func TestMergeCommandConfig(t *testing.T) {
	base := CommandConfig{
		Command:          []string{"base", "cmd"},
		IdempotentErrors: []string{"already installed"},
	}

	override := CommandConfig{
		Command: []string{"override", "cmd"},
	}

	result := mergeCommandConfig(base, override)

	if !reflect.DeepEqual(result.Command, override.Command) {
		t.Fatalf("expected command %v, got %v", override.Command, result.Command)
	}

	if !reflect.DeepEqual(result.IdempotentErrors, base.IdempotentErrors) {
		t.Fatalf("expected idempotent errors %v, got %v", base.IdempotentErrors, result.IdempotentErrors)
	}
}

func TestMergeListConfig(t *testing.T) {
	base := ListConfig{
		Command:       []string{"base", "list"},
		Parse:         "lines",
		ParseStrategy: "lines",
		JSONField:     "name",
	}

	override := ListConfig{
		Command:   []string{"override", "list"},
		Parse:     "json",
		JSONField: "version",
	}

	result := mergeListConfig(base, override)

	if !reflect.DeepEqual(result.Command, override.Command) {
		t.Fatalf("expected command %v, got %v", override.Command, result.Command)
	}

	if result.Parse != "json" {
		t.Fatalf("expected parse json, got %s", result.Parse)
	}

	if result.ParseStrategy != base.ParseStrategy {
		t.Fatalf("expected parse strategy %s, got %s", base.ParseStrategy, result.ParseStrategy)
	}

	if result.JSONField != "version" {
		t.Fatalf("expected json field version, got %s", result.JSONField)
	}
}

func TestMergeMetadataExtractors(t *testing.T) {
	base := map[string]MetadataExtractorConfig{
		"scope": {
			Pattern: "^@([^/]+)/",
			Group:   1,
			Source:  "name",
		},
		"version": {
			Source: "json_field",
			Field:  "version",
		},
	}

	override := map[string]MetadataExtractorConfig{
		"version": {
			Source: "json_field",
			Field:  "pkg_version",
		},
		"full_name": {
			Source: "name",
		},
	}

	result := mergeMetadataExtractors(base, override)

	if len(result) != 3 {
		t.Fatalf("expected 3 metadata extractors, got %d", len(result))
	}

	if result["scope"] != base["scope"] {
		t.Fatalf("expected scope extractor to be unchanged")
	}

	if result["version"] != override["version"] {
		t.Fatalf("expected version extractor to be overridden")
	}

	if result["full_name"] != override["full_name"] {
		t.Fatalf("expected full_name extractor to be from override")
	}
}

func TestMergeManagerConfig(t *testing.T) {
	base := ManagerConfig{
		Binary: "base-binary",
		List: ListConfig{
			Command: []string{"base", "list"},
			Parse:   "lines",
		},
		Install: CommandConfig{
			Command:          []string{"base", "install"},
			IdempotentErrors: []string{"already installed"},
		},
		Description:   "Base manager",
		InstallHint:   "Base hint",
		HelpURL:       "https://base.example.com",
		UpgradeTarget: "name",
		MetadataExtractors: map[string]MetadataExtractorConfig{
			"scope": {
				Pattern: "^@([^/]+)/",
				Group:   1,
				Source:  "name",
			},
		},
	}

	override := ManagerConfig{
		Binary: "override-binary",
		List: ListConfig{
			Command: []string{"override", "list"},
		},
		Install: CommandConfig{
			Command: []string{"override", "install"},
		},
		Description: "Override manager",
		MetadataExtractors: map[string]MetadataExtractorConfig{
			"full_name": {
				Source: "name",
			},
		},
	}

	result := MergeManagerConfig(base, override)

	if result.Binary != "override-binary" {
		t.Fatalf("expected binary override-binary, got %s", result.Binary)
	}

	if !reflect.DeepEqual(result.List.Command, override.List.Command) {
		t.Fatalf("expected list command %v, got %v", override.List.Command, result.List.Command)
	}

	if !reflect.DeepEqual(result.Install.Command, override.Install.Command) {
		t.Fatalf("expected install command %v, got %v", override.Install.Command, result.Install.Command)
	}

	if !reflect.DeepEqual(result.Install.IdempotentErrors, base.Install.IdempotentErrors) {
		t.Fatalf("expected install idempotent errors %v, got %v", base.Install.IdempotentErrors, result.Install.IdempotentErrors)
	}

	if result.Description != "Override manager" {
		t.Fatalf("expected description Override manager, got %s", result.Description)
	}

	if result.InstallHint != base.InstallHint {
		t.Fatalf("expected install hint %s, got %s", base.InstallHint, result.InstallHint)
	}

	if len(result.MetadataExtractors) != 2 {
		t.Fatalf("expected 2 metadata extractors, got %d", len(result.MetadataExtractors))
	}

	if result.MetadataExtractors["scope"] != base.MetadataExtractors["scope"] {
		t.Fatalf("expected scope extractor to be inherited from base")
	}

	if result.MetadataExtractors["full_name"] != override.MetadataExtractors["full_name"] {
		t.Fatalf("expected full_name extractor to be from override")
	}
}
