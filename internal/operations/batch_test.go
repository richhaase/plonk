// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package operations

import (
	"context"
	"errors"
	"testing"
	"time"
)

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

func TestGenericBatchProcessor_ProcessItems(t *testing.T) {
	tests := []struct {
		name            string
		items           []string
		processor       ItemProcessor
		options         BatchProcessorOptions
		expectedResults int
		expectError     bool
	}{
		{
			name:  "successful processing of all items",
			items: []string{"item1", "item2", "item3"},
			processor: func(ctx context.Context, item string) OperationResult {
				return OperationResult{
					Name:   item,
					Status: "added",
				}
			},
			options: BatchProcessorOptions{
				ItemType:               "test",
				Operation:              "add",
				ShowIndividualProgress: false,
				Timeout:                time.Second * 5,
				ContinueOnError:        boolPtr(true),
			},
			expectedResults: 3,
			expectError:     false,
		},
		{
			name:  "continue on error",
			items: []string{"item1", "error-item", "item3"},
			processor: func(ctx context.Context, item string) OperationResult {
				if item == "error-item" {
					return OperationResult{
						Name:   item,
						Status: "failed",
						Error:  errors.New("test error"),
					}
				}
				return OperationResult{
					Name:   item,
					Status: "added",
				}
			},
			options: BatchProcessorOptions{
				ItemType:               "test",
				Operation:              "add",
				ShowIndividualProgress: false,
				Timeout:                time.Second * 5,
				ContinueOnError:        boolPtr(true),
			},
			expectedResults: 3,
			expectError:     false,
		},
		{
			name:  "stop on error",
			items: []string{"item1", "error-item", "item3"},
			processor: func(ctx context.Context, item string) OperationResult {
				if item == "error-item" {
					return OperationResult{
						Name:   item,
						Status: "failed",
						Error:  errors.New("test error"),
					}
				}
				return OperationResult{
					Name:   item,
					Status: "added",
				}
			},
			options: BatchProcessorOptions{
				ItemType:               "test",
				Operation:              "add",
				ShowIndividualProgress: false,
				Timeout:                time.Second * 5,
				ContinueOnError:        boolPtr(false),
			},
			expectedResults: 2, // Should stop after error
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewBatchProcessor(tt.processor, tt.options)

			results, err := processor.ProcessItems(context.Background(), tt.items)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(results) != tt.expectedResults {
				t.Errorf("Expected %d results, got %d", tt.expectedResults, len(results))
			}
		})
	}
}

func TestStandardBatchWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		items           []string
		processor       ItemProcessor
		options         BatchProcessorOptions
		expectError     bool
		expectedResults int
	}{
		{
			name:  "successful workflow",
			items: []string{"item1", "item2"},
			processor: func(ctx context.Context, item string) OperationResult {
				return OperationResult{
					Name:   item,
					Status: "added",
				}
			},
			options: BatchProcessorOptions{
				ItemType:  "test",
				Operation: "add",
				Timeout:   time.Second * 5,
			},
			expectError:     false,
			expectedResults: 2,
		},
		{
			name:  "workflow with failures",
			items: []string{"item1", "error-item"},
			processor: func(ctx context.Context, item string) OperationResult {
				if item == "error-item" {
					return OperationResult{
						Name:   item,
						Status: "failed",
						Error:  errors.New("test error"),
					}
				}
				return OperationResult{
					Name:   item,
					Status: "added",
				}
			},
			options: BatchProcessorOptions{
				ItemType:  "test",
				Operation: "add",
				Timeout:   time.Second * 5,
			},
			expectError:     false, // Changed to false since DetermineExitCode handles this
			expectedResults: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := StandardBatchWorkflow(context.Background(), tt.items, tt.processor, tt.options)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(results) != tt.expectedResults {
				t.Errorf("Expected %d results, got %d", tt.expectedResults, len(results))
			}
		})
	}
}

func TestSimpleProcessor(t *testing.T) {
	processFn := func(ctx context.Context, item string) OperationResult {
		return OperationResult{
			Name:   item,
			Status: "added",
		}
	}

	processor := SimpleProcessor(processFn)

	result := processor(context.Background(), "test-item")

	if result.Name != "test-item" {
		t.Errorf("Expected name 'test-item', got '%s'", result.Name)
	}
	if result.Status != "added" {
		t.Errorf("Expected status 'added', got '%s'", result.Status)
	}
}

func TestPackageProcessor(t *testing.T) {
	processFn := func(ctx context.Context, packageName, manager string) OperationResult {
		return OperationResult{
			Name:    packageName,
			Manager: manager,
			Status:  "added",
		}
	}

	processor := PackageProcessor(processFn, "homebrew")

	result := processor(context.Background(), "test-package")

	if result.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", result.Name)
	}
	if result.Manager != "homebrew" {
		t.Errorf("Expected manager 'homebrew', got '%s'", result.Manager)
	}
	if result.Status != "added" {
		t.Errorf("Expected status 'added', got '%s'", result.Status)
	}
}
