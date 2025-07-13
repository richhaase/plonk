// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

func TestNewCommandPipeline(t *testing.T) {
	// Create a test command with flags
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("output", "table", "Output format")
	cmd.Flags().Bool("dry-run", false, "Dry run")
	cmd.Flags().Bool("force", false, "Force")
	cmd.Flags().Bool("verbose", false, "Verbose")

	pipeline, err := NewCommandPipeline(cmd, "package")
	if err != nil {
		t.Fatalf("unexpected error creating pipeline: %v", err)
	}

	if pipeline.itemType != "package" {
		t.Errorf("expected itemType 'package', got '%s'", pipeline.itemType)
	}

	if pipeline.format != OutputTable {
		t.Errorf("expected format OutputTable, got %v", pipeline.format)
	}

	if pipeline.flags == nil {
		t.Error("expected flags to be parsed")
	}
}

func TestNewSimpleCommandPipeline(t *testing.T) {
	// Create a test command with minimal flags
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("output", "json", "Output format")

	pipeline, err := NewSimpleCommandPipeline(cmd, "info")
	if err != nil {
		t.Fatalf("unexpected error creating simple pipeline: %v", err)
	}

	if pipeline.itemType != "info" {
		t.Errorf("expected itemType 'info', got '%s'", pipeline.itemType)
	}

	if pipeline.format != OutputJSON {
		t.Errorf("expected format OutputJSON, got %v", pipeline.format)
	}

	if pipeline.flags != nil {
		t.Error("expected flags to be nil for simple pipeline")
	}
}

func TestCommandPipeline_ExecuteWithResults(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("output", "json", "Output format")
	cmd.Flags().Bool("dry-run", false, "Dry run")

	pipeline, err := NewCommandPipeline(cmd, "package")
	if err != nil {
		t.Fatalf("unexpected error creating pipeline: %v", err)
	}

	// Create a test processor that returns successful results
	processor := func(ctx context.Context, args []string, flags *SimpleFlags) ([]operations.OperationResult, error) {
		return []operations.OperationResult{
			{
				Name:   "test-package",
				Status: "added",
				Metadata: map[string]interface{}{
					"manager": "homebrew",
				},
			},
		}, nil
	}

	// Execute the pipeline
	err = pipeline.ExecuteWithResults(context.Background(), processor, []string{"test-package"})
	if err != nil {
		t.Errorf("unexpected error executing pipeline: %v", err)
	}
}

func TestCommandPipeline_ExecuteWithData(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("output", "table", "Output format")

	pipeline, err := NewSimpleCommandPipeline(cmd, "status")
	if err != nil {
		t.Fatalf("unexpected error creating pipeline: %v", err)
	}

	// Create a test processor that returns output data
	processor := func(ctx context.Context, args []string) (OutputData, error) {
		return &testOutputData{message: "test output"}, nil
	}

	// Execute the pipeline
	err = pipeline.ExecuteWithData(context.Background(), processor, []string{})
	if err != nil {
		t.Errorf("unexpected error executing pipeline: %v", err)
	}
}

func TestCommandPipeline_GetAccessors(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("output", "yaml", "Output format")
	cmd.Flags().Bool("verbose", true, "Verbose")

	pipeline, err := NewCommandPipeline(cmd, "dotfile")
	if err != nil {
		t.Fatalf("unexpected error creating pipeline: %v", err)
	}

	// Test accessors
	if pipeline.GetFormat() != OutputYAML {
		t.Errorf("expected format OutputYAML, got %v", pipeline.GetFormat())
	}

	flags := pipeline.GetFlags()
	if flags == nil {
		t.Error("expected flags to be returned")
	}
	if !flags.Verbose {
		t.Error("expected verbose flag to be true")
	}

	if pipeline.GetCommand() != cmd {
		t.Error("expected command to be returned")
	}
}

func TestCountStatus(t *testing.T) {
	results := []operations.OperationResult{
		{Status: "added"},
		{Status: "failed"},
		{Status: "added"},
		{Status: "already-configured"},
	}

	addedCount := countStatus(results, "added")
	if addedCount != 2 {
		t.Errorf("expected 2 'added' results, got %d", addedCount)
	}

	failedCount := countStatus(results, "failed")
	if failedCount != 1 {
		t.Errorf("expected 1 'failed' result, got %d", failedCount)
	}

	successCount := countStatus(results, "added", "already-configured")
	if successCount != 3 {
		t.Errorf("expected 3 successful results, got %d", successCount)
	}
}

func TestGetMetadataString(t *testing.T) {
	result := operations.OperationResult{
		Metadata: map[string]interface{}{
			"manager": "homebrew",
			"number":  42,
		},
	}

	// Test valid string metadata
	manager := getMetadataString(result, "manager")
	if manager != "homebrew" {
		t.Errorf("expected 'homebrew', got '%s'", manager)
	}

	// Test non-string metadata
	number := getMetadataString(result, "number")
	if number != "" {
		t.Errorf("expected empty string for non-string metadata, got '%s'", number)
	}

	// Test missing metadata
	missing := getMetadataString(result, "missing")
	if missing != "" {
		t.Errorf("expected empty string for missing metadata, got '%s'", missing)
	}

	// Test nil metadata
	result.Metadata = nil
	nilMeta := getMetadataString(result, "manager")
	if nilMeta != "" {
		t.Errorf("expected empty string for nil metadata, got '%s'", nilMeta)
	}
}

func TestGenerateActionMessage(t *testing.T) {
	tests := []struct {
		name     string
		result   operations.OperationResult
		expected string
	}{
		{
			name:     "added status",
			result:   operations.OperationResult{Name: "pkg1", Status: "added"},
			expected: "Added pkg1",
		},
		{
			name:     "already configured",
			result:   operations.OperationResult{Name: "pkg2", Status: "already-configured"},
			expected: "pkg2 already configured",
		},
		{
			name:     "failed with error",
			result:   operations.OperationResult{Name: "pkg3", Status: "failed", Error: &testError{"test error"}},
			expected: "Failed: test error",
		},
		{
			name:     "unknown status",
			result:   operations.OperationResult{Name: "pkg4", Status: "unknown"},
			expected: "Processed pkg4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := generateActionMessage(tt.result)
			if message != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, message)
			}
		})
	}
}

// Test helper types

type testOutputData struct {
	message string
}

func (t *testOutputData) TableOutput() string {
	return t.message
}

func (t *testOutputData) StructuredData() any {
	return map[string]string{"message": t.message}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
