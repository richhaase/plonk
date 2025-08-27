// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/testutil"
)

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("Test spinner")
	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}
	if spinner.text != "Test spinner" {
		t.Errorf("Expected text 'Test spinner', got %q", spinner.text)
	}
	if spinner.running {
		t.Error("New spinner should not be running")
	}
	if spinner.done == nil {
		t.Error("New spinner should have done channel initialized")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Test")

	// Test Start
	result := spinner.Start()
	if result != spinner {
		t.Error("Start should return the spinner for chaining")
	}
	if !spinner.running {
		t.Error("Spinner should be running after Start")
	}

	// Give spinner time to write output
	time.Sleep(150 * time.Millisecond)

	// Test Stop
	spinner.Stop()
	if spinner.running {
		t.Error("Spinner should not be running after Stop")
	}

	output := buf.String()
	if !strings.Contains(output, "Test") {
		t.Errorf("Output should contain spinner text, got: %q", output)
	}
}

func TestSpinner_MultipleStart(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Test")

	// Start spinner
	spinner.Start()
	if !spinner.running {
		t.Error("Spinner should be running after first Start")
	}

	// Try to start again - should be idempotent
	result := spinner.Start()
	if result != spinner {
		t.Error("Start should return spinner even when already running")
	}
	if !spinner.running {
		t.Error("Spinner should still be running")
	}

	spinner.Stop()
}

func TestSpinner_MultipleStop(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Test")

	// Stop without starting - should be safe
	spinner.Stop()

	// Start and stop normally
	spinner.Start()
	time.Sleep(50 * time.Millisecond)
	spinner.Stop()

	// Stop again - should be safe
	spinner.Stop()
}

func TestSpinner_Success(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Loading")
	spinner.Start()
	time.Sleep(50 * time.Millisecond)

	spinner.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, IconSuccess) {
		t.Errorf("Success message should contain success icon, got: %q", output)
	}
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success message should contain the message, got: %q", output)
	}
	if spinner.running {
		t.Error("Spinner should be stopped after Success")
	}
}

func TestSpinner_Error(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Loading")
	spinner.Start()
	time.Sleep(50 * time.Millisecond)

	spinner.Error("Operation failed")

	output := buf.String()
	if !strings.Contains(output, IconError) {
		t.Errorf("Error message should contain error icon, got: %q", output)
	}
	if !strings.Contains(output, "Operation failed") {
		t.Errorf("Error message should contain the message, got: %q", output)
	}
	if spinner.running {
		t.Error("Spinner should be stopped after Error")
	}
}

func TestSpinner_UpdateText(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Initial text")
	spinner.Start()

	// Give spinner time to display initial text
	time.Sleep(150 * time.Millisecond)

	spinner.UpdateText("Updated text")

	// Give spinner time to display updated text
	time.Sleep(150 * time.Millisecond)

	spinner.Stop()

	output := buf.String()
	if !strings.Contains(output, "Updated text") {
		t.Errorf("Output should contain updated text, got: %q", output)
	}
}

func TestSpinner_NonTerminal(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	// Use non-terminal writer
	buf := testutil.NewBufferWriter(false)
	progressWriter = buf

	spinner := NewSpinner("Non-terminal test")
	spinner.Start()

	// For non-terminal, output is immediate
	time.Sleep(50 * time.Millisecond)
	spinner.Stop()

	output := buf.String()
	// In non-terminal mode, should just print the text once with newline
	if output != "Non-terminal test\n" {
		t.Errorf("Non-terminal output should be simple text with newline, got: %q", output)
	}
}

func TestSpinner_Animation(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Animating")
	spinner.Start()

	// Let it animate through multiple frames
	time.Sleep(350 * time.Millisecond)

	spinner.Stop()

	output := buf.String()

	// Check that we see multiple spinner characters
	foundChars := 0
	for _, char := range SpinnerChars {
		if strings.Contains(output, char) {
			foundChars++
		}
	}

	if foundChars < 2 {
		t.Errorf("Expected to see multiple spinner characters, found %d in output: %q", foundChars, output)
	}
}

func TestSpinner_ConcurrentOperations(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Concurrent test")

	// Start spinner
	spinner.Start()

	// Perform concurrent operations
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		spinner.UpdateText("Update 1")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		spinner.UpdateText("Update 2")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		spinner.UpdateText("Update 3")
	}()

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Should be safe to stop
	spinner.Stop()

	// Verify final text is one of the updates
	if spinner.text != "Update 1" && spinner.text != "Update 2" && spinner.text != "Update 3" {
		t.Errorf("Expected text to be one of the updates, got: %q", spinner.text)
	}
}

func TestSpinnerManager_NewSpinnerManager(t *testing.T) {
	manager := NewSpinnerManager(5)
	if manager == nil {
		t.Fatal("NewSpinnerManager returned nil")
	}
	if manager.totalItems != 5 {
		t.Errorf("Expected totalItems to be 5, got %d", manager.totalItems)
	}
	if manager.current != 0 {
		t.Errorf("Expected current to be 0, got %d", manager.current)
	}
}

func TestSpinnerManager_StartSpinner(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	tests := []struct {
		name      string
		total     int
		operation string
		item      string
		expected  string
	}{
		{
			name:      "single item",
			total:     1,
			operation: "Installing",
			item:      "package",
			expected:  "Installing: package",
		},
		{
			name:      "multiple items first",
			total:     3,
			operation: "Processing",
			item:      "file1.txt",
			expected:  "[1/3] Processing: file1.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := testutil.NewBufferWriter(true)
			progressWriter = buf

			manager := NewSpinnerManager(tt.total)
			spinner := manager.StartSpinner(tt.operation, tt.item)

			if spinner == nil {
				t.Fatal("StartSpinner returned nil")
			}

			// Let spinner run briefly
			time.Sleep(150 * time.Millisecond)
			spinner.Stop()

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expected, output)
			}
		})
	}
}

func TestSpinnerManager_MultipleSpinners(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	manager := NewSpinnerManager(3)

	// Test that each spinner gets the correct progress text
	testCases := []struct {
		operation string
		item      string
		expected  string
	}{
		{"Installing", "package1", "[1/3] Installing: package1"},
		{"Installing", "package2", "[2/3] Installing: package2"},
		{"Installing", "package3", "[3/3] Installing: package3"},
	}

	for i, tc := range testCases {
		spinner := manager.StartSpinner(tc.operation, tc.item)

		// Verify the spinner has the correct formatted text
		if spinner.text != tc.expected {
			t.Errorf("Test %d: Expected spinner text %q, got: %q", i+1, tc.expected, spinner.text)
		}

		// Stop spinner to clean up
		spinner.Stop()
	}

	// Verify counter incremented correctly
	if manager.current != 3 {
		t.Errorf("Expected current to be 3, got %d", manager.current)
	}
}

func TestWithSpinner(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	executed := false
	err := WithSpinner("Processing", func(ctx context.Context) error {
		executed = true
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinner returned unexpected error: %v", err)
	}

	if !executed {
		t.Error("Function was not executed")
	}

	// WithSpinner clears the spinner when done, so we won't see the text
	// The test should verify execution and no error, which it does
}

func TestWithSpinner_Error(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	expectedErr := errors.New("test error")
	err := WithSpinner("Processing", func(ctx context.Context) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestWithSpinner_Context(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	var receivedCtx context.Context
	err := WithSpinner("Processing", func(ctx context.Context) error {
		receivedCtx = ctx
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinner returned unexpected error: %v", err)
	}

	if receivedCtx == nil {
		t.Error("Function did not receive a context")
	}
}

func TestWithSpinnerResult_Success(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	executed := false
	err := WithSpinnerResult("Installing package", func(ctx context.Context) error {
		executed = true
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinnerResult returned unexpected error: %v", err)
	}

	if !executed {
		t.Error("Function was not executed")
	}

	output := buf.String()
	if !strings.Contains(output, IconSuccess) {
		t.Errorf("Output should contain success icon, got: %q", output)
	}
	if !strings.Contains(output, "Completed: Installing package") {
		t.Errorf("Output should contain success message, got: %q", output)
	}
}

func TestWithSpinnerResult_Error(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	expectedErr := errors.New("installation failed")
	err := WithSpinnerResult("Installing package", func(ctx context.Context) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	output := buf.String()
	if !strings.Contains(output, IconError) {
		t.Errorf("Output should contain error icon, got: %q", output)
	}
	if !strings.Contains(output, "Failed: Installing package") {
		t.Errorf("Output should contain error message, got: %q", output)
	}
}

func TestSpinner_RaceConditions(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	spinner := NewSpinner("Race test")

	// Try to create race conditions with concurrent operations
	var wg sync.WaitGroup

	// Multiple goroutines trying to start
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			spinner.Start()
		}()
	}

	// Multiple goroutines trying to update text
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			spinner.UpdateText(fmt.Sprintf("Update %d", n))
		}(i)
	}

	// Multiple goroutines trying to stop
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i) * time.Millisecond)
			spinner.Stop()
		}()
	}

	wg.Wait()

	// Final stop to ensure cleanup
	spinner.Stop()

	// If we get here without deadlock or panic, the test passes
}

func TestSpinnerManager_ConcurrentStartSpinner(t *testing.T) {
	originalWriter := progressWriter
	defer func() { progressWriter = originalWriter }()

	buf := testutil.NewBufferWriter(true)
	progressWriter = buf

	manager := NewSpinnerManager(100)

	var wg sync.WaitGroup
	spinners := make([]*Spinner, 10)

	// Start multiple spinners concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			spinners[index] = manager.StartSpinner("Processing", fmt.Sprintf("item%d", index))
		}(i)
	}

	wg.Wait()

	// Stop all spinners
	for _, spinner := range spinners {
		if spinner != nil {
			spinner.Stop()
		}
	}

	// Verify counter incremented correctly
	if manager.current != 10 {
		t.Errorf("Expected current to be 10, got %d", manager.current)
	}
}
