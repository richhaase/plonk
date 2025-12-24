// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/output"
	packages "github.com/richhaase/plonk/internal/packages"
	"github.com/richhaase/plonk/internal/testutil"
)

// CLITestEnv carries handles for configuring a CLI test run.
type CLITestEnv struct {
	T         *testing.T
	HomeDir   string
	ConfigDir string
	Writer    *testutil.BufferWriter
	Executor  *packages.MockCommandExecutor
}

// RunCLI runs the root cobra command with the provided args in an isolated
// environment (temp HOME/PLONK_DIR, NO_COLOR). It captures output via a
// test writer and uses a MockCommandExecutor for package manager calls.
//
// The optional setup callback can customize the mock executor responses or
// seed files in the temp config/home directories before execution.
func RunCLI(t *testing.T, args []string, setup func(env CLITestEnv)) (string, error) {
	t.Helper()

	// Create isolated HOME and PLONK_DIR
	tempHome := t.TempDir()
	tempConfig := filepath.Join(tempHome, ".config", "plonk")
	if err := os.MkdirAll(tempConfig, 0o755); err != nil {
		t.Fatalf("failed to create temp config dir: %v", err)
	}

	// Set environment for the command run
	t.Setenv("HOME", tempHome)
	t.Setenv("PLONK_DIR", tempConfig)
	t.Setenv("NO_COLOR", "1")

	// Capture output via test writer
	w := testutil.NewBufferWriter(false)
	output.SetWriter(w)
	t.Cleanup(func() { output.SetWriter(&output.StdoutWriter{}) })

	// Use mock executor for all package manager commands
	mockExec := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{}}
	packages.SetDefaultExecutor(mockExec)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	env := CLITestEnv{
		T:         t,
		HomeDir:   tempHome,
		ConfigDir: tempConfig,
		Writer:    w,
		Executor:  mockExec,
	}

	if setup != nil {
		setup(env)
	}

	// If caller didn't specify an output format, default to table to avoid sticky flag state
	hasOutputFlag := false
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" || args[i] == "--output" {
			hasOutputFlag = true
			break
		}
	}
	if !hasOutputFlag {
		// Best effort: reset persistent flag
		_ = rootCmd.PersistentFlags().Set("output", "table")
	}
	// Reset status flags (value + Changed=false)
	for _, name := range []string{"missing"} {
		if f := statusCmd.Flags().Lookup(name); f != nil {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
	}

	// Reset dotfiles flags: value + Changed=false
	for _, name := range []string{"missing"} {
		if f := dotfilesCmd.Flags().Lookup(name); f != nil {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
	}

	// Capture stdout (table outputs are printed directly to stdout)
	oldStdout := os.Stdout
	r, wpipe, errPipe := os.Pipe()
	if errPipe != nil {
		t.Fatalf("failed to create pipe: %v", errPipe)
	}
	os.Stdout = wpipe

	// Run the CLI
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	// Restore stdout
	_ = wpipe.Close()
	os.Stdout = oldStdout

	// Read captured stdout
	var stdoutBuf bytes.Buffer
	if _, copyErr := io.Copy(&stdoutBuf, r); copyErr != nil {
		t.Fatalf("failed to read captured stdout: %v", copyErr)
	}
	_ = r.Close()

	// Combine stdout (table) with writer buffer (progress/messages)
	combined := stdoutBuf.String()
	if wb := w.String(); wb != "" {
		combined += wb
	}

	return combined, err
}
