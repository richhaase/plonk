// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SpinnerChars defines the animation frames for the spinner
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner represents a progress spinner
type Spinner struct {
	text       string
	writer     Writer
	mu         sync.Mutex
	running    bool
	done       chan struct{}
	wg         sync.WaitGroup
	spinnerIdx int
}

// NewSpinner creates a new spinner with the given text
func NewSpinner(text string) *Spinner {
	return &Spinner{
		text:   text,
		writer: progressWriter,
		done:   make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return s
	}

	s.running = true
	s.done = make(chan struct{})
	s.wg.Add(1)

	go s.spin()
	return s
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	s.running = false
	close(s.done)
	s.mu.Unlock()

	// Wait for the spinner goroutine to finish (without holding the mutex)
	s.wg.Wait()

	// Clear the spinner line
	if s.writer.IsTerminal() {
		s.writer.Printf("\r\033[K")
	}
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	icon := GetStatusIcon("success")
	s.writer.Printf("%s %s\n", icon, message)
}

// Error stops the spinner and shows an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	icon := GetStatusIcon("failed")
	s.writer.Printf("%s %s\n", icon, message)
}

// UpdateText updates the spinner text while it's running
func (s *Spinner) UpdateText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text = text
}

// spin runs the spinner animation loop
func (s *Spinner) spin() {
	defer s.wg.Done()

	if !s.writer.IsTerminal() {
		// If not a terminal, just print the text once
		s.writer.Printf("%s\n", s.text)
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			if s.running {
				char := SpinnerChars[s.spinnerIdx%len(SpinnerChars)]
				s.writer.Printf("\r%s %s", char, s.text)
				s.spinnerIdx++
			}
			s.mu.Unlock()
		}
	}
}

// SpinnerManager manages multiple spinners for batch operations
type SpinnerManager struct {
	totalItems int
	current    int
	mu         sync.Mutex
}

// NewSpinnerManager creates a new spinner manager for batch operations
func NewSpinnerManager(totalItems int) *SpinnerManager {
	return &SpinnerManager{
		totalItems: totalItems,
		current:    0,
	}
}

// StartSpinner starts a spinner for a specific item with progress indicator
func (sm *SpinnerManager) StartSpinner(operation, item string) *Spinner {
	sm.mu.Lock()
	sm.current++
	current := sm.current
	total := sm.totalItems
	sm.mu.Unlock()

	var text string
	if total > 1 {
		text = fmt.Sprintf("[%d/%d] %s: %s", current, total, operation, item)
	} else {
		text = fmt.Sprintf("%s: %s", operation, item)
	}

	return NewSpinner(text).Start()
}

// WithSpinner executes a function with a spinner running
func WithSpinner(text string, fn func(context.Context) error) error {
	spinner := NewSpinner(text).Start()
	defer spinner.Stop()

	// Create a context for the operation
	ctx := context.Background()
	return fn(ctx)
}

// WithSpinnerResult executes a function with a spinner and handles the result
func WithSpinnerResult(text string, fn func(context.Context) error) error {
	spinner := NewSpinner(text).Start()

	err := fn(context.Background())
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed: %s", text))
		return err
	}

	spinner.Success(fmt.Sprintf("Completed: %s", text))
	return nil
}
